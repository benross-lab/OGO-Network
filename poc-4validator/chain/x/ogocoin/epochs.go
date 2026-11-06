// Epoch timer mirror for the local prototype.
//
// This file implements a faithful mirror of the Cosmos SDK x/epochs
// module (v0.53.4, the pinned version) inside x/ogocoin, because the
// Milestone B scaffold is a custom ABCI application, not a full
// depinject SDK application, and real x/epochs cannot be wired into
// it without the full keeper wiring that is Milestone C work. See
// chain/EPOCHS_GAP.md for exactly what is mirrored, what is not, and
// what remains to wire real x/epochs.
//
// What is mirrored exactly, read from the pinned SDK source
// (cosmos-sdk@v0.53.4, x/epochs/keeper/abci.go, keeper/hooks.go,
// keeper/keeper.go, types/hooks.go, types/genesis.pb.go):
//   - One epoch timer per identifier; the monetary identifier is
//     "monetary". Hooks filter on the identifier.
//   - Epoch numbers are 1-indexed: the first epoch is epoch 1. The
//     first tick is the first block whose block time is not before
//     the configured start time. Schedule periods are 0-indexed, so
//     epoch e mints the amount for schedule period e-1.
//   - The epoch grid is anchored at the genesis start time. When an
//     epoch ends, the next epoch starts at the prior epoch end time
//     (start time plus duration), never at the ticking block time.
//     An epoch ends at the first block whose block time is strictly
//     after the epoch end time (start time plus duration).
//   - At most ONE epoch ticks per block per identifier. After a halt,
//     missed epochs tick one per block on restart, in epoch order,
//     with no lump mint and no skipped epoch.
//   - Hooks: AfterEpochEnd(identifier, ending epoch number) and
//     BeforeEpochStart(identifier, starting epoch number), run in
//     registration order. A hook failure is logged and recorded and
//     does NOT halt the chain: the epoch still advances. That is the
//     upstream behaviour (the keeper runs hooks in a cached context
//     and discards their writes on error). The mirror records the
//     failure in LastHookError so a skipped issuance is visible
//     instead of silent.
//   - Events: epoch_end carries the ending epoch number, epoch_start
//     carries the new epoch number and its start time.
//
// What is NOT real x/epochs: there is no SDK keeper, no store key,
// no depinject wiring, no gRPC query service, and no StoreUpgrade.
// The timer state lives in the application's own state file, not in
// an SDK store. Do not present this as SDK x/epochs integration.
//
// Determinism: epoch ticks are driven by block time only, and the
// monetary rules depend only on the epoch number and the fixed
// duration, never on wall-clock time. Two nodes with the same state
// agree on the current epoch number.
//
// All parameters here are TEST-ONLY when used in the localnet. No
// production value is selected anywhere in this file.
package ogocoin

import (
	"fmt"
	"time"

	"cosmossdk.io/math"
)

// EpochIdentifierMonetary is the epoch identifier for the monetary
// schedule. It is the only identifier x/ogocoin listens to. Other
// identifiers may be added later (for example a daily identifier for
// incentive distribution), but the monetary hook ignores them.
const EpochIdentifierMonetary = "monetary"

// EpochConfig is the genesis configuration for one epoch timer. It is
// carried in the genesis app_state under the "epochs" key, separate
// from the ogocoin genesis state, because this file must not redefine
// any symbol owned by the monetary interface files (types.go,
// genesis.go, schedule.go, ledger.go).
type EpochConfig struct {
	// Identifier is the timer name, for example "monetary".
	Identifier string `json:"identifier"`
	// DurationSeconds is D, the fixed epoch duration in seconds.
	DurationSeconds int64 `json:"duration_seconds"`
}

// Disabled reports whether the epoch timer is off. A duration of zero
// (or an absent config) means epoch mode is disabled and the app
// keeps the legacy block-based MintScheduled path.
func (c EpochConfig) Disabled() bool {
	return c.DurationSeconds <= 0
}

