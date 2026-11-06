# OGO Network Architecture

Status: Architecture specification. Planning only. OGO-Network repository created 2026-11-06 (public, benross-lab/OGO-Network) by Benross's explicit authorization. No blockchain deployed, no token or smart contract deployed, no public token sale, no testnet launched, no production infrastructure changed, no money spent, no monetary parameter selected, no launch announcement authorized.
Date: 2026-11-06
Companion document: `ogo-network-native-coin-architecture-implementation-plan-v1.0.md` (the overview and index for these eleven specifications)
Architecture direction: CONFIRMED AND LOCKED, Option B. OGO Network launches with its own native blockchain and native OGO Coin. Do not reopen Base, Ethereum, or Solana token-first alternatives unless new evidence creates a compelling reason.

## 1. Purpose

This document describes the architecture of OGO Network: what the components are, who is responsible for what, where the security boundaries sit, what the repositories look like, and what development stack is recommended. It is the architecture companion to the implementation plan. The plan states the decisions and the backlog; this document states the shape of the system.

## 2. System overview

OGO Network is an independent, open-source Layer 1 blockchain. Its native asset is OGO Coin. It is the monetary and settlement layer for the wider OGO ecosystem: AI agents, developers, applications, digital services, marketplace transactions, and eventually telecommunications products. The OGO application is a client of the network, not its owner. The network must operate independently of OGO Systems, and must be usable without an OGO account, through independent wallets, exchanges, developers, and third-party services.

```
                          +-------------------+
                          |   OGO Application |
                          |  (benross-lab/OGO)|
                          |  web, Worker,     |
                          |  Supabase         |
                          +--------+----------+
                                   | public RPC / LCD / indexer API only
                                   | (never a private chain connection)
              +--------------------+--------------------+
              |                                         |
   +----------v-----------+                  +----------v-----------+
   |  External wallets    |                  |  External developers |
   |  Keplr, Leap,        |                  |  and services        |
   |  Cosmostation,       |                  |  (APIs and SDKs)     |
   |  OGO Wallet          |                  +----------+-----------+
   +----------+-----------+                             |
              |                                         |
              +------------+---------------+------------+
                           |
              +------------v-----------------------------+
              |            OGO Network                   |
              |  +-------------------------------------+ |
              |  | CometBFT consensus and networking   | |
              |  +-------------------------------------+ |
              |  | Cosmos SDK application framework    | |
              |  |  x/auth x/bank x/staking x/slashing | |
              |  |  x/distribution x/gov x/upgrade     | |
              |  |  x/epochs ibc-go wasmd              | |
              |  +-------------------------------------+ |
              |  | Custom modules                      | |
              |  |  x/monetary x/treasury x/rewards    | |
              |  +-------------------------------------+ |
              |  | Public interfaces                   | |
              |  |  RPC, LCD, indexer, explorer        | |
              |  +-------------------------------------+ |
              +------------------------------------------+
                           |
              +------------v-----------------------------+
              |  Validators (independent operators)      |
              |  Sentry nodes, remote signing, monitoring |
              +------------------------------------------+
```

Three money systems exist and never merge: native OGO Coin (on-chain, in x/bank), OGO Coins (account-based, in Supabase), and OGO Credits (account-based, in Supabase). No combined balance is ever shown. No automatic conversion exists. Any future conversion is a separate decision requiring legal review and explicit approval.

## 3. Component responsibility map

