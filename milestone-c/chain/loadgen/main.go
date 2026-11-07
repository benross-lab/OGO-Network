// Command loadgen is the Milestone C load-test driver (P2/P3/P4).
// It funds N sender accounts from one master key in a single
// multi-message transaction, then has each sender fire simple
// bank transfers at a target aggregate rate, recording submit time
// and txhash for every accepted transaction. After the run it
// samples committed transactions and reports finality percentiles
// (block time minus submit time) and achieved throughput.
//
// Usage (from chain/):
//
//	go run ./loadgen -node http://127.0.0.1:27657 \
//	  -mnemonic-file localnet4-data/loadmaster-mnemonic.txt \
//	  -senders 100 -duration 10m -tps 200 \
//	  -out loadgen-p2.jsonl
//
// TEST-ONLY local staging. The mnemonic file stays in the gitignored
// localnet data directory and is never committed.
package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	stdmath "math"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/client"
	clienttx "github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/go-bip39"
)

const denom = "uogo"

type accountInfo struct {
	Number   uint64 `json:"account_number"`
	Sequence uint64 `json:"sequence"`
}

var httpClient = &http.Client{Timeout: 15 * time.Second}

func getJSON(url string, out any) error {
	resp, err := httpClient.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, out)
}

func getAccount(node, addr string) (accountInfo, error) {
	var res struct {
		Account struct {
			AccountNumber string `json:"account_number"`
			Sequence      string `json:"sequence"`
		} `json:"account"`
	}
	if err := getJSON(node+"/cosmos/auth/v1beta1/accounts/"+addr, &res); err != nil {
		return accountInfo{}, err
	}
	num, err := strconv.ParseUint(res.Account.AccountNumber, 10, 64)
	if err != nil {
		return accountInfo{}, fmt.Errorf("account_number %q: %w", res.Account.AccountNumber, err)
	}
	seq, err := strconv.ParseUint(res.Account.Sequence, 10, 64)
	if err != nil {
		return accountInfo{}, fmt.Errorf("sequence %q: %w", res.Account.Sequence, err)
	}
	return accountInfo{Number: num, Sequence: seq}, nil
}