// Validate checks the epoch configuration. Duration must be positive
// and, per the upstream guidance, greater than the expected block
// time; the localnet uses durations far above the ~1s block time.
func (c EpochConfig) Validate() error {
	if c.Identifier == "" {
		return fmt.Errorf("ogocoin: epoch identifier must not be empty")
	}
	if c.DurationSeconds <= 0 {
		return fmt.Errorf("ogocoin: epoch duration_seconds must be > 0, got %d", c.DurationSeconds)
	}
	return nil
}

// EpochInfo mirrors the SDK x/epochs types.EpochInfo field for field,
// with the same meaning. See the package comment for the semantics.
type EpochInfo struct {
	Identifier              string        `json:"identifier"`
	StartTime               time.Time     `json:"start_time"`
	Duration                time.Duration `json:"duration"`
	CurrentEpoch            int64         `json:"current_epoch"`
	CurrentEpochStartHeight int64         `json:"current_epoch_start_height"`
	CurrentEpochStartTime   time.Time     `json:"current_epoch_start_time"`
	EpochCountingStarted    bool          `json:"epoch_counting_started"`
}

// EpochHooks mirrors the SDK x/epochs types.EpochHooks interface. A
// receiving module must filter on the epoch identifier and execute
// only for its own identifier.
type EpochHooks interface {
	// AfterEpochEnd runs when an epoch ends. epochNumber is the number
	// of the epoch that is ending.
	AfterEpochEnd(identifier string, epochNumber int64) error
	// BeforeEpochStart runs when a new epoch starts. epochNumber is
	// the number of the epoch that is starting.
	BeforeEpochStart(identifier string, epochNumber int64) error
}

// EpochTickResult is what one Tick call did. At most one epoch ends
// and one epoch starts per block.
type EpochTickResult struct {
	// EndedEpoch is the epoch number that ended, or 0 if none ended.
	EndedEpoch int64
	// StartedEpoch is the epoch number that started, or 0 if none started.
	StartedEpoch int64
	// HookFired is true when AfterEpochEnd was invoked on the
	// registered hooks in this tick.
	HookFired bool
	// HookErr is the hook error, if the hook failed. The epoch still
	// advanced; the error is recorded, not hidden.
	HookErr error
}

// EpochTimerState is the persistable part of an EpochTimer: the epoch
// info plus hook observability counters. Hooks themselves are not
// persisted; the application re-registers them after a restart.
type EpochTimerState struct {
	Info          EpochInfo `json:"info"`
	LastHookError string    `json:"last_hook_error,omitempty"`
	HookFireCount int64     `json:"hook_fire_count"`
}

// EpochTimer is the epoch timer mirror. It holds one EpochInfo and a
// list of registered hooks, and advances the timer exactly the way
// the pinned SDK x/epochs BeginBlocker does.
type EpochTimer struct {
	state EpochTimerState
	hooks []EpochHooks
}

// NewEpochTimer creates a timer from genesis configuration. The epoch
// grid is anchored at genesisTime: epoch 1 covers
// (genesisTime, genesisTime + D].
func NewEpochTimer(cfg EpochConfig, genesisTime time.Time) *EpochTimer {
	return &EpochTimer{
		state: EpochTimerState{
			Info: EpochInfo{
				Identifier:           cfg.Identifier,
				StartTime:            genesisTime,
				Duration:             time.Duration(cfg.DurationSeconds) * time.Second,
				CurrentEpoch:         0,
				EpochCountingStarted: false,
			},
		},
	}
}

// NewEpochTimerFromState restores a timer from persisted state. Hooks
// must be re-registered by the caller after restore.
func NewEpochTimerFromState(state EpochTimerState) *EpochTimer {
	return &EpochTimer{state: state}
}

// State returns the persistable timer state.
func (t *EpochTimer) State() EpochTimerState {
	return t.state
}

// Info returns the current epoch info.
func (t *EpochTimer) Info() EpochInfo {
	return t.state.Info
}

// RegisterHook adds a hook. Hooks run in registration order.
func (t *EpochTimer) RegisterHook(h EpochHooks) {
	t.hooks = append(t.hooks, h)
}

