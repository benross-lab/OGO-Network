# OGO Network, Milestone C: Handoff for GPT GitHub Review

## Purpose

This package gives GPT complete read access to the current Milestone C
implementation of OGO Network, so it can review the code and continue
local-staging development without losing existing work. It was assembled
2026-11-07 (America/New_York) from the local staging workspace at the exact
state recorded below. Nothing in this package was pushed, published, deployed,
or sent anywhere. It is a local file handoff only.

## Source workspace and exact state

- Source workspace:
  `~/workspace/goals/ogo-ai-builder-and-network/files/ogo-network-milestone-c/`
- HEAD commit (full SHA): `a5613fc6eb3aef474e6be3ee54be4f8aaba659d9`
- Commit message: `Milestone C: transaction authentication tests (CheckTx + FinalizeBlock)`
- Branch: `main`
- Remote: none configured (`git remote -v` returns empty). There is no remote
  to push to, and nothing here is authorized to be pushed.
- Recent history:
  - `a5613fc` Milestone C: transaction authentication tests (CheckTx + FinalizeBlock)
  - `b988fb7` C2/C7: remove epoch timer mirror; S7 wasm VM check; fix upgrade verify script
  - `aef4d05` C6 fix 2: monetary InitGenesis migration guard + init version map merge
  - `c5efd84` C6 fix: register upgrade handlers only in the v2 binary
  - `2155137` Add wired ogod CLI (Gaia-adapted root command, autocli) and fix .gitignore anchoring
- Git status at assembly time: 85 entries, all preserved and none part of this
  package's source. 69 tracked modifications, all under
  `chain/localnet-data/` (runtime validator data from the running localnet).
  16 untracked entries: `c7-evidence-notes.md`, `chain/loadgen-calib-*`,
  `chain/loadgen/`, `chain/localnet4-data/`, `chain/p1-24h.out`,
  `chain/upgrade-v2/`, `cleanbuild/`, `cosmos-sdk-cosmovisor-src/`,
  `o1-runbook.md`, `sdk-0.55-dry-run-estimate.md`. Of those, the
  `c7-evidence-notes.md`, `chain/loadgen-calib-*`, `chain/loadgen/`, and
  `chain/upgrade-v2/` items are included in this package; the rest are listed
  under Excluded below.

## What Milestone C is

Milestone C is the implementation milestone for the OGO Network Layer 1
blockchain: a real Cosmos SDK application with depinject wiring (see
`chain/app/app.go` and `chain/app/app_config.go`). Standard SDK modules are
wired: auth, bank, staking, slashing, distribution, gov, upgrade, epochs,
consensus, evidence, feegrant, authz, genutil, vesting, circuit. The custom
module is `x/monetary` (scheduled issuance, treasury, supply invariants),
registered through the documented hybrid hook. `x/mint` is deliberately NOT
included: OGO has no discretionary minting. IBC and CosmWasm remain deferred;
they are not wired and are not dependencies of this work.

The most recent work is transaction authentication (commit `a5613fc`): 15 new
tests in `chain/app/auth_test.go` proving the standard SDK ante handler
enforces signature verification, signer/public-key/account matching, chain ID
and account number validation, sequence validation and advancement, and replay
protection, on both the CheckTx and FinalizeBlock paths. No application code
was changed in that commit: the ante handler already enforced every tested
property. The full evidence is in `docs/ogo-network-milestone-c-security-note.md`.

Do not confuse this chain with the older three-validator prototype. The
prototype's `chain/app/app.go` contains an unsigned JSON transaction path
(`executeTx`) that trusts the supplied `from` field. That path does not exist
in the Milestone C chain and is not the implementation target.

## Package layout

- `README.md` (this file)
- `BUILD.md` (build, test, and localnet instructions)
- `MANIFEST.md` (SHA-256 and byte size for every file, plus the exclusion list)
- `chain/` (copy of the Milestone C chain source, pruned as described below)
- `ref/` (reference copies of simapp and Gaia wiring files used during the build)
- `docs/` (architecture specs, monetary design, Milestone C plan, deployment
  architecture, acceptance matrix, blockers, and the security note)
- `test-logs/` (fresh build/test logs generated from the `chain/` copy in this
  package; each log records its exact command and exit code)
- `c7-evidence-notes.md` (untracked workspace note that predates the
  authentication work)
- `milestone-c-security-note-pointer.md` (pointer from the staging repo root to
  the security note; the note itself is in `docs/`)

## What is in chain/

A copy of `ogo-network-milestone-c/chain/`, including:

