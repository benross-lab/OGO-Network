# OGO Network development acceptance matrix

**Branch:** `manus/dev-network-api`  
**Repository HEAD before changes:** `80a80b513c1d381b92e577136be4d79624b9451a`  
**Scope:** isolated development work only. No public network, production deployment, token issuance, genesis change, or monetary-parameter change.

Statuses are intentionally limited to **PASS**, **FAIL**, or **NOT TESTED**. “PASS — existing evidence” means the repository already contains the supporting log or report; it is not a claim that this isolated sandbox reproduced the result.

| Criterion | Status | Evidence / reason |
| --- | --- | --- |
| Milestone C clean build | PASS | Fresh run with repository-pinned Go 1.25.7: `go build ./...` exit 0. |
| Milestone C `go vet` | PASS | Fresh run with repository-pinned Go 1.25.7: `go vet ./...` exit 0. |
| Milestone C complete test suite | PASS | Fresh run with repository-pinned Go 1.25.7: `go test -count=1 ./...` exit 0; app, wasmcheck, and monetary packages passed. |
| Transaction authentication tests | PASS — existing evidence | `milestone-c/chain/app/auth_test.go` and the handoff record fifteen authentication tests passing. |
| Monetary bypass suite | PASS — existing evidence | Handoff record lists seventeen bypass/governance/wasm tests passing. |
| P1: 24-hour four-validator observation | NOT TESTED | The repository explicitly records only short observation windows; a complete 24-hour run remains open. No new 24-hour run was started in the ephemeral sandbox. |
| S5: four-validator liveness and recovery | NOT TESTED | Existing PoC records liveness and restart/rejoin evidence, but the full criterion including jailing, slashing, and unjailed rejoin remains open. No new four-host run was possible here. |
| M2: Cosmovisor upgrade with store migration | NOT TESTED | Existing acceptance evidence says no genuine Cosmovisor upgrade was executed. This work did not alter the chain or create an upgrade. |
| M3: monetary bypass suite after upgrade | NOT TESTED | Depends on M2; no upgrade was executed. |
| M5: failed upgrade rollback and rejoin | NOT TESTED | Requires a real coordinated upgrade environment; not run. |
| O1: independent validators, sentry, runbook execution | NOT TESTED | Existing PoC has partial single-host evidence. Independent hosts and sentry topology are not testable in this sandbox. |
| O2: 24-hour monitoring and alerting | NOT TESTED | No persistent four-validator environment or alert sink is available here. |
| O5: delete/restore validator from snapshot and catch up | NOT TESTED | Existing PoC contains partial snapshot/restore evidence; this work did not touch its evidence or run a new validator recovery. |
| Network API origin unit/integration tests | PASS | `network-api/main_test.go`: fake upstream tests verify real-state proxying, explicit no-simulation behavior, client-signed broadcast, and idempotency. Command: `/usr/local/go/bin/go test ./...` from `network-api/`. |
| Four-validator localnet agreement smoke test | PASS | Isolated `localnet4-smoke` run reached height 10/11 with the same app hash on all four nodes; `localnet4.sh verify` completed successfully. Data was outside the repository and stopped after the test. |
| Network API against real localnet RPC | PASS | API origin queried the isolated localnet at `127.0.0.1:27657`; `/healthz`, `/v1/network`, and `/v1/validators` returned live `ogo-local-1` state with four validators. |
| Cloudflare Worker syntax | PASS | `node --check cloudflare/ogo-network-api/src/index.js` passed. No deployment was made. |
| Live LCD balance queries | NOT TESTED | No live LCD endpoint is running in this sandbox. The API returns explicit 501 rather than inventing balances. |
| Live transaction indexer/history | NOT TESTED | No indexer exists in the repository or sandbox. The API returns explicit 501 rather than approximating history. |
| Live wallet integration in OGO Systems | NOT TESTED | OGO Systems application source is not present in this workspace. A framework-neutral TypeScript client contract was added, but it is not an app integration. |
| Public RPC/API exposure | NOT TESTED | Explicitly prohibited without approval and external hosting. |

## Interpretation

The isolated API work is testable and passing. The multi-validator operational gates are not failures of the API implementation; they require persistent hosts, independent failure domains, a real upgrade path, monitoring, and backup infrastructure. Existing prototype and four-validator PoC evidence remains untouched.