// Tick advances the timer for one block, mirroring the pinned SDK
// x/epochs BeginBlocker exactly:
//   - If block time is before the start time, nothing happens.
//   - If counting has not started, epoch 1 starts (no AfterEpochEnd
//     fires for a non-existent epoch 0). BeforeEpochStart fires.
//   - Otherwise, if block time is strictly after the current epoch
//     end time (current start time plus duration), the current epoch
//     ends: AfterEpochEnd fires with the ending epoch number, then
//     the next epoch starts at the prior end time. Only one epoch
//     ends per Tick call, so missed epochs catch up one per block.
//   - A hook error is recorded in LastHookError and does not stop
//     the epoch from advancing (upstream behaviour).
func (t *EpochTimer) Tick(blockTime time.Time, height int64) EpochTickResult {
	info := &t.state.Info
	if blockTime.Before(info.StartTime) {
		return EpochTickResult{}
	}
	if !info.EpochCountingStarted {
		info.EpochCountingStarted = true
		info.CurrentEpoch = 1
		info.CurrentEpochStartHeight = height
		info.CurrentEpochStartTime = info.StartTime
		t.runBeforeEpochStart(info.Identifier, info.CurrentEpoch)
		return EpochTickResult{StartedEpoch: 1}
	}
	epochEndTime := info.CurrentEpochStartTime.Add(info.Duration)
	if !blockTime.After(epochEndTime) {
		return EpochTickResult{}
	}
	ended := info.CurrentEpoch
	result := EpochTickResult{EndedEpoch: ended, HookFired: true}
	if err := t.runAfterEpochEnd(info.Identifier, ended); err != nil {
		result.HookErr = err
		t.state.LastHookError = fmt.Sprintf("epoch %d: %v", ended, err)
	} else {
		t.state.LastHookError = ""
	}
	t.state.HookFireCount++
	info.CurrentEpoch++
	info.CurrentEpochStartTime = info.CurrentEpochStartTime.Add(info.Duration)
	info.CurrentEpochStartHeight = height
	t.runBeforeEpochStart(info.Identifier, info.CurrentEpoch)
	result.StartedEpoch = info.CurrentEpoch
	return result
}

// runAfterEpochEnd invokes AfterEpochEnd on every registered hook and
// joins any errors. A hook failure does not propagate as a Tick
// failure: the caller records it and the epoch still advances.
func (t *EpochTimer) runAfterEpochEnd(identifier string, epochNumber int64) error {
	var errs error
	for _, h := range t.hooks {
		if err := h.AfterEpochEnd(identifier, epochNumber); err != nil {
			errs = joinErr(errs, err)
		}
	}
	return errs
}

// runBeforeEpochStart invokes BeforeEpochStart on every registered
// hook. Errors are recorded the same way as AfterEpochEnd errors.
func (t *EpochTimer) runBeforeEpochStart(identifier string, epochNumber int64) {
	if err := func() error {
		var errs error
		for _, h := range t.hooks {
			if err := h.BeforeEpochStart(identifier, epochNumber); err != nil {
				errs = joinErr(errs, err)
			}
		}
		return errs
	}(); err != nil {
		t.state.LastHookError = fmt.Sprintf("before epoch %d start: %v", epochNumber, err)
	}
}

// joinErr appends err to errs, the way errors.Join does, without
// importing errors here for a single use.
func joinErr(errs, err error) error {
	if errs == nil {
		return err
	}
	return fmt.Errorf("%v; %v", errs, err)
}

// MonetaryEpochHook is the x/ogocoin hook that executes scheduled
// issuance at the end of each monetary epoch. It filters on the
// monetary identifier and ignores all others.
type MonetaryEpochHook struct {
	// Ledger is the ledger the hook mints into. The application must
	// re-point this hook at the reloaded ledger after a restart.
	Ledger *Ledger
}