- `app/` (`app.go`, `app_config.go`, `modules.go`, `upgrades.go`, `export.go`,
  `config.go`, `app_test.go`, `auth_test.go`)
- `x/monetary/` (keeper, ledger, models, schedule, treasury, genesis, hooks,
  bypass, IBC stub, module, and all tests)
- `cmd/ogod/` (wired CLI)
- `scripts/` (`localnet.sh`, `localnet4.sh`, `upgrade-v2.sh`, `p1_observe.py`)
- `Makefile`, `README.md`, `DEPENDENCIES.md`, `MODULE_BOUNDARIES.md`,
  `EPOCHS_GAP.md`, `BATCHING_RESULTS.md`, `.gitignore`
- `go.mod`, `go.sum`, `gaia-v28.3.1-go.mod` (dependency lockfiles)
- `batchdemo/`, `loadgen/` (load generator source), `wasmcheck/`
- `upgrade-v2/` (upgrade-test evidence: proposal, pre/post state, height)
- `p1-24h.log`, `p1-24h.out` (P1 observation run artifacts)
- `loadgen-calib-*` and `loadgen-p2-*` (untracked calibration files; see
  manifest note)

Verified after copying: `diff -r` of the package `chain/` against the staging
`chain/`, excluding only the pruned paths below, reports no differences.

## What is in docs/

- `ogo-network-milestone-c-security-note.md` (the chain security note)
- `ogo-network-native-coin-architecture-implementation-plan-v1.1.md` (v1.1 is
  the current plan; v1.0 is kept intact in staging as the reviewed version and
  is NOT in this package)
- All 18 files from `ogo-network-architecture/`: the 11 specs
  (`OGO_NETWORK_ARCHITECTURE`, `OGO_CONSENSUS_AND_EPOCH_SPEC`,
  `OGO_MONETARY_ENGINE_SPEC`, `OGO_VALIDATOR_AND_STAKING_SPEC`,
  `OGO_TREASURY_AND_SUSTAINABILITY`, `OGO_NATIVE_WALLET_SPEC`,
  `OGO_FEE_MARKET_SPEC`, `OGO_NETWORK_API_SPEC`, `OGO_NETWORK_SECURITY_MODEL`,
  `OGO_NETWORK_IMPLEMENTATION_ROADMAP`, `OGO_NETWORK_DECISION_REGISTER`),
  plus `OGO_NETWORK_RECONCILIATION_REPORT.md`,
  `OGO_NETWORK_ACCEPTANCE_CRITERIA.md` (the acceptance matrix),
  `OGO_NETWORK_RECONCILIATION_LOG.md`, `OGO_NETWORK_PROTOTYPE_SPEC.md`,
  `OGO_NETWORK_PROTOTYPE_TASKS.md`,
  `OGO_NETWORK_PROTOTYPE_TEST_REQUIREMENTS.md`,
  `OGO_NETWORK_PROTOTYPE_BLOCKERS.md` (blockers)
- Monetary design: `ogo-network-monetary-constitution-v0.1.md`,
  `ogo-network-issuance-mathematics-v0.1.md`,
  `ogo-network-geometric-vs-hybrid-deep-dive-v0.1.md`,
  `ogo-network-monetary-design-decision-report-v0.1.md`,
  `ogo-network-hybrid-terminal-and-treasury-v0.1.md`,
  `ogo-network-treasury-sustainability-v0.2.md`,
  `ogo-network-monetary-architecture-decision-preparation-v0.1.md`,
  `ogo-network-discrepancy-resolutions-2026-11-06.md`
- `ogo-network-milestone-c-plan-v0.1.md`,
  `ogo-network-deployment-architecture-v0.1.md` and its diagram SVG

## Fresh test results (generated for this package)

All commands were run inside the `chain/` copy in this package with Go 1.25.7
(`go version go1.25.7 linux/amd64`), `TMPDIR` set to a directory under
`~/workspace` (see BUILD.md for why `/tmp` must not be used). Full output is
in `test-logs/`; each log records its command and exit code.

| Log | Command | Exit | Result |
| --- | --- | --- | --- |
| `01-build.log` | `go build ./...` | 0 | builds clean |
| `02-vet.log` | `go vet ./...` | 0 | vet clean |
| `03-test-all.log` | `go test -count=1 ./...` | 0 | app ok (12.893s), wasmcheck ok, x/monetary ok; cmd/ogod, cmd/ogod/cmd, loadgen have no test files |
| `04-auth-tests.log` | `go test -count=1 -v ./app/ -run TestAuth` | 0 | 15 passed, 0 failed |
| `05-bypass-tests.log` | `go test -count=1 -v ./x/monetary/ -run 'TestBypass\|TestGov\|TestWasm'` | 0 | 17 passed, 0 failed |

