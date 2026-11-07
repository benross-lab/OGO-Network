# OGO Network Milestone B: Dependency Matrix (verified in build)

Status: Pinned in `go.mod` / `go.sum` in this directory. Entire matrix
treated as one tested unit (Gaia v28.3.1 reference).

## Pinned versions

| Component | Version | Source |
|---|---|---|
| Go toolchain | 1.25.7 | `go version` in build report; tarball SHA256 `12e6d6a191091ae27dc31f6efc630e3a3b8ba409baf3573d955b196fdf086005` (dl.google.com) |
| Go minimum (documented only) | 1.23 | Original wasmd minimum. NOT a verified build target. |
| Cosmos SDK | v0.53.6 required, v0.53.4 effective via replace | Gaia v28.3.1 go.mod (sha256 `72b985a58a88efea8193e2bf77be47b51f227da0055427cc57dd21877c3db872`) |
| CometBFT | v0.38.23 | Gaia v28.3.1 go.mod |
| wasmd | v0.60.9 | Gaia v28.3.1 go.mod |
| wasmvm/v2 | v2.3.5 | Gaia v28.3.1 go.mod |
| ibc-go/v10 | v10.7.0 | Gaia v28.3.1 go.mod |
| ibc-go 08-wasm | v10.5.0 | Gaia v28.3.1 go.mod |
| cometbft-db | v1.0.4 | Gaia v28.3.1 go.mod |
| cosmos-db | v1.1.3 | Gaia v28.3.1 go.mod |
| IAVL | v1.2.6 | Gaia v28.3.1 go.mod |
| cosmos-proto | v1.0.0-beta.5 | Gaia v28.3.1 go.mod |
| Cosmovisor | v1.7.3 | Cosmos Docs (separate binary, not a Go dep) |

## Why the SDK replace exists

Gaia v28.3.1's go.mod requires `cosmos-sdk v0.53.6` and then replaces it
with `v0.53.4`, with Gaia's own comment: "wasmd calls in sdk v0.53.6:
this is required until Gaia upgrades to v0.53.6 or higher". In other
words, wasmd v0.60.9 depends on SDK behaviour that changed in v0.53.6,
and Gaia's chain code had not yet moved to it, so Gaia pins the SDK
back to v0.53.4 while recording the v0.53.6 requirement. Copying only
the require line without the replace builds a different chain than Gaia
v28.3.1. OGO copies the whole matrix including the replace, and the
effective SDK version in the built binary is v0.53.4 (verify with
`go list -m github.com/cosmos/cosmos-sdk`).

All five Gaia replaces are preserved in `go.mod` (keyring, SDK,
jwt-go, gin, goleveldb). No individual dependency was upgraded.

## What is NOT adopted

- SDK v0.53.8 / CometBFT v0.38.25 / Go 1.26 (Gaia v29): multisig
  `--gas auto` broken in SDK v0.53.8. Evaluation only.
- 2026.1 family (SDK 0.55 / CometBFT 0.40 / ibc-go v11): evaluation only.
- CosmWasm 3.0 / wasmd 0.61: evaluation only.

## Verification performed in Milestone B

- `go list -m all` output recorded in the build report (effective versions).
- `go build ./...` and `go test ./...` run on Go 1.25.7.
- Reproducibility check: second build from a clean copy, hash compared.
