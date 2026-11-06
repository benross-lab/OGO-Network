// OgoApp is the Milestone 1-2 local prototype: a CometBFT ABCI 2.0
// application whose state machine is x/ogocoin.
//
// LIMITATIONS (stated plainly, not hidden):
//   - Transactions are JSON and are NOT signature-verified in this
//     scaffold. The ledger enforces balances and supply, but anyone
//     with RPC access can submit a tx naming any sender. Real auth
//     (x/auth signatures) arrives with the full SDK depinject wiring
//     in Milestone C. Do not use this binary where adversarial
//     submitters exist.
//   - The wasm VM is not instantiated; contract execution is pending.
//   - Live IBC relaying is not wired; escrow accounting is tested at
//     the ledger level only.
//   - Epoch execution uses the epoch timer mirror in x/ogocoin
//     (epochs.go), which reproduces the pinned SDK x/epochs
//     BeginBlocker semantics. It is NOT the SDK x/epochs module: no
//     keeper, no depinject wiring, no StoreUpgrade yet. See
//     chain/EPOCHS_GAP.md. When the epoch timer is disabled
//     (no epochs config in genesis), issuance falls back to the
//     legacy block-based MintScheduled path from Milestone B.
package app

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"cosmossdk.io/math"
	abcitypes "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/libs/log"

	ogocoin "github.com/benross-lab/ogo-network/x/ogocoin"
)

// Tx is the JSON transaction format accepted by this prototype.
type Tx struct {
	Type   string `json:"type"` // "send" or "burn"
	From   string `json:"from"`
	To     string `json:"to,omitempty"`
	Amount string `json:"amount"`
	Fee    string `json:"fee,omitempty"`
}

// OgoApp implements abcitypes.Application.
type OgoApp struct {
	logger    log.Logger
	homeDir   string
	stateFile string
	// appStateFile persists the last committed height and the epoch
	// timer state. It is separate from the ledger file because the
	// ledger file format is owned by x/ogocoin and must not change
	// shape here. Without a persisted height, a restart would report
	// height 0 and CometBFT would replay blocks, re-applying
	// transactions (Milestone B survived this only because the
	// replayed range held no transactions).
	appStateFile string

	mu      sync.Mutex
	ledger  *ogocoin.Ledger
	height  int64
	appHash []byte

	// epochTimer is nil when epoch mode is disabled (no epochs config
	// in genesis). When set, issuance runs only through the epoch
	// hook; the legacy block-based MintScheduled path is skipped.
	epochTimer *ogocoin.EpochTimer
}

// persistedAppState is the on-disk app state (height + epoch timer).
type persistedAppState struct {
	Height int64                    `json:"height"`
	Epoch  *ogocoin.EpochTimerState `json:"epoch,omitempty"`
}

// NewOgoApp loads the ledger from the state file if present; otherwise
// the ledger is created by InitChain from the genesis app_state.
func NewOgoApp(logger log.Logger, homeDir string) *OgoApp {
	return &OgoApp{
		logger:       logger,
		homeDir:      homeDir,
		stateFile:    filepath.Join(homeDir, "data", "ogocoin", "state.json"),
		appStateFile: filepath.Join(homeDir, "data", "ogocoin", "app.json"),
	}
}

func (a *OgoApp) persistLocked() error {
	if err := a.persistLedgerLocked(); err != nil {
		return err
	}
	return a.persistAppStateLocked()
}

func (a *OgoApp) persistLedgerLocked() error {
	b, err := a.ledger.Marshal()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(a.stateFile), 0o755); err != nil {
		return err
	}
	tmp := a.stateFile + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, a.stateFile)
}

// persistAppStateLocked writes the last committed height and the
// epoch timer state. It must be called with a.mu held, after the
// ledger is persisted.
func (a *OgoApp) persistAppStateLocked() error {
	st := persistedAppState{Height: a.height}
	if a.epochTimer != nil {
		s := a.epochTimer.State()
		st.Epoch = &s
	}
	b, err := json.Marshal(st)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(a.appStateFile), 0o755); err != nil {
		return err
	}
	tmp := a.appStateFile + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, a.appStateFile)
}

