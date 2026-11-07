# Epochs Gap: Mirror vs Real x/epochs

Status: Milestone 2 record. Read this before claiming anything about
x/epochs in this prototype.

## The one-sentence version

This prototype does NOT integrate the real SDK x/epochs module. It
runs an epoch timer mirror (`x/ogocoin/epochs.go`) that reproduces
the pinned SDK x/epochs BeginBlocker semantics exactly, wired into
the custom ABCI app. The hook-fires test proves the mirror's hook
executes in the wired application. Nothing here proves anything
about real x/epochs wiring: that proof belongs to Milestone C, when
the full depinject SDK app exists.

## Why the mirror exists

The Milestone B scaffold is a custom ABCI 2.0 application, not a
full Cosmos SDK application. There is no depinject app, no BaseApp,
no SDK stores, and no keepers. Real x/epochs is an SDK module: its
keeper needs a KVStoreService, its AppModule is provided through
depinject, and its hooks are registered through depinject's
`InvokeSetHooks`. None of that machinery exists in this scaffold.
Wiring real x/epochs here would mean building the full SDK app,
which is Milestone C work, not a Milestone 2 patch. The mirror keeps
Milestone 2 honest: epoch execution is demonstrated now, with the
exact semantics, and the gap to the real module is written down
instead of papered over.

## What was read from the pinned source

All of the following was read from the pinned SDK source in the
module cache (`cosmos-sdk@v0.53.4`, the effective version under the
locked Gaia matrix), not from memory or documentation summaries:

- `x/epochs/keeper/abci.go`: `BeginBlocker`. One epoch tick per
  block per identifier (the Walk callback ends one epoch and starts
  the next, with no loop). Tick when block time is strictly after
  the epoch end time (`CurrentEpochStartTime + Duration`). The next
  epoch starts at the prior end time, never at the ticking block
  time. The first tick starts epoch 1 when counting has not started
  (and block time is not before `StartTime`); no `AfterEpochEnd`
  fires for a non-existent epoch 0. Hook errors are swallowed: the
  hooks run in a cached context and their writes are discarded on
  error, the failure is logged, and the epoch still advances.
- `x/epochs/types/genesis.pb.go`: `EpochInfo` field comments.
  "Only one timer tick for a given identifier can occur per block."
  The halt example is explicit: with start 10 and duration 5, a
  chain offline at t=14 coming back at t=30 starts the epoch for
  (10, 15] at t=30, (15, 20] at t=31, (20, 25] at t=32. That is the
  one-per-block catch-up rule, confirmed in the pinned source.
- `x/epochs/types/hooks.go`: the `EpochHooks` interface
  (`AfterEpochEnd(ctx, identifier, epochNumber int64)`,
  `BeforeEpochStart(ctx, identifier, epochNumber int64)`),
  `MultiEpochHooks` (runs hooks in array sequence, joins errors),
  and `EpochHooksWrapper` (the depinject `OnePerModuleType` carrier).
- `x/epochs/keeper/keeper.go`: the `Keeper` struct holds `hooks`
  as a plain field; `SetHooks` panics if called twice.
- `x/epochs/keeper/hooks.go`: `Hooks()` returns an empty
  `MultiEpochHooks` when no hooks are set; `AfterEpochEnd` and
  `BeforeEpochStart` delegate to it.
- `x/epochs/module.go`: `NewAppModule(keeper keeper.Keeper)` takes
  the keeper BY VALUE. `BeginBlock` delegates to
  `am.keeper.BeginBlocker`.
- `x/epochs/depinject.go`: `ProvideModule` creates the keeper `k`,
  then `NewAppModule(k)`, and returns both as `ModuleOutputs`.
  `InvokeSetHooks(keeper *keeper.Keeper, hooks
  map[string]types.EpochHooksWrapper)` sorts hook modules by name,
  builds a `MultiEpochHooks`, and calls `keeper.SetHooks` on the
  keeper pointer depinject holds.

## The depinject bug, precisely (issues #24770, #25425)

