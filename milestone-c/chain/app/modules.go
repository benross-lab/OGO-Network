// Package app wires the OGO Network Milestone C wired application.
//
// Milestone C status (honest): the standard SDK modules below are
// WIRED through depinject (auth, bank, staking, slashing,
// distribution, gov, upgrade, epochs, consensus, evidence, feegrant,
// authz, genutil, vesting, circuit) and their keepers are reachable
// through their query paths. x/monetary is wired through the
// RegisterModules hybrid hook. x/mint is deliberately absent. IBC and
// CosmWasm are not wired in this build (see the build report).
package app

import (
	// Pinned-compiled only (Milestone C): wasmd and ibc-go are part
	// of the locked Gaia v28.3.1 matrix and are compiled into this
	// module to prove the matrix builds as one unit, but their
	// keepers are NOT wired in this build (see ModuleRegistry).
	// Full keeper wiring of wasm is the S7 capability check; live
	// IBC relaying stays deferred (B8).
	_ "github.com/CosmWasm/wasmd/x/wasm"
	_ "github.com/cosmos/ibc-go/modules/light-clients/08-wasm/v10"
	_ "github.com/cosmos/ibc-go/v10/modules/apps/transfer"
	_ "github.com/cosmos/ibc-go/v10/modules/core"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	epochstypes "github.com/cosmos/cosmos-sdk/x/epochs/types"
	evidencetypes "cosmossdk.io/x/evidence/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	monetary "github.com/benross-lab/ogo-network/x/monetary"
)

// ModuleStatus records what is real in this wired app.
type ModuleStatus struct {
	Name    string
	Version string
	Status  string // "wired" | "wired-hybrid" | "absent-deliberate" | "not-wired"
	Notes   string
}

// ModuleRegistry is the wiring status of every module in the OGO app.
var ModuleRegistry = []ModuleStatus{
	{Name: monetary.ModuleName, Version: "local (this repo)", Status: "wired-hybrid", Notes: "Fixed-supply enforcement via x/bank minting; epoch hook into real x/epochs; strict M = G + S_actual genesis. Registered via RegisterModules/RegisterStores."},
	{Name: epochstypes.ModuleName, Version: "SDK v0.53.4 effective", Status: "wired", Notes: "Real x/epochs keeper, store, and queries via depinject; hook registered by the SetHooks-before-AppModule swap (SDK #25425 workaround, proven by the hook-fires test)."},
	{Name: authtypes.ModuleName, Version: "SDK v0.53.4 effective", Status: "wired", Notes: "Signature verification via the standard ante handler; unsigned and wrongly-signed txs rejected before execution."},
	{Name: banktypes.ModuleName, Version: "SDK v0.53.4 effective", Status: "wired", Notes: "Balances, supply, and module-account minting for x/monetary."},
	{Name: stakingtypes.ModuleName, Version: "SDK v0.53.4 effective", Status: "wired", Notes: "Validator operations and delegation; bond denom uogo (TEST-ONLY genesis)."},
	{Name: slashingtypes.ModuleName, Version: "SDK v0.53.4 effective", Status: "wired", Notes: "Downtime tracking, jailing, and recovery."},
	{Name: distrtypes.ModuleName, Version: "SDK v0.53.4 effective", Status: "wired", Notes: "Distribution; staking hooks wired."},
	{Name: govtypes.ModuleName, Version: "SDK v0.53.4 effective", Status: "wired", Notes: "Governance proposals, including software upgrades; monetary parameters are not changeable through gov."},
	{Name: upgradetypes.ModuleName, Version: "SDK v0.53.4 effective", Status: "wired", Notes: "x/upgrade with Cosmovisor; v2 upgrade adds the monetary-meta store; monetary-change upgrades are rejected."},
	{Name: consensustypes.ModuleName, Version: "SDK v0.53.4 effective", Status: "wired", Notes: "Consensus params keeper."},
	{Name: evidencetypes.ModuleName, Version: "SDK v0.53.4 effective", Status: "wired", Notes: "Evidence handling (double-sign evidence path)."},
	{Name: "mint", Version: "SDK v0.53.4 effective", Status: "absent-deliberate", Notes: "Deliberately not included: OGO has no discretionary minting; scheduled issuance is x/monetary only."},
	{Name: "wasm (wasmd)", Version: "v0.60.9", Status: "not-wired", Notes: "VM not instantiated in this build. S7 capability check is a separate step; see the build report."},
	{Name: "ibc-go core + transfer", Version: "v10.7.0", Status: "not-wired", Notes: "Live IBC relaying deferred (B8). Escrow/voucher accounting remains at the ledger level in x/monetary tests."},
}