| Component | Owner | Responsibilities | Must never |
|---|---|---|---|
| CometBFT | Upstream (Cosmos) | Byzantine fault tolerant consensus, block production, networking, finality | N/A (consensus is not OGO code) |
| Cosmos SDK | Upstream (Cosmos) | Accounts, auth, bank, staking, slashing, distribution, governance, upgrades, epochs | N/A (framework is not OGO code) |
| x/monetary | OGO | Supply cap, issuance schedule, mint operation, supply accounting, issued-by-epoch set | Mint outside the schedule; accept an amount argument; allow a governance schedule change |
| x/treasury | OGO | Treasury accounting, R-F spending rule, emergency reserve floor, mode transition | Mint; hold coins not accounted for in x/bank; spend above the rule |
| x/rewards | OGO | Per-epoch and per-block split accounting for issuance and fees | Mint; decide the split (it executes the split) |
| ibc-go | Upstream (Cosmos) | IBC clients, connections, channels, transfers, IBC v2 | Decide supply accounting (that lives in x/monetary) |
| wasmd | Upstream (CosmWasm) | Smart contract execution, sandboxing, gas metering | Expose a mint method; allow a contract to overspend |
| OGO Wallet | OGO | Key generation, storage, signing, recovery, agent authorization | Store a seed or private key on any server; sign without user confirmation |
| Indexer | OGO | Block, transaction, and event indexing for history and the explorer | Be treated as authoritative when it disagrees with the chain |
| Explorer | OGO | Public display of blocks, transactions, accounts, validators, supply | Require an OGO account; display a value without a chain source |
| OGO application | OGO | Economy dashboard, wallet linking, payment requests, agent permissions, developer APIs, marketplace settlement | Hold keys, sign for users, write chain state, broadcast, trust client claims, act as a validator, move treasury funds |
| Cloudflare Worker | OGO | Application server: chain reads, nonce issue and signature verification, unsigned transaction preparation, on-chain verification before fulfillment | Everything in the row above, plus writing to the platform ledger as a side effect of a chain operation |
| Supabase | OGO | Application authentication, profiles, application data, Coins and Credits ledger | Own or control a blockchain balance; be the source of truth for native balances |
| Validators | Independent operators | Block production, signing, staking, upgrades, monitoring | Be appointed by OGO Systems (the set is ranked by bonded stake, not allowlisted) |

## 4. Proposed repository structure

No repository is created by this plan or by these specifications. Creation needs explicit approval for each repository. The structure below is the plan for when creation is approved.

| Repository | Purpose | Stack |
|---|---|---|
| OGO (existing, benross-lab/OGO) | Main application and Economy interface. Talks to the network through public RPC and API services only. | Existing stack (TypeScript, React, Supabase, Cloudflare Worker) |
| OGO-Network | Blockchain node, runtime and modules, genesis, protocol | Go 1.25.7, Cosmos SDK, CometBFT, wasmd, ibc-go, Cosmovisor |
| OGO-Wallet | Wallet interface and key management | Decided in M4. Must support Keplr, Leap, and Cosmostation integration and the hardware signing seam. |
| OGO-Explorer | Blocks, transactions, accounts, validators, supply | Decided in M6. Reads from the chain and the indexer. Usable without an OGO account. |
| OGO-Network-Docs | Protocol documentation, validator guides, API specifications | Markdown, published as a static site. Canonical protocol reference. |

