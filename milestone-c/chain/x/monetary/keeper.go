// Keeper is the wired x/monetary keeper for the Milestone C SDK app.
//
// It is intentionally small and honest about what it does and does not
// do:
//
//   - It holds the monetary Params (max supply, denom, schedule, model)
//     and the issued-by-epoch set in its own SDK store, so both are
//     covered by the app hash. The legacy in-memory Ledger in ledger.go
//     remains the pure state machine used by the unit tests; this keeper
//     is the store-backed counterpart that runs on chain, and it mints
//     through x/bank (MintCoins into the monetary module account, the
//     rewards pool) instead of mutating an in-memory balance map.
//   - The ONLY minting path is MintForEpoch, called from the epoch hook
//     with the exact model amount for one epoch, once. There is no
//     discretionary mint, no MsgServer that changes params, and no
//     governance path that changes the schedule: SetParams refuses any
//     change to MaxSupply or the schedule after genesis.
//   - Genesis balances in the monetary genesis state are a DECLARATION
//     of G (the genesis circulating supply) for the strict
//     M = G + S_actual check. The actual funded balances live in x/bank
//     genesis. InitGenesis cross-checks the two: the x/bank uogo supply
//     at init time must equal the declared G, or genesis fails.
//
// All monetary values in code and tests are TEST-ONLY and labelled so.
// No production monetary parameter is selected here.
package monetary

import (
	"context"
	"encoding/json"
	"fmt"

	"cosmossdk.io/core/store"
	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// BankKeeper is the minimal x/bank surface the monetary keeper needs.
// bankkeeper.Keeper satisfies it.
type BankKeeper interface {
	MintCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
	GetSupply(ctx context.Context, denom string) sdk.Coin
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
}

// Keeper holds monetary state in its SDK store.
type Keeper struct {
	storeService store.KVStoreService
	cdc          codec.BinaryCodec
	bankKeeper   BankKeeper
}

// NewKeeper creates a monetary keeper. The bank keeper is supplied by
// the app after depinject builds x/bank.
func NewKeeper(storeService store.KVStoreService, cdc codec.BinaryCodec, bankKeeper BankKeeper) Keeper {
	return Keeper{storeService: storeService, cdc: cdc, bankKeeper: bankKeeper}
}

// IssuanceScheduleValue returns the schedule as a value (zero value
// when the pointer is nil), for schedule comparisons.
func (p Params) IssuanceScheduleValue() IssuanceSchedule {
	if p.IssuanceSchedule == nil {
		return IssuanceSchedule{}
	}
	return *p.IssuanceSchedule
}

// RewardsPoolAddress is the monetary module account: scheduled issuance
// is minted into this account, which is the rewards pool in the wired
// chain (the prototype's separate "ogo1rewardspool" string was not a
// real bech32 address and could hold no funds).
func RewardsPoolAddress() sdk.AccAddress {
	return authtypes.NewModuleAddress(ModuleName)
}

func paramsKey() []byte      { return []byte("params") }
func totalIssuedKey() []byte { return []byte("total_issued") }
func issuedKey(epoch int64) []byte {
	return []byte(fmt.Sprintf("issued/%d", epoch))
}
func upgradeMarkerKey() []byte { return []byte("upgrade_marker") }

// GetParams returns the stored params, or an error if genesis has not run.
func (k Keeper) GetParams(ctx context.Context) (Params, error) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(paramsKey())
	if err != nil {
		return Params{}, err
	}
	if bz == nil {
		return Params{}, fmt.Errorf("monetary: params not set (genesis has not run)")
	}
	var p Params
	if err := json.Unmarshal(bz, &p); err != nil {
		return Params{}, err
	}
	return p, nil
}

// SetParams stores params. It is used only by InitGenesis and by the
// upgrade handler for non-monetary migrations; it refuses any change
// to MaxSupply or the schedule once params exist, so no governance or
// upgrade path can move a monetary parameter.
func (k Keeper) SetParams(ctx context.Context, p Params) error {
	store := k.storeService.OpenKVStore(ctx)
	existing, err := store.Get(paramsKey())
	if err != nil {
		return err
	}
	if existing != nil {
		var old Params
		if err := json.Unmarshal(existing, &old); err != nil {
			return err
		}
		if old.MaxSupply == nil || p.MaxSupply == nil || !old.MaxSupply.Equal(*p.MaxSupply) {
			return fmt.Errorf("%w: max_supply is immutable", ErrUpgradeRejected)
		}
		if !schedulesEqual(old.IssuanceScheduleValue(), p.IssuanceScheduleValue()) {
			return fmt.Errorf("%w: issuance schedule is immutable", ErrUpgradeRejected)
		}
		if old.IssuanceModel != p.IssuanceModel {
			return fmt.Errorf("%w: issuance model is immutable", ErrUpgradeRejected)
		}
	}
	bz, err := json.Marshal(p)
	if err != nil {
		return err
	}
	return store.Set(paramsKey(), bz)
}

// GetTotalIssued returns the cumulative scheduled issuance.
func (k Keeper) GetTotalIssued(ctx context.Context) (math.Int, error) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(totalIssuedKey())
	if err != nil {
		return math.NewInt(0), err
	}
	if bz == nil {
		return math.NewInt(0), nil
	}
	i, ok := math.NewIntFromString(string(bz))
	if !ok {
		return math.NewInt(0), fmt.Errorf("monetary: corrupt total_issued %q", string(bz))
	}
	return i, nil
}

