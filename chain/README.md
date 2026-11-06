# OGO Network Local Prototype (Milestones 1-2)

Local development only. Chain ID `ogo-local-1`. No real value, no public
network, no token issuance. All values are TEST-ONLY and labelled so.

## Milestones 1-2: what is new (2026-11-06)

- **Milestone 1, repository initialization.** `Makefile` (build, vet,
  test, localnet targets), `MODULE_BOUNDARIES.md` (implemented vs
  pinned-compiled vs planned modules), this README, and the prototype
  README at `../README.md`. The app builds with `go build` and passes
  `go vet`.
- **Milestone 2, consensus and epoch execution.** A 3-validator
  localnet (`scripts/localnet.sh`) that produces blocks and agrees on
  the same ledger: same balances, same supply, same app hash at the
  same height on all 3 nodes, verified through the CLI query paths.
  Three validators is the decided count for the local prototype:
  it is distinct from the minimum operational set of 4 validators
  (the smallest set that tolerates one faulty validator and can
  demonstrate liveness with one down) and from the decentralization
  target of 50 or more validators for a public network. With 3
  validators the network halts when one goes down, which is the
  expected BFT behaviour for this count; the liveness action in
  `scripts/localnet.sh` therefore tests halt-and-resume, not
  continued production with a validator down.
  Epoch mode: pass `--epoch-duration-seconds` to `init` and issuance
  runs only through the monetary `AfterEpochEnd` hook. The hook-fires
  test (`app/app_epochs_test.go`) proves the hook executes in the
  wired application.
- **Epoch timer mirror, stated plainly.** The scaffold is a custom
  ABCI app, not a full depinject SDK app, so real x/epochs cannot be
  wired into it yet. `x/ogocoin/epochs.go` is a faithful mirror of
  the pinned SDK x/epochs BeginBlocker semantics (fixed grid anchored
  at genesis, 1-indexed epochs, one tick per block, hook interface,
  events), read from the pinned SDK v0.53.4 source. It is NOT the SDK
  module. `EPOCHS_GAP.md` records exactly what is mirrored, what is
  not, and the StoreUpgrade path for wiring real x/epochs in the
  full SDK app (Milestone C).
- **Restart safety fix.** Height and epoch state now persist in
  `data/ogocoin/app.json`, so a restart resumes at the right height
  instead of reporting height 0 and replaying blocks (which would
  re-apply transactions). Covered by
  `TestAppRestartRestoresHeightAndEpochs`.
- **New CLI.** `init --epoch-duration-seconds`,
  `set-genesis-validators` (build one shared genesis for the 3
  validators), `show-node-id`, `show-validator`, `start` with
  `--rpc-laddr`, `--p2p-laddr`, `--moniker`, `--persistent-peers`,
  and `query epochs`.
- **Full run:** `make localnet` (or `../demo.sh` for the narrated
  end-to-end demo). `make test` runs 90 tests with 0 failures: 54
  Milestone B tests, 13 new Milestone 2 epoch tests (8 timer and
  mint tests in `x/ogocoin`, 5 wired-app tests in `app`, including
  the hook-fires test), and 23 tests from parallel Milestone 3/4/5
  work by other agents in the same tree (batching, issuance models,
  monetary engine, treasury), which this milestone does not claim.
  Keplr and live IBC are not prerequisites for this prototype, and
  the monetary prototype is independent of smart contracts;
  CosmWasm is evaluated separately before adoption.

## What is real in this scaffold (Milestone B, unchanged)

- `x/ogocoin`: fixed-supply monetary module under Monetary Constitution
  v0.1 (fixed max supply + SCHEDULED issuance, zero reserved founder
  allocation). Genesis validation (max supply mandatory, no default;
  issuance schedule mandatory and validated; founder/reserved purposes
  rejected; genesis + schedule total must fit under max), supply ledger
  with schedule-only minting (`MintScheduled`: exactly the schedule
  amount, once per period, into the rewards pool; every other mint
  fails), burn, fees, IBC escrow / voucher accounting, and
  bypass-attempt surfaces for module accounts, governance upgrades
  (including schedule changes), and CosmWasm bindings. Fully tested
  (`go test ./...`, 54 tests).
- `app`: CometBFT ABCI 2.0 application running x/ogocoin. Real blocks,
  real transfers, real scheduled issuance per block, real supply and
  schedule queries over RPC.
- `cmd/ogod`: the `ogod` binary: init, keys, genesis accounts, start,
  tx send/burn, query balance/supply/params/schedule/global-supply.


## What is NOT done (stated plainly)

- Transactions are JSON and NOT signature-verified in this scaffold.
  Anyone with RPC access can submit a tx naming any sender. Real auth
  (x/auth) arrives with full SDK depinject wiring (Milestone C).
- The wasm VM is not instantiated; live contract execution is pending.
- Live IBC relaying is not wired; escrow accounting is ledger-level only.
- Standard SDK modules (auth/bank/staking/gov/distribution) are pinned
  and compiled via blank imports in `app/modules.go` to prove the matrix
  builds, but their keepers are not yet wired. See `app/modules.go`.

## Build

Prerequisites: Go 1.25.7 (see `../REPRODUCIBLE_BUILD.md`).

```sh
export PATH=$HOME/workspace/tools/go-1.25.7/bin:$PATH
go version   # must print go1.25.7
go mod tidy
go build ./...
go test ./...
go build -o ogod ./cmd/ogod
```

## Local node (test values only)

```sh
./ogod init --home ./localnet-data --max-supply 1000010000 \
  --schedule-start 1 --schedule-period-length 5 --schedule-periods 4 \
  --schedule-initial-per-period 1000 --schedule-reduction 9/10
./ogod keys add alice --home ./localnet-data
./ogod keys add bob --home ./localnet-data
./ogod add-genesis-account <alice-addr> 600000000 --purpose public --home ./localnet-data
./ogod add-genesis-account <bob-addr> 400000000 --purpose public --home ./localnet-data
./ogod validate-genesis --home ./localnet-data
./ogod start --home ./localnet-data
# in another terminal:
./ogod tx send --from-key alice --to <bob-addr> --amount 1000 --home ./localnet-data
./ogod query supply
./ogod query schedule
./ogod query balance <bob-addr>
```

The `--max-supply` and schedule values above are TEST-ONLY for
localnet. The production max supply and issuance schedule remain TBD
and have no defaults anywhere. Note: genesis supply + total scheduled
issuance must fit under max, so a genesis totalling exactly max is
rejected when the schedule is non-empty. Genesis allocations need a
public `--purpose` label; founder/reserved purposes are rejected.
