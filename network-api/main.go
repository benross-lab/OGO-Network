package main

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const maxBody = 1 << 20

type Config struct {
	RPCURL         string
	LCDURL         string
	IndexerURL     string
	EpochsPath     string
	AllowBroadcast bool
	ListenAddr     string
	RequestTimeout time.Duration
}

func configFromEnv() Config {
	allow := strings.EqualFold(os.Getenv("OGO_API_ALLOW_BROADCAST"), "true")
	timeout := 10 * time.Second
	if n, err := strconv.Atoi(os.Getenv("OGO_API_TIMEOUT_SECONDS")); err == nil && n > 0 && n <= 120 {
		timeout = time.Duration(n) * time.Second
	}
	return Config{
		RPCURL:         strings.TrimRight(envOr("OGO_API_RPC_URL", "http://127.0.0.1:26657"), "/"),
		LCDURL:         strings.TrimRight(os.Getenv("OGO_API_LCD_URL"), "/"),
		IndexerURL:     strings.TrimRight(os.Getenv("OGO_API_INDEXER_URL"), "/"),
		EpochsPath:     envOr("OGO_API_EPOCHS_PATH", "/ogo/epochs/v1/current"),
		AllowBroadcast: allow,
		ListenAddr:     envOr("OGO_API_LISTEN", "127.0.0.1:8787"),
		RequestTimeout: timeout,
	}
}

func envOr(name, fallback string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return fallback
}

type Server struct {
	cfg         Config
	http        *http.Client
	mu          sync.Mutex
	idempotency map[string]broadcastResult
}

type broadcastResult struct {
	Fingerprint string         `json:"-"`
	Response    map[string]any `json:"response"`
}

func main() {
	cfg := configFromEnv()
	s := &Server{cfg: cfg, http: &http.Client{Timeout: cfg.RequestTimeout}, idempotency: make(map[string]broadcastResult)}
	log.Printf("OGO Network API development server listening on %s; rpc=%s broadcast_enabled=%t", cfg.ListenAddr, cfg.RPCURL, cfg.AllowBroadcast)
	if err := http.ListenAndServe(cfg.ListenAddr, s.routes()); err != nil {
		log.Fatal(err)
	}
}

func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.health)
	mux.HandleFunc("/v1/network", s.network)
	mux.HandleFunc("/v1/balances/", s.balance)
	mux.HandleFunc("/v1/transactions/prepare", s.prepare)
	mux.HandleFunc("/v1/transactions/broadcast", s.broadcast)
	mux.HandleFunc("/v1/transactions/", s.transaction)
	mux.HandleFunc("/v1/addresses/", s.addressTransactions)
	mux.HandleFunc("/v1/validators", s.validators)
	mux.HandleFunc("/v1/epochs", s.epochs)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Idempotency-Key")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.URL.Path != "/healthz" && !strings.HasPrefix(r.URL.Path, "/v1/") {
			errorJSON(w, 404, "not_found", "endpoint not found")
			return
		}
		mux.ServeHTTP(w, r)
	})
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]any{"ok": true, "service": "ogo-network-api", "development_only": true})
}

func (s *Server) network(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	status, _, err := s.rpc(r.Context(), "/status", nil)
	if err != nil {
		upstreamError(w, err)
		return
	}
	net, _, _ := s.rpc(r.Context(), "/net_info", nil)
	writeJSON(w, 200, map[string]any{"network": json.RawMessage(status), "net_info": json.RawMessage(net)})
}

func (s *Server) balance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	address := strings.TrimPrefix(r.URL.Path, "/v1/balances/")
	if address == "" || strings.Contains(address, "/") {
		errorJSON(w, 400, "invalid_address", "a single address is required")
		return
	}
	if s.cfg.LCDURL == "" {
		errorJSON(w, 501, "lcd_unavailable", "balance queries require OGO_API_LCD_URL; no balance is simulated")
		return
	}
	body, status, err := s.upstream(r.Context(), s.cfg.LCDURL+"/cosmos/bank/v1beta1/balances/"+url.PathEscape(address), http.MethodGet, nil, "")
	if err != nil {
		upstreamError(w, err)
		return
	}
	writeRaw(w, status, body)
}

