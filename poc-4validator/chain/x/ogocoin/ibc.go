package ogocoin

import (
	"fmt"

	"cosmossdk.io/math"
)

// IBC escrow / voucher accounting.
//
// The rule (locked 2026-11-05): native OGO escrowed on one chain and
// represented as a voucher on another chain must NOT count as two
// independent units of economic supply.
//
// Model, for a transfer of X uogo from OGO Network (chain A) to a
// counterparty chain (chain B) over a channel:
//   - A moves X from the sender's balance into escrow. A's chain-local
//     supply is UNCHANGED: escrowed coins still exist on A, held by the
//     escrow account, so they remain part of A's bank supply.
//   - B mints a voucher of X in the IBC voucher denom (ibc/<hash>).
//     The voucher is a CLAIM on A's escrow, not new OGO.
//   - Global economic supply = (A local supply - escrowed) + vouchers.
//     When vouchers == escrowed, global == A local supply. A naive sum
//     (A local + vouchers) would double-count and is WRONG.
//
// Invariants:
//   - vouchers[channel] <= escrow[channel], always.
//   - inbound native OGO that would push A local supply over max_supply
//     is rejected.
//   - OGO Credits never appear in any of these figures.

// EscrowForIBC locks amount from a user's balance into escrow for a
// channel. Chain-local total supply is unchanged (balance -> escrow).
func (l *Ledger) EscrowForIBC(from, channelID string, amount math.Int) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if amount.IsNil() || !amount.IsPositive() {
		return fmt.Errorf("%w: escrow amount %s", ErrZeroAmount, amount.String())
	}
	bal := l.Balances[from]
	if bal.IsNil() {
		bal = math.NewInt(0)
	}
	if bal.LT(amount) {
		return fmt.Errorf("%w: %s holds %s, tried to escrow %s", ErrInsufficient, from, bal.String(), amount.String())
	}
	l.Balances[from] = bal.Sub(amount)
	esc := l.Escrow[channelID]
	if esc.IsNil() {
		esc = math.NewInt(0)
	}
	l.Escrow[channelID] = esc.Add(amount)
	return l.CheckInvariants()
}

// RegisterVoucher records vouchers minted on the counterparty chain for an
// escrow. It fails if vouchers would exceed the escrow backing them.
func (l *Ledger) RegisterVoucher(channelID string, amount math.Int) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if amount.IsNil() || amount.IsNegative() {
		return fmt.Errorf("%w: voucher amount %s", ErrNegativeAmount, amount.String())
	}
	esc := l.Escrow[channelID]
	if esc.IsNil() {
		esc = math.NewInt(0)
	}
	v := l.Vouchers[channelID]
	if v.IsNil() {
		v = math.NewInt(0)
	}
	if v.Add(amount).GT(esc) {
		return fmt.Errorf("%w: channel %s escrow %s, vouchers would be %s", ErrVoucherOverEscrow, channelID, esc.String(), v.Add(amount).String())
	}
	l.Vouchers[channelID] = v.Add(amount)
	return nil
}

// ReleaseEscrow burns vouchers on the counterparty (recorded here as a
// voucher decrease) and returns escrowed native OGO to a user on this
// chain. Chain-local total supply is unchanged (escrow -> balance).
func (l *Ledger) ReleaseEscrow(to, channelID string, amount math.Int) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if amount.IsNil() || !amount.IsPositive() {
		return fmt.Errorf("%w: release amount %s", ErrZeroAmount, amount.String())
	}
	v := l.Vouchers[channelID]
	if v.IsNil() {
		v = math.NewInt(0)
	}
	if v.LT(amount) {
		return fmt.Errorf("%w: channel %s has %s vouchers outstanding, tried to release %s", ErrVoucherOverEscrow, channelID, v.String(), amount.String())
	}
	esc := l.Escrow[channelID]
	if esc.IsNil() {
		esc = math.NewInt(0)
	}
	if esc.LT(amount) {
		return fmt.Errorf("ogocoin: internal error: escrow %s below release %s on channel %s", esc.String(), amount.String(), channelID)
	}
	l.Vouchers[channelID] = v.Sub(amount)
	l.Escrow[channelID] = esc.Sub(amount)
	toBal := l.Balances[to]
	if toBal.IsNil() {
		toBal = math.NewInt(0)
	}
	l.Balances[to] = toBal.Add(amount)
	return l.CheckInvariants()
}

