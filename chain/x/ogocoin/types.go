// Package ogocoin implements OGO Network's fixed-supply monetary module.
//
// Monetary Constitution v0.1 (2026-11-05, draft for review) SUPERSEDES the
// earlier "minting disabled after genesis" rule. The current rule is:
// fixed maximum supply + SCHEDULED issuance, zero reserved founder
// allocation.
//
// The protocol may mint ONLY exactly as a predetermined, deterministic
// schedule dictates (see IssuanceSchedule). There is no discretionary,
// admin, founder, or governance minting. Schedule modification is not
// available through ordinary governance. Issuance stops permanently at
// max supply. This package constrains x/bank; it does not replace it.
//
// Naming (locked 2026-11-05): Network OGO Network, native coin OGO,
// ticker OGO, Wallet OGO Wallet, Explorer OGO Scan.
//
// OGO Credits are a separate, app-internal accounting unit. They are never
// counted in, converted to, or mixed with OGO Coin supply here.
package ogocoin

import (
	"errors"

	"cosmossdk.io/math"
)

// ModuleName is the x/ogocoin module name.
const ModuleName = "ogocoin"

// Denom is the proposed base denomination (proposed, not final).
const Denom = "uogo"

// Decimals is the proposed decimal precision (proposed, not final).
// 1 OGO = 10^6 uogo.
const Decimals uint32 = 6

// ChainID is the localnet-only chain ID. A mainnet chain ID must not be
// claimed until the genesis spec is approved.
const ChainID = "ogo-local-1"

// FeeCollector is the well-known account that receives transaction fees.
// Fees are paid in OGO from existing supply; they never mint.
const FeeCollector = "ogo1feecollector"

// RewardsPool is the well-known account that receives scheduled issuance.
// Distribution from the pool to validators/delegators is a separate,
// TBD mechanism; the pool itself is the only mint recipient.
const RewardsPool = "ogo1rewardspool"

// Errors returned by genesis validation and the ledger.
var (
	// ErrMaxSupplyMissing: max_supply is a mandatory genesis parameter.
	// There is no default value anywhere in this module.
	ErrMaxSupplyMissing = errors.New("ogocoin: max_supply is mandatory in genesis and was not provided (no default exists)")
	ErrMaxSupplyInvalid = errors.New("ogocoin: max_supply must be greater than zero")
	ErrDenomEmpty       = errors.New("ogocoin: denom must not be empty")
	ErrDenomMismatch    = errors.New("ogocoin: denom does not match the declared network denom (uogo, proposed)")
	ErrDecimalsMismatch = errors.New("ogocoin: decimals must be 6 (proposed)")
	ErrMintingEnabled   = errors.New("ogocoin: minting must be disabled after genesis")
	ErrCapNotEnforced   = errors.New("ogocoin: supply_cap_enforced must be true")
	ErrSupplyOverMax    = errors.New("ogocoin: total supply would exceed max_supply")
	ErrNegativeAmount   = errors.New("ogocoin: amount must not be negative")
	ErrZeroAmount       = errors.New("ogocoin: amount must be greater than zero")
	ErrInsufficient     = errors.New("ogocoin: insufficient balance")
	ErrUnknownAccount   = errors.New("ogocoin: unknown account")
	// ErrMintDisabled: discretionary minting is prohibited. The only
	// minting the protocol permits is scheduled issuance (see
	// MintScheduled), for exactly the schedule amount at the scheduled
	// height. This error is the discretionary-mint failure.
	ErrMintDisabled       = errors.New("ogocoin: discretionary minting is prohibited; only scheduled issuance may create coins")
	ErrUpgradeRejected    = errors.New("ogocoin: upgrade rejected: monetary parameters are immutable")
	ErrVoucherOverEscrow  = errors.New("ogocoin: voucher supply exceeds escrow backing; vouchers cannot exceed escrow")
	ErrUnknownMethod      = errors.New("ogocoin: unknown method")
	ErrScheduleMissing    = errors.New("ogocoin: issuance schedule is mandatory in genesis and was not provided")
	ErrScheduleInvalid    = errors.New("ogocoin: issuance schedule is invalid")
	ErrScheduleExhausted  = errors.New("ogocoin: issuance schedule has ended; issuance stops permanently at max supply")
	ErrScheduleNotStarted = errors.New("ogocoin: issuance schedule has not started at this height")
	ErrDoubleIssuance     = errors.New("ogocoin: issuance for this period was already claimed")
	ErrFounderAllocation  = errors.New("ogocoin: founder/reserved allocation is prohibited (zero reserved founder allocation)")
	ErrPurposeMissing     = errors.New("ogocoin: genesis allocation purpose is mandatory (public disclosure)")
)