// loadPersistedLocked reloads the ledger, height, and epoch timer
// state from disk. It must be called with a.mu held.
func (a *OgoApp) loadPersistedLocked() {
	if b, err := os.ReadFile(a.stateFile); err == nil {
		if l, err := ogocoin.UnmarshalLedger(b); err == nil {
			a.ledger = l
		}
	}
	if b, err := os.ReadFile(a.appStateFile); err == nil {
		var st persistedAppState
		if err := json.Unmarshal(b, &st); err == nil {
			a.height = st.Height
			if st.Epoch != nil {
				a.epochTimer = ogocoin.NewEpochTimerFromState(*st.Epoch)
			}
		}
	}
	a.rewireEpochHookLocked()
}

// rewireEpochHookLocked (re)registers the monetary epoch hook against
// the current ledger. The hook holds a ledger pointer, and a restart
// reloads a new ledger instance, so the hook must be re-pointed every
// time the ledger is (re)loaded. It must be called with a.mu held.
func (a *OgoApp) rewireEpochHookLocked() {
	if a.epochTimer == nil || a.ledger == nil {
		return
	}
	// The timer is rebuilt from persisted state on reload, so its
	// hook list is empty; on first InitChain it is also empty.
	// Register exactly one monetary hook per ledger instance.
	a.epochTimer = ogocoin.NewEpochTimerFromState(a.epochTimer.State())
	a.epochTimer.RegisterHook(&ogocoin.MonetaryEpochHook{Ledger: a.ledger})
}

func (a *OgoApp) apphashLocked() []byte {
	b, err := a.ledger.Marshal()
	if err != nil {
		return a.appHash
	}
	h := sha256.Sum256(b)
	return h[:]
}

// executeTx applies one JSON tx to the ledger. It is shared by CheckTx
// (on a clone, read-only) and FinalizeBlock (for real).
func executeTx(l *ogocoin.Ledger, raw []byte) (uint32, string) {
	var tx Tx
	if err := json.Unmarshal(raw, &tx); err != nil {
		return 1, fmt.Sprintf("bad tx json: %v", err)
	}
	amt, ok := parseAmount(tx.Amount)
	if !ok {
		return 1, fmt.Sprintf("bad amount %q", tx.Amount)
	}
	fee, ok := parseAmount(tx.Fee)
	if !ok {
		return 1, fmt.Sprintf("bad fee %q", tx.Fee)
	}
	switch tx.Type {
	case "send":
		if tx.From == "" || tx.To == "" {
			return 1, "send requires from and to"
		}
		if err := l.DeductFee(tx.From, fee); err != nil {
			return 1, err.Error()
		}
		if err := l.Send(tx.From, tx.To, amt); err != nil {
			return 1, err.Error()
		}
		return 0, fmt.Sprintf("sent %s %s from %s to %s (fee %s)", amt.String(), l.Denom, tx.From, tx.To, fee.String())
	case "burn":
		if tx.From == "" {
			return 1, "burn requires from"
		}
		if err := l.Burn(tx.From, amt); err != nil {
			return 1, err.Error()
		}
		return 0, fmt.Sprintf("burned %s %s from %s", amt.String(), l.Denom, tx.From)
	default:
		return 1, fmt.Sprintf("unknown tx type %q", tx.Type)
	}
}

func parseAmount(s string) (math.Int, bool) {
	if s == "" {
		return math.NewInt(0), true
	}
	amt, ok := math.NewIntFromString(s)
	if !ok {
		return math.Int{}, false
	}
	return amt, true
}