// GetIssued returns the amount issued for epoch n (0-indexed), and
// whether that epoch has been claimed.
func (k Keeper) GetIssued(ctx context.Context, epoch int64) (math.Int, bool, error) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(issuedKey(epoch))
	if err != nil {
		return math.NewInt(0), false, err
	}
	if bz == nil {
		return math.NewInt(0), false, nil
	}
	i, ok := math.NewIntFromString(string(bz))
	if !ok {
		return math.NewInt(0), false, fmt.Errorf("monetary: corrupt issued amount for epoch %d", epoch)
	}
	return i, true, nil
}

// InitGenesis validates the monetary genesis state strictly
// (M = G + S_actual, no headroom) and stores the params. It also
// cross-checks the declared genesis supply against the x/bank uogo
// supply at init time: bank genesis must have run first (the app's
// InitGenesis order puts bank before monetary), and the two must agree.
//
// Migration guard: if params are already stored, this is not a fresh
// genesis but a RunMigrations call for a module that was missing from
// x/upgrade's stored version map (the depinject-built init version
// map does not include modules registered via the RegisterModules
// hybrid hook, so monetary is absent until the init-map merge in
// app.go). RunMigrations calls InitGenesis with DefaultGenesis (no
// max supply) for such a module. Re-initialising would wipe the
// live params, so when state already exists InitGenesis is a no-op
// and the stored params, issued set, and total issued are preserved.
func (k Keeper) InitGenesis(ctx sdk.Context, gs GenesisState) error {
	if _, err := k.GetParams(ctx); err == nil {
		return nil
	}
	if err := ValidateGenesisExact(gs); err != nil {
		return err
	}
	if k.bankKeeper != nil {
		declared := TotalGenesisSupply(gs)
		supply := k.bankKeeper.GetSupply(ctx, gs.Params.Denom)
		if !supply.Amount.Equal(declared) {
			return fmt.Errorf("monetary: declared genesis supply %s does not match x/bank %s supply %s (bank genesis must list the same allocations as the monetary genesis declaration)",
				declared.String(), gs.Params.Denom, supply.Amount.String())
		}
	}
	if err := k.SetParams(ctx, gs.Params); err != nil {
		return err
	}
	store := k.storeService.OpenKVStore(ctx)
	return store.Set(totalIssuedKey(), []byte("0"))
}

// ExportGenesis returns the stored params plus the issued-by-epoch set.
func (k Keeper) ExportGenesis(ctx context.Context) (GenesisState, error) {
	p, err := k.GetParams(ctx)
	if err != nil {
		return GenesisState{}, err
	}
	return GenesisState{Params: p, Balances: nil}, nil
}

// MintForEpoch mints exactly R_n for epoch n (0-indexed) from the model
// selected by genesis params, once, into the monetary module account
// (the rewards pool). It is the single model-aware mint entrypoint in
// the wired app (the legacy ledger entrypoints remain for unit tests
// only). A second claim of the same epoch fails with ErrDoubleIssuance
// and mints nothing.
func (k Keeper) MintForEpoch(ctx sdk.Context, epoch int64) (math.Int, error) {
	p, err := k.GetParams(ctx)
	if err != nil {
		return math.NewInt(0), err
	}
	model, err := p.SelectedModel()
	if err != nil {
		return math.NewInt(0), err
	}
	if epoch < 0 || epoch >= model.Periods() {
		return math.NewInt(0), ErrScheduleExhausted
	}
	if _, claimed, err := k.GetIssued(ctx, epoch); err != nil {
		return math.NewInt(0), err
	} else if claimed {
		return math.NewInt(0), fmt.Errorf("%w: epoch %d already issued", ErrDoubleIssuance, epoch)
	}
	amount, err := model.AmountForEpoch(epoch)
	if err != nil {
		return math.NewInt(0), err
	}
	store := k.storeService.OpenKVStore(ctx)
	if err := store.Set(issuedKey(epoch), []byte(amount.String())); err != nil {
		return math.NewInt(0), err
	}
	if amount.IsZero() {
		return amount, nil
	}
	if k.bankKeeper == nil {
		return math.NewInt(0), fmt.Errorf("monetary: bank keeper not wired")
	}
	supply := k.bankKeeper.GetSupply(ctx, p.Denom)
	if supply.Amount.Add(amount).GT(*p.MaxSupply) {
		_ = store.Delete(issuedKey(epoch))
		return math.NewInt(0), fmt.Errorf("%w: scheduled issuance %s for epoch %d would exceed max %s", ErrSupplyOverMax, amount.String(), epoch, p.MaxSupply.String())
	}
	coins := sdk.NewCoins(sdk.NewCoin(p.Denom, amount))
	if err := k.bankKeeper.MintCoins(ctx, ModuleName, coins); err != nil {
		_ = store.Delete(issuedKey(epoch))
		return math.NewInt(0), err
	}
	total, err := k.GetTotalIssued(ctx)
	if err != nil {
		return math.NewInt(0), err
	}
	if err := store.Set(totalIssuedKey(), []byte(total.Add(amount).String())); err != nil {
		return math.NewInt(0), err
	}
	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"monetary_epoch_issuance",
		sdk.NewAttribute("epoch_number", fmt.Sprintf("%d", epoch)),
		sdk.NewAttribute("amount", amount.String()),
		sdk.NewAttribute("recipient", RewardsPoolAddress().String()),
	))
	return amount, nil
}

// SetUpgradeMarker records the v2 upgrade marker in the monetary store.
// It is written by the v2 upgrade handler, not by any mint path.
func (k Keeper) SetUpgradeMarker(ctx context.Context, value string) error {
	store := k.storeService.OpenKVStore(ctx)
	return store.Set(upgradeMarkerKey(), []byte(value))
}

// GetUpgradeMarker returns the v2 upgrade marker, or "" if unset.
func (k Keeper) GetUpgradeMarker(ctx context.Context) (string, error) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(upgradeMarkerKey())
	if err != nil {
		return "", err
	}
	return string(bz), nil
}
