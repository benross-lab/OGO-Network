package ogocoin

import (
	"fmt"

	"cosmossdk.io/math"
)

// Bypass vectors.
//
// Guiding principle: OGO's monetary rules must be enforced by the
// protocol, not by trusting OGO Systems. Each function below models one
// way an attacker (or a bug) might try to get around the fixed supply,
// and each one must FAIL. The tests in bypass_test.go attempt all of them
// and assert supply is unchanged afterwards.
//
// Vector 1: Privileged module accounts.
// Vector 2: Governance upgrades.
// Vector 3: Smart-contract (CosmWasm) interfaces.

// ---------------------------------------------------------------------------
// Vector 1: privileged module accounts
// ---------------------------------------------------------------------------

// ModuleAccount names a privileged account that exists in a standard
// Cosmos SDK chain. In a full SDK app these hold funds for modules
// (fee collector, staking, gov, distribution, wasm). None of them may
// mint. This scaffold models the attempt surface; the full SDK wiring
// (Milestone C) must preserve the same property through x/bank keepers.
var KnownModuleAccounts = []string{
	"fee_collector", // x/auth fee collector
	"bonded_pool",   // x/staking bonded pool
	"not_bonded_pool",
	"gov",          // x/gov
	"distribution", // x/distribution
	"wasm",         // x/wasm (wasmd)
	"mint",         // NOTE: OGO has NO x/mint module. This name exists
	// only so the test can prove that even asking for it fails.
	"admin", // no such account exists; any attempt using it must fail
}

// MintFromModuleAccount simulates x/bank MintCoins called from a module
// account. It always fails: OGO has no mint module and no minter key.
func (l *Ledger) MintFromModuleAccount(module string, amount math.Int) error {
	return fmt.Errorf("%w: module account %q attempted to mint %s", ErrMintDisabled, module, amount.String())
}

// BurnFromModuleAccountToMint simulates a module account burning coins
// and then claiming the burned amount as remintable credit. Burning is
// one-way; this always fails and changes nothing.
func (l *Ledger) BurnFromModuleAccountToMint(_ string, _ math.Int) error {
	return fmt.Errorf("%w: burn does not create remintable credit", ErrMintDisabled)
}

// SendFromModuleAccount simulates x/bank SendCoins from a module account.
// It may only move funds the module account actually holds, and it can
// never increase total supply. A send of more than the module balance
// fails; a send within balance succeeds via the normal Send path.
func (l *Ledger) SendFromModuleAccount(module, to string, amount math.Int) error {
	return l.Send(module, to, amount)
}

// ---------------------------------------------------------------------------
// Vector 2: governance upgrades
// ---------------------------------------------------------------------------

// UpgradeProposal models a governance-approved upgrade handler's monetary
// effects. In a full SDK app this would arrive via x/gov + x/upgrade.
// Monetary parameters are immutable by design: ANY proposal that touches
// them must be rejected by the invariant check, and state must be
// unchanged after the failed upgrade. Per Constitution Article VI, a
// monetary schedule change requires a separately documented
// constitutional process, NOT an ordinary governance upgrade.
type UpgradeProposal struct {
	Name string
	// ChangeMaxSupply: if non-nil, the proposal tries to set a new max.
	ChangeMaxSupply *math.Int
	// EnableMinting: the proposal tries to turn discretionary minting on.
	EnableMinting bool
	// AddMintPath: the proposal tries to add a mint entrypoint/module.
	AddMintPath string
	// ChangeSchedule: if non-nil, the proposal tries to change the
	// predetermined issuance schedule through ordinary governance.
	ChangeSchedule *IssuanceSchedule
	// NonMonetaryNote: free text for changes that do not touch money.
	NonMonetaryNote string
}

// ApplyUpgrade enforces the rule. It returns an error and changes nothing
// for any monetary change. A purely non-monetary upgrade is recorded
// (no supply effect) and returns nil.
func (l *Ledger) ApplyUpgrade(u UpgradeProposal) error {
	if u.ChangeMaxSupply != nil {
		return fmt.Errorf("%w: upgrade %q attempted to change max_supply to %s", ErrUpgradeRejected, u.Name, u.ChangeMaxSupply.String())
	}
	if u.EnableMinting {
		return fmt.Errorf("%w: upgrade %q attempted to enable minting", ErrUpgradeRejected, u.Name)
	}
	if u.AddMintPath != "" {
		return fmt.Errorf("%w: upgrade %q attempted to add mint path %q", ErrUpgradeRejected, u.Name, u.AddMintPath)
	}
	if u.ChangeSchedule != nil {
		return fmt.Errorf("%w: upgrade %q attempted to change the issuance schedule through ordinary governance", ErrUpgradeRejected, u.Name)
	}
	// Non-monetary upgrade: accepted, no monetary state change.
	return nil
}

// ---------------------------------------------------------------------------
// Vector 3: smart-contract (CosmWasm) interfaces
// ---------------------------------------------------------------------------

// WasmBindings is the designed CosmWasm binding surface for OGO Coin.
// wasmd is pinned (v0.60.9) but the wasm VM is not instantiated in this
// Milestone B scaffold; live contract execution is Milestone C work.
// What IS enforced here, at the interface definition, is that no mint
// entrypoint exists in the binding surface at all.
//
// Exposed (safe) methods:
//   - BankSend: move coins between accounts (total preserved)
//   - BankBurn: burn coins (total decreases, one-way)
//   - QueryBalance / QuerySupply: read-only
//
// Deliberately ABSENT (any attempt to call one must fail):
//   - BankMint / MintCoins / CreateDenom / any mint entrypoint
var WasmExposedMethods = []string{"BankSend", "BankBurn", "QueryBalance", "QuerySupply"}

var wasmForbiddenMethods = map[string]bool{
	"BankMint": true, "MintCoins": true, "Mint": true,
	"CreateDenom": true, "MsgMint": true, "ForceMint": true,
}

// WasmMethodExposed reports whether a method name is in the binding surface.
func WasmMethodExposed(method string) bool {
	for _, m := range WasmExposedMethods {
		if m == method {
			return true
		}
	}
	return false
}

// WasmCall dispatches a contract call. Unknown or forbidden methods fail;
// exposed methods run through the ledger's normal paths.
func (l *Ledger) WasmCall(contract, method, from, to string, amount math.Int) error {
	if wasmForbiddenMethods[method] {
		return fmt.Errorf("%w: contract %q attempted forbidden mint method %q", ErrMintDisabled, contract, method)
	}
	if !WasmMethodExposed(method) {
		return fmt.Errorf("%w: contract %q called unknown method %q", ErrUnknownMethod, contract, method)
	}
	switch method {
	case "BankSend":
		return l.Send(from, to, amount)
	case "BankBurn":
		return l.Burn(from, amount)
	case "QueryBalance", "QuerySupply":
		return fmt.Errorf("ogocoin: %q is a query, not a transaction", method)
	default:
		return fmt.Errorf("%w: %q", ErrUnknownMethod, method)
	}
}

// WasmMintAttempt is the direct test hook: a contract attempting to mint.
func (l *Ledger) WasmMintAttempt(contract string, amount math.Int) error {
	return l.WasmCall(contract, "BankMint", contract, contract, amount)
}
