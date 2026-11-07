// Upgrade handlers for the Milestone C wired app (M2).
//
// v2 is the first on-chain upgrade of the wired chain. It is a
// NON-MONETARY upgrade: it adds one new state field, the upgrade
// marker in the new "monetary-meta" store, and changes no monetary
// parameter, no schedule, and no balance. The v2 binary is the same
// source built with -ldflags "-X .../app.UpgradeV2Enabled=true", so
// it mounts the monetary-meta store; the v1 binary does not.
//
// The handlers are registered ONLY in the v2 binary (see app.go):
// x/upgrade's PreBlocker panics with "BINARY UPDATED BEFORE TRIGGER"
// as soon as a plan exists if the running binary already has a
// handler for it, so a v1 binary registering the v2 handler halts
// the chain the moment the plan is set, before the upgrade height.
// (Observed 2026-11-06: halted at height 858, one block before the
// planned 859.)
//
// "monetary-change" is the rejection path for M3: an upgrade whose
// handler would change a monetary parameter is registered to return
// an error. A governance proposal naming it can be submitted, but
// the handler refuses to run, so the chain does not apply it. The
// rejection is tested in-process (app_test.go); it is NOT executed
// live, because a failing upgrade handler at the upgrade height
// halts the chain by design.
package app

import (
	"context"
	"fmt"

	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/types/module"

	monetary "github.com/benross-lab/ogo-network/x/monetary"
)

// UpgradeV2Name is the on-chain upgrade name for the v2 upgrade.
const UpgradeV2Name = "v2"

// MonetaryChangeUpgradeName is the upgrade name whose handler always
// rejects: it stands for "change a monetary parameter by upgrade".
const MonetaryChangeUpgradeName = "monetary-change"

// V2UpgradeHandler is the v2 upgrade handler (non-monetary). It is a
// named function so tests can invoke it directly.
func (app *OgoApp) V2UpgradeHandler(ctx context.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
	// Non-monetary migration: write the upgrade marker into the new
	// monetary-meta store (v2 binary only; the v1 binary never runs
	// this handler because Cosmovisor swaps it out at the upgrade
	// height).
	if UpgradeV2Enabled == "true" {
		metaKey := app.GetKey(monetary.MetaStoreKey)
		if metaKey == nil {
			return nil, fmt.Errorf("v2 upgrade: monetary-meta store key not mounted")
		}
		metaService := runtime.NewKVStoreService(metaKey)
		store := metaService.OpenKVStore(ctx)
		if err := store.Set([]byte("upgrade_marker"), []byte("v2")); err != nil {
			return nil, err
		}
		// Also record the marker in the monetary store, so it is
		// visible through the monetary keeper.
		if err := app.MonetaryKeeper.SetUpgradeMarker(ctx, "v2"); err != nil {
			return nil, err
		}
	}
	return app.ModuleManager.RunMigrations(ctx, app.Configurator(), fromVM)
}

// MonetaryChangeUpgradeHandler is the rejection handler for M3: an
// upgrade that would change a monetary parameter always fails.
func (app *OgoApp) MonetaryChangeUpgradeHandler(_ context.Context, _ upgradetypes.Plan, _ module.VersionMap) (module.VersionMap, error) {
	return nil, fmt.Errorf("%w: an upgrade may not change max supply, the issuance schedule, or the issuance model", monetary.ErrUpgradeRejected)
}

// RegisterUpgradeHandlers registers the upgrade handlers and the
// store loader for the v2 upgrade.
func (app *OgoApp) RegisterUpgradeHandlers() {
	app.UpgradeKeeper.SetUpgradeHandler(UpgradeV2Name, app.V2UpgradeHandler)
	app.UpgradeKeeper.SetUpgradeHandler(MonetaryChangeUpgradeName, app.MonetaryChangeUpgradeHandler)

	upgradeInfo, err := app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(err)
	}

	if upgradeInfo.Name == UpgradeV2Name && !app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		storeUpgrades := storetypes.StoreUpgrades{
			Added: []string{monetary.MetaStoreKey},
		}
		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &storeUpgrades))
	}
}
