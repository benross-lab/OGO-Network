package monetary

import (
	"encoding/json"
	"fmt"
	"sync"

	"cosmossdk.io/math"
)

// Ledger is the in-memory fixed-supply state machine for the local
// prototype. It is the single source of truth for uogo balances in the
// ABCI app; CometBFT persists blocks, the ledger persists supply state.
//
// Monetary Constitution v0.1: the ONLY minting this ledger permits is
// scheduled issuance (MintScheduled), for exactly the schedule amount,
// once per period, into the rewards pool. Discretionary minting of any
// kind always fails (see Mint).
//
// Invariants (checked by CheckInvariants, called after every mutation):
//   - total supply (balances + escrow) <= max_supply
//   - no mutation can increase total supply except scheduled issuance,
//     which is itself capped by the genesis-validated schedule total
//   - burned coins never return
type Ledger struct {
	mu sync.Mutex

	MaxSupply math.Int `json:"max_supply"`
	Denom     string   `json:"denom"`

	// Balances maps address -> uogo balance. The fee collector is a
	// normal entry here; fees move coins, they never create them.
	Balances map[string]math.Int `json:"balances"`
	// Escrow maps channelID -> amount of native uogo locked on this chain
	// backing vouchers on a counterparty chain. Escrowed coins are still
	// part of this chain's bank supply (they sit in the escrow account),
	// so they are included in TotalSupply.
	Escrow map[string]math.Int `json:"escrow"`
	// Vouchers maps channelID -> voucher amount outstanding on the
	// counterparty chain for the escrow in Escrow[channelID]. Vouchers
	// are claims on escrow, not additional supply.
	Vouchers map[string]math.Int `json:"vouchers"`
	// Burned is the cumulative amount burned. One-way sink; burning
	// never creates remintable credit.
	Burned math.Int `json:"burned"`
	// Schedule is the predetermined issuance schedule (mandatory).
	Schedule IssuanceSchedule `json:"issuance_schedule"`
	// IssuedByPeriod records the amount issued per period index, so a
	// period can never be claimed twice.
	IssuedByPeriod map[int64]math.Int `json:"issued_by_period"`
	// TotalIssued is the cumulative scheduled issuance.
	TotalIssued math.Int `json:"total_issued"`
}

// NewLedger builds a ledger from validated genesis state. Milestone C:
// the default genesis path is the STRICT check (ValidateGenesisExact),
// so headroom above G + S_actual is rejected here, not only in a
// separate helper.
func NewLedger(gs GenesisState) (*Ledger, error) {
	if err := ValidateGenesisExact(gs); err != nil {
		return nil, err
	}
	l := &Ledger{
		MaxSupply:      *gs.Params.MaxSupply,
		Denom:          gs.Params.Denom,
		Balances:       make(map[string]math.Int, len(gs.Balances)),
		Escrow:         make(map[string]math.Int),
		Vouchers:       make(map[string]math.Int),
		Burned:         math.NewInt(0),
		Schedule:       *gs.Params.IssuanceSchedule,
		IssuedByPeriod: make(map[int64]math.Int),
		TotalIssued:    math.NewInt(0),
	}
	for _, b := range gs.Balances {
		l.Balances[b.Address] = b.Amount
	}
	return l, nil
}

// Clone returns a deep copy (used by tests to prove state is unchanged
// after a rejected operation).
func (l *Ledger) Clone() *Ledger {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.cloneLocked()
}

func (l *Ledger) cloneLocked() *Ledger {
	c := &Ledger{
		MaxSupply:      l.MaxSupply,
		Denom:          l.Denom,
		Balances:       make(map[string]math.Int, len(l.Balances)),
		Escrow:         make(map[string]math.Int, len(l.Escrow)),
		Vouchers:       make(map[string]math.Int, len(l.Vouchers)),
		Burned:         l.Burned,
		Schedule:       l.Schedule,
		IssuedByPeriod: make(map[int64]math.Int, len(l.IssuedByPeriod)),
		TotalIssued:    l.TotalIssued,
	}
	for k, v := range l.Balances {
		c.Balances[k] = v
	}
	for k, v := range l.Escrow {
		c.Escrow[k] = v
	}
	for k, v := range l.Vouchers {
		c.Vouchers[k] = v
	}
	for k, v := range l.IssuedByPeriod {
		c.IssuedByPeriod[k] = v
	}
	return c
}

