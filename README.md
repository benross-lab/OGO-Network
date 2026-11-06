# OGO Network Local Prototype (Milestones 1-2)

Status: local prototype, Milestones 1 (repository initialization)
and 2 (consensus and epoch execution). Local only. No deployment,
no token issuance, no testnet, no mainnet, no production change, no
spending. All monetary values are TEST-ONLY and are labelled so
everywhere they appear.

## What the prototype is

A runnable local OGO Network chain that demonstrates the
architecture: a CometBFT ABCI 2.0 application whose state machine is
`x/ogocoin` (fixed maximum supply plus scheduled issuance, zero
reserved founder allocation), running on THREE local validators that
produce blocks and agree on the same ledger. Epoch issuance executes
through the monetary `AfterEpochEnd` hook, proven by a hook-fires
test in the wired application, not by a compile check.

The network operates independently of the OGO application. The
application is a client of the network, not its authority: nothing
in this prototype reads from or writes to Supabase, the OGO
Cloudflare Worker, or benross-lab/OGO. There is no code path from
the chain to the application at all.

## Validator count: three for the prototype, larger targets for later

Benross decided the authorized local prototype uses THREE
validators. This is the execution requirement for the local
prototype only, and it is distinct from the larger validator
targets proposed for later networks:

- **3 validators, local prototype (this repo).** The decided count
  for demonstrating consensus, epoch execution, and transfers
  locally.
- **4 validators, minimum operational set.** BFT consensus needs
  3f+1 validators to tolerate f faulty validators, so 4 is the
  smallest set that tolerates one faulty validator and can
  demonstrate liveness with one validator down (the S5 test:
  take 1 of 4 offline, the remaining 3 keep producing blocks).
  That is a target for a later multi-validator testnet, not this
  prototype.
- **50 or more validators, decentralization target.** The
  sufficiently-decentralized target for a public network
  (Nakamoto coefficient at least 10, top 10 under 33 percent).
  Also a later-network target, not this prototype.

With 3 validators the network tolerates zero faulty validators:
stopping one leaves 2 of 3 (exactly two thirds, not more), so there
is no quorum and block production halts. That halt is expected BFT
behaviour for this count, not a failure. The localnet's liveness
action therefore tests halt-and-resume (stop one validator, confirm
the halt, restart it, confirm blocks resume and nodes agree).

## Scope confirmations

- Keplr and live IBC are NOT prerequisites for this prototype.
  Wallet integration (Keplr, Leap) and live IBC relaying are later
  milestones; the prototype demonstrates consensus, epochs, and
  transfers without them.
- The first monetary prototype is independent of smart contracts.
  CosmWasm is evaluated separately before adoption; the wasmd
  module is pinned and compiled but the VM is not instantiated.
- The epoch timer is a mirror, not the SDK module. See
  `chain/EPOCHS_GAP.md`: `x/ogocoin/epochs.go` reproduces the pinned
  SDK x/epochs BeginBlocker semantics exactly, because the scaffold
  is a custom ABCI app and real x/epochs needs the full depinject
  SDK app (Milestone C). Do not present the mirror as SDK x/epochs
  integration.

## Layout

- `chain/`: the Go module: `app/` (ABCI app), `cmd/ogod` (CLI),
  `x/ogocoin/` (monetary module plus the epoch timer mirror),
  `scripts/localnet.sh` (3-validator localnet), `Makefile`,
  `MODULE_BOUNDARIES.md`, `EPOCHS_GAP.md`, `README.md`.
- `demo.sh`: the narrated end-to-end demonstration (this file's
  next section).

## Build

Prerequisites: Go 1.25.7 at `~/workspace/tools/go-1.25.7/bin/go`
(the locked prototype toolchain). Go commands need the workspace
temp and build caches because the /tmp tmpfs is only 512 MB:

```sh
cd chain
export TMPDIR=~/workspace/.go-tmp GOCACHE=~/workspace/.go-build-cache
export PATH=$HOME/workspace/tools/go-1.25.7/bin:$PATH
make build   # builds ./ogod
make vet     # go vet ./...
make test    # go test -count=1 ./... (90 tests, 0 failures)
```

Do NOT edit `chain/go.mod` or `chain/go.sum`: the dependency matrix
is LOCKED (Cosmos SDK v0.53.6 required with Gaia's replace to
v0.53.4, CometBFT v0.38.23, wasmd v0.60.9, wasmvm/v2 v2.3.5,
ibc-go/v10 v10.7.0). No single dependency may be upgraded in
isolation.

## Run: 3-validator localnet

```sh
cd chain
./scripts/localnet.sh up       # build + init + start + verify
./scripts/localnet.sh status   # height and app hash per node
./scripts/localnet.sh verify   # same ledger state on all 3 nodes
./scripts/localnet.sh liveness # halt-and-resume test
./scripts/localnet.sh stop     # stop all 3 nodes
./scripts/localnet.sh clean    # stop and delete localnet data
```