// AfterEpochEnd mints exactly the schedule amount for the ending
// epoch, once, into the rewards pool. Epoch e maps to schedule
// period e-1 (epochs are 1-indexed, periods are 0-indexed). When the
// schedule is exhausted (epoch beyond the last period) there is
// nothing to mint and the hook succeeds without minting: epochs keep
// ticking after issuance ends, they just issue nothing.
func (h *MonetaryEpochHook) AfterEpochEnd(identifier string, epochNumber int64) error {
	if identifier != EpochIdentifierMonetary {
		return nil
	}
	if h.Ledger == nil {
		return fmt.Errorf("ogocoin: monetary epoch hook has no ledger")
	}
	if err := h.Ledger.MintScheduledForEpoch(epochNumber); err != nil {
		if isExhausted(err) {
			return nil
		}
		return err
	}
	return nil
}

// BeforeEpochStart does nothing for the monetary module: issuance
// happens at epoch end, not at epoch start.
func (h *MonetaryEpochHook) BeforeEpochStart(_ string, _ int64) error {
	return nil
}

// isExhausted reports whether err is (or wraps) ErrScheduleExhausted.
func isExhausted(err error) bool {
	for err != nil {
		if err == ErrScheduleExhausted {
			return true
		}
		u, ok := err.(interface{ Unwrap() error })
		if !ok {
			return false
		}
		err = u.Unwrap()
	}
	return false
}

// MintScheduledForEpoch issues exactly the schedule amount for the
// schedule period of an epoch, once, into the rewards pool. It is the
// epoch-mode counterpart of MintScheduled (which is keyed by block
// height). Both share the IssuedByPeriod set, keyed by period index,
// so a period can never be claimed twice no matter which path claims
// it first.
//
// Epoch numbers are 1-indexed (SDK x/epochs: the first epoch is
// epoch 1); schedule periods are 0-indexed, so epoch e mints
// AmountForPeriod(e-1). It fails when the epoch is past the last
// period (exhausted; issuance stops permanently), when the period
// was already claimed (double issuance), or when the amount would
// push total supply over max_supply.
func (l *Ledger) MintScheduledForEpoch(epochNumber int64) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if epochNumber <= 0 {
		return fmt.Errorf("%w: epoch number %d", ErrScheduleInvalid, epochNumber)
	}
	period := uint64(epochNumber - 1)
	if period >= l.Schedule.Periods {
		return ErrScheduleExhausted
	}
	if _, claimed := l.IssuedByPeriod[int64(period)]; claimed {
		return fmt.Errorf("%w: epoch %d (period %d) already issued", ErrDoubleIssuance, epochNumber, period)
	}
	amount := l.Schedule.AmountForPeriod(period)
	// A period whose floored amount is zero issues nothing; record it
	// so the period cannot be re-attempted, and return success.
	l.IssuedByPeriod[int64(period)] = amount
	if amount.IsZero() {
		return nil
	}
	if l.totalLocked().Add(amount).GT(l.MaxSupply) {
		delete(l.IssuedByPeriod, int64(period))
		return fmt.Errorf("%w: scheduled issuance %s for epoch %d would exceed max %s", ErrSupplyOverMax, amount.String(), epochNumber, l.MaxSupply.String())
	}
	pool := l.Balances[RewardsPool]
	if pool.IsNil() {
		pool = math.NewInt(0)
	}
	l.Balances[RewardsPool] = pool.Add(amount)
	l.TotalIssued = l.TotalIssued.Add(amount)
	return l.CheckInvariants()
}

// TotalIssuedAmount returns the cumulative scheduled issuance.
func (l *Ledger) TotalIssuedAmount() math.Int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.TotalIssued
}

// IssuedByEpochCopy returns a copy of the issued-by-period set, keyed
// by period index (which equals epoch number minus one). This is the
// duplicate-issuance guard, exposed for queries and tests.
func (l *Ledger) IssuedByEpochCopy() map[int64]math.Int {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make(map[int64]math.Int, len(l.IssuedByPeriod))
	for k, v := range l.IssuedByPeriod {
		out[k] = v
	}
	return out
}