// Equal reports whether two ledgers hold identical monetary state.
func (l *Ledger) Equal(o *Ledger) bool {
	a, b := l.Clone(), o.Clone()
	if !a.MaxSupply.Equal(b.MaxSupply) || a.Denom != b.Denom || !a.Burned.Equal(b.Burned) {
		return false
	}
	if !a.TotalIssued.Equal(b.TotalIssued) || !schedulesEqual(a.Schedule, b.Schedule) {
		return false
	}
	if len(a.Balances) != len(b.Balances) || len(a.Escrow) != len(b.Escrow) || len(a.Vouchers) != len(b.Vouchers) || len(a.IssuedByPeriod) != len(b.IssuedByPeriod) {
		return false
	}
	for k, v := range a.Balances {
		if ov, ok := b.Balances[k]; !ok || !v.Equal(ov) {
			return false
		}
	}
	for k, v := range a.Escrow {
		if ov, ok := b.Escrow[k]; !ok || !v.Equal(ov) {
			return false
		}
	}
	for k, v := range a.Vouchers {
		if ov, ok := b.Vouchers[k]; !ok || !v.Equal(ov) {
			return false
		}
	}
	for k, v := range a.IssuedByPeriod {
		if ov, ok := b.IssuedByPeriod[k]; !ok || !v.Equal(ov) {
			return false
		}
	}
	return true
}

// BurnedAmount returns the cumulative amount burned.
// Note: Denom and MaxSupply are exported struct fields (needed for JSON
// persistence), so there are intentionally no Denom()/MaxSupply() methods:
// a Go type cannot have a field and a method with the same name.
func (l *Ledger) BurnedAmount() math.Int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.Burned
}

// schedulesEqual compares two schedules by value (math.Int holds a
// pointer, so struct == would compare pointers, not values).
func schedulesEqual(a, b IssuanceSchedule) bool {
	return a.ScheduleType == b.ScheduleType &&
		a.StartHeight == b.StartHeight &&
		a.PeriodLength == b.PeriodLength &&
		a.Periods == b.Periods &&
		a.InitialPerPeriod.Equal(b.InitialPerPeriod) &&
		a.ReductionNumerator == b.ReductionNumerator &&
		a.ReductionDenominator == b.ReductionDenominator &&
		intEqual(a.BasePerPeriod, b.BasePerPeriod) &&
		a.TaperEnabled == b.TaperEnabled &&
		a.TaperNumerator == b.TaperNumerator &&
		a.TaperDenominator == b.TaperDenominator &&
		a.TaperPeriods == b.TaperPeriods
}

// intEqual compares two math.Int values treating a nil Int as zero, so
// schedules built before the hybrid fields existed (BasePerPeriod
// unset) still compare correctly.
func intEqual(a, b math.Int) bool {
	if a.IsNil() && b.IsNil() {
		return true
	}
	if a.IsNil() || b.IsNil() {
		return false
	}
	return a.Equal(b)
}

// BalanceOf returns an account's balance (zero if absent).
func (l *Ledger) BalanceOf(addr string) math.Int {
	l.mu.Lock()
	defer l.mu.Unlock()
	if b, ok := l.Balances[addr]; ok {
		return b
	}
	return math.NewInt(0)
}

// TotalSupply returns chain-local supply: sum(balances) + sum(escrow).
// Escrowed coins exist on this chain, so they count here.
func (l *Ledger) TotalSupply() math.Int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.totalLocked()
}

func (l *Ledger) totalLocked() math.Int {
	t := math.NewInt(0)
	for _, v := range l.Balances {
		t = t.Add(v)
	}
	for _, v := range l.Escrow {
		t = t.Add(v)
	}
	return t
}

// Headroom is max_supply - total supply. It is unminted headroom, not a
// reserve: minting is disabled, so headroom can never be drawn on.
func (l *Ledger) Headroom() math.Int {
	return l.MaxSupply.Sub(l.TotalSupply())
}

// CheckInvariants verifies total <= max. Called after every mutation.
func (l *Ledger) CheckInvariants() error {
	if l.totalLocked().GT(l.MaxSupply) {
		return fmt.Errorf("%w: total %s exceeds max %s", ErrSupplyOverMax, l.totalLocked().String(), l.MaxSupply.String())
	}
	return nil
}