OGO-Network layout (corrected: three custom modules only, no x/governance and no x/fees; x/monetary is the production name for the scaffold's x/ogocoin):

```
OGO-Network/
  app/            app.go, app_setup.go, modules.go
  cmd/ogod/       main.go
  x/monetary/     module.go, keeper/ (keeper, schedule, genesis, invariants),
                  types/ (params, schedule, errors), client/cli/ (query, tx)
  x/treasury/     module.go, keeper/ (keeper, spending, genesis), types/
  x/rewards/      module.go, keeper/ (keeper, split, genesis), types/
  proto/ogo/      monetary/v1, treasury/v1, rewards/v1
  genesis/        mainnet.example.json, testnet.example.json, localnet.json
  config/         config.toml, app.toml
  scripts/        build.sh, test.sh, localnet.sh, upgrade.sh
  docs/           ARCHITECTURE.md, GENESIS_SPEC.md, TOKENOMICS.md,
                  GOVERNANCE.md, SECURITY.md, REPRODUCIBLE_BUILD.md
  tests/integration/  localnet_test.go, ibc_test.go, upgrade_test.go
  Makefile, go.mod, go.sum, README.md
```

Notes: `genesis/` example files contain no real parameters until Benross approves them. The localnet genesis uses TEST-ONLY parameters, labelled so. The `docs/` directory in OGO-Network mirrors the canonical docs in OGO-Network-Docs; if the two disagree, OGO-Network-Docs wins.

## 5. Dependency and module map

### 5.1 Verified dependency matrix (Milestone B, 2026-11-06)

The matrix moves as one tested unit. No single dependency may be upgraded in isolation. Verified with `go list -m` against Gaia v28.3.1 (go.mod SHA256 `72b985a58a88efea8193e2bf77be47b51f227da0055427cc57dd21877c3db872`).

| Component | Version | Notes |
|---|---|---|
| Go | 1.25.7 | Locked prototype toolchain. Go 1.23 was the blueprint proposal; Milestone B actually used 1.25.7. |
| Cosmos SDK | v0.53.6 required, v0.53.4 effective through Gaia's replace | Replace exists because wasmd calls in SDK v0.53.6. |
| CometBFT | v0.38.23 | Includes the median-time fix (GHSA-c32p-wcqj-j677, fixed in v0.38.21). CometBFT v1.x is retracted. |
| CosmWasm (wasmd) | v0.60.9 | wasmvm/v2 v2.3.5. CGO embeds one absolute module cache path; byte-identical builds require the canonical default GOMODCACHE. |
| IBC | ibc-go/v10 v10.7.0 | IBC Classic plus IBC v2. 08-wasm v10.5.0. |
| Cosmovisor | v1.7.3 | Separate binary, for coordinated binary upgrades. |
| Reference | Gaia v28.3.1 | Source of the matrix. The Hub ran gaia v28.3.1 on 2026-10-15 (SDK v0.53.4, CometBFT v0.38.23, ibc-go v10.7.0, wasmd v0.60.9). |

Do not adopt SDK v0.53.8 or Gaia v29 for the first localnet: multisig `--gas auto` is broken in SDK v0.53.8. Gaia v29 was approved by Hub Proposal 1059 (99.41 percent) and scheduled for block 33,618,500 on 28 October 2026; verify its current status before relying on it for anything.

### 5.2 Current maintained versions, checked 2026-11-06

Only two Cosmos release families are supported. Everything below SDK v0.53 and CometBFT v0.38 is end of life and receives no maintenance, security patches, or compatibility support.

| Release family | Cosmos SDK | CometBFT | Status |
|---|---|---|---|
| 2025.1 | 0.53.x (latest patch v0.53.5, 2025-12-12; v0.53.6 required by Gaia, v0.53.4 effective) | 0.38.x (v0.38.23 in the matrix) | Supported. The build target for Milestone C. |
| 2026.1 | 0.55.x (v0.55.0, 2026-07-27; v0.55.1 patch builds appearing November 2026) | 0.40.x | Supported. Separate evaluated upgrade after Milestone C, not a build target for the first localnet. |

Intermediate release v0.54.0 (2026-04-10) exists between the two families. The v0.55 upgrade guide documents a direct 0.53 to 0.55 path; whether to step through v0.54 is part of the upgrade evaluation (proof-of-concept criterion M4), not decided here.

What v0.55 changes that matters for OGO: native ML-DSA post-quantum signatures (key type `ml_dsa_65`; a chain counts as post-quantum once two thirds of voting power uses ML-DSA consensus keys); in-place validator consensus key rotation; cosmos/kms v0.1.0; removal of `SIGN_MODE_TEXTUAL`; removal of the `x/protocolpool` module (added in v0.53, removed in v0.55, so any code written against it in the 0.53 matrix would need rework on upgrade); and the `x/epochs` NewAppModule fix described in section 6.3. Ledger Security 2026.1 (2026-10-08) bundles SDK v0.55.0, CometBFT v0.40.0, the enterprise proof-of-authority module v1.1.0, and cosmos/kms v0.1.0.

Evaluation track only, not build targets: wasmd v0.61 with wasmvm v3 (seen live on Terra Classic and Uptick in October 2026); CosmWasm 3.0; SDK 0.55 with CometBFT 0.40. Each needs its own compatibility check against the whole matrix before adoption.

### 5.3 Module map

Standard modules used, and what each is for: x/auth (accounts, signature verification, sequences); x/bank (native denom balances, sends, supply queries, burn); x/staking (validator set, bonding, delegation, unbonding, redelegation); x/slashing (downtime and equivocation, jailing, tombstoning); x/distribution (validator and delegator reward distribution, community pool accounting); x/gov (proposals, deposits, voting, tally, execution); x/upgrade (upgrade plans, store migration coordination); x/epochs (named timers for epoch-based periods, see section 6); x/evidence (equivocation evidence, feeding x/slashing); ibc-go (IBC client, connection, channel, transfer, 08-wasm client); wasmd (CosmWasm contracts).

Custom modules, only three: x/monetary (supply cap, issuance schedule, mint, supply accounting), x/treasury (treasury accounting, R-F spending rule, emergency reserve floor, mode transition), x/rewards (per-epoch and per-block split accounting for issuance and fees). x/monetary is the only module that can mint. x/treasury does not mint. x/rewards does not mint and does not decide the split; it executes the split.

Explicitly NOT created: x/governance (x/gov already provides governance; a second governance module would create a divergent voting path) and x/fees (fee collection uses the standard ante handler plus x/distribution, and fee splitting is handled by x/rewards). No x/bridge module: bridges are third-party trust and are not part of the chain.

## 6. The epochs question: investigation findings

The plan recommended epoch-based periods via an epochs module. This section records what was verified about that module on 2026-11-06, because the plan described it as "Osmosis x/epochs style" and the actual state is better than that description: x/epochs is now a first-class Cosmos SDK module.

### 6.1 Finding: x/epochs is in the SDK core from v0.53.0

x/epochs was upstreamed from Osmosis into the Cosmos SDK in v0.53.0 (SDK changelog, PR #23815, "Upstream x/epochs from Osmosis"). It is documented in the Cosmos SDK docs as `x/epochs`, and it is listed in the v0.53 upgrade reference as an optional new module. OGO does not need to vendor the Osmosis module or write its own timer. The module is available in the same 0.53 matrix the scaffold already uses.

Revision note, 2026-11-06: the first-class status of x/epochs from SDK v0.53.0, upstreamed through PR #23815, is confirmed and retained. The mandatory depinject hook-fires test that proves the module works in the wired application is specified in section 6.2, condition 1, and in `OGO_CONSENSUS_AND_EPOCH_SPEC.md` section 13.

How it works, from the SDK documentation:

- The module defines on-chain timers that execute at fixed time intervals. Other modules register logic to be executed at timer ticks through hooks.
- Each timer has a unique identifier. Each epoch has a start time and an end time, where end time equals start time plus the timer interval.
- The timer ticks at the first block whose block time is greater than the timer end time, and sets the start as the prior timer end time.
- The module keeps a single `EpochInfo` per identifier. EpochInfos are initialized as part of genesis initialization or upgrade logic.
- Hooks: `BeforeEpochStart` and `AfterEpochEnd`, called with the epoch identifier and the epoch number. A receiving module must filter on the epoch identifier and execute only for its own identifier.
- Events: `epoch_start` with the epoch number, and the module provides queries for the current epoch infos (`<appd> query epochs epoch-infos`).

### 6.2 Verdict: appropriate for OGO's monetary schedule, with four conditions

x/epochs is appropriate for OGO's monetary schedule, subject to the following conditions, each of which must be satisfied in Milestone C and proven by the proof of concept:

1. **Hook registration must be proven to fire, not just wired.** There is a known depinject bug in this area: SDK issue #24770 (register of epoch hooks in `InvokeSetHooks`, fixed in v0.53.2) and issue #25425 (`InvokeSetHooks` called with a nil keeper, and the AppModule containing a copy instead of a pointer, so hooks set post-creating the AppModule through depinject did not apply because it is a different instance). The v0.54 upgrade guide changes `NewAppModule` to require the epoch keeper by pointer instead of value, fixing the bug. In the 0.53 matrix, OGO must verify with a mandatory depinject hook-fires test, in the wired application, that the x/monetary `AfterEpochEnd` hook actually executes on an epoch tick. The test is release-blocking for the monetary prototype and must confirm all five of the following, in order: (1) the epoch module initializes in the wired application; (2) the monetary hook is registered in the wired hook set; (3) the hook actually fires at an epoch transition, minting issuance and emitting the epoch event; (4) exactly one scheduled issuance occurs for that epoch, equal to exactly R_n, with the epoch recorded exactly once in the issued-by-epoch set; (5) a duplicate hook execution cannot issue twice. Compilation alone is not an acceptable test, and a keeper-level unit test alone is not an acceptable test. A module that is wired but whose hook never fires is a silent monetary failure: no issuance would ever happen, and the failure would look like the schedule working correctly with zero issuance.

Revision note, 2026-11-06: condition 1 above replaces the earlier, looser wording, which only asked to "verify with a test that the hook actually executes". The five confirmations are now mandatory acceptance criteria, and the test is release-blocking for the monetary prototype. The full test specification is in `OGO_CONSENSUS_AND_EPOCH_SPEC.md` section 13.
2. **Adding x/epochs requires a StoreUpgrade.** The upgrade reference states this explicitly. The wiring is: keeper instantiation, store key addition, hooks registration, app module registration, an entry in `SetOrderBeginBlockers`, an entry in `SetGenesisModuleOrder`, and an entry in `SetExportModuleOrder`. Because the scaffold was built without x/epochs, Milestone C must add it through a store upgrade, with the migration tested on the proof-of-concept chain before it is used for any real schedule.
3. **Epoch timing is driven by block time.** The timer ticks at the first block whose block time is greater than the end time, and block time is the weighted median of validator timestamps. The schedule's determinism therefore comes from epoch numbers, not from wall-clock time. This is acceptable, and it is strictly better than block-based periods (which drift with block time) and better than pure calendar time (which is directly manipulable). But it means the epoch boundary inherits the timestamp trust assumptions of CometBFT, including the median-time manipulation addressed by GHSA-c32p-wcqj-j677 (fixed in CometBFT v0.38.21, included in the matrix's v0.38.23). The monetary rules must never depend on the exact wall-clock time of a tick, only on the epoch number and the fixed duration D.
4. **Halt and catch-up behaviour must be verified against the pinned implementation, not assumed.** The plan requires missed epochs to tick one per block on restart, in epoch order, with no lump mint, no skipped epoch, and no double mint (proof-of-concept criterion S4). Whether the pinned x/epochs BeginBlocker advances one epoch per block or attempts to catch up multiple epochs in a single block must be read from the actual SDK source at the pinned version during Milestone C, and the S4 test (30-minute halt, restart, observe) is the acceptance test. If the pinned implementation does not advance one epoch per block, x/monetary must enforce the one-per-block catch-up itself, by claiming at most one epoch per block and leaving the rest for subsequent blocks.

### 6.3 Alternatives considered

- **SDK x/mint epoch path:** x/mint provides inflation-based issuance, not a fixed-supply scheduled issuance. It cannot express OGO's schedule (fixed max, schedule-only mint, no discretionary mint), so it is not an alternative for the monetary schedule. It is not used.
- **Vendoring Osmosis x/epochs directly:** unnecessary now that the module is in the SDK core. Vendoring would add a second copy of the same code and a second maintenance burden. Not recommended.
- **A custom epochs module inside x/monetary:** unnecessary while the SDK module exists and is maintained. If the SDK module were ever deprecated or removed, a custom module would be the fallback, and the hook interface (epoch identifier plus epoch number) is small enough to reimplement. Not needed now.
- **Block-based periods:** rejected for the production schedule (Solana drift precedent, 276 days behind by June 2026). The scaffold uses block-based periods today; Milestone C migrates the schedule code to epoch-based periods. See `OGO_CONSENSUS_AND_EPOCH_SPEC.md` for the full comparison.

## 7. Security boundaries

Five boundaries, each enforced in code or in architecture, not by convention:

1. **Consensus vs application:** CometBFT decides ordering and finality. The SDK application decides state transitions. No application code can change consensus rules, and no consensus configuration can mint.
2. **Monetary vs governance:** monetary parameters (max supply, schedule, final issuance period, treasury rule) live in x/monetary and x/treasury state, not in a generic parameter store x/gov can change. A governance proposal cannot change them. An upgrade handler that attempts to modify monetary state is rejected by a denylist gate. The mint entrypoint takes no amount argument. An every-block invariant checks the supply rules. The schedule is publicly queryable. See `OGO_NETWORK_SECURITY_MODEL.md`.
3. **Chain vs application backend:** the OGO application talks to the network only through public RPC and API services. The Worker never holds keys, never signs, never writes chain state, never broadcasts, never trusts client claims, never writes to the platform ledger as a side effect of a chain operation, never acts as a validator, and never moves treasury funds. The blockchain is the source of truth for native balances; Supabase is the source of truth for application data.
4. **User keys vs OGO servers:** signing happens only in the user's wallet. No OGO server ever sees a private key or a seed phrase. See `OGO_NATIVE_WALLET_SPEC.md`.
5. **Protocol revenue vs company revenue:** protocol revenue (issuance shares, transaction fees, treasury income) is kept separate from OGO company revenue (subscriptions, AI inference, hosted apps, marketplace revenue). Company revenue must not be counted as protocol revenue unless an explicit, implemented, disclosed mechanism routes a portion into the network. See `OGO_TREASURY_AND_SUSTAINABILITY.md`.

## 8. Recommended development stack

| Layer | Stack | Notes |
|---|---|---|
| Chain | Go 1.25.7, Cosmos SDK 0.53 matrix, CometBFT 0.38.23, wasmd 0.60.9, ibc-go v10.7.0, Cosmovisor 1.7.3 | Locked for the prototype. Upgrade to the 2026.1 family is a separate evaluated decision after Milestone C. |
| Wallet | Decided in M4 | Must support Keplr, Leap, Cosmostation; abstract signing seam for hardware wallets. |
| Indexer | Decided in M6 | Indexes blocks, transactions, events. Reconciled against the chain (gate G9). |
| Explorer | Decided in M6 | Reads from chain and indexer. Usable without an OGO account. |
| Application | Existing OGO stack (TypeScript, React, Supabase, Cloudflare Worker) | Network integration is a module in the existing app, not a new application. |
| Docs | Markdown in OGO-Network-Docs, published as a static site | Canonical protocol reference. |
| CI and builds | Reproducible builds from source, pinned dependencies with hashes, clean-copy build verification | Byte-identical binary under the canonical GOMODCACHE condition (Milestone B finding). |
| Remote signing | cosmos/kms v0.1.0 | TMKMS will be removed from the Cosmos Bug Bounty Program at the start of Q1 2027; operators have about three months from the Ledger Security 2026.1 release (2026-10-08) to migrate. Do not build new validator operations on TMKMS. |

## 9. What is not finalized

The framework is the leading candidate, not a finalized decision. Cosmos SDK plus CometBFT is recommended, continuing from the verified Milestone B scaffold, but final selection requires the proof of concept (plan section 2.6) to pass: 21 MUST criteria (P1 to P5, S1 to S8, M1 to M5, O1, O2, O5) covering performance, security, maintenance, and operations. Revision note, 2026-11-06: the earlier figure of 26 was a count error, corrected in the reconciliation pass, and O3 and O4 were then removed from the framework-selection criteria by Benross's decision (Phase 2 Execution Directive), bringing the count from 23 to 21. If a MUST criterion fails for a reason structural to the SDK or CometBFT, the next step is a time-boxed Substrate spike that re-implements x/monetary as a FRAME pallet and re-runs the bypass suite, not a full rewrite. That spike is not authorized by these specifications.

Open items that this document does not decide: epoch duration D; maximum supply; issuance model and parameters; validator, delegator, community, and treasury split; fee policy; treasury spend level and emergency reserve floor; governance thresholds; network name and chain id; denomination and decimals; bech32 prefix and derivation path. All are listed in `OGO_NETWORK_DECISION_REGISTER.md`.

## 10. Sources

- Cosmos SDK changelog and upgrade references, checked 2026-11-06: x/epochs upstreamed from Osmosis in v0.53.0 (PR #23815); x/epochs bug fixes #24610, #24770, #25087, #25425; v0.54 upgrade guide (NewAppModule keeper-by-pointer change); v0.55.0 release 2026-07-27; v0.54.0 release 2026-04-10; v0.53.5 release 2025-12-12; release families page (2026.1: SDK 0.55.x; 2025.1: SDK 0.53.x; SDK v0.50.x and lower end of life); Ledger Security 2026.1 (2026-10-08).
- Cosmos Hub operator docs, checked 2026-10-15 and rechecked 2026-11-06: gaia v28.3.1 (SDK v0.53.4 with go.mod requiring v0.53.6, CometBFT v0.38.23, ibc-go v10.7.0, wasmd v0.60.9); Gaia v29 approved by Proposal 1059, scheduled for block 33,618,500 on 28 October 2026.
- Milestone B build report and reproducible build guide, in `ogo-network-milestone-b/`.
- Consolidated plan, `ogo-network-native-coin-architecture-implementation-plan-v1.0.md`, sections 2 and 3.
