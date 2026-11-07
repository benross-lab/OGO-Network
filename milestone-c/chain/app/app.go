// OgoApp is the Milestone C wired Cosmos SDK application.
//
// Wiring, stated plainly:
//   - The standard SDK modules (auth, bank, staking, slashing,
//     distribution, gov, upgrade, epochs, consensus, evidence,
//     feegrant, authz, genutil, vesting, circuit) are wired through
//     depinject from app_config.go (the simapp app_di.go pattern).
//     x/mint is deliberately NOT included: OGO has no discretionary
//     minting.
//   - x/monetary is registered through the runtime App's
//     RegisterModules/RegisterStores hybrid hook (the hook the SDK
//     documents for modules not registered through the app config),
//     before Load. Its keeper uses the same KVStoreService pattern.
//   - x/epochs hook registration: the pinned SDK v0.53.4 depinject
//     wiring has the copy-versus-pointer defect documented in
//     EPOCHS_GAP.md (issues #24770, #25425): ProvideModule copies the
//     keeper into the AppModule by value, while InvokeSetHooks sets
//     hooks on the depinject-held keeper, so the AppModule's copy
//     never receives them. This app therefore does what simapp's
//     manual app.go does (lines 495-501 there): it calls SetHooks on
//     the epochs keeper and then builds the epochs AppModule FROM
//     that hooked keeper, replacing the depinject-built module in the
//     ModuleManager before any block runs. The hook-fires test in
//     app_test.go proves the hook actually fires in the wired app;
//     compilation alone would not catch the defect.
//   - IBC and CosmWasm are NOT wired in this build (see DECISIONS.md
//     and the build report): IBC relaying stays deferred (B8), and
//     the wasm VM is a separate capability check (S7).
package app

import (
	"io"

	dbm "github.com/cosmos/cosmos-db"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	"cosmossdk.io/client/v2/autocli"
	"cosmossdk.io/core/appmodule"

	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	circuitkeeper "cosmossdk.io/x/circuit/keeper"
	evidencekeeper "cosmossdk.io/x/evidence/keeper"
	feegrantkeeper "cosmossdk.io/x/feegrant/keeper"
	upgradekeeper "cosmossdk.io/x/upgrade/keeper"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/address"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	runtimeservices "github.com/cosmos/cosmos-sdk/runtime/services"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authsims "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	consensuskeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	"github.com/cosmos/cosmos-sdk/x/epochs"
	epochskeeper "github.com/cosmos/cosmos-sdk/x/epochs/keeper"
	epochstypes "github.com/cosmos/cosmos-sdk/x/epochs/types"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"

	monetary "github.com/benross-lab/ogo-network/x/monetary"
)

var (
	_ runtime.AppI            = (*OgoApp)(nil)
	_ servertypes.Application = (*OgoApp)(nil)
)

// UpgradeV2Enabled is set to "true" at link time for the v2 upgrade
// binary only (-ldflags "-X .../app.UpgradeV2Enabled=true"). The v1
// binary has it "false". The v2 binary mounts the monetary-meta store
// key and can run the v2 upgrade handler's store migration.
var UpgradeV2Enabled = "false"

// OgoApp extends the depinject-built runtime App with the OGO keepers.
type OgoApp struct {
	*runtime.App
	legacyAmino       *codec.LegacyAmino
	appCodec          codec.Codec
	txConfig          client.TxConfig
	interfaceRegistry codectypes.InterfaceRegistry

	AccountKeeper         authkeeper.AccountKeeper
	BankKeeper            bankkeeper.BaseKeeper
	StakingKeeper         *stakingkeeper.Keeper
	SlashingKeeper        slashingkeeper.Keeper
	DistrKeeper           distrkeeper.Keeper
	GovKeeper             *govkeeper.Keeper
	UpgradeKeeper         *upgradekeeper.Keeper
	EvidenceKeeper        evidencekeeper.Keeper
	ConsensusParamsKeeper consensuskeeper.Keeper
	CircuitKeeper         circuitkeeper.Keeper
	FeeGrantKeeper        feegrantkeeper.Keeper
	AuthzKeeper           authzkeeper.Keeper
	EpochsKeeper          epochskeeper.Keeper

	MonetaryKeeper monetary.Keeper
	MonetaryModule monetary.AppModule

	// simulation manager (nil until built in NewOgoApp)
	sm *module.SimulationManager
}