// Send moves amount from one account to another. Total supply is exactly
// preserved. Amount must be > 0 and the sender must hold it.
func (l *Ledger) Send(from, to string, amount math.Int) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if amount.IsNil() || !amount.IsPositive() {
		return fmt.Errorf("%w: send amount %s", ErrZeroAmount, amount.String())
	}
	fromBal := l.Balances[from]
	if fromBal.IsNil() {
		fromBal = math.NewInt(0)
	}
	if fromBal.LT(amount) {
		return fmt.Errorf("%w: %s holds %s, tried to send %s", ErrInsufficient, from, fromBal.String(), amount.String())
	}
	before := l.totalLocked()
	l.Balances[from] = fromBal.Sub(amount)
	toBal := l.Balances[to]
	if toBal.IsNil() {
		toBal = math.NewInt(0)
	}
	l.Balances[to] = toBal.Add(amount)
	if !l.totalLocked().Equal(before) {
		return fmt.Errorf("monetary: internal error: transfer changed total supply")
	}
	return l.CheckInvariants()
}

// Burn destroys amount from an account. It reduces total supply by
// exactly that amount, increments the cumulative burned counter, and
// creates no remintable credit.
func (l *Ledger) Burn(addr string, amount math.Int) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if amount.IsNil() || !amount.IsPositive() {
		return fmt.Errorf("%w: burn amount %s", ErrZeroAmount, amount.String())
	}
	bal := l.Balances[addr]
	if bal.IsNil() {
		bal = math.NewInt(0)
	}
	if bal.LT(amount) {
		return fmt.Errorf("%w: %s holds %s, tried to burn %s", ErrInsufficient, addr, bal.String(), amount.String())
	}
	l.Balances[addr] = bal.Sub(amount)
	l.Burned = l.Burned.Add(amount)
	return l.CheckInvariants()
}

// DeductFee moves a fee from payer to the fee collector. Fees are paid in
// OGO from existing supply; they never mint. Total supply is unchanged.
func (l *Ledger) DeductFee(payer string, fee math.Int) error {
	if fee.IsNil() || fee.IsZero() {
		return nil
	}
	return l.Send(payer, FeeCollector, fee)
}

// Mint is discretionary minting and always fails. No founder, admin,
// module account, governance proposal, or contract may create coins this
// way. This method exists only so the failure is explicit and testable,
// rather than a missing method callers might re-invent.
func (l *Ledger) Mint(_ string, _ math.Int) error {
	return ErrMintDisabled
}

// ScheduleAmountForHeight returns the exact issuance amount the schedule
// dictates at a height, and whether the schedule is active there.
func (l *Ledger) ScheduleAmountForHeight(height int64) (math.Int, bool) {
	period, ok := l.Schedule.PeriodForHeight(height)
	if !ok {
		return math.NewInt(0), false
	}
	return l.Schedule.AmountForPeriod(period), true
}

// MintScheduled issues exactly the schedule amount for the period
// containing height, once, into the rewards pool. It is the ONLY minting
// the protocol permits. It fails when:
//   - the height is before the schedule start (not started),
//   - the height is past the last period (exhausted; issuance stops),
//   - the period was already claimed (double issuance),
//   - the amount would push total supply over max_supply.
func (l *Ledger) MintScheduled(height int64) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if height < l.Schedule.StartHeight {
		return ErrScheduleNotStarted
	}
	period, ok := l.Schedule.PeriodForHeight(height)
	if !ok {
		return ErrScheduleExhausted
	}
	if _, claimed := l.IssuedByPeriod[int64(period)]; claimed {
		return fmt.Errorf("%w: period %d already issued", ErrDoubleIssuance, period)
	}
	amount := l.Schedule.AmountForPeriod(period)
	// A period whose floored amount is zero issues nothing; record it so
	// the period cannot be re-attempted, and return success.
	l.IssuedByPeriod[int64(period)] = amount
	if amount.IsZero() {
		return nil
	}
	if l.totalLocked().Add(amount).GT(l.MaxSupply) {
		delete(l.IssuedByPeriod, int64(period))
		return fmt.Errorf("%w: scheduled issuance %s at height %d would exceed max %s", ErrSupplyOverMax, amount.String(), height, l.MaxSupply.String())
	}
	pool := l.Balances[RewardsPool]
	if pool.IsNil() {
		pool = math.NewInt(0)
	}
	l.Balances[RewardsPool] = pool.Add(amount)
	l.TotalIssued = l.TotalIssued.Add(amount)
	return l.CheckInvariants()
}