// Info implements abci.Application.
func (a *OgoApp) Info(_ context.Context, req *abcitypes.RequestInfo) (*abcitypes.ResponseInfo, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	// Reload persisted state if present (node restarts).
	a.loadPersistedLocked()
	if a.ledger != nil {
		a.appHash = a.apphashLocked()
	}
	return &abcitypes.ResponseInfo{
		Data:             "ogo-network",
		Version:          "0.2.0-milestones-1-2",
		AppVersion:       2,
		LastBlockHeight:  a.height,
		LastBlockAppHash: a.appHash,
	}, nil
}

// Query implements abci.Application. Paths:
//   - "balance": data is the bech32 address, returns {"address","balance","denom"}
//   - "supply": returns {"total_supply","max_supply","headroom","burned","denom"}
//   - "params": returns the genesis params (max_supply included)
//   - "global-supply": data is a channel id, returns escrow/voucher/global figures
//   - "epochs": returns the epoch timer state (or {"enabled":"false"}
//     when epoch mode is off), including the issued-by-epoch set
func (a *OgoApp) Query(_ context.Context, req *abcitypes.RequestQuery) (*abcitypes.ResponseQuery, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.ledger == nil {
		return &abcitypes.ResponseQuery{Code: 1, Log: "ledger not initialised"}, nil
	}
	// RPC clients may pass paths with or without a leading slash.
	path := req.Path
	if len(path) > 0 && path[0] == '/' {
		path = path[1:]
	}
	switch path {
	case "balance":
		addr := string(req.Data)
		return &abcitypes.ResponseQuery{Code: 0, Value: mustJSON(map[string]string{
			"address": addr,
			"balance": a.ledger.BalanceOf(addr).String(),
			"denom":   a.ledger.Denom,
		})}, nil
	case "supply":
		return &abcitypes.ResponseQuery{Code: 0, Value: mustJSON(map[string]string{
			"total_supply": a.ledger.TotalSupply().String(),
			"max_supply":   a.ledger.MaxSupply.String(),
			"headroom":     a.ledger.Headroom().String(),
			"burned":       a.ledger.BurnedAmount().String(),
			"denom":        a.ledger.Denom,
		})}, nil
	case "params":
		return &abcitypes.ResponseQuery{Code: 0, Value: mustJSON(map[string]string{
			"max_supply":          a.ledger.MaxSupply.String(),
			"denom":               a.ledger.Denom,
			"inflation_rate":      "0",
			"minting_enabled":     "false",
			"supply_cap_enforced": "true",
		})}, nil
	case "schedule":
		// The predetermined issuance schedule (Constitution Article III)
		// plus issuance progress, for independent verification.
		sched := a.ledger.Schedule
		return &abcitypes.ResponseQuery{Code: 0, Value: mustJSON(map[string]string{
			"schedule_type":        sched.ScheduleType,
			"start_height":         fmt.Sprintf("%d", sched.StartHeight),
			"period_length_blocks": fmt.Sprintf("%d", sched.PeriodLength),
			"periods":              fmt.Sprintf("%d", sched.Periods),
			"initial_per_period":   sched.InitialPerPeriod.String(),
			"reduction":            fmt.Sprintf("%d/%d", sched.ReductionNumerator, sched.ReductionDenominator),
			"total_scheduled":      sched.TotalIssuance().String(),
			"total_issued":         a.ledger.TotalIssued.String(),
			"total_supply":         a.ledger.TotalSupply().String(),
			"max_supply":           a.ledger.MaxSupply.String(),
		})}, nil
	case "global-supply":
		ch := string(req.Data)
		esc := a.ledger.Escrowed(ch)
		v := a.ledger.VouchersOutstanding(ch)
		global := ogocoin.GlobalEconomicSupply(a.ledger.TotalSupply(), esc, v)
		return &abcitypes.ResponseQuery{Code: 0, Value: mustJSON(map[string]string{
			"channel":                ch,
			"chain_local_supply":     a.ledger.TotalSupply().String(),
			"escrowed":               esc.String(),
			"vouchers_outstanding":   v.String(),
			"global_economic_supply": global.String(),
			"naive_double_count":     ogocoin.NaiveDoubleCount(a.ledger.TotalSupply(), v).String(),
		})}, nil
	case "epochs":
		// The epoch timer state. "implementation" is stated plainly:
		// this is the epoch timer mirror, not the SDK x/epochs module.
		if a.epochTimer == nil {
			return &abcitypes.ResponseQuery{Code: 0, Value: mustJSON(map[string]string{
				"enabled":        "false",
				"implementation": "epoch-timer-mirror (not SDK x/epochs); see EPOCHS_GAP.md",
			})}, nil
		}
		info := a.epochTimer.Info()
		state := a.epochTimer.State()
		issued := a.ledger.IssuedByEpochCopy()
		periods := make([]int64, 0, len(issued))
		for p := range issued {
			periods = append(periods, p)
		}
		sort.Slice(periods, func(i, j int) bool { return periods[i] < periods[j] })
		var issuedList []string
		for _, p := range periods {
			issuedList = append(issuedList, fmt.Sprintf("%d:%s", p, issued[p].String()))
		}
		return &abcitypes.ResponseQuery{Code: 0, Value: mustJSON(map[string]string{
			"enabled":                    "true",
			"implementation":             "epoch-timer-mirror (not SDK x/epochs); see EPOCHS_GAP.md",
			"identifier":                 info.Identifier,
			"duration_seconds":           fmt.Sprintf("%d", int64(info.Duration/time.Second)),
			"start_time":                 info.StartTime.Format(time.RFC3339),
			"current_epoch":              fmt.Sprintf("%d", info.CurrentEpoch),
			"current_epoch_start_time":   info.CurrentEpochStartTime.Format(time.RFC3339),
			"current_epoch_start_height": fmt.Sprintf("%d", info.CurrentEpochStartHeight),
			"epoch_counting_started":     fmt.Sprintf("%v", info.EpochCountingStarted),
			"hook_fire_count":            fmt.Sprintf("%d", state.HookFireCount),
			"last_hook_error":            state.LastHookError,
			"total_issued":               a.ledger.TotalIssuedAmount().String(),
			"issued_by_period":           strings.Join(issuedList, ","),
		})}, nil
	default:
		return &abcitypes.ResponseQuery{Code: 1, Log: fmt.Sprintf("unknown query path %q", req.Path)}, nil
	}
}

