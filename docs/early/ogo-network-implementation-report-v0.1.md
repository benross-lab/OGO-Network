# OGO Network: Technical Implementation Report v0.1

Status: For review only. Design phase. No code written, no files changed in benross-lab/OGO, no repo created, no deploy, no publish, no spend, no token issuance, no testnet, no mainnet.
Date: 2026-11-05
Directive: OGO Network Next Engineering Directive (Benross, 2026-11-05)
Prior doc: ogo-network-blueprint-v0.1.md (same folder)

## 1. Decisions taken in this pass

1. Max supply stays TBD, and is a mandatory genesis parameter. No number was chosen to move development forward. The chain must fail genesis validation if max_supply is missing, zero, or non-positive. Monetary rules first, monetary quantity second.
2. Prototype baseline retained as a family: Cosmos SDK 0.53.x, CometBFT 0.38.x, wasmd 0.60.x. Exact pins are in section 2, taken from one coherent live reference, not mixed from different releases.
3. SDK 0.55 / CometBFT 0.40 and CosmWasm 3.0 / wasmd 0.61 stay as separate evaluation tracks. No pin, no migration.
4. All work stays inside benross-lab/OGO. No new repository. Proposed docs target `docs/network/` and `docs/blockchain/` in that repo. Drafts for review are in `ogo-network-proposed-docs/` alongside this report; nothing has been copied into `~/workspace/ogo/` yet.
5. No allocation percentages, validator economics, or governance thresholds were invented. Where the charter needs a number, the draft says TBD and names who must decide it.

## 2. Exact compatible dependency matrix: verified 2026-11-05

Rule applied: pin from one coherent build, not from each component's minimum. Mixing minima is how chains compile and then fail at runtime.

### 2a. Recommended prototype pin: Gaia v28.3.1 matrix

This is the live Cosmos Hub build. It is the only set below verified as a whole from a single go.mod.

- Go toolchain: 1.25.7 (gaia v28.3.1 go.mod: `go 1.25.7`)
- Cosmos SDK: v0.53.6 required, effectively v0.53.4 via replace (see note below)
- CometBFT: v0.38.23
- wasmd: v0.60.9
- wasmvm/v2: v2.3.5
- ibc-go/v10: v10.7.0
- ibc-go 08-wasm: v10.5.0
- ibc-apps packet-forward-middleware: v10.6.0
- ibc-apps rate-limiting: v10.1.0
- cometbft-db: v1.0.4
- cosmos-db: v1.1.3
- IAVL: v1.2.6
- cosmos-proto: v1.0.0-beta.5
- Cosmovisor: v1.7.3 (independent of SDK version)

Gaia replace block, copied verbatim in meaning:
- `github.com/99designs/keyring => github.com/cosmos/keyring v1.2.0`
- `github.com/cosmos/cosmos-sdk => github.com/cosmos/cosmos-sdk v0.53.4` with the comment "wasmd calls in sdk v0.53.6: this is required until Gaia upgrades to v0.53.6 or higher"
- `github.com/dgrijalva/jwt-go => github.com/golang-jwt/jwt/v4 v4.4.2`
- `github.com/gin-gonic/gin => github.com/gin-gonic/gin v1.8.1`
- `github.com/syndtr/goleveldb => github.com/syndtr/goleveldb v1.0.1-0.20210819022825-2ae1ddf74ef7`

That replace is the trap: the go.mod says SDK v0.53.6 in `require` and then replaces it with v0.53.4. Anyone pinning "SDK 0.53.6" without copying the replace block builds a different chain than Gaia v28.3.1. OGO must copy the whole matrix including replaces, or pin explicitly to v0.53.4 and record why, at scaffold time. Do not resolve this by editing the number in isolation.

### 2b. What wasmd 0.60.0 alone would give: do not use as the pin

wasmd v0.60.0 go.mod on its own:
- go 1.23.6, SDK v0.53.0, CometBFT v0.38.17, ibc-go/v10 v10.0.0, wasmvm/v2 v2.2.1

wasmd v0.60.9 go.mod on its own:
- go 1.23.8, toolchain go1.24.1, SDK v0.53.6, CometBFT v0.38.21, ibc-go/v10 v10.5.0, wasmvm/v2 v2.3.5

These are minima for wasmd, not a chain pin. Gaia v28.3.1 sits above them (CometBFT 0.38.23, ibc-go 10.7.0). The OGO prototype must use the Gaia set in 2a, not the wasmd 0.60.0 set.

### 2c. Go version: a real discrepancy to decide

The directive says Go 1.23. That is the wasmd 0.60.0 minimum (`go 1.23.6`) and wasmd 0.60.9 still declares `go 1.23.8`.

Gaia v28.3.1 declares `go 1.25.7`.

Recommendation: keep "Go 1.23" as the minimum language version, but build the prototype with Go 1.25.7, the Gaia toolchain. Building the Gaia matrix with Go 1.23 is not verified and should not be assumed to work. This needs Benross's acknowledgement, not a silent change: the report flags it here rather than rewriting the directive.