// MintScheduledEpoch is the amount-less, epoch-numbered mint
// entrypoint from the consensus spec (section 5): MintScheduled takes
// no amount argument, it computes R_n for epoch n from the schedule.
// It is the epoch-based counterpart of the height-based
// MintScheduled. It fails if the epoch is out of range or already
// claimed, and it mints exactly R_n into the rewards pool, once.
// This is the entrypoint the epoch hook calls. In this prototype the
// caller is the epoch-timer mirror in epochs.go, whose AfterEpochEnd
// hook is wired into the app and proven to fire by
// TestEpochHookFiresInWiredApp. Real SDK x/epochs wiring (keeper,
// store, depinject) remains Milestone C work; see EPOCHS_GAP.md.
func (l *Ledger) MintScheduledEpoch(epoch int64) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.mintScheduledEpochLocked(epoch)
}

func (l *Ledger) mintScheduledEpochLocked(epoch int64) error {
	if epoch < 0 || epoch >= int64(l.Schedule.Periods) {
		return ErrScheduleExhausted
	}
	if _, claimed := l.IssuedByPeriod[epoch]; claimed {
		return fmt.Errorf("%w: epoch %d already issued", ErrDoubleIssuance, epoch)
	}
	amount := l.Schedule.AmountForPeriod(uint64(epoch))
	l.IssuedByPeriod[epoch] = amount
	if amount.IsZero() {
		return nil
	}
	if l.totalLocked().Add(amount).GT(l.MaxSupply) {
		delete(l.IssuedByPeriod, epoch)
		return fmt.Errorf("%w: scheduled issuance %s for epoch %d would exceed max %s", ErrSupplyOverMax, amount.String(), epoch, l.MaxSupply.String())
	}
	pool := l.Balances[RewardsPool]
	if pool.IsNil() {
		pool = math.NewInt(0)
	}
	l.Balances[RewardsPool] = pool.Add(amount)
	l.TotalIssued = l.TotalIssued.Add(amount)
	return l.CheckInvariants()
}

// MintScheduledModel mints exactly R_n for epoch n from an
// IssuanceModel (geometric, hybrid, or tapered-hybrid), once, into
// the rewards pool. It is the model-based mint path: the legacy
// height-based MintScheduled can only express the geometric schedule,
// while this path runs any model selected through genesis
// configuration, with the same duplicate guard and cap check.
func (l *Ledger) MintScheduledModel(model IssuanceModel, epoch int64) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if model == nil {
		return ErrScheduleMissing
	}
	if epoch < 0 || epoch >= model.Periods() {
		return ErrScheduleExhausted
	}
	if _, claimed := l.IssuedByPeriod[epoch]; claimed {
		return fmt.Errorf("%w: epoch %d already issued", ErrDoubleIssuance, epoch)
	}
	amount, err := model.AmountForEpoch(epoch)
	if err != nil {
		return err
	}
	l.IssuedByPeriod[epoch] = amount
	if amount.IsZero() {
		return nil
	}
	if l.totalLocked().Add(amount).GT(l.MaxSupply) {
		delete(l.IssuedByPeriod, epoch)
		return fmt.Errorf("%w: scheduled issuance %s for epoch %d would exceed max %s", ErrSupplyOverMax, amount.String(), epoch, l.MaxSupply.String())
	}
	pool := l.Balances[RewardsPool]
	if pool.IsNil() {
		pool = math.NewInt(0)
	}
	l.Balances[RewardsPool] = pool.Add(amount)
	l.TotalIssued = l.TotalIssued.Add(amount)
	return l.CheckInvariants()
}

// NextUnclaimedEpoch returns the oldest epoch that has not been
// claimed yet, or -1 if every epoch in the schedule is claimed.
func (l *Ledger) NextUnclaimedEpoch() int64 {
	l.mu.Lock()
	defer l.mu.Unlock()
	for epoch := int64(0); epoch < int64(l.Schedule.Periods); epoch++ {
		if _, claimed := l.IssuedByPeriod[epoch]; !claimed {
			return epoch
		}
	}
	return -1
}

