// Epoch hook: x/monetary's AfterEpochEnd hook, wired into real
// x/epochs in the Milestone C app.
//
// The hook receives 1-indexed SDK epoch numbers; the issuance models
// are 0-indexed, so epoch n ends period n-1. Only the "monetary"
// identifier mints; all other identifiers are ignored. A duplicate
// claim (same epoch twice) fails in the keeper and mints nothing, so
// a second execution cannot issue twice: the issued-by-epoch set in
// the monetary store is the guard, and the hook runs in the epochs
// keeper's cached context, so a failed hook's writes are discarded
// and the epoch still advances (upstream behaviour).
package monetary

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	epochstypes "github.com/cosmos/cosmos-sdk/x/epochs/types"
)

// MonetaryEpochHook is the x/epochs hook for x/monetary.
type MonetaryEpochHook struct {
	Keeper Keeper
}

var _ epochstypes.EpochHooks = MonetaryEpochHook{}

// AfterEpochEnd mints the scheduled issuance for the epoch that just
// ended, exactly once.
func (h MonetaryEpochHook) AfterEpochEnd(ctx context.Context, identifier string, epochNumber int64) error {
	if identifier != EpochIdentifierMonetary {
		return nil
	}
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	amount, err := h.Keeper.MintForEpoch(sdkCtx, epochNumber-1)
	if err != nil {
		return fmt.Errorf("monetary: epoch %d issuance failed: %w", epochNumber, err)
	}
	sdkCtx.Logger().Info("monetary epoch issuance", "epoch", epochNumber, "amount", amount.String())
	return nil
}

// BeforeEpochStart does nothing for x/monetary.
func (h MonetaryEpochHook) BeforeEpochStart(_ context.Context, _ string, _ int64) error {
	return nil
}

// EpochHooksWrapper returns the depinject EpochHooksWrapper for this
// module, for use when the module gains a protobuf module config and
// moves into the depinject app config. Until then the app wires the
// hook directly (see app/app.go, which follows the simapp pattern:
// SetHooks on the epochs keeper before the epochs AppModule is built
// from that keeper).
func (h MonetaryEpochHook) EpochHooksWrapper() epochstypes.EpochHooksWrapper {
	return epochstypes.EpochHooksWrapper{EpochHooks: h}
}