// CheckTx implements abci.Application: validate on a clone, no mutation.
func (a *OgoApp) CheckTx(_ context.Context, req *abcitypes.RequestCheckTx) (*abcitypes.ResponseCheckTx, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.ledger == nil {
		return &abcitypes.ResponseCheckTx{Code: 1, Log: "ledger not initialised"}, nil
	}
	code, logMsg := executeTx(a.ledger.Clone(), req.Tx)
	return &abcitypes.ResponseCheckTx{Code: code, Log: logMsg, GasWanted: 1000}, nil
}

// InitChain implements abci.Application.
func (a *OgoApp) InitChain(_ context.Context, req *abcitypes.RequestInitChain) (*abcitypes.ResponseInitChain, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	var appState struct {
		Ogocoin ogocoin.GenesisState `json:"ogocoin"`
		Epochs  *ogocoin.EpochConfig `json:"epochs"`
	}
	if err := json.Unmarshal(req.AppStateBytes, &appState); err != nil {
		return nil, fmt.Errorf("ogocoin: bad app_state: %w", err)
	}
	l, err := ogocoin.NewLedger(appState.Ogocoin)
	if err != nil {
		return nil, fmt.Errorf("ogocoin: genesis validation failed: %w", err)
	}
	a.ledger = l
	a.height = 0
	a.appHash = a.apphashLocked()
	// Epoch mode: an epochs config in genesis starts the epoch timer
	// mirror, anchored at the genesis time. Without it, epoch mode
	// stays off and issuance uses the legacy block-based path.
	a.epochTimer = nil
	if appState.Epochs != nil && !appState.Epochs.Disabled() {
		if err := appState.Epochs.Validate(); err != nil {
			return nil, fmt.Errorf("ogocoin: bad epochs config: %w", err)
		}
		a.epochTimer = ogocoin.NewEpochTimer(*appState.Epochs, req.Time)
	}
	a.rewireEpochHookLocked()
	if err := a.persistLocked(); err != nil {
		return nil, err
	}
	a.logger.Info("ogo-network InitChain", "chain_id", req.ChainId, "max_supply", l.MaxSupply.String(), "epochs_enabled", a.epochTimer != nil)
	return &abcitypes.ResponseInitChain{Validators: req.Validators}, nil
}

