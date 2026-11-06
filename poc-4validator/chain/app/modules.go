// Package app wires the OGO Network local prototype.
//
// Honest status of the wiring (Milestone B):
//   - The RUNNING localnet binary is a CometBFT ABCI 2.0 application
//     (see app.go) whose state machine is x/ogocoin. It produces real
//     blocks, real transfers, and real supply queries.
//   - The standard Cosmos SDK modules below are PINNED to the Gaia
//     v28.3.1 matrix and COMPILED into this module via the blank imports
//     (proving the pinned matrix builds as one unit), but their keepers
//     are NOT yet wired through depinject. Full keeper wiring of
//     auth/bank/staking/gov/distribution + wasmd + ibc-go is the next
//     engineering step (Milestone C). Nothing in this file claims that
//     wiring is done.
//   - x/ogocoin is the only fully implemented module in this scaffold,
//     and it is the module that enforces the monetary rules.
package app

import (
	_ "github.com/CosmWasm/wasmd/x/wasm"
	_ "github.com/cosmos/cosmos-sdk/x/auth"
	_ "github.com/cosmos/cosmos-sdk/x/bank"
	_ "github.com/cosmos/cosmos-sdk/x/distribution"
	_ "github.com/cosmos/cosmos-sdk/x/gov"
	_ "github.com/cosmos/cosmos-sdk/x/staking"
	_ "github.com/cosmos/ibc-go/modules/light-clients/08-wasm/v10"
	_ "github.com/cosmos/ibc-go/v10/modules/apps/transfer"
	_ "github.com/cosmos/ibc-go/v10/modules/core"
)

// ModuleStatus records what is real in this scaffold and what is pending.
type ModuleStatus struct {
	Name    string
	Version string
	Status  string // "implemented" | "pinned-compiled" | "pending"
	Notes   string
}

// ModuleRegistry is the wiring status of every module in the OGO app.
var ModuleRegistry = []ModuleStatus{
	{Name: "x/ogocoin", Version: "local (this repo)", Status: "implemented", Notes: "Fixed-supply enforcement: genesis validation, ledger, IBC escrow accounting, bypass surface. Fully tested."},
	{Name: "x/auth", Version: "SDK v0.53.4 effective (v0.53.6 required)", Status: "pinned-compiled", Notes: "Package compiled to prove the matrix builds. Keeper wiring via depinject is pending (Milestone C). Prototype ABCI app does not yet verify signatures; see README limitations."},
	{Name: "x/bank", Version: "SDK v0.53.4 effective", Status: "pinned-compiled", Notes: "Constrained by x/ogocoin in the target design. In this scaffold the ledger implements bank-like balances directly; x/bank keeper wiring is pending."},
	{Name: "x/staking", Version: "SDK v0.53.4 effective", Status: "pinned-compiled", Notes: "Keeper wiring pending. Validator rewards come from existing supply only (no mint); rates TBD."},
	{Name: "x/gov", Version: "SDK v0.53.4 effective", Status: "pinned-compiled", Notes: "Keeper wiring pending. The upgrade-rejection rule is enforced in x/ogocoin.ApplyUpgrade and tested; on-chain gov proposals are pending."},
	{Name: "x/distribution", Version: "SDK v0.53.4 effective", Status: "pinned-compiled", Notes: "Keeper wiring pending."},
	{Name: "x/wasm (wasmd)", Version: "v0.60.9", Status: "pinned-compiled", Notes: "VM not instantiated in this scaffold. The binding surface (no mint) is defined in x/ogocoin/bypass.go and tested. Live contract execution is Milestone C."},
	{Name: "ibc-go core + transfer", Version: "v10.7.0", Status: "pinned-compiled", Notes: "Escrow/voucher accounting is implemented and tested at the ledger level. Live IBC relaying is pending."},
	{Name: "ibc-go 08-wasm", Version: "v10.5.0", Status: "pinned-compiled", Notes: "Light client package compiled; live use pending."},
}