### 2d. Do not jump to these yet

- SDK v0.53.8 / CometBFT v0.38.25 / Go 1.26 (Gaia v29, activated 2026-10-28 at block 33,618,500): newer patch line in the same 0.53 family, but with a known regression, as of SDK v0.53.8, `--gas auto` gas simulation is broken for multisig senders; multisig must specify fixed `--gas` until fixed upstream. Evaluate separately; do not adopt for the first localnet.
- 2026.1 family: SDK 0.55.x, CometBFT 0.40.x, IBC Go v11.x.y. Evaluation track only.
- CosmWasm 3.0 / wasmd 0.61: IBCv2 support, contracts from 1.x / 2.x / 3.0 remain executable unless they use new IBCv2 features. Evaluation track only. Terra Classic's SDK 0.53 migration used wasmd 0.61.x with ibc-go v10.x, which shows the two can mix, but that is not a pin for OGO.

## 3. Monetary module: fixed supply as a mandatory genesis parameter

Module: `x/ogocoin` (proposed), on top of standard `x/bank`. It does not replace `x/bank`; it constrains it.

### 3a. Genesis parameters (exact)

| Parameter | Value | Status |
|---|---|---|
| max_supply | TBD | Mandatory, must be present and > 0 or genesis fails |
| denom | uogo | Proposed, not final |
| decimals | 6 | Proposed, not final |
| inflation_rate | 0 | Fixed |
| minting_enabled | false after genesis | Fixed |
| supply_cap_enforced | true | Fixed |

No default max_supply. There is no fallback number in code, config, or docs. A genesis file without max_supply is invalid by design.

Denom and decimals are proposals only. Changing them before genesis is a spec change; changing them after genesis is a breaking migration and must go through governance + upgrade handler, not a quiet edit.

### 3b. Genesis validation: must fail closed

On `ValidateGenesis` and on `InitGenesis`, in order:
1. Reject if max_supply is missing, zero, or negative.
2. Reject if denom is empty or does not match the declared denom.
3. Reject if any genesis allocation is in a different denom.
4. Reject if sum(all genesis allocations) > max_supply.
5. Reject if any allocation is negative or zero-amount dust that would break accounting (define the dust rule in GENESIS_SPEC.md before code).
6. Reject if minting_enabled is true after genesis, or if any genesis field enables a mint path.
7. Record max_supply in store as immutable params. No message, proposal, or upgrade handler may change it without an explicit, separately approved migration, and the default upgrade path must refuse to change it.

Initial supply may be less than max_supply. The gap is unminted headroom, not a reserve to draw on later without a governance-approved, publicly recorded action. Whether that headroom can ever be minted is TBD and must be decided before genesis: the safe default in code is that it cannot, because minting is disabled after genesis.

### 3c. Supply invariants: enforced at runtime, not just in tests

- Invariant 1: `total_supply(uogo) <= max_supply` at every block, including after every BeginBlocker / EndBlocker.
- Invariant 2: no transaction, message, contract call, IBC packet, or upgrade handler may increase total supply. `x/bank` mint is not exposed to any OGO message or CosmWasm binding.
- Invariant 3: no hidden admin mint. There is no admin address, no minter key, no privileged `MsgMint`, no backdoor in `x/ogocoin`. If a future feature needs minting, it is a new module + governance + audit, not a flag flip.
- Invariant 4: burning reduces total supply and does not create remintable credit. Burn is a one-way sink; the burned amount does not increase headroom that can be re-minted. Record burn events with amount, burner, height, and tx hash.
- Invariant 5: IBC inflows of uogo count toward total supply the same as local mints. An IBC transfer cannot push total supply above max_supply; the packet must fail if it would. IBC denom tracing must not create a second spendable representation of the same uogo that escapes the cap.
- Invariant 6: fees paid in OGO do not mint. Fee collection, distribution, and validator rewards come from existing supply (fees + an explicitly allocated reserve, if approved). No perpetual newly minted staking reward alongside zero inflation.

### 3d. Automated tests: before issuance functionality

Design tests before implementing issuance. Minimum suite, all in `chain/x/ogocoin/`:
- Genesis rejects missing / zero / negative max_supply.
- Genesis rejects sum(allocations) > max_supply, including a single allocation equal to max_supply + 1 uogo.
- Genesis accepts sum(allocations) == max_supply and sum < max_supply.
- Transfer preserves total supply exactly.
- A direct mint attempt (any path) fails and leaves supply unchanged.
- Burn reduces supply by exactly the burned amount and cannot be reversed by a second burn or a remint call.
- IBC inflow that would exceed the cap fails; IBC inflow under the cap succeeds and is counted.
- Upgrade handler that attempts to change max_supply is rejected by the invariant check; state after a failed upgrade is unchanged.
- Supply query returns the same total from `x/bank` supply, the module store, and a recomputed sum over accounts; any mismatch fails the test.
- Fuzz / property test: random sequences of transfers, burns, and IBC packets never violate `total <= max`.