The 15 authentication tests: `TestAuthForgedSenderRejected`,
`TestAuthUnsignedRejected`, `TestAuthInvalidSignatureRejected`,
`TestAuthPubkeyMismatchRejected`, `TestAuthWrongChainIDRejected`,
`TestAuthWrongAccountNumberRejected`, `TestAuthReplayedSequenceRejected`,
`TestAuthDuplicateTransactionRejected`, `TestAuthValidTransactionAccepted`,
`TestAuthInsufficientFundsRejected`, `TestAuthInvalidFeeRejected`,
`TestAuthMalformedTransactionRejected`, `TestAuthZeroFeeBehaviorDocumented`,
`TestAuthPathsAgree`, `TestAuthRejectedTxCannotAlterSupply`.

Known semantics the tests pin down (see the security note for detail):
CheckTx never executes messages, so an insufficient-funds transaction is
admitted by CheckTx and rejected at block execution (code 5). A transaction
that passes the ante handler but fails message execution still pays its fee
(100 uogo to the fee collector) and consumes its sequence, while supply and
total issued stay unchanged. Zero-fee transactions are accepted on both paths
because no minimum gas price is set in app code; operators set
`--minimum-gas-prices`.

## Boundaries (read before doing anything)

- Local staging only. Do not push, publish, deploy, create a public testnet,
  or modify the public OGO-Network repository.
- Do not modify the three-validator prototype
  (`~/workspace/goals/ogo-ai-builder-and-network/files/ogo-network-prototype/`,
  commit `b29f830`) or the four-validator PoC
  (`~/workspace/goals/ogo-ai-builder-and-network/files/ogo-network-poc-4validator/`).
  They are referenced here by path and commit only; their trees are not in
  this package.
- Do not issue or distribute tokens. Do not change monetary parameters, supply
  rules, or consensus behavior. Do not select a max supply or any monetary
  parameter; those remain separate decisions for Benross.
- Do not overwrite the public repo's `main` branch. This package is not a
  branch, a PR, or a push. It is a read-only handoff.
- The wallet implementation is a separate track under `~/workspace/ogo`. It is
  not in this package; do not mix its scope into the chain work.
- No independent security audit has occurred. Implementation tests are not an
  independent audit, and the security note does not claim otherwise.

## What is NOT included and why

See MANIFEST.md for the full exclusion list. The short version:

- `chain/localnet-data/` and `chain/localnet4-data/`: runtime validator data.
  `localnet4-data/` contains a loadmaster mnemonic file and validator private
  keys. Never copy, publish, or share these. They were preserved in place in
  staging and are not in this package.
- The `ogod` and `batchdemo-bin` binaries and `chain/build/` outputs: build
  artifacts, reproducible from source with BUILD.md.
- Plan v1.0: kept intact in staging as the reviewed version; v1.1 is the
  current plan and is the one included.
- Exchange readiness report and companions, whitepaper draft, finance
  research, economy files, wallet contract: separate tracks, not Milestone C
  chain scope.
- Staging-root extras `cleanbuild/`, `cosmos-sdk-cosmovisor-src/`,
  `o1-runbook.md`, `sdk-0.55-dry-run-estimate.md`: adjacent work products not
  in the requested handoff scope; they remain in the staging workspace and can
  be added on request.
- The prototype and PoC trees: referenced by path and commit only, per the
  boundary above.

## How GPT should use this package

1. Read this README, then `docs/OGO_NETWORK_ARCHITECTURE.md` and
   `docs/ogo-network-milestone-c-plan-v0.1.md` for the design and the plan.
2. Read `docs/ogo-network-milestone-c-security-note.md` for what is verified,
   what is implemented but not adequately tested, what is designed but not
   implemented, and the open risks.
3. Read `chain/app/app.go`, `chain/app/app_config.go`, and
   `chain/x/monetary/` for the implementation; `chain/app/auth_test.go` for
   the authentication proof.
4. Reproduce the test results with BUILD.md before changing anything, and
   record the baseline.
5. Work in the staging workspace
   (`~/workspace/goals/ogo-ai-builder-and-network/files/ogo-network-milestone-c/`),
   not in this package. This package is a snapshot for review; edits belong in
   staging, under the same local-staging-only boundaries.
6. If a review finding needs the excluded material (for example the O1 runbook
   or the SDK 0.55 dry-run estimate), ask for it by name; it is still in
   staging. Never ask for the localnet validator data or the mnemonic.
