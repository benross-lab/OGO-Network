# OGO Network

OGO Network is an independent open-source Layer 1 blockchain with native OGO Coin: fixed maximum supply, scheduled issuance, zero reserved founder allocation, community governance. OGO Systems creates the initial technology. The OGO community ultimately governs the network.

Status: local prototype only. The chain in this repository runs on one machine as a development network. There is no mainnet, no public testnet, no token issued, no spending, and no production infrastructure. A public testnet is NOT authorized: do not present one as available. All monetary values in the prototype are TEST-ONLY and are labelled so everywhere they appear.

## What is here

- `chain/`: the Go implementation of the prototype chain. A CometBFT ABCI 2.0 application whose state machine is `x/ogocoin` (fixed maximum supply plus scheduled issuance, zero reserved founder allocation), running on three local validators that produce blocks and agree on the same ledger. Epoch issuance executes through the monetary `AfterEpochEnd` hook.
- `demo.sh`: the narrated end-to-end demonstration of the three-validator prototype.
- `poc-4validator/`: the separate four-validator fault-tolerance proof of concept, kept as reference evidence. It demonstrates liveness with one validator down (S5), plus snapshot and restore (O1). See its README and RUNBOOK.
- `docs/`: the finished network documentation. Architecture plan v1.1 and the eleven specifications (`docs/architecture/`), the monetary research (`docs/monetary/`), the deployment architecture (`docs/deployment/`), the early design-phase records (`docs/early/`), the prototype readme as of 2026-11-06 (`docs/prototype-readme-2026-11-06.md`), and the discrepancy resolutions (`docs/discrepancy-resolutions-2026-11-06.md` with recomputation scripts in `docs/discrepancy-resolutions/`).
- `logs/`: the run logs from the 2026-11-06 local prototype runs, kept as evidence.

## What is NOT here

- OGO Wallet: a separate repository, not created yet. The wallet lives in the OGO application for now and will move to its own repository later.
- OGO Explorer: a future repository, not created yet. No placeholder.
- Milestone C: the implementation milestone (real SDK modules, real x/epochs wiring, signature verification, staking, slashing, upgrades) is in progress in local staging and is NOT in this repository. It lands when its own completion verification passes.
- The whitepaper draft, the exchange readiness report, the finance provider research, and the economy work: not network implementation scope, not published here.
- No license has been selected yet. This code is published without a license pending the founder's decision. Do not assume a license.

## Build

Prerequisites: Go 1.25.7, the locked prototype toolchain. Go commands need the workspace temp and build caches because the /tmp tmpfs is only 512 MB:

```sh
cd chain
export TMPDIR=~/workspace/.go-tmp GOCACHE=~/workspace/.go-build-cache
export PATH=$HOME/workspace/tools/go-1.25.7/bin:$PATH
make build   # builds ./ogod
make vet     # go vet ./...
make test    # go test -count=1 ./... (90 tests, 0 failures)
```