// Escrowed returns the amount escrowed on a channel.
func (l *Ledger) Escrowed(channelID string) math.Int {
	l.mu.Lock()
	defer l.mu.Unlock()
	if v, ok := l.Escrow[channelID]; ok {
		return v
	}
	return math.NewInt(0)
}

// VouchersOutstanding returns vouchers outstanding for a channel.
func (l *Ledger) VouchersOutstanding(channelID string) math.Int {
	l.mu.Lock()
	defer l.mu.Unlock()
	if v, ok := l.Vouchers[channelID]; ok {
		return v
	}
	return math.NewInt(0)
}

// GlobalEconomicSupply computes the economic supply across both sides of
// an IBC channel WITHOUT double-counting: circulating native on this chain
// (local supply minus escrow) plus outstanding vouchers. When vouchers
// equal escrow, this equals the chain-local supply exactly.
func GlobalEconomicSupply(chainLocalSupply, escrowed, vouchersOutstanding math.Int) math.Int {
	return chainLocalSupply.Sub(escrowed).Add(vouchersOutstanding)
}

// NaiveDoubleCount is the WRONG sum (local + vouchers) that would count
// escrow-backed vouchers as new supply. It exists only so tests can prove
// the correct function differs from it whenever vouchers are outstanding.
func NaiveDoubleCount(chainLocalSupply, vouchersOutstanding math.Int) math.Int {
	return chainLocalSupply.Add(vouchersOutstanding)
}

// CheckGlobalSupply asserts the IBC accounting rule for one channel:
// vouchers <= escrow, and global economic supply == chain-local supply
// when vouchers == escrow. It returns an error on any violation.
func (l *Ledger) CheckGlobalSupply(channelID string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	esc := l.Escrow[channelID]
	if esc.IsNil() {
		esc = math.NewInt(0)
	}
	v := l.Vouchers[channelID]
	if v.IsNil() {
		v = math.NewInt(0)
	}
	if v.GT(esc) {
		return fmt.Errorf("%w: channel %s escrow %s, vouchers %s", ErrVoucherOverEscrow, channelID, esc.String(), v.String())
	}
	global := GlobalEconomicSupply(l.totalLocked(), esc, v)
	// When vouchers == escrow, global must equal local supply exactly.
	// When vouchers < escrow (e.g. a voucher was burned without release
	// yet), global is BELOW local supply, which is also safe (never above).
	if global.GT(l.totalLocked()) {
		return fmt.Errorf("ogocoin: global economic supply %s exceeds chain-local supply %s", global.String(), l.totalLocked().String())
	}
	return nil
}

// InboundTransfer credits native OGO received from another chain, but
// only if the chain-local total would stay within max_supply. This is the
// cap check on IBC inflows.
func (l *Ledger) InboundTransfer(to string, amount math.Int) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if amount.IsNil() || !amount.IsPositive() {
		return fmt.Errorf("%w: inbound amount %s", ErrZeroAmount, amount.String())
	}
	if l.totalLocked().Add(amount).GT(l.MaxSupply) {
		return fmt.Errorf("%w: inbound %s would push total %s over max %s", ErrSupplyOverMax, amount.String(), l.totalLocked().String(), l.MaxSupply.String())
	}
	toBal := l.Balances[to]
	if toBal.IsNil() {
		toBal = math.NewInt(0)
	}
	l.Balances[to] = toBal.Add(amount)
	return l.CheckInvariants()
}