### 3e. Independent verification: how anyone checks supply without trusting OGO

- `ogod query bank total --denom uogo` returns total supply; it must equal the module's recorded total and the sum of all account balances.
- Genesis file is published with its SHA256; anyone can recompute sum(allocations) and compare to max_supply.
- Every block's supply is queryable via RPC / REST / gRPC; an independent verifier can run a node, replay, and check invariants without OGO's server.
- No supply figure is published from a dashboard alone. Dashboards must cite the chain query and height they read.
- OGO Credits are never counted in this supply. Credits live in the OGO app database; OGO Coin lives on chain. A conversion rate, if ever approved, is a separate, publicly specified bridge, not an accounting merge.

## 4. Proposed changes for review: target paths in benross-lab/OGO

No files have been created in the repo yet. Drafts are in this goal's `ogo-network-proposed-docs/` folder, mirroring these target paths:

- `docs/network/FOUNDING_CHARTER.md`: charter v0.1, status Proposed, authorization restrictions intact
- `docs/network/TECHNICAL_BLUEPRINT.md`: pointer to the blueprint + this report
- `docs/network/TOKENOMICS.md`: fixed supply, 0% inflation, no allocation numbers (all TBD, explicitly listed)
- `docs/network/GOVERNANCE.md`: roles, public proposals, timelocks, no invented thresholds
- `docs/network/SECURITY.md`: audit, reproducible builds, validator concentration, multisig, recovery
- `docs/blockchain/ARCHITECTURE.md`: topology, modules, API surface, OGO app adapter boundary
- `docs/blockchain/DEPENDENCIES.md`: the exact matrix in section 2, with sources
- `docs/blockchain/GENESIS_SPEC.md`: the genesis parameter table + validation + invariants in section 3

Local check: `~/workspace/ogo/docs/` currently holds 29 files, no `network/` or `blockchain/` subfolders. These eight files would be new files, not edits to existing docs. No existing doc is modified by this proposal.

What this report does not propose: any Go code, any `chain/` scaffold, any node, any wallet, any explorer, any Supabase / Worker change, any branch, any push, any publish.

## 5. What was verified, what was not

Verified in this pass:
- Gaia v28.3.1 go.mod (fetched 2026-11-06): SDK, CometBFT, wasmd, wasmvm, ibc-go, db versions and replaces, Go 1.25.7
- wasmd v0.60.0 go.mod: SDK v0.53.0, CometBFT v0.38.17, ibc-go v10.0.0, wasmvm v2.2.1, Go 1.23.6
- wasmd v0.60.9 go.mod: SDK v0.53.6, CometBFT v0.38.21, ibc-go v10.5.0, wasmvm v2.3.5, Go 1.23.8
- Cosmos Docs Release Families: 2025.1 = SDK 0.53.x + ibc-go v10.x.y; 2026.1 = SDK 0.55.x + ibc-go v11.x.y
- Gaia v29: SDK v0.53.8, CometBFT 0.38.25, Go 1.26, multisig `--gas auto` regression in 0.53.8

Not verified, and not claimed:
- That the Gaia matrix builds on this machine (no build run in this pass; building is Milestone B, needs explicit go-ahead)
- Exact ibc-go patch for a from-scratch OGO scaffold (must be taken from `go mod tidy` against the Gaia pin at scaffold time, then frozen)
- CosmWasm 3.0 / wasmd 0.61 exact go.mod (evaluation track only)
- Any validator, fee, or governance number (all TBD by design)

## 6. Next step: needs Benross

1. Review this report + the eight draft docs.
2. Confirm the Go call: build toolchain Go 1.25.7 for the Gaia-aligned pin, with Go 1.23 as minimum only.
3. Confirm the SDK pin approach: copy the Gaia v28.3.1 matrix including its replace block, vs pin cleanly to v0.53.4 and record the reason.
4. After that, Milestone B can be scoped: localnet scaffold only, no public network, with the invariant test suite in section 3d written before any issuance code.

No release action follows from this report alone.

## Addendum: Monetary Constitution v0.1 (2026-11-05, supersedes §3 monetary rule)

The zero-inflation / mint-disabled-after-genesis rule in this report is SUPERSEDED by the Monetary Constitution v0.1: fixed max supply + scheduled issuance, zero reserved founder allocation. Issuance only exactly as a predetermined schedule dictates; no discretionary/admin/founder/governance minting; schedule not changeable by ordinary governance; issuance stops permanently at max. Max supply, issuance duration, reward schedule, treasury funding, and genesis circulating supply remain TBD. See ogo-network-monetary-constitution-v0.1.md. Milestone B was instructed to revise to the scheduled-issuance framework with schedule parameters TBD and no invented numbers. Unchanged: fixed cap, mandatory genesis max, cap enforced every block, burn one-way, IBC no double-count, bypass tests, reproducible build, Acceptance Standard v0.1.