// PrepareProposal implements abci.Application.
func (a *OgoApp) PrepareProposal(_ context.Context, req *abcitypes.RequestPrepareProposal) (*abcitypes.ResponsePrepareProposal, error) {
	return &abcitypes.ResponsePrepareProposal{Txs: req.Txs}, nil
}

// ProcessProposal implements abci.Application.
func (a *OgoApp) ProcessProposal(_ context.Context, _ *abcitypes.RequestProcessProposal) (*abcitypes.ResponseProcessProposal, error) {
	return &abcitypes.ResponseProcessProposal{Status: abcitypes.ResponseProcessProposal_ACCEPT}, nil
}

// FinalizeBlock implements abci.Application.
func (a *OgoApp) FinalizeBlock(_ context.Context, req *abcitypes.RequestFinalizeBlock) (*abcitypes.ResponseFinalizeBlock, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.ledger == nil {
		return nil, fmt.Errorf("ogocoin: ledger not initialised (InitChain did not run)")
	}
	// Epoch tick first (mirrors the SDK BeginBlocker, which runs
	// before transactions). In epoch mode the AfterEpochEnd hook is
	// the ONLY issuance path; in legacy mode the block-based
	// MintScheduled path below runs instead. Never both.
	var events []abcitypes.Event
	if a.epochTimer != nil {
		issuedBefore := a.ledger.TotalIssuedAmount()
		tick := a.epochTimer.Tick(req.Time, req.Height)
		if tick.HookErr != nil {
			// Upstream behaviour: a failed hook is logged and the
			// epoch still advances. The failure is recorded on the
			// timer (LastHookError) and visible in `query epochs`.
			a.logger.Error("ogo-network epoch hook failed", "epoch", tick.EndedEpoch, "err", tick.HookErr)
		}
		if tick.EndedEpoch != 0 {
			events = append(events, abcitypes.Event{
				Type: "epoch_end",
				Attributes: []abcitypes.EventAttribute{
					{Key: "identifier", Value: a.epochTimer.Info().Identifier},
					{Key: "epoch_number", Value: fmt.Sprintf("%d", tick.EndedEpoch)},
				},
			})
		}
		if tick.StartedEpoch != 0 {
			events = append(events, abcitypes.Event{
				Type: "epoch_start",
				Attributes: []abcitypes.EventAttribute{
					{Key: "identifier", Value: a.epochTimer.Info().Identifier},
					{Key: "epoch_number", Value: fmt.Sprintf("%d", tick.StartedEpoch)},
					{Key: "epoch_start_time", Value: a.epochTimer.Info().CurrentEpochStartTime.Format(time.RFC3339)},
				},
			})
		}
		if tick.HookFired {
			issued := a.ledger.TotalIssuedAmount().Sub(issuedBefore)
			if issued.IsPositive() {
				events = append(events, abcitypes.Event{
					Type: "epoch_issuance",
					Attributes: []abcitypes.EventAttribute{
						{Key: "identifier", Value: a.epochTimer.Info().Identifier},
						{Key: "epoch_number", Value: fmt.Sprintf("%d", tick.EndedEpoch)},
						{Key: "amount", Value: issued.String()},
						{Key: "recipient", Value: ogocoin.RewardsPool},
					},
				})
			}
		}
	}
	results := make([]*abcitypes.ExecTxResult, 0, len(req.Txs))
	for _, raw := range req.Txs {
		code, logMsg := executeTx(a.ledger, raw)
		results = append(results, &abcitypes.ExecTxResult{Code: code, Log: logMsg})
	}
	if a.epochTimer == nil {
		// Legacy block-based scheduled issuance (Constitution
		// Article III): the protocol issues exactly the schedule
		// amount for this block's period, once, into the rewards
		// pool. Not-started / exhausted / already-issued are
		// expected skips; any other error (e.g. cap) fails the block.
		if err := a.ledger.MintScheduled(req.Height); err != nil {
			if !errors.Is(err, ogocoin.ErrScheduleNotStarted) && !errors.Is(err, ogocoin.ErrScheduleExhausted) && !errors.Is(err, ogocoin.ErrDoubleIssuance) {
				return nil, fmt.Errorf("ogocoin: scheduled issuance failed in block %d: %w", req.Height, err)
			}
		}
	}
	if err := a.ledger.CheckInvariants(); err != nil {
		return nil, fmt.Errorf("ogocoin: supply invariant violated in block %d: %w", req.Height, err)
	}
	a.height = req.Height
	a.appHash = a.apphashLocked()
	if err := a.persistLocked(); err != nil {
		return nil, err
	}
	return &abcitypes.ResponseFinalizeBlock{TxResults: results, AppHash: a.appHash, Events: events}, nil
}