In the pinned v0.53.4 wiring, `ProvideModule` copies the keeper
into the AppModule by value (`NewAppModule(k)`), while
`InvokeSetHooks` sets hooks on the keeper instance depinject holds
(a pointer to the stored keeper value). The AppModule's keeper copy
never receives the hooks. Its `BeginBlocker` therefore calls
`k.Hooks()`, gets an empty `MultiEpochHooks`, and no registered
hook ever fires. The chain compiles, starts, produces blocks, and
ticks epochs, and issuance silently never happens. That is why the
hook-fires test exists: only a test that advances the wired
application past an epoch boundary and observes issuance can catch
this. Issue #24770 was the earlier registration fix (in v0.53.2);
issue #25425 is the copy-versus-pointer defect. The v0.54 upgrade
guide changes `NewAppModule` to take the keeper by pointer, which
removes the defect structurally. On the 0.53 matrix the defect must
be caught by the hook-fires test, not by inspection.

## What the mirror reproduces exactly

- One timer per identifier; the monetary identifier is `monetary`.
  The monetary hook filters on it and ignores all others.
- Epoch numbers are 1-indexed: the first epoch is epoch 1.
- The grid is anchored at the genesis start time. An epoch ends at
  the first block whose block time is strictly after
  `CurrentEpochStartTime + Duration`. The next epoch starts at the
  prior end time.
- At most one epoch ticks per block. After a halt, missed epochs
  tick one per block on restart, in epoch order.
- Hooks `AfterEpochEnd` / `BeforeEpochStart` with the identifier
  and the epoch number, in registration order.
- Events `epoch_end` (ending epoch number), `epoch_start` (new
  epoch number and start time), and `epoch_issuance` (epoch number,
  amount, recipient). The SDK emits typed `EventEpochEnd` /
  `EventEpochStart` events; the mirror emits ABCI events with the
  same names and the same numbered content.
- Hook-failure behaviour: the epoch still advances. The mirror
  improves on the upstream in one way only: it records the failure
  in `LastHookError` (visible in `query epochs`) instead of only
  logging it, so a skipped issuance is visible rather than silent.

## What the mirror does NOT reproduce

- No SDK keeper, no store key, no collections schema, no
  KVStoreService. Timer state lives in the application's own
  `data/ogocoin/app.json`, not in an SDK store.
- No depinject wiring, no `ProvideModule`, no `InvokeSetHooks`, no
  `EpochHooksWrapper`. Hooks are registered by a direct method call
  in the application. The depinject bug therefore cannot occur in
  the mirror, and the mirror's hook-fires test proves nothing about
  depinject wiring. That proof must be re-run against the real
  module in Milestone C.
- No gRPC query service (`query epochs epoch-infos`). The mirror
  exposes the same information over the ABCI `epochs` query path.
- No genesis init or export through the keeper. Epoch config comes
  from the genesis app_state `epochs` key, read by the application.
- Epoch state is NOT covered by the app hash. The app hash is the
  hash of the ledger only, so a node whose epoch counter diverged
  would still report the same app hash until the next mint made the
  ledgers diverge. In the real module, `EpochInfo` is store state
  and is covered by the app hash. This is a known mirror gap.
- The mirror's first-tick rule follows the pinned code (a block
  whose time is not before `StartTime` starts epoch 1), while the
  `EpochInfo` doc comment says the first tick is the first block
  whose block time is greater than `StartTime`. The code wins; the
  difference is one block at the exact start second.
- The mirror maps epoch e to schedule period e-1 (epochs are
  1-indexed, periods are 0-indexed). The consensus and epoch
  specification uses 0-indexed periods; the mapping is stated here
  and in `x/ogocoin/epochs.go` so the two numberings are never
  confused.

## What remains to wire real x/epochs (Milestone C)

Each step must be done in the full depinject SDK app, and each must
be proven by a test, not by a compile check.

1. Build the full SDK app (BaseApp plus depinject), with stores for
   every module. Add the epochs store key.
2. Register the epochs module: keeper instantiation through
   `ProvideModule`, AppModule registration, an entry in
   `SetOrderBeginBlockers` (epochs runs in BeginBlock, before
   transactions, as the mirror does at the top of FinalizeBlock),
   and entries in `SetGenesisModuleOrder` and
   `SetExportModuleOrder`.
3. Register the x/monetary hook: x/monetary provides an
   `EpochHooksWrapper` through depinject (one per module type), so
   `InvokeSetHooks` wires it into the epochs keeper. Because the
   pinned v0.53.4 `NewAppModule` takes the keeper by value, this
   step MUST be verified by the hook-fires test below; do not trust
   inspection. The v0.54 upgrade (keeper by pointer) removes the
   hazard and should be evaluated as part of the upgrade dry run.
