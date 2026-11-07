// AppModule is the x/monetary module for the Milestone C SDK app.
//
// Registration note, recorded honestly: x/monetary is registered
// through the runtime App's RegisterModules/RegisterStores hybrid hook
// (see app/app.go), not through a depinject app-config entry, because
// a depinject module config needs a generated protobuf Module message
// and this module's genesis and state are plain JSON (no protobuf
// types exist for it yet). The standard SDK modules ARE wired through
// depinject (app/app_config.go). The monetary keeper itself is built
// with the same KVStoreService pattern the depinject modules use, and
// its store key is mounted before Load. Adding a protobuf module
// config so monetary can move into app_config is a follow-up, not a
// Milestone C blocker: the keeper, its store, its genesis order, and
// its epoch hook are all live in the wired app either way.
package monetary

import (
	"encoding/json"
	"fmt"

	gwruntime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"

	"cosmossdk.io/core/appmodule"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
)

// ConsensusVersion is the module consensus version. It stays at 1
// across the v2 upgrade: the v2 store migration adds the
// monetary-meta store but does not change the monetary module's
// consensus version, so no per-module migration runs for it.
const ConsensusVersion = 1

// AppModule implements the x/monetary module.
type AppModule struct {
	keeper Keeper
}

// NewAppModule creates an AppModule from a keeper.
func NewAppModule(keeper Keeper) AppModule {
	return AppModule{keeper: keeper}
}

var (
	_ module.AppModule      = AppModule{}
	_ module.HasGenesis     = AppModule{}
	_ appmodule.AppModule   = AppModule{}
	_ appmodule.HasServices = AppModule{}
)

// IsAppModule implements appmodule.AppModule.
func (am AppModule) IsAppModule() {}

// IsOnePerModuleType implements depinject.OnePerModuleType.
func (am AppModule) IsOnePerModuleType() {}

// Name returns the module name.
func (am AppModule) Name() string { return ModuleName }

// RegisterLegacyAminoCodec registers the module's types (none: no msgs).
func (am AppModule) RegisterLegacyAminoCodec(_ *codec.LegacyAmino) {}

// RegisterInterfaces registers the module's interface types (none).
func (am AppModule) RegisterInterfaces(_ codectypes.InterfaceRegistry) {}

// RegisterGRPCGatewayRoutes registers REST routes (none: no query
// service yet; monetary state is readable through x/bank balances and
// supply, x/epochs epoch-infos, and export genesis. A dedicated gRPC
// query service for params and the issued set is a follow-up).
func (am AppModule) RegisterGRPCGatewayRoutes(_ client.Context, _ *gwruntime.ServeMux) {}

// RegisterServices registers module services (none yet; see above).
func (am AppModule) RegisterServices(_ grpc.ServiceRegistrar) error { return nil }

// DefaultGenesis returns the default genesis state. It deliberately
// has no max supply and no schedule: there is no default monetary
// parameter, and a genesis file must be edited to set one before the
// chain can start. Genesis is plain JSON (not protobuf), so encoding
// uses encoding/json directly, not the proto codec.
func (am AppModule) DefaultGenesis(_ codec.JSONCodec) json.RawMessage {
	gs := GenesisState{Params: DefaultParams(), Balances: nil}
	bz, err := json.Marshal(gs)
	if err != nil {
		panic(err)
	}
	return bz
}

// ValidateGenesis validates the monetary genesis state. The wired
// chain uses the strict check (M = G + S_actual).
func (am AppModule) ValidateGenesis(_ codec.JSONCodec, _ client.TxEncodingConfig, bz json.RawMessage) error {
	var gs GenesisState
	if err := json.Unmarshal(bz, &gs); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", ModuleName, err)
	}
	return ValidateGenesisExact(gs)
}

// InitGenesis initialises the monetary module from genesis.
func (am AppModule) InitGenesis(ctx sdk.Context, _ codec.JSONCodec, bz json.RawMessage) {
	var gs GenesisState
	if err := json.Unmarshal(bz, &gs); err != nil {
		panic(fmt.Errorf("failed to unmarshal %s genesis state: %w", ModuleName, err))
	}
	if err := am.keeper.InitGenesis(ctx, gs); err != nil {
		panic(fmt.Errorf("monetary InitGenesis failed: %w", err))
	}
}

// ExportGenesis exports the monetary genesis state.
func (am AppModule) ExportGenesis(ctx sdk.Context, _ codec.JSONCodec) json.RawMessage {
	gs, err := am.keeper.ExportGenesis(ctx)
	if err != nil {
		panic(err)
	}
	bz, err := json.Marshal(gs)
	if err != nil {
		panic(err)
	}
	return bz
}

// ConsensusVersion returns the module consensus version.
func (am AppModule) ConsensusVersion() uint64 { return ConsensusVersion }

// Keeper returns the module keeper (used by the app for the epoch hook
// and the upgrade handler).
func (am AppModule) Keeper() Keeper { return am.keeper }