func broadcastSync(node string, txBytes []byte) (string, uint32, string, error) {
	payload, _ := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "broadcast_tx_sync",
		"params":  map[string]string{"tx": base64.StdEncoding.EncodeToString(txBytes)},
	})
	resp, err := httpClient.Post(node, "application/json", bytes.NewReader(payload))
	if err != nil {
		return "", 0, "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", 0, "", err
	}
	var res struct {
		Result struct {
			Hash string `json:"hash"`
			Code uint32 `json:"code"`
			Log  string `json:"log"`
		} `json:"result"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(body, &res); err != nil {
		return "", 0, "", fmt.Errorf("broadcast decode: %w (%s)", err, string(body))
	}
	if res.Error != nil {
		return "", 0, "", fmt.Errorf("broadcast rpc error: %s", res.Error.Message)
	}
	return res.Result.Hash, res.Result.Code, res.Result.Log, nil
}

func txHeight(node, hash string) (int64, error) {
	var res struct {
		Result struct {
			Height string `json:"height"`
		} `json:"result"`
	}
	if err := getJSON(node+"/tx?hash=0x"+hash, &res); err != nil {
		return 0, err
	}
	if res.Result.Height == "" {
		return 0, fmt.Errorf("tx %s not found yet", hash)
	}
	var h int64
	if _, err := fmt.Sscanf(res.Result.Height, "%d", &h); err != nil {
		return 0, err
	}
	return h, nil
}

func blockTime(node string, height int64) (time.Time, error) {
	var res struct {
		Result struct {
			Block struct {
				Header struct {
					Time time.Time `json:"time"`
				} `json:"header"`
			} `json:"block"`
		} `json:"result"`
	}
	if err := getJSON(fmt.Sprintf("%s/block?height=%d", node, height), &res); err != nil {
		return time.Time{}, err
	}
	return res.Result.Block.Header.Time, nil
}

func latestHeight(node string) (int64, error) {
	var res struct {
		Result struct {
			SyncInfo struct {
				LatestBlockHeight string `json:"latest_block_height"`
			} `json:"sync_info"`
		} `json:"result"`
	}
	if err := getJSON(node+"/status", &res); err != nil {
		return 0, err
	}
	return strconv.ParseInt(res.Result.SyncInfo.LatestBlockHeight, 10, 64)
}

func waitCommitted(node, hash string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if _, err := txHeight(node, hash); err == nil {
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("tx %s not committed within %s", hash, timeout)
}

type signer struct {
	priv cryptotypes.PrivKey
	addr sdk.AccAddress
	acc  uint64
	seq  uint64
	mu   sync.Mutex
}

func (s *signer) signAndBroadcast(ctx context.Context, node string, txConfig client.TxConfig, msgs []sdk.Msg, gas uint64, fee math.Int) (string, uint32, string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	builder := txConfig.NewTxBuilder()
	if err := builder.SetMsgs(msgs...); err != nil {
		return "", 0, "", err
	}
	builder.SetGasLimit(gas)
	builder.SetFeeAmount(sdk.NewCoins(sdk.NewCoin(denom, fee)))
	sigV2 := signing.SignatureV2{
		PubKey: s.priv.PubKey(),
		Data: &signing.SingleSignatureData{
			SignMode:  signing.SignMode_SIGN_MODE_DIRECT,
			Signature: nil,
		},
		Sequence: s.seq,
	}
	if err := builder.SetSignatures(sigV2); err != nil {
		return "", 0, "", err
	}
	signerData := authsigning.SignerData{
		Address:       s.addr.String(),
		ChainID:       chainID,
		AccountNumber: s.acc,
		Sequence:      s.seq,
		PubKey:        s.priv.PubKey(),
	}
	sigV2, err := clienttx.SignWithPrivKey(ctx, signing.SignMode_SIGN_MODE_DIRECT, signerData, builder, s.priv, txConfig, s.seq)
	if err != nil {
		return "", 0, "", err
	}
	if err := builder.SetSignatures(sigV2); err != nil {
		return "", 0, "", err
	}
	bz, err := txConfig.TxEncoder()(builder.GetTx())
	if err != nil {
		return "", 0, "", err
	}
	hash, code, log, err := broadcastSync(node, bz)
	if err != nil {
		return "", 0, "", err
	}
	if code == 0 {
		s.seq++
	}
	return hash, code, log, nil
}

var chainID string

type record struct {
	Hash       string  `json:"hash"`
	SubmitMs   int64   `json:"submit_unix_ms"`
	SubmitH    int64   `json:"submit_height"`
	Worker     int     `json:"worker"`
	Committed  bool    `json:"committed,omitempty"`
	CommitH    int64   `json:"commit_height,omitempty"`
	FinalityS  float64 `json:"finality_s,omitempty"`
	HeightLag  int64   `json:"height_lag,omitempty"`
}

func main() {
	node := flag.String("node", "http://127.0.0.1:27657", "RPC endpoint of one validator")
	api := flag.String("api", "http://127.0.0.1:1417", "REST API endpoint of one validator")
	chain := flag.String("chain-id", "ogo-local-1", "chain id")
	mnemonicFile := flag.String("mnemonic-file", "", "path to the master-key mnemonic (TEST-ONLY)")
	senders := flag.Int("senders", 100, "number of sender accounts")
	duration := flag.Duration("duration", 60*time.Second, "load duration")
	tps := flag.Float64("tps", 50, "target aggregate transfers per second")
	fundEach := flag.Int64("fund-each", 1_000_000, "uogo funded to each sender")
	amount := flag.Int64("amount", 1, "uogo per load transfer")
	out := flag.String("out", "loadgen.jsonl", "results file (one JSON record per accepted tx)")
	recipient := flag.String("recipient", "ogo1km69yz8p0kw2ervdls40h7xssrvhnk4lmgr9n9", "transfer recipient (alice)")
	flag.Parse()
	chainID = *chain
	// OGO bech32 prefixes (must match app/config.go).
	sdk.GetConfig().SetBech32PrefixForAccount("ogo", "ogopub")
	sdk.GetConfig().SetBech32PrefixForValidator("ogovaloper", "ogovaloperpub")
	sdk.GetConfig().SetBech32PrefixForConsensusNode("ogovalcons", "ogovalconspub")
	sdk.GetConfig().SetCoinType(118)
	sdk.GetConfig().Seal()
	if *mnemonicFile == "" {
		fmt.Fprintln(os.Stderr, "-mnemonic-file is required")
		os.Exit(2)
	}
	mnemonicBytes, err := os.ReadFile(*mnemonicFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read mnemonic: %v\n", err)
		os.Exit(1)
	}
	mnemonic := strings.TrimSpace(string(mnemonicBytes))
	if !bip39.IsMnemonicValid(mnemonic) {
		fmt.Fprintln(os.Stderr, "mnemonic is not valid bip39")
		os.Exit(1)
	}

	ir := codectypes.NewInterfaceRegistry()
	banktypes.RegisterInterfaces(ir)
	authtypes.RegisterInterfaces(ir)
	cdc := codec.NewProtoCodec(ir)
	txConfig := authtx.NewTxConfig(cdc, authtx.DefaultSignModes)
	_ = types.Config{} // keep sdk import used via types below

	// Derive the master key (standard Cosmos path m/44'/118'/0'/0/0).
	priv, err := hd.Secp256k1.Derive()(mnemonic, "", "m/44'/118'/0'/0/0")
	if err != nil {
		fmt.Fprintf(os.Stderr, "derive master key: %v\n", err)
		os.Exit(1)
	}
	masterPriv := secp256k1.PrivKey{Key: priv}
	masterAddr := sdk.AccAddress(masterPriv.PubKey().Address())
	masterAcc, err := getAccount(*api, masterAddr.String())
	if err != nil {
		fmt.Fprintf(os.Stderr, "query master account: %v\n", err)
		os.Exit(1)
	}
	master := &signer{priv: &masterPriv, addr: masterAddr, acc: masterAcc.Number, seq: masterAcc.Sequence}
	fmt.Printf("master %s acc=%d seq=%d\n", masterAddr, masterAcc.Number, masterAcc.Sequence)

	// Create sender keys and fund them in one multi-message tx.
	senderKeys := make([]*secp256k1.PrivKey, *senders)
	msgs := make([]sdk.Msg, 0, *senders)
	for i := 0; i < *senders; i++ {
		k := secp256k1.GenPrivKey()
		senderKeys[i] = k
		msgs = append(msgs, banktypes.NewMsgSend(masterAddr, sdk.AccAddress(k.PubKey().Address()), sdk.NewCoins(sdk.NewCoin(denom, math.NewInt(*fundEach)))))
	}
	fundGas := uint64(50_000 * len(msgs))
	fundFee := math.NewInt(int64(fundGas) / 1000) // min gas price 0.001uogo
	hash, code, log, err := master.signAndBroadcast(context.Background(), *node, txConfig, msgs, fundGas, fundFee)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fund tx: %v\n", err)
		os.Exit(1)
	}
	if code != 0 {
		fmt.Fprintf(os.Stderr, "fund tx rejected code=%d log=%s\n", code, log)
		os.Exit(1)
	}
	fmt.Printf("fund tx %s (senders=%d, gas=%d, fee=%s)\n", hash, *senders, fundGas, fundFee)
	if err := waitCommitted(*node, hash, 60*time.Second); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	// Build per-sender signers (account numbers assigned at funding).
	workers := make([]*signer, *senders)
	for i, k := range senderKeys {
		addr := sdk.AccAddress(k.PubKey().Address())
		acc, err := getAccount(*api, addr.String())
		if err != nil {
			fmt.Fprintf(os.Stderr, "query sender %d account: %v\n", i, err)
			os.Exit(1)
		}
		workers[i] = &signer{priv: k, addr: addr, acc: acc.Number, seq: acc.Sequence}
	}
	recipientAddr, err := sdk.AccAddressFromBech32(*recipient)
	if err != nil {
		fmt.Fprintf(os.Stderr, "recipient: %v\n", err)
		os.Exit(1)
	}

	outFile, err := os.Create(*out)
	if err != nil {
		fmt.Fprintf(os.Stderr, "create out: %v\n", err)
		os.Exit(1)
	}
	defer outFile.Close()
	enc := json.NewEncoder(outFile)

	var mu sync.Mutex
	var records []*record
	var submitted, accepted, failed int64
	stop := time.Now().Add(*duration)
	perWorkerInterval := time.Duration(float64(time.Second) * float64(*senders) / *tps)
	fmt.Printf("starting load: senders=%d target_tps=%.1f duration=%s interval/worker=%s\n", *senders, *tps, *duration, perWorkerInterval)

	var wg sync.WaitGroup
	for i, w := range workers {
		wg.Add(1)
		go func(idx int, s *signer) {
			defer wg.Done()
			ticker := time.NewTicker(perWorkerInterval)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					if time.Now().After(stop) {
						return
					}
					msg := banktypes.NewMsgSend(s.addr, recipientAddr, sdk.NewCoins(sdk.NewCoin(denom, math.NewInt(*amount))))
					submit := time.Now()
					submitH, _ := latestHeight(*node)
					h, code, log, err := s.signAndBroadcast(context.Background(), *node, txConfig, []sdk.Msg{msg}, 200_000, math.NewInt(200))
					mu.Lock()
					submitted++
					if err != nil {
						failed++
						fmt.Fprintf(os.Stderr, "worker %d broadcast error: %v\n", idx, err)
					} else if code != 0 {
						failed++
						// Sequence desync: CheckTx accepted some of this
						// sender's txs that never landed on chain, so the
						// local sequence ran ahead of the chain's. Resync
						// from the chain instead of failing forever.
						if code == 32 {
							if acc, qerr := getAccount(*api, s.addr.String()); qerr == nil {
								s.mu.Lock()
								s.seq = acc.Sequence
								s.mu.Unlock()
							}
						}
						fmt.Fprintf(os.Stderr, "worker %d rejected code=%d log=%s\n", idx, code, log)
					} else {
						accepted++
						r := &record{Hash: h, SubmitMs: submit.UnixMilli(), SubmitH: submitH, Worker: idx}
						records = append(records, r)
						_ = enc.Encode(r)
					}
					mu.Unlock()
				}
			}
		}(i, w)
	}
	wg.Wait()
	elapsed := time.Since(stop.Add(-*duration))
	fmt.Printf("load done: submitted=%d accepted=%d failed=%d elapsed=%s achieved_tps=%.1f\n", submitted, accepted, failed, elapsed, float64(accepted)/elapsed.Seconds())

	// Finality sampling: up to 200 evenly spaced accepted txs.
	sample := 200
	if len(records) < sample {
		sample = len(records)
	}
	latencies := make([]float64, 0, sample)
	lags := make([]float64, 0, sample)
	committed := 0
	for i := 0; i < sample; i++ {
		r := records[i*len(records)/sample]
		h, err := txHeight(*node, r.Hash)
		if err != nil {
			continue
		}
		bt, err := blockTime(*node, h)
		if err != nil {
			continue
		}
		lat := bt.Sub(time.UnixMilli(r.SubmitMs)).Seconds()
		latencies = append(latencies, lat)
		lags = append(lags, float64(h-r.SubmitH))
		committed++
		r.Committed = true
		r.CommitH = h
		r.FinalityS = lat
		r.HeightLag = h - r.SubmitH
	}
	sort.Float64s(latencies)
	sort.Float64s(lags)
	summary := map[string]any{
		"submitted":     submitted,
		"accepted":      accepted,
		"failed":        failed,
		"duration_s":    elapsed.Seconds(),
		"achieved_tps":  float64(accepted) / elapsed.Seconds(),
		"sampled":       len(latencies),
		"sampled_ok":    committed,
		"finality_p50":  percentile(latencies, 0.50),
		"finality_p95":  percentile(latencies, 0.95),
		"finality_max":  percentile(latencies, 1.0),
		"height_lag_p50": percentile(lags, 0.50),
		"height_lag_p95": percentile(lags, 0.95),
		"height_lag_max": percentile(lags, 1.0),
	}
	sb, _ := json.MarshalIndent(summary, "", "  ")
	fmt.Println(string(sb))
	_ = os.WriteFile(strings.TrimSuffix(*out, ".jsonl")+"-summary.json", sb, 0o644)
}

func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return stdmath.NaN()
	}
	idx := int(stdmath.Ceil(p*float64(len(sorted)))) - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}