The localnet uses TEST-ONLY parameters: max supply 1,000,010,000
uogo, genesis alice 600,000,000 and bob 400,000,000, schedule
reducing-per-period (4 periods, initial 1000, reduction 9/10, total
scheduled 3439), epoch duration 20 seconds. The production max
supply, schedule, and epoch duration are all TBD and have no
defaults anywhere.

Manual equivalent (same values as the script):

```sh
./ogod init --home ./localnet-data/val1 --max-supply 1000010000 \
  --schedule-start 1 --schedule-period-length 100 --schedule-periods 4 \
  --schedule-initial-per-period 1000 --schedule-reduction 9/10 \
  --epoch-duration-seconds 20 --epoch-identifier monetary
# ... init val2 and val3 the same way, add keys and genesis accounts
# to val1, set the shared validator set, copy genesis to val2/val3 ...
./ogod start --home ./localnet-data/val1 --moniker val1 \
  --rpc-laddr tcp://127.0.0.1:26657 --p2p-laddr tcp://127.0.0.1:26656 \
  --persistent-peers <id2@127.0.0.1:26666,id3@127.0.0.1:26676>
```

## Demo

```sh
./demo.sh
```

The demo starts 3 validators, waits for the first epoch boundary
and shows the hook minting into the rewards pool, transfers
123,456 uogo from alice to bob (fee 100, broadcast time measured),
verifies the same balances from all 3 nodes, runs the
halt-and-resume test, and runs the batching demo binary
(`chain/batchdemo`) if it is present (another agent builds it; the
demo only calls it). It stops the nodes when it finishes, so it is
safe to re-run.

## Test

```sh
cd chain
make test
```

90 tests, 0 failures: 54 Milestone B tests (genesis, schedule,
ledger, IBC, bypass), 13 new Milestone 2 epoch tests (8 timer and
mint tests in `x/ogocoin/epochs_test.go`, 5 wired-app tests in
`app/app_epochs_test.go`, including the hook-fires test), and 23
tests from parallel Milestone 3/4/5 work by other agents in the
same tree (batching, issuance models, monetary engine, treasury).
The Milestone 2 claim is the 13 epoch tests plus the live localnet
evidence below; the other 23 are not this milestone's work.

The critical test is `TestEpochHookFiresInWiredApp`: it advances
the wired application past epoch boundaries and asserts issuance
minted (rewards pool 1000 after epoch 1, 1900 after epoch 2) and
epoch events emitted. It exists because of the depinject
hook-registration bug (SDK issues #24770, #25425): a compile check
is not sufficient. `TestEpochHaltCatchUpInWiredApp` proves missed
epochs tick one per block in order with no lump mint and no double
mint. `TestAppRestartRestoresHeightAndEpochs` proves a restart does
not replay blocks or double-mint.

## What works, what does not (honest status)

Implemented and executed successfully on the live 3-validator
localnet: block production and consensus on the same ledger (same
balances, same supply, same app hash at the same height on all 3
nodes); epoch execution through the monetary hook (live: 2 epochs,
1900 uogo issued, supply 1,000,001,900); transfers reconciled on
all 3 nodes (alice 599,876,444, bob 400,123,456 after a 123,456
transfer with a 100 fee); halt-and-resume after a validator stop;
full restart with no replay and no double mint.

Tested under failure: duplicate epoch claim fails with
`ErrDoubleIssuance` and changes nothing; halt catch-up ticks one
epoch per block in order; a failed hook does not halt the chain and
is recorded; restart restores height and epoch state.

Not implemented: real SDK x/epochs wiring (mirror instead, see
`chain/EPOCHS_GAP.md`); the hybrid issuance model through the app's
epoch path (the hybrid and tapered-hybrid models now exist as code
in `x/ogocoin/models.go`, parallel Milestone 3 work by another
agent, selectable through the genesis `issuance_model` field, but
the app's epoch hook still mints the legacy geometric schedule
only, and the `init` CLI has no flags to express a hybrid genesis,
so a hybrid schedule cannot be produced or executed by this
prototype yet; the seam is named in `chain/EPOCHS_GAP.md`);
signature-verified transactions (JSON transactions, not
signature-checked); the wasm VM; live IBC relaying; staking,
governance, and distribution keepers (pinned-compiled only);
x/treasury and x/rewards (planned, see
`chain/MODULE_BOUNDARIES.md`); gas metering (flat fee per
transaction; CheckTx reports a hardcoded GasWanted of 1000).

## Known limitations

- Epoch state is not covered by the app hash in the mirror (the
  hash covers the ledger only). In the real x/epochs module the
  epoch info is store state and is covered. See `EPOCHS_GAP.md`.
- A node killed mid-block (after the app persists state but before
  CometBFT marks the block committed) can restart with its app
  height one ahead of its block store. Clean SIGTERM stops do not
  hit this; it is a crash-consistency window, not a routine path.
- Keys under `<home>/keys` are plaintext, LOCALNET ONLY. Never use
  for real value.