// ProcessNextEpoch is the halt-recovery catch-up step (consensus spec
// section 7): it claims at most ONE epoch per block, the oldest
// unclaimed epoch, and only if that epoch has ended (epoch <=
// currentEpoch). It returns the amount minted for that epoch, or 0
// with no error if there is nothing to catch up. It never mints a
// lump sum and never skips an epoch: the caller must call it once
// per block until it reports nothing left, in epoch order.
// currentEpoch is the epoch number whose end time has passed at this
// block. blockHeight is recorded only for the error message; the
// amount is a function of the epoch number alone.
func (l *Ledger) ProcessNextEpoch(currentEpoch int64, blockHeight int64) (math.Int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	next := int64(-1)
	for epoch := int64(0); epoch < int64(l.Schedule.Periods); epoch++ {
		if _, claimed := l.IssuedByPeriod[epoch]; !claimed {
			next = epoch
			break
		}
	}
	if next == -1 {
		return math.NewInt(0), nil
	}
	if next > currentEpoch {
		return math.NewInt(0), nil
	}
	amount := l.Schedule.AmountForPeriod(uint64(next))
	before := l.TotalIssued
	if err := l.mintScheduledEpochLocked(next); err != nil {
		return math.NewInt(0), fmt.Errorf("monetary: catch-up at height %d for epoch %d failed: %w", blockHeight, next, err)
	}
	minted := l.TotalIssued.Sub(before)
	if !minted.Equal(amount) {
		return math.NewInt(0), fmt.Errorf("monetary: internal error: catch-up minted %s, expected %s", minted.String(), amount.String())
	}
	return amount, nil
}

// CheckMonetaryInvariants checks the ledger-level supply invariants
// from monetary engine spec section 10 that are testable here:
//  1. total supply never exceeds max supply M.
//  2. total issued never exceeds total scheduled (legacy schedule).
//  3. each epoch is claimed at most once and the issued-by-epoch
//     amounts sum to TotalIssued, with each amount exactly R_n.
//  4. total supply equals genesis supply plus total issued minus
//     total burned, where genesis supply is recovered as
//     total - TotalIssued + Burned (balances + escrow - issued +
//     burned); this is checked as the identity
//     TotalIssued == sum(IssuedByPeriod) and TotalSupply <= Max,
//     because the ledger does not store genesis supply separately.
//  5. dust is never issued: TotalIssued <= TotalScheduled, and no
//     amount beyond the schedule total can ever be minted.
//
// Invariants 7 (no discretionary mint through other modules), 8
// (governance cannot change the schedule), 9 (IBC), and 11
// (every-block invariant in the wired app) are covered by the bypass
// and IBC suites and by the app's FinalizeBlock, not here.
func (l *Ledger) CheckMonetaryInvariants() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if err := l.CheckInvariants(); err != nil {
		return err
	}
	scheduled := l.Schedule.TotalIssuance()
	if l.TotalIssued.GT(scheduled) {
		return fmt.Errorf("%w: total issued %s exceeds total scheduled %s", ErrSupplyOverMax, l.TotalIssued.String(), scheduled.String())
	}
	sum := math.NewInt(0)
	for epoch, amt := range l.IssuedByPeriod {
		if epoch < 0 || epoch >= int64(l.Schedule.Periods) {
			return fmt.Errorf("monetary: issued-by-epoch contains out-of-range epoch %d", epoch)
		}
		want := l.Schedule.AmountForPeriod(uint64(epoch))
		if !amt.Equal(want) {
			return fmt.Errorf("monetary: epoch %d issued %s, schedule says %s", epoch, amt.String(), want.String())
		}
		sum = sum.Add(amt)
	}
	if !sum.Equal(l.TotalIssued) {
		return fmt.Errorf("monetary: issued-by-epoch sum %s does not equal total issued %s", sum.String(), l.TotalIssued.String())
	}
	return nil
}

// Marshal serialises the ledger for disk persistence (Commit).
func (l *Ledger) Marshal() ([]byte, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return json.Marshal(l.cloneLocked())
}

// UnmarshalLedger restores a ledger from disk. It re-validates the cap.
func UnmarshalLedger(b []byte) (*Ledger, error) {
	l := &Ledger{}
	if err := json.Unmarshal(b, l); err != nil {
		return nil, err
	}
	if l.Balances == nil {
		l.Balances = make(map[string]math.Int)
	}
	if l.Escrow == nil {
		l.Escrow = make(map[string]math.Int)
	}
	if l.Vouchers == nil {
		l.Vouchers = make(map[string]math.Int)
	}
	if l.IssuedByPeriod == nil {
		l.IssuedByPeriod = make(map[int64]math.Int)
	}
	if l.MaxSupply.IsNil() || !l.MaxSupply.IsPositive() {
		return nil, ErrMaxSupplyInvalid
	}
	if err := l.CheckInvariants(); err != nil {
		return nil, err
	}
	return l, nil
}
