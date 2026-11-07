package main

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func testServer(t *testing.T, upstream http.Handler) (*Server, *httptest.Server) {
	t.Helper()
	u := httptest.NewServer(upstream)
	s := &Server{cfg: Config{RPCURL: u.URL, RequestTimeout: time.Second, AllowBroadcast: true}, http: u.Client(), idempotency: map[string]broadcastResult{}}
	return s, u
}

func TestNetworkUsesUpstreamChainState(t *testing.T) {
	s, upstream := testServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/status" {
			_, _ = w.Write([]byte(`{"result":{"node_info":{"network":"ogo-local-1"},"sync_info":{"latest_block_height":"42"}}}`))
			return
		}
		if r.URL.Path == "/net_info" {
			_, _ = w.Write([]byte(`{"result":{"n_peers":"3"}}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer upstream.Close()
	r := httptest.NewRequest(http.MethodGet, "/v1/network", nil)
	w := httptest.NewRecorder()
	s.network(w, r)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "ogo-local-1") || !strings.Contains(w.Body.String(), "42") {
		t.Fatalf("unexpected network response %d %s", w.Code, w.Body.String())
	}
}

func TestBalanceDoesNotSimulateWithoutLCD(t *testing.T) {
	s := &Server{cfg: Config{RPCURL: "http://unused", RequestTimeout: time.Second}, http: http.DefaultClient, idempotency: map[string]broadcastResult{}}
	r := httptest.NewRequest(http.MethodGet, "/v1/balances/ogo1abc", nil)
	w := httptest.NewRecorder()
	s.balance(w, r)
	if w.Code != http.StatusNotImplemented || !strings.Contains(w.Body.String(), "no balance is simulated") {
		t.Fatalf("unexpected response %d %s", w.Code, w.Body.String())
	}
}

func TestBroadcastClientSignedAndIdempotent(t *testing.T) {
	var calls int
	s, upstream := testServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/broadcast_tx_sync" || r.Method != http.MethodPost {
			t.Fatalf("unexpected upstream request: %s %s", r.Method, r.URL.Path)
		}
		calls++
		var payload map[string]string
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		if payload["mode"] != "sync" {
			t.Fatalf("unexpected mode: %#v", payload)
		}
		_, _ = w.Write([]byte(`{"result":{"code":0,"hash":"ABC123"}}`))
	}))
	defer upstream.Close()
	encoded := base64.StdEncoding.EncodeToString([]byte("client-signed-bytes"))
	body := `{"signed_tx":"` + encoded + `"}`
	for i := 0; i < 2; i++ {
		r := httptest.NewRequest(http.MethodPost, "/v1/transactions/broadcast", strings.NewReader(body))
		r.Header.Set("Idempotency-Key", "payment-1")
		w := httptest.NewRecorder()
		s.broadcast(w, r)
		if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "ABC123") {
			t.Fatalf("unexpected response %d %s", w.Code, w.Body.String())
		}
	}
	if calls != 1 {
		t.Fatalf("expected one upstream call, got %d", calls)
	}
}

func TestBroadcastRejectsConflictingIdempotency(t *testing.T) {
	s, upstream := testServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(`{"result":{"code":0}}`)) }))
	defer upstream.Close()
	first := base64.StdEncoding.EncodeToString([]byte("one"))
	second := base64.StdEncoding.EncodeToString([]byte("two"))
	for _, encoded := range []string{first, second} {
		r := httptest.NewRequest(http.MethodPost, "/v1/transactions/broadcast", strings.NewReader(`{"signed_tx":"`+encoded+`"}`))
		r.Header.Set("Idempotency-Key", "same-key")
		w := httptest.NewRecorder()
		s.broadcast(w, r)
		if encoded == first && w.Code != http.StatusOK {
			t.Fatalf("first request failed: %d %s", w.Code, w.Body.String())
		}
		if encoded == second && w.Code != http.StatusConflict {
			t.Fatalf("second request should conflict: %d %s", w.Code, w.Body.String())
		}
	}
}