// Commit implements abci.Application.
func (a *OgoApp) Commit(_ context.Context, _ *abcitypes.RequestCommit) (*abcitypes.ResponseCommit, error) {
	return &abcitypes.ResponseCommit{RetainHeight: 0}, nil
}

// ListSnapshots implements abci.Application.
func (a *OgoApp) ListSnapshots(_ context.Context, _ *abcitypes.RequestListSnapshots) (*abcitypes.ResponseListSnapshots, error) {
	return &abcitypes.ResponseListSnapshots{}, nil
}

// OfferSnapshot implements abci.Application.
func (a *OgoApp) OfferSnapshot(_ context.Context, _ *abcitypes.RequestOfferSnapshot) (*abcitypes.ResponseOfferSnapshot, error) {
	return &abcitypes.ResponseOfferSnapshot{Result: abcitypes.ResponseOfferSnapshot_REJECT}, nil
}

// LoadSnapshotChunk implements abci.Application.
func (a *OgoApp) LoadSnapshotChunk(_ context.Context, _ *abcitypes.RequestLoadSnapshotChunk) (*abcitypes.ResponseLoadSnapshotChunk, error) {
	return &abcitypes.ResponseLoadSnapshotChunk{}, nil
}

// ApplySnapshotChunk implements abci.Application.
func (a *OgoApp) ApplySnapshotChunk(_ context.Context, _ *abcitypes.RequestApplySnapshotChunk) (*abcitypes.ResponseApplySnapshotChunk, error) {
	return &abcitypes.ResponseApplySnapshotChunk{Result: abcitypes.ResponseApplySnapshotChunk_ABORT}, nil
}

// ExtendVote implements abci.Application.
func (a *OgoApp) ExtendVote(_ context.Context, _ *abcitypes.RequestExtendVote) (*abcitypes.ResponseExtendVote, error) {
	return &abcitypes.ResponseExtendVote{}, nil
}

// VerifyVoteExtension implements abci.Application.
func (a *OgoApp) VerifyVoteExtension(_ context.Context, _ *abcitypes.RequestVerifyVoteExtension) (*abcitypes.ResponseVerifyVoteExtension, error) {
	return &abcitypes.ResponseVerifyVoteExtension{Status: abcitypes.ResponseVerifyVoteExtension_ACCEPT}, nil
}

func mustJSON(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		return []byte("{}")
	}
	return b
}