4. StoreUpgrade path for adding x/epochs to a chain built without
   it (this prototype's chain, or any chain in the same state):
   an upgrade handler with a `StoreUpgrades` entry that adds the
   epochs store key, plus a migration that initializes the
   `EpochInfo` from the mirror's persisted state so the grid
   continues without a reset: identifier `monetary`, duration D
   from the genesis epochs config, start time = the mirror's grid
   anchor (the genesis time), current epoch = the mirror's current
   epoch, current epoch start time and height = the mirror's
   values, counting started = true. The x/monetary issued-by-epoch
   set lives in x/monetary's own store and is untouched by this
   migration. The migration must be tested on a copy of the
   proof-of-concept chain (same genesis, same history) before it is
   used for any real schedule: run the upgrade, then verify the
   epoch number, the issued set, total issued, and total supply are
   identical before and after.
5. Re-run, against the real module in the wired app: the hook-fires
   test (advance past an epoch boundary, assert issuance minted and
   the epoch event emitted), the duplicate-claim test (claim epoch
   n, re-claim it in the same block and in a later block, both
   fail, total issued unchanged), the S4 halt and catch-up test
   (halt 30 minutes, restart, missed epochs tick one per block in
   order, no lump mint, no skipped epoch, no double mint), and the
   bypass suite.
6. Remove the mirror (`x/ogocoin/epochs.go` timer and the app's
   `app.json` epoch state) only after the real module passes all
   of the above on the proof-of-concept chain. Do not run the
   mirror and the real module at the same time: both would claim
   the same periods through the shared issued-by-epoch set, and the
   second claim would fail as a double issuance.

## The model seam (added 2026-11-06, parallel Milestone 3 work)

While this milestone was being built, the monetary-interface agent
landed the issuance models in the same tree (`x/ogocoin/models.go`,
plus `IssuanceModel` on `Params` and `BasePerPeriod` / taper fields
on `IssuanceSchedule` in `types.go`, model validation in
`genesis.go`, and `MintScheduledEpoch`, `MintScheduledModel`,
`ProcessNextEpoch`, and `CheckMonetaryInvariants` in `ledger.go`).
That work is not this milestone's, and it changes the hybrid
status: the hybrid and tapered-hybrid models now EXIST as code and
are selectable through the genesis `issuance_model` field, but they
are NOT wired into this prototype's epoch path. The app's epoch
hook still mints the legacy geometric schedule only, and the `init`
CLI has no flags to express a hybrid genesis. The remaining work,
for the follow-up that wires them, is:

- Rewire the monetary hook to the model-aware mint
  (`MintScheduledModel` with the model selected from genesis
  params). The hook receives 1-indexed SDK epoch numbers; the
  models are 0-indexed, so the call is
  `MintScheduledModel(model, epochNumber - 1)`.
- Add `init` CLI flags for the model selection and the hybrid
  parameters (`--issuance-model`, `--base-per-period`, taper
  flags), with genesis validation already present.
- Reconcile the two arithmetic definitions before the rewire:
  the legacy `AmountForPeriod` floors at every multiplication step
  (period 9 = 386, total 6508 for the test schedule), while the
  models use the exact floor(R0 * r^n) formula (period 9 = 387).
  The two give identical amounts for the demo's 4-period schedule
  but diverge from period 7 onward. Switching the hook changes
  issuance amounts, so the choice is a monetary design decision,
  not a wiring detail. Do not rewire silently.
- Unify the epoch mint entrypoints. There are now two, sharing one
  issued-by-epoch set: `MintScheduledForEpoch` (this file,
  1-indexed SDK epochs, legacy geometric amounts) and
  `MintScheduledEpoch` / `MintScheduledModel` (ledger.go,
  0-indexed epochs, model-aware amounts). They coincide at the
  ledger level today (both key the set by period index and mint the
  same amounts for a geometric genesis), so there is no runtime
  conflict: only this file's entrypoint is called by the app. In
  Milestone C there must be exactly one, and it should be the
  model-aware one, called by the real x/epochs hook.
- Correct a stale comment in `ledger.go`: it says the hook is not
  wired into the app and "hook-fires is NOT implemented in this
  prototype". That was true of the model-aware entrypoint when it
  was written, and it remains true that the model-aware entrypoint
  is not called by the app. But the mirror's hook IS wired and IS
  proven to fire (`TestEpochHookFiresInWiredApp`). The comment
  needs to say which hook it means, or a reader will conclude no
  hook fires at all. This file cannot fix it: `ledger.go` is owned
  by the monetary-interface agent.
