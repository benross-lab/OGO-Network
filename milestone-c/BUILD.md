# OGO Network, Milestone C: Build Instructions

Local staging only. Nothing in this document authorizes a push, a deploy, a
public testnet, or any token issuance.

## Toolchain

- Go 1.25.7, at `~/workspace/tools/go-1.25.7/bin/go` on the staging machine.
  Verify with `go version` (expect `go version go1.25.7 linux/amd64`).
- The module is `github.com/benross-lab/ogo-network`. Dependency versions are
  pinned in `chain/go.mod` and `chain/go.sum`; `chain/gaia-v28.3.1-go.mod` is
  the Gaia v28.3.1 dependency reference. Do not upgrade Cosmos SDK or any
  pinned dependency independently; the locked matrix is Cosmos SDK v0.53.6
  requirement with Gaia's replace to SDK v0.53.4 (effective SDK v0.53.4),
  CometBFT v0.38.23, wasmd v0.60.9, wasmvm/v2 v2.3.5, ibc-go/v10 v10.7.0.

## Important: TMPDIR

`/tmp` on the staging machine is a 512M tmpfs and the Go linker runs out of
space there when building the test binaries. Always set TMPDIR to a directory
under `~/workspace` before building or testing, for example:

```sh
mkdir -p ~/workspace/tmp-gotest
export TMPDIR=~/workspace/tmp-gotest
```

Delete the scratch directory when finished.

## Build, vet, test

From inside `chain/`:

```sh
export PATH=~/workspace/tools/go-1.25.7/bin:$PATH
export TMPDIR=~/workspace/tmp-gotest

go build ./...
go vet ./...
go test -count=1 ./...
```

Expected baseline (verified 2026-11-07 against the copy in this package; logs
in `test-logs/`): build exit 0, vet exit 0, full test suite exit 0 with the
`app`, `wasmcheck`, and `x/monetary` packages ok. `cmd/ogod`, `cmd/ogod/cmd`,
and `loadgen` have no test files.

Targeted runs:

```sh
# Authentication tests only (15 tests)
go test -count=1 -v ./app/ -run TestAuth

# Monetary bypass selection (17 tests)
go test -count=1 -v ./x/monetary/ -run 'TestBypass|TestGov|TestWasm'
```

## Running the localnet (development network only)

The scripts live in `chain/scripts/`. They start and stop nodes automatically
on the local machine. No public testnet is authorized; these are local
development networks only, with test accounts and test-only OGO Coins.

```sh
# Three-validator localnet
bash chain/scripts/localnet.sh

# Four-validator localnet
bash chain/scripts/localnet4.sh
```

Do not run these against the staging workspace's existing `localnet-data/`
unless you intend to reuse or reset that data; the staging `localnet-data/`
runtime files are preserved working state, not part of this package. If you
start a fresh localnet from this package copy, it creates its own data
directories; stop it cleanly and do not copy its validator keys anywhere.

The P1 observation helper is `chain/scripts/p1_observe.py`. The upgrade
procedure is `chain/scripts/upgrade-v2.sh`; see `chain/upgrade-v2/` for the
evidence from the last upgrade test run.

## What the chain is (short version)

- Entry point: `chain/cmd/ogod/` (wired CLI, Gaia-adapted root command).
- Application: `chain/app/app.go` (depinject-wired `OgoApp`), `app_config.go`,
  `modules.go`, `upgrades.go`, `export.go`, `config.go`.
- Custom module: `chain/x/monetary/` (scheduled issuance, treasury, supply
  invariants, ledger, hooks, bypass suite).
- Authentication: standard SDK ante handler wired in `app.go` (decorator
  chain: SetUpContext, ExtensionOptions, ValidateBasic, TxTimeoutHeight,
  ValidateMemo, ConsumeGasForTxSize, DeductFee, SetPubKey, ValidateSigCount,
  SigGasConsume, SigVerification, IncrementSequence). No custom signature
  scheme. Proof: `chain/app/auth_test.go`.
- Epochs: real SDK `x/epochs` integration with the hook-registration
  workaround documented in `chain/EPOCHS_GAP.md` (SDK v0.53.4 depinject
  copy-versus-pointer defect). Do not "fix" this by removing the workaround;
  the hook-fires test in `app_test.go` proves the hook actually fires.
- Deliberately excluded: `x/mint` (no discretionary minting), IBC and
  CosmWasm (deferred; not wired, not dependencies of this work).

## Boundaries

- Work in the staging workspace, not in this package. This package is a
  review snapshot.
- No push, no publish, no deploy, no public testnet, no tokens, no monetary
  parameter changes, no consensus behavior changes.
- Never copy, publish, or share `localnet-data/` or `localnet4-data/`; they
  contain validator private keys and a mnemonic.