Do NOT edit `chain/go.mod` or `chain/go.sum`: the dependency matrix is LOCKED (Cosmos SDK v0.53.6 required with Gaia's replace to v0.53.4, CometBFT v0.38.23, wasmd v0.60.9, wasmvm/v2 v2.3.5, ibc-go/v10 v10.7.0). No single dependency may be upgraded in isolation.

## Run: 3-validator local development network

```sh
./demo.sh
```

The demo starts 3 validators, waits for the first epoch boundary and shows the hook minting into the rewards pool, transfers 123,456 uogo from alice to bob (fee 100), verifies the same balances from all 3 nodes, and runs the halt-and-resume test. It stops the nodes when it finishes, so it is safe to re-run. Expected outcome: VERIFY OK, supply 1,000,001,900 after 2 epochs (1,900 issued), alice 599,876,444, bob 400,123,456.

Manual localnet control:

```sh
cd chain
./scripts/localnet.sh up       # build + init + start + verify
./scripts/localnet.sh status   # height and app hash per node
./scripts/localnet.sh verify   # same ledger state on all 3 nodes
./scripts/localnet.sh liveness # halt-and-resume test
./scripts/localnet.sh stop     # stop all 3 nodes
./scripts/localnet.sh clean    # stop and delete localnet data
```

The localnet uses TEST-ONLY parameters: max supply 1,000,010,000 uogo, genesis alice 600,000,000 and bob 400,000,000, schedule reducing-per-period (4 periods, initial 1000, reduction 9/10, total scheduled 3439), epoch duration 20 seconds. The production max supply, schedule, and epoch duration are all TBD and have no defaults anywhere.

## Run: 4-validator fault-tolerance proof of concept

```sh
cd poc-4validator
./run-poc.sh
```

Four validators, chain-id `ogo-poc-4val-1`. Expected outcome: S5 liveness demonstrated (blocks continue with val4 down, val4 restarts and catches up), VERIFY OK with identical app hash on all four nodes. See `poc-4validator/README.md`, `poc-4validator/RUNBOOK.md`, and `poc-4validator/ogo-network-poc-4validator-report.md`.

## Validator counts, kept distinct

- 3 validators: the decided local prototype in this repository.
- 4 validators: the minimum operational set (smallest set tolerating one faulty validator) and the separate fault-tolerance proof of concept.
- 50 or more validators: the proposed decentralization target for a public network, NOT an approved launch parameter.

With 3 validators the network tolerates zero faulty validators: stopping one leaves 2 of 3 (exactly two thirds, not more), so there is no quorum and block production halts. That halt is expected BFT behaviour for this count, not a failure.

## What works, what does not (honest status)

Implemented and executed on the live 3-validator localnet: block production and consensus on the same ledger (same balances, same supply, same app hash at the same height on all 3 nodes); epoch execution through the monetary hook (live: 2 epochs, 1900 uogo issued); transfers reconciled on all 3 nodes; halt-and-resume after a validator stop; full restart with no replay and no double mint.

Not implemented: real SDK x/epochs wiring (the epoch timer is a mirror, see `chain/EPOCHS_GAP.md`); signature-verified transactions (JSON transactions, not signature-checked); the wasm VM; live IBC relaying; staking, governance, and distribution keepers (pinned and compiled only); x/treasury and x/rewards (planned, see `chain/MODULE_BOUNDARIES.md`); gas metering (flat fee per transaction). Framework selection is NOT closed: Cosmos SDK plus CometBFT is a candidate, pending a genuine Cosmovisor upgrade and state migration (M2), the 24-hour continuous run (P1), and the remaining S5 and O1 work.

## Provenance and publication notes

- The prototype was built in local staging and committed there as `b29f830` ("OGO Network local prototype, Phase 2 Milestones 1-5", 2026-11-06). That commit is preserved as the first commit of this repository; the staging history is retained locally.
- This repository was created 2026-11-06 as public `benross-lab/OGO-Network` by Benross's explicit authorization ("Ok create the github.. OGO Network", "And push it all", scope confirmed as OGO-Network only). Only the authorized scope was pushed: the prototype chain, the four-validator PoC as reference, and the finished network documentation.
- The prototype and PoC reports in this repository are records of the 2026-11-06 local runs. Their "local staging only" status lines describe those runs, not the current publication state.
- Two discrepancies flagged in the first monthly exchange readiness report were resolved by recomputation before publication: hybrid dust is 22 uogo (not 18), and the low-fee R-F binding delta is 0 units (not 4). See `docs/discrepancy-resolutions-2026-11-06.md`. The exchange readiness report itself is a published first edition and was not silently rewritten; its next monthly edition should pick up the resolutions.
- One pre-existing em dash in a code comment in `chain/x/ogocoin/bypass.go` (and the PoC copy) was replaced with a colon in the published copy to satisfy the OGO no-em-dash rule. The change is comment-only and does not affect behaviour; the staging prototype is unchanged.
- No em dashes are used in any file in this repository, per the OGO copy rule.