func (s *Server) transaction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/v1/transactions/"), "/")
	if len(parts) == 2 && parts[1] == "verify" {
		if parts[0] == "" {
			errorJSON(w, 400, "invalid_hash", "a transaction hash is required")
			return
		}
		body, status, err := s.rpc(r.Context(), "/tx?hash="+url.QueryEscape(parts[0]), nil)
		if err != nil {
			upstreamError(w, err)
			return
		}
		var tx struct {
			Result struct {
				TxResult struct {
					Code uint32 `json:"code"`
					Hash string `json:"hash"`
					Log  string `json:"log"`
				} `json:"tx_result"`
			} `json:"result"`
		}
		if json.Unmarshal(body, &tx) != nil {
			errorJSON(w, 502, "invalid_upstream_response", "RPC returned an invalid transaction response")
			return
		}
		writeJSON(w, status, map[string]any{"committed": status < 400 && tx.Result.TxResult.Code == 0, "hash": tx.Result.TxResult.Hash, "code": tx.Result.TxResult.Code, "log": tx.Result.TxResult.Log, "transaction": json.RawMessage(body)})
		return
	}
	if len(parts) != 1 || parts[0] == "" {
		errorJSON(w, 400, "invalid_hash", "a transaction hash is required")
		return
	}
	hash := url.QueryEscape(parts[0])
	body, status, err := s.rpc(r.Context(), "/tx?hash="+hash, nil)
	if err != nil {
		upstreamError(w, err)
		return
	}
	writeRaw(w, status, body)
}

func (s *Server) validators(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	body, status, err := s.rpc(r.Context(), "/validators?per_page=100", nil)
	if err != nil {
		upstreamError(w, err)
		return
	}
	writeRaw(w, status, body)
}

func (s *Server) epochs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	if s.cfg.LCDURL == "" {
		errorJSON(w, 501, "lcd_unavailable", "epoch queries require OGO_API_LCD_URL; no epoch data is invented")
		return
	}
	body, status, err := s.upstream(r.Context(), s.cfg.LCDURL+s.cfg.EpochsPath, http.MethodGet, nil, "")
	if err != nil {
		upstreamError(w, err)
		return
	}
	writeRaw(w, status, body)
}

func (s *Server) addressTransactions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	if s.cfg.IndexerURL == "" {
		errorJSON(w, 501, "indexer_unavailable", "transaction history requires OGO_API_INDEXER_URL; history is never approximated")
		return
	}
	address := strings.TrimPrefix(r.URL.Path, "/v1/addresses/")
	address = strings.TrimSuffix(address, "/transactions")
	if address == "" || strings.Contains(address, "/") {
		errorJSON(w, 400, "invalid_address", "a single address is required")
		return
	}
	body, status, err := s.upstream(r.Context(), s.cfg.IndexerURL+"/v1/addresses/"+url.PathEscape(address)+"/transactions?"+r.URL.RawQuery, http.MethodGet, nil, "")
	if err != nil {
		upstreamError(w, err)
		return
	}
	writeRaw(w, status, body)
}

type prepareRequest struct {
	ChainID       string            `json:"chain_id"`
	From          string            `json:"from"`
	Messages      []json.RawMessage `json:"messages"`
	Fee           json.RawMessage   `json:"fee"`
	Memo          string            `json:"memo"`
	AccountNumber uint64            `json:"account_number"`
	Sequence      uint64            `json:"sequence"`
}

func (s *Server) prepare(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	var req prepareRequest
	if err := readJSON(w, r, &req); err != nil {
		errorJSON(w, 400, "invalid_json", err.Error())
		return
	}
	if req.From == "" || len(req.Messages) == 0 {
		errorJSON(w, 400, "invalid_request", "from and at least one message are required")
		return
	}
	if req.ChainID == "" {
		var status struct {
			Result struct {
				NodeInfo struct {
					Network string `json:"network"`
				} `json:"node_info"`
			} `json:"result"`
		}
		if err := s.rpcJSON(r.Context(), "/status", nil, &status); err == nil {
			req.ChainID = status.Result.NodeInfo.Network
		}
	}
	// This is deliberately an unsigned signing request, not a signed Cosmos transaction.
	writeJSON(w, 200, map[string]any{
		"type":             "ogo.unsigned_transaction.v1",
		"development_only": true,
		"chain_id":         req.ChainID,
		"from":             req.From,
		"messages":         req.Messages,
		"fee":              req.Fee,
		"memo":             req.Memo,
		"account_number":   req.AccountNumber,
		"sequence":         req.Sequence,
		"signing":          map[string]any{"client_side_only": true, "mode": "direct", "private_keys_requested": false},
	})
}