// IssuanceSchedule is the predetermined, deterministic issuance schedule
// (Monetary Constitution v0.1, Article III). All VALUES are TBD; the
// STRUCTURE is the framework. No schedule numbers are invented here:
// a schedule with missing or invalid values fails genesis validation.
//
// Model: issuance runs in periods. Period p starts at
// StartHeight + p*PeriodLength and lasts PeriodLength blocks. The amount
// issued in period p is InitialPerPeriod * (ReductionNumerator /
// ReductionDenominator)^p, computed with integer floor division at each
// step, so issuance strictly decreases while the ratio is below 1.
// Total scheduled issuance is the sum over all periods. The schedule is
// validated at genesis so that genesis supply + total scheduled issuance
// never exceeds max_supply; issuance therefore stops permanently at max.
type IssuanceSchedule struct {
	// ScheduleType is the only supported type in this framework.
	ScheduleType string `json:"schedule_type"` // must be "reducing-per-period"
	// StartHeight is the first block height at which issuance may occur.
	StartHeight int64 `json:"start_height"`
	// PeriodLength is the number of blocks in one issuance period.
	PeriodLength int64 `json:"period_length_blocks"`
	// Periods is the number of issuance periods. After the last period,
	// issuance stops permanently.
	Periods uint64 `json:"periods"`
	// InitialPerPeriod is the issuance amount (uogo) in period 0.
	InitialPerPeriod math.Int `json:"initial_per_period"`
	// ReductionNumerator / ReductionDenominator is the per-period
	// reduction factor. Must be strictly below 1 (numerator < denominator)
	// so issuance gradually decreases, per the constitution.
	ReductionNumerator   int64 `json:"reduction_numerator"`
	ReductionDenominator int64 `json:"reduction_denominator"`
	// BasePerPeriod is the hybrid-only fixed absolute base rate per
	// period (uogo). Zero for a geometric schedule. For a hybrid
	// schedule, InitialPerPeriod is R0 = base + B0, so the geometric
	// bonus initial is InitialPerPeriod minus BasePerPeriod.
	// TEST-ONLY values only; production values are TBD.
	BasePerPeriod math.Int `json:"base_per_period,omitempty"`
	// TaperEnabled, TaperNumerator, TaperDenominator, TaperPeriods
	// are the tapered-hybrid-only parameters (spec section 3.2):
	// a gentle geometric taper over the final TaperPeriods periods.
	TaperEnabled     bool   `json:"taper_enabled,omitempty"`
	TaperNumerator   int64  `json:"taper_numerator,omitempty"`
	TaperDenominator int64  `json:"taper_denominator,omitempty"`
	TaperPeriods     uint64 `json:"taper_periods,omitempty"`
}

// Params are the monetary parameters recorded at genesis. MaxSupply and
// IssuanceSchedule are pointers so "missing" (nil) is distinguishable
// from zero: both fail validation, and neither is ever defaulted.
type Params struct {
	// MaxSupply is the fixed maximum supply in uogo. Mandatory, TBD in
	// production. TEST-ONLY values may appear in tests only, clearly
	// labeled, and must never be presented as the chosen supply.
	MaxSupply *math.Int `json:"max_supply"`
	// Denom is the base denomination. Proposed: "uogo".
	Denom string `json:"denom"`
	// Decimals is the display precision. Proposed: 6.
	Decimals uint32 `json:"decimals"`
	// InflationRate is DISCRETIONARY inflation, fixed at "0". Scheduled
	// issuance is separate and is governed by IssuanceSchedule.
	InflationRate string `json:"inflation_rate"`
	// MintingEnabled refers to DISCRETIONARY minting and must be false.
	// Scheduled issuance is the only minting the protocol permits.
	MintingEnabled bool `json:"minting_enabled"`
	// SupplyCapEnforced must be true.
	SupplyCapEnforced bool `json:"supply_cap_enforced"`
	// IssuanceSchedule is the mandatory predetermined issuance schedule.
	// All schedule values are TBD in production; the schedule must be
	// present and valid or genesis fails.
	IssuanceSchedule *IssuanceSchedule `json:"issuance_schedule"`
	// IssuanceModel selects the issuance model through genesis
	// configuration (spec section 3.1): "geometric", "hybrid", or
	// "tapered-hybrid". Empty means the legacy Milestone B genesis,
	// which selects geometric. See SelectedModel in models.go.
	// Planned rename: x/ogocoin becomes x/monetary; this field and
	// the model enum stay the same across that rename.
	IssuanceModel string `json:"issuance_model,omitempty"`
}

// Balance is one genesis allocation. Denom is carried per balance so a
// wrong-denom allocation is detectable by validation. Purpose is a
// mandatory public-disclosure label; founder/reserved purposes are
// rejected (zero reserved founder allocation, Constitution Article IV).
type Balance struct {
	Address string   `json:"address"`
	Denom   string   `json:"denom"`
	Amount  math.Int `json:"amount"`
	Purpose string   `json:"purpose"`
}

// ForbiddenPurposes are genesis allocation purposes that are rejected:
// any reserved founder / insider / presale allocation.
var ForbiddenPurposes = map[string]bool{
	"founder": true, "founder_reserve": true, "founders": true,
	"ogo_systems": true, "ogo_systems_reserve": true,
	"team": true, "team_reserve": true,
	"insider": true, "insider_reserve": true,
	"presale": true, "private_sale": true, "private-presale": true,
}

// GenesisState is the x/ogocoin genesis state.
type GenesisState struct {
	Params   Params    `json:"params"`
	Balances []Balance `json:"balances"`
}

// DefaultParams returns the fixed (non-supply) parameters. MaxSupply is
// deliberately nil: there is no default maximum supply.
func DefaultParams() Params {
	return Params{
		MaxSupply:         nil,
		Denom:             Denom,
		Decimals:          Decimals,
		InflationRate:     "0",
		MintingEnabled:    false,
		SupplyCapEnforced: true,
	}
}

// ZeroInt is a convenience for a zero amount.
func ZeroInt() math.Int { return math.NewInt(0) }