// NewOgoApp returns an initialized OgoApp.
func NewOgoApp(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	loadLatest bool,
	appOpts servertypes.AppOptions,
	baseAppOptions ...func(*baseapp.BaseApp),
) *OgoApp {
	var (
		app        = &OgoApp{}
		appBuilder *runtime.AppBuilder
		appConfig  = depinject.Configs(
			AppConfig,
			depinject.Supply(
				appOpts,
				logger,
			),
		)
	)

	if err := depinject.Inject(appConfig,
		&appBuilder,
		&app.appCodec,
		&app.legacyAmino,
		&app.txConfig,
		&app.interfaceRegistry,
		&app.AccountKeeper,
		&app.BankKeeper,
		&app.StakingKeeper,
		&app.SlashingKeeper,
		&app.DistrKeeper,
		&app.GovKeeper,
		&app.UpgradeKeeper,
		&app.AuthzKeeper,
		&app.EvidenceKeeper,
		&app.FeeGrantKeeper,
		&app.ConsensusParamsKeeper,
		&app.CircuitKeeper,
		&app.EpochsKeeper,
	); err != nil {
		panic(err)
	}

	app.App = appBuilder.Build(db, traceStore, baseAppOptions...)

	// x/monetary is registered via the RegisterModules hybrid hook,
	// not via depinject, so the depinject-built init version map
	// (which x/upgrade stores at genesis) does not include it.
	// Without this merge, a later RunMigrations treats monetary as
	// a brand-new module and calls InitGenesis with DefaultGenesis.
	// Merge it in so future chains record it from genesis.
	if vm := app.UpgradeKeeper.GetInitVersionMap(); vm != nil {
		vm[monetary.ModuleName] = monetary.ConsensusVersion
		app.UpgradeKeeper.SetInitVersionMap(vm)
	}

	if err := app.RegisterStreamingServices(appOpts, app.kvStoreKeys()); err != nil {
		panic(err)
	}

	// x/monetary: store key, keeper, and module registration (hybrid
	// hook, before Load so genesis orders include it).
	monetaryKey := storetypes.NewKVStoreKey(monetary.StoreKey)
	keys := []storetypes.StoreKey{monetaryKey}
	if UpgradeV2Enabled == "true" {
		keys = append(keys, storetypes.NewKVStoreKey(monetary.MetaStoreKey))
	}
	if err := app.RegisterStores(keys...); err != nil {
		panic(err)
	}
	app.MonetaryKeeper = monetary.NewKeeper(
		runtime.NewKVStoreService(monetaryKey),
		app.appCodec,
		app.BankKeeper,
	)
	app.MonetaryModule = monetary.NewAppModule(app.MonetaryKeeper)
	if err := app.RegisterModules(app.MonetaryModule); err != nil {
		panic(err)
	}

	// x/epochs hook fix (see the package doc): SetHooks on the
	// depinject-built keeper, then rebuild the epochs AppModule from
	// that hooked keeper and replace the depinject-built module in
	// the manager. The depinject InvokeSetHooks path alone would
	// leave the AppModule's keeper copy hookless (SDK #25425).
	app.EpochsKeeper.SetHooks(epochstypes.NewMultiEpochHooks(
		monetary.MonetaryEpochHook{Keeper: app.MonetaryKeeper},
	))
	app.ModuleManager.Modules[epochstypes.ModuleName] = epochs.NewAppModule(app.EpochsKeeper)

	// Ante handler: the standard SDK handler, which verifies
	// signatures (x/auth SigVerificationDecorator), checks sequences,
	// meters gas, and deducts fees. Unsigned and wrongly-signed
	// transactions are rejected here, before execution. There is no
	// hardcoded GasWanted anywhere in this app.
	app.setAnteHandler(app.txConfig)

	// Upgrade handlers are registered ONLY in the v2 binary
	// (UpgradeV2Enabled == "true"). The x/upgrade PreBlocker panics
	// with "BINARY UPDATED BEFORE TRIGGER" as soon as a plan exists
	// if the running binary already has a handler for it, so a v1
	// binary that registered the v2 handler would halt the chain the
	// moment the upgrade plan is set, before the upgrade height.
	// The standard Cosmovisor flow is: v1 (no handler) halts at the
	// upgrade height with "UPGRADE NEEDED" and dumps upgrade-info.json;
	// Cosmovisor swaps in v2 (handler registered), which applies the
	// migration. (Observed 2026-11-06: the ungated registration halted
	// the chain at height 858, one block before the planned 859.)
	if UpgradeV2Enabled == "true" {
		app.RegisterUpgradeHandlers()
	}

	if err := app.Load(loadLatest); err != nil {
		panic(err)
	}

	// Simulation manager, for the simulation framework only.
	overrideModules := map[string]module.AppModuleSimulation{
		authtypes.ModuleName: auth.NewAppModule(app.appCodec, app.AccountKeeper, authsims.RandomGenesisAccounts, nil),
	}
	app.sm = module.NewSimulationManagerFromAppModules(app.ModuleManager.Modules, overrideModules)
	app.sm.RegisterStoreDecoders()

	return app
}