type broadcastRequest struct {
	SignedTx string `json:"signed_tx"`
}

func (s *Server) broadcast(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	if !s.cfg.AllowBroadcast {
		errorJSON(w, 503, "broadcast_disabled", "broadcast is disabled unless OGO_API_ALLOW_BROADCAST=true in development")
		return
	}
	key := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
	if key == "" || len(key) > 200 {
		errorJSON(w, 400, "missing_idempotency_key", "Idempotency-Key is required for broadcast")
		return
	}
	var req broadcastRequest
	if err := readJSON(w, r, &req); err != nil {
		errorJSON(w, 400, "invalid_json", err.Error())
		return
	}
	if req.SignedTx == "" {
		errorJSON(w, 400, "missing_signed_tx", "signed_tx is required")
		return
	}
	if _, err := base64.StdEncoding.DecodeString(req.SignedTx); err != nil {
		errorJSON(w, 400, "invalid_signed_tx", "signed_tx must be standard base64")
		return
	}
	fingerprint := sha256.Sum256([]byte(req.SignedTx))
	fp := hex.EncodeToString(fingerprint[:])
	s.mu.Lock()
	if old, ok := s.idempotency[key]; ok {
		s.mu.Unlock()
		if old.Fingerprint != fp {
			errorJSON(w, 409, "idempotency_conflict", "the idempotency key was already used for different signed bytes")
			return
		}
		writeJSON(w, 200, old.Response)
		return
	}
	s.mu.Unlock()
	payload, _ := json.Marshal(map[string]string{"tx_bytes": req.SignedTx, "mode": "sync"})
	body, status, err := s.rpc(r.Context(), "/broadcast_tx_sync", payload)
	if err != nil {
		upstreamError(w, err)
		return
	}
	var response map[string]any
	if json.Unmarshal(body, &response) != nil {
		response = map[string]any{"raw": string(body)}
	}
	s.mu.Lock()
	s.idempotency[key] = broadcastResult{Fingerprint: fp, Response: response}
	s.mu.Unlock()
	writeJSON(w, status, response)
}

func (s *Server) rpc(ctx context.Context, path string, body []byte) ([]byte, int, error) {
	method := http.MethodGet
	if body != nil {
		method = http.MethodPost
	}
	return s.upstream(ctx, s.cfg.RPCURL+path, method, body, "application/json")
}

func (s *Server) rpcJSON(ctx context.Context, path string, body []byte, out any) error {
	b, status, err := s.rpc(ctx, path, body)
	if err != nil {
		return err
	}
	if status >= 400 {
		return fmt.Errorf("rpc returned HTTP %d", status)
	}
	return json.Unmarshal(b, out)
}

func (s *Server) upstream(ctx context.Context, endpoint, method string, body []byte, contentType string) ([]byte, int, error) {
	if body == nil && method == http.MethodGet { /* normal GET */
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, strings.NewReader(string(body)))
	if err != nil {
		return nil, 0, err
	}
	if body != nil {
		req.Header.Set("Content-Type", contentType)
	}
	resp, err := s.http.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxBody))
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return data, resp.StatusCode, nil
}

func readJSON(w http.ResponseWriter, r *http.Request, out any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxBody)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(out); err != nil {
		return err
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func writeRaw(w http.ResponseWriter, status int, body []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(body)
}
func errorJSON(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{"error": map[string]string{"code": code, "message": message}})
}
func upstreamError(w http.ResponseWriter, err error) {
	if errors.Is(err, context.DeadlineExceeded) {
		errorJSON(w, 504, "upstream_timeout", err.Error())
		return
	}
	errorJSON(w, 502, "upstream_error", err.Error())
}
func methodNotAllowed(w http.ResponseWriter) {
	errorJSON(w, 405, "method_not_allowed", "method not allowed")
}
