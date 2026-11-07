package monetary

import (
	"fmt"
	"strings"

	"cosmossdk.io/math"
)

// ValidateGenesis fails closed. It is called by the app's InitChain and by
// the ogod CLI before any node starts. The checks run in this order:
//
//  1. max_supply present and > 0 (missing or zero/negative fails; no default)
//  2. denom non-empty and equal to the declared denom
//  3. decimals equal to the proposed value
//  4. discretionary minting disabled (MintingEnabled must be false)
//  5. supply cap enforced
//  6. issuance schedule present and valid (mandatory, deterministic)
//  7. every allocation: non-empty address, correct denom, non-negative
//     amount, non-empty public purpose, purpose not in the forbidden
//     founder/reserved set (zero reserved founder allocation)
//  8. sum(allocations) + total scheduled issuance <= max_supply
//
// Initial supply may be less than max_supply. The gap is covered only by
// the predetermined schedule: genesis supply + total scheduled issuance
// must fit under max, and once the schedule ends issuance stops
// permanently. There is no discretionary path to the headroom.
func ValidateGenesis(gs GenesisState) error {
	p := gs.Params

	if p.MaxSupply == nil || p.MaxSupply.IsNil() {
		return ErrMaxSupplyMissing
	}
	if !p.MaxSupply.IsPositive() {
		return fmt.Errorf("%w: got %s", ErrMaxSupplyInvalid, p.MaxSupply.String())
	}
	if p.Denom == "" {
		return ErrDenomEmpty
	}
	if p.Denom != Denom {
		return fmt.Errorf("%w: got %q, want %q", ErrDenomMismatch, p.Denom, Denom)
	}
	if p.Decimals != Decimals {
		return fmt.Errorf("%w: got %d, want %d", ErrDecimalsMismatch, p.Decimals, Decimals)
	}
	if p.MintingEnabled {
		return ErrMintingEnabled
	}
	if !p.SupplyCapEnforced {
		return ErrCapNotEnforced
	}
	if p.IssuanceSchedule == nil {
		return ErrScheduleMissing
	}
	if err := p.IssuanceSchedule.Validate(); err != nil {
		return err
	}
	// Milestone 3: the issuance model is selected through genesis
	// configuration. An empty value is the legacy Milestone B genesis
	// and selects geometric; a present unknown value is rejected.
	// Hybrid / tapered-hybrid genesis must carry their extra
	// parameters in the schedule (base_per_period, taper fields);
	// SelectedModel validates them.
	if _, err := p.SelectedModel(); err != nil {
		return err
	}

	seen := make(map[string]bool, len(gs.Balances))
	total := math.NewInt(0)
	for i, b := range gs.Balances {
		if b.Address == "" {
			return fmt.Errorf("monetary: genesis balance %d has an empty address", i)
		}
		if seen[b.Address] {
			return fmt.Errorf("monetary: duplicate genesis allocation for %s", b.Address)
		}
		seen[b.Address] = true
		if b.Denom != p.Denom {
			return fmt.Errorf("%w: balance %d for %s uses denom %q", ErrDenomMismatch, i, b.Address, b.Denom)
		}
		if b.Amount.IsNil() {
			return fmt.Errorf("monetary: genesis balance %d for %s has a missing amount", i, b.Address)
		}
		if b.Amount.IsNegative() {
			return fmt.Errorf("%w: balance %d for %s is %s", ErrNegativeAmount, i, b.Address, b.Amount.String())
		}
		if strings.TrimSpace(b.Purpose) == "" {
			return fmt.Errorf("%w: balance %d for %s", ErrPurposeMissing, i, b.Address)
		}
		if ForbiddenPurposes[strings.ToLower(strings.TrimSpace(b.Purpose))] {
			return fmt.Errorf("%w: balance %d for %s has purpose %q", ErrFounderAllocation, i, b.Address, b.Purpose)
		}
		total = total.Add(b.Amount)
	}
	if total.GT(*p.MaxSupply) {
		return fmt.Errorf("%w: genesis allocations sum to %s, max_supply is %s", ErrSupplyOverMax, total.String(), p.MaxSupply.String())
	}
	// The schedule must fit under the cap together with genesis supply.
	// The scheduled total is always the selected model's exact
	// S_actual (single exact rounding rule, decision D-B4); the legacy
	// schedule TotalIssuance is now the same exact total.
	scheduled, err := p.SelectedScheduledTotal()
	if err != nil {
		return err
	}
	if total.Add(scheduled).GT(*p.MaxSupply) {
		return fmt.Errorf("%w: genesis supply %s + total scheduled issuance %s exceeds max_supply %s",
			ErrSupplyOverMax, total.String(), scheduled.String(), p.MaxSupply.String())
	}
	return nil
}

// ComputeMaxSupply returns M = G + S_actual for a genesis state,
// where G is the genesis circulating supply (sum of allocations) and
// S_actual is the exact integer sum of the floored per-period amounts
// of the model selected by genesis configuration. It is the spec
// section 2 rule 5 / section 3.1 definition: M is derived at genesis,
// not chosen first. S_actual is the exact-formula total (single
// rounding rule, decision D-B4); for the TEST-ONLY schedule it is
// 6511.
func ComputeMaxSupply(gs GenesisState) (math.Int, error) {
	if gs.Params.IssuanceSchedule == nil {
		return math.NewInt(0), ErrScheduleMissing
	}
	scheduled, err := gs.Params.SelectedScheduledTotal()
	if err != nil {
		return math.NewInt(0), err
	}
	return TotalGenesisSupply(gs).Add(scheduled), nil
}

// ValidateGenesisExact is the strict supply-invariant genesis check
// (spec section 3.1 / section 10): it runs ValidateGenesis and then
// additionally requires max_supply to equal exactly G + S_actual,
// computed by exact integer summation. A max supply that is merely
// below the cap, or above G + S_actual, is rejected here.
//
// Status (Milestone C): this strict check IS now the default
// genesis path. NewLedger and the wired x/monetary keeper InitGenesis
// both call ValidateGenesisExact, so a genesis file whose max_supply
// carries headroom above G + S_actual is rejected at startup.
// ValidateGenesis (lenient, G + S_actual <= max) remains only as a
// building block inside ValidateGenesisExact and for tests that
// exercise individual validation failures; it is not a genesis path
// any chain may start from. Dust (M_target - S_actual, where a
// nominal target exists) is left permanently unissued.
//
func ValidateGenesisExact(gs GenesisState) error {
	if err := ValidateGenesis(gs); err != nil {
		return err
	}
	if gs.Params.MaxSupply == nil || gs.Params.MaxSupply.IsNil() {
		return ErrMaxSupplyMissing
	}
	want, err := ComputeMaxSupply(gs)
	if err != nil {
		return err
	}
	if !gs.Params.MaxSupply.Equal(want) {
		return fmt.Errorf("%w: max_supply %s does not equal G + S_actual %s (genesis supply %s + total scheduled %s)",
			ErrSupplyOverMax, gs.Params.MaxSupply.String(), want.String(),
			TotalGenesisSupply(gs).String(), want.Sub(TotalGenesisSupply(gs)).String())
	}
	return nil
}

// TotalGenesisSupply returns sum(allocations). It assumes ValidateGenesis
// has already passed.
func TotalGenesisSupply(gs GenesisState) math.Int {
	total := math.NewInt(0)
	for _, b := range gs.Balances {
		total = total.Add(b.Amount)
	}
	return total
}