// SimulationManager implements the SimulationApp interface.
func (app *OgoApp) SimulationManager() *module.SimulationManager {
	return app.sm
}

func (app *OgoApp) setAnteHandler(txConfig client.TxConfig) {
	anteHandler, err := ante.NewAnteHandler(ante.HandlerOptions{
		AccountKeeper:   app.AccountKeeper,
		BankKeeper:      app.BankKeeper,
		SignModeHandler: txConfig.SignModeHandler(),
		FeegrantKeeper:  app.FeeGrantKeeper,
		SigGasConsumer:  ante.DefaultSigVerificationGasConsumer,
	})
	if err != nil {
		panic(err)
	}
	app.SetAnteHandler(anteHandler)
}

// LegacyAmino returns the app's amino codec.
func (app *OgoApp) LegacyAmino() *codec.LegacyAmino { return app.legacyAmino }

// AppCodec returns the app's codec.
func (app *OgoApp) AppCodec() codec.Codec { return app.appCodec }

// InterfaceRegistry returns the app's interface registry.
func (app *OgoApp) InterfaceRegistry() codectypes.InterfaceRegistry {
	return app.interfaceRegistry
}

// TxConfig returns the app's tx config.
func (app *OgoApp) TxConfig() client.TxConfig { return app.txConfig }

// AutoCliOpts returns the autocli options for the app.
func (app *OgoApp) AutoCliOpts() autocli.AppOptions {
	modules := make(map[string]appmodule.AppModule, 0)
	for _, m := range app.ModuleManager.Modules {
		if moduleWithName, ok := m.(module.HasName); ok {
			moduleName := moduleWithName.Name()
			if appModule, ok := moduleWithName.(appmodule.AppModule); ok {
				modules[moduleName] = appModule
			}
		}
	}

	return autocli.AppOptions{
		Modules:               modules,
		ModuleOptions:         runtimeservices.ExtractAutoCLIOptions(app.ModuleManager.Modules),
		AddressCodec:          address.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix()),
		ValidatorAddressCodec: address.NewBech32Codec(sdk.GetConfig().GetBech32ValidatorAddrPrefix()),
		ConsensusAddressCodec: address.NewBech32Codec(sdk.GetConfig().GetBech32ConsensusAddrPrefix()),
	}
}

var _ = autocliv1.ModuleOptions{}

// GetKey returns the KVStoreKey for the provided store key.
func (app *OgoApp) GetKey(storeKey string) *storetypes.KVStoreKey {
	sk := app.UnsafeFindStoreKey(storeKey)
	kvStoreKey, ok := sk.(*storetypes.KVStoreKey)
	if !ok {
		return nil
	}
	return kvStoreKey
}

func (app *OgoApp) kvStoreKeys() map[string]*storetypes.KVStoreKey {
	keys := make(map[string]*storetypes.KVStoreKey)
	for _, k := range app.GetStoreKeys() {
		if kv, ok := k.(*storetypes.KVStoreKey); ok {
			keys[kv.Name()] = kv
		}
	}
	return keys
}

// RegisterAPIRoutes registers all application module routes.
func (app *OgoApp) RegisterAPIRoutes(apiSvr *api.Server, apiConfig config.APIConfig) {
	app.App.RegisterAPIRoutes(apiSvr, apiConfig)
	if err := server.RegisterSwaggerAPI(apiSvr.ClientCtx, apiSvr.Router, apiConfig.Swagger); err != nil {
		panic(err)
	}
}

