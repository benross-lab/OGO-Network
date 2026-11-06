# OGO Network: Native Coin Architecture and Implementation Plan v1.1

Status: Architecture and implementation planning only. OGO-Network repository created 2026-11-06 (public, benross-lab/OGO-Network) by Benross's explicit authorization. No blockchain deployed, no token or smart contract deployed, no public token sale, no testnet launched, no production infrastructure changed, no database migration applied, no money spent, no infrastructure purchased, no monetary parameter selected, no publication or launch announcement authorized. Implementation work must not be merged into main without separate approval.
Date: 2026-11-06
Version: v1.1 (v1.0 kept intact as the reviewed version)
Project: OGO Network
Native asset: OGO Coin
Architecture direction: Independent Layer 1 blockchain
Initial framework candidate: Cosmos SDK + CometBFT (recommendation, pending proof of concept)

## Revision log, v1.0 to v1.1 (2026-11-06)

v1.1 incorporates Benross's review of v1.0, the four research findings, the three inconsistency resolutions, and his Phase 2 Execution Directive. Nothing in v1.0 was silently overwritten: v1.0 stays intact as the reviewed version, and every change below is marked with a revision note in the text.

1. **Validator count: 3 for the local prototype (Benross decided).** The reconciliation pass had resolved M1 to 4 validators with BFT reasoning; Benross's decision supersedes that for the local prototype. M1 states 3 throughout, with the distinction recorded: 3 is the decided local-prototype execution requirement, 4 is the minimum operational set, 50 or more is the decentralization target. Section 12.3 now says 3 local validator nodes, not 3 to 4.
2. **Keplr and IBC removed from framework-selection tests (Benross decided).** O3 and O4 are DEFERRED, not prerequisites for the initial monetary prototype unless a specific technical dependency demonstrates necessity. The proof of concept has 21 MUST criteria (P1 to P5, S1 to S8, M1 to M5, O1, O2, O5), not 23; the earlier figure of 26 was a count error. O3 and O4 rows are retained as definitions of what those tests must show when they happen.
3. **Proof-of-concept validator count (P1, S5, M2, O1): DECIDED, Option 1 (Benross, 2026-11-06).** Those criteria are specified for a 4-validator network, which S5 needs. Benross chose a separate four-validator framework-selection proof of concept, in addition to the unchanged three-validator Phase 2 prototype; S5 is NOT redefined as a halt-and-resume demonstration. The PoC ran on 2026-11-06: S5 liveness demonstrated (blocks continued with 1 of 4 offline, restart and sync verified), P1 partially executed (24-hour duration not completed in one session), M2 not executed (no Cosmovisor or upgrade-handler support in the scaffold), O1 partially executed (runbook and snapshot/restore executed; no sentry topology on a single machine). Full results in `ogo-network-poc-4validator/ogo-network-poc-4validator-report.md`. This closes decision blocker D1.
4. **Mandatory hook-fires test (finding a).** S3 now specifies the five ordered confirmations (module initializes, hook registered, hook actually fires, exactly one issuance, duplicate cannot issue twice). Compilation alone is not an acceptable test. x/epochs is a first-class SDK module from v0.53.0 (PR #23815).
5. **Treasury sustainability (finding b).** Desired spending and sustainable spending are separated; the low-fee finding (raising desired spend from 1,500 to 2,500 changes actual spending by 0 units out of 114,511; both runs spend 114,511) is recorded, with the six-figure reporting requirement (desired, actual permitted, inflows, reserve balance, unfunded, shortfall periods).
6. **Agent fee market (finding c).** Agent micropayment viability is an explicit protocol design requirement; agent-native batching is a core protocol consideration; the fee architecture distinguishes ordinary transactions from agent micropayments; fee sponsorship is evaluated as an additional mechanism. The 1,000 uogo fee example and the approximately one-million-transactions-per-day finding are recorded.
7. **Hybrid valley (finding d).** The 26-period result is corrected: it assumes fee growth begins only after issuance ends; the valley is 0 periods when fee growth begins at genesis. Every scenario table must identify its fee-growth start assumption.
8. **Stale delivery language.** Section 9.1 now states the eleven specifications have been delivered 2026-11-06 to the architecture workspace.
9. **CosmWasm sequencing.** The first monetary prototype remains independent of smart contracts; S7 and P3 are framework-selection capability checks; CosmWasm adoption is evaluated separately after M3 against the eight criteria in section 9.4.
10. **Framework status.** Cosmos SDK plus CometBFT remains a candidate, not a finalized selection, pending the proof of concept.
11. **GitHub publication and discrepancy resolutions (2026-11-06).** Benross authorized creating the public benross-lab/OGO-Network repository and pushing the authorized scope (prototype chain, four-validator PoC as reference, finished network documentation). Two discrepancies flagged in the exchange readiness report are resolved here: hybrid dust is 22 uogo (not 18), and the low-fee R-F delta is 0 units (not 4). See `docs/discrepancy-resolutions-2026-11-06.md`. OGO-Wallet, OGO-Explorer, and OGO-Network-Docs are not created; Milestone C work in progress is not pushed; no public testnet is authorized.

## How this plan was built, and what was verified

This plan consolidates six parallel workstreams: repository inspection, framework evaluation, protocol specification and monetary integration, wallet architecture, application integration and engineering plan, and security and threat model. Each workstream draft was checked by the coordinator against the prior reports and, where the claim was about a live system, against the live system itself.

Verified by the coordinator independently of the workstream drafts:

- Live main of benross-lab/OGO, 2026-11-06: HEAD commit `0e30dcd459745f194839a8100d5c2665456ef509`, 2026-11-06T06:12:19Z, "OGO Universe: W/arrows tap-to-step on keydown, matching D-pad". Commit `e2fd900dbcf9`, 2026-11-06T02:41:09Z, "Add account-secured OGO Coins and Credits wallet", exists and introduced the economy wallet.
- The wallet UI finding is confirmed: `src/OgoEconomy.tsx` (111 lines) exists on live main, and `src/App.tsx` (3,377 lines), `src/main.tsx`, and `src/SubscriptionSection.tsx` contain no reference to it. The migration, Worker endpoints, and Stripe webhook are live in main; the wallet UI component appears to have no mount point. This needs one live UI confirmation (see section 11).
- Milestone B scaffold facts, rechecked against its build report: Go 1.25.7, SDK v0.53.6 required with v0.53.4 effective, CometBFT v0.38.23, wasmd v0.60.9, wasmvm/v2 v2.3.5, ibc-go/v10 v10.7.0, 54 tests passing, single-node localnet to height 68, reproducible build byte-identical under the canonical GOMODCACHE condition.
- Polkadot SDK stable2609: node v1.25.0 confirmed via public release reporting, 2026-10-26. The framework draft stated 88 breaking changes; the stable2609-rc1 release notes record 87. This plan uses 87.

Discrepancies between workstream drafts, resolved in this plan (not smoothed over):

- Module list: the protocol draft justifies NOT creating x/governance and x/fees (they duplicate x/gov, x/auth, and x/rewards). The integration draft's repository layout still listed them. This plan follows the protocol draft: three custom modules only, x/monetary, x/treasury, x/rewards. See section 3.
- Module naming: the production monetary module is x/monetary, which extends and renames the scaffold's x/ogocoin. The integration draft uses x/ogocoin in places. This plan uses x/monetary throughout, with the descent noted.
- Go toolchain: the blueprint proposed Go 1.23, but Milestone B actually used and verified Go 1.25.7, and the Milestone B authorization locks Go 1.25.7 as the prototype toolchain. This plan states Go 1.25.7. Any M0 text still citing Go 1.23 is superseded.
- Internal cross-references in the protocol draft (binding surface, wiring remaining) pointed at the wrong section numbers. This plan uses the correct references.

## Decisions recorded by Benross, 2026-11-06

Two decisions resolve questions left open by earlier work. They are recorded here as decisions, not proposals.

1. Architecture direction: OGO Network is an independent Layer 1 blockchain and OGO Coin is its native currency, not an ERC-20 or SPL token on another chain. This is the feasibility report's option B (OGO Network as the native asset's home from inception). The conditional Base recommendation in the feasibility report is superseded for the native coin. Guiding principle: OGO owns its monetary infrastructure, while users own their coins and private keys.
2. Repository architecture: the blockchain lives in separate repositories from the existing OGO application: OGO-Network (chain), OGO-Wallet (wallet), OGO-Explorer (explorer), OGO-Network-Docs (protocol docs). This resolves the earlier repo tension (the charter proposed a standalone ogo-network layout; the standing repo-first rule said benross-lab/OGO only). For the network, separate repositories are now the decision. No repositories are created by this plan; creation needs explicit approval for each repository.

Everything else in this plan is a recommendation, a candidate, or an open decision. No monetary parameter is selected anywhere in this document.

## 1. Architecture Decision Record

### ADR-1: OGO Coin is native to its own independent Layer 1

Context: OGO needs an open economic layer for AI agents, developers, applications, digital services, marketplace transactions, and eventually telecommunications products. A token on another chain would make OGO's core monetary system dependent on that chain's governance, fee market, and upgrade path. The founding charter states OGO Network is an independent, open-source Layer 1 with a fixed maximum coin supply and community governance, and that no OGO account is needed to own, transfer, or use OGO Coin.

Decision: OGO Network is an independent Layer 1 blockchain. OGO Coin is its native currency, the same way BTC belongs to Bitcoin. There is no token on Base, Ethereum, or Solana, and no wrapped or bridged form of OGO Coin.

Consequences: OGO must operate its own validator set, consensus, upgrades, and security. OGO owns its monetary infrastructure. Users own their coins and private keys. The network must be usable without an OGO account, through independent wallets, exchanges, developers, and third-party services. If the OGO application disappears, the network continues operating.

Alternatives rejected: a public token on an existing chain (feasibility report options A and C) is rejected for the native coin, because it would make the core monetary system dependent on another chain. It remains a legitimate question for other assets, but it is not the home of OGO Coin.

### ADR-2: Separate repositories for the network

Context: the charter proposed a standalone ogo-network layout, while the standing repo-first rule said all OGO work goes in benross-lab/OGO only. The two rules conflict for a blockchain, which needs its own release cycle and security boundary.

Decision: five repositories. OGO (main application and Economy interface), OGO-Network (blockchain node, runtime and modules, genesis, protocol), OGO-Wallet (wallet interface and key management), OGO-Explorer (blocks, transactions, accounts, validators, supply), OGO-Network-Docs (protocol documentation, validator guides, API specifications). No repository is created by this plan.

Consequences: the chain has its own release cycle, its own CI, and its own security boundary. The OGO application communicates with the network through public RPC and API services only. Supabase remains responsible for application authentication, profiles, and application data. The blockchain remains the source of truth for native coin balances.

### ADR-3: Cosmos SDK + CometBFT as the framework candidate (recommendation, pending proof of concept)

Context: see section 2. The recommendation is to continue from the verified Milestone B scaffold. Final framework selection requires the proof of concept in section 2 to pass.

Decision (recommended, not final): Cosmos SDK + CometBFT, Go implementation, with IBC and CosmWasm. A Polkadot parachain is rejected as the deployment form, because it uses shared relay-chain security rather than OGO's own independent validator security and consensus.

### ADR-4: Epoch-based periods (recommendation, pending confirmation)

Context: Decision Preparation v0.1 section 1. Block-based periods drift with block time (Solana was 276 days behind its intended inflation timeline by June 2026). Pure calendar time is fragile because block timestamps are a weighted median under validator control (Tachyon GHSA-c32p-wcqj-j677, February 2026).

Decision (recommended): epoch-based periods via an epochs module. A period is one epoch of a named timer, with fixed duration D in seconds, fixed in genesis. Issuance for epoch n is R_n, claimed once, in the epoch hook, keyed by epoch number. No duration D is recommended here. See section 3.

### ADR-5: The platform ledger stays separate

Context: the existing OGO Coins and Credits implementation is an account-based database ledger in Supabase, verified on live main (section 7 of the repo inspection in the workstream drafts; summarized in section 7 of this plan). It is not a self-custody wallet.

Decision: the Supabase Coins/Credits ledger remains a separate application-level accounting system. It does not become the blockchain ledger, and there is no automatic conversion between it and native OGO Coin. Any future conversion mechanism is a separate decision requiring legal review and explicit approval.

### ADR-6: Monetary rules carried forward

The following are carried forward from prior research as recommendations or candidates, not decisions. Each still needs Benross's explicit approval.

- Three-number accounting, recommended, model-independent: M = G + S_actual, computed at genesis by exact integer summation, with dust left permanently unissued. See section 4.
- R-F treasury rule, recommended default: spend = min(D, max(Trail_4, floor(0.05 x B)), max(0, B - F)), one formula for both modes, transition keyed to the genesis-fixed final issuance period. See section 4.
- Gentle geometric taper, candidate if hybrid is chosen: factor about 0.95 over the final half of the schedule, renormalized to the same total. Short linear tapers are rejected as worse than abrupt. See section 4.
- P4 combined treasury policy, candidate, not approved: the most conservative option, with a known post-issuance freeze defect. See section 4.
- Geometric and hybrid remain the two monetary candidates. No model or parameter is selected.

## 2. Framework comparison

### 2.1 The verified Cosmos scaffold baseline

Milestone B is not a paper design. It is a working Cosmos SDK scaffold in staging at `ogo-network-milestone-b/chain/`, and this evaluation starts from it rather than from zero. Any recommendation to change frameworks must justify discarding this verified work.

Verified, 2026-11-06:

- Toolchain: Go 1.25.7 linux/amd64, from `~/workspace/tools/go-1.25.7`. Tarball SHA256 verified against the official checksum.
- Dependency matrix, verified with `go list -m`, as one tested unit: Cosmos SDK v0.53.6 required, v0.53.4 effective through Gaia's replace; CometBFT v0.38.23; wasmd v0.60.9; wasmvm/v2 v2.3.5; ibc-go/v10 v10.7.0, with 08-wasm v10.5.0; cometbft-db v1.0.4; cosmos-db v1.1.3; IAVL v1.2.6; cosmos-proto v1.0.0-beta.5; cosmossdk.io/math v1.5.3; Cosmovisor v1.7.3 as a separate binary. Reference: Gaia v28.3.1 go.mod, SHA256 `72b985a58a88efea8193e2bf77be47b51f227da0055427cc57dd21877c3db872`. The SDK replace exists because Gaia v28.3.1 carries the comment that wasmd calls in SDK v0.53.6 require the pin back to v0.53.4 until Gaia upgrades to v0.53.6 or higher. No dependency was independently upgraded.
- Tests: 54 tests pass, 0 fail, in x/ogocoin, covering genesis validation (16), schedule math (7), ledger invariants (8), IBC escrow and voucher accounting (6), and bypass attempts (17). `go vet ./...` exits 0.
- Monetary rules already encoded and tested: fixed maximum supply with no default, mandatory issuance schedule, discretionary mint always fails, scheduled mint issues exactly the schedule amount once per period into the rewards pool (`ogo1rewardspool`), founder and reserved purposes rejected, governance cannot change max supply, minting, or the schedule through an ordinary upgrade, and the wasm binding surface exposes no mint method. The binding surface tested is exactly [BankSend, BankBurn, QueryBalance, QuerySupply].
- Local node proof: single-node localnet, chain id `ogo-local-1`, binary `ogod` 0.1.0-milestone-b, produced blocks to height 68 across two runs (first run stopped at height 33 by a VM service restart, then resumed cleanly). A real transfer of 123456 uogo with fee 100 and a real burn of 1000 uogo both returned code 0, with exact balance, supply, and schedule query results. Genesis 1,000,000,000 plus scheduled 3,439 equalled total supply 1,000,003,439 exactly. All numbers in the proof are labelled TEST-ONLY.
- Reproducible build: clean-copy build from a fresh module cache exits 0, vet 0, tests 54 pass and 0 fail, in about 18.5 minutes. Byte-identical binary hash `5794dc3707c692e2159f7dab9b7a7ad541ff2eb6c66562ea5e2cb7653243be04` is achieved only with the canonical default GOMODCACHE path, because wasmvm v2.3.5 CGO embeds one absolute module cache path that `-trimpath` does not rewrite. With a custom GOMODCACHE the build still succeeds and all tests still pass, but the hash differs by that one embedded path. This is a documented reproducibility condition, not a supply-rule issue.

What the scaffold does NOT yet do (Milestone C work, stated as limits, not defects):

- Transactions are JSON and are NOT signature-verified. Anyone with RPC access can submit a transaction naming any sender. Real authentication arrives with full SDK depinject wiring.
- Standard SDK modules (auth, bank, staking, gov, distribution, fees, wasmd, ibc-go) are pinned and compiled through blank imports. Their keepers are NOT yet wired.
- The wasm VM is not instantiated. Live contract execution is pending. The bypass tests cover the binding surface, not live contract execution.
- Live IBC relaying is not wired. Escrow and voucher accounting is ledger-level only.
- Keys under `<home>/keys` are plaintext, localnet only.
- The scaffold implements block-based periods. The recommended architecture is epoch-based periods, so the schedule code needs a period-architecture migration, not just keeper wiring.

Upgrade path already identified: only two Cosmos release families were supported at the blueprint check: 2025.1 (SDK 0.53.x plus CometBFT 0.38.x) and 2026.1 (SDK 0.55.x plus CometBFT 0.40.x). Everything below SDK 0.53 and CometBFT 0.38 is end of life, and CometBFT v1.x was retracted. Public sources checked 2026-11-06 confirm the 2026.1 family: SDK v0.55.0 released 2026-07-27, CometBFT v0.40.0, ML-DSA post-quantum signatures, in-place validator consensus key rotation, cosmos/kms v0.1.0, bundled as Ledger Security 2026.1 on 2026-10-08. The v0.55 upgrade guide documents a direct 0.53 to 0.55 path. Build Milestone C on the verified 0.53 matrix, then evaluate 0.55 as a planned upgrade, rather than restarting the scaffold on the newest family mid-build.

### 2.2 The three options, in brief

Option A: Cosmos SDK plus CometBFT (Go). A modular Go framework for application-specific chains. CometBFT provides Byzantine fault tolerant consensus, networking, and finality. The SDK provides auth, bank, staking, gov, distribution, and fee modules, plus a keeper and capability model for custom modules. IBC is provided by ibc-go. Smart contracts run through CosmWasm (wasmd plus wasmvm). Production users include Cosmos Hub, Osmosis, dYdX Chain, Injective, Sei, and Celestia. The SDK is used by more than 200 production chains, according to Cosmos ecosystem reporting checked 2026-11-06.

Option B: Substrate, now Polkadot SDK (Rust). A modular Rust framework for custom blockchains, built by Parity Technologies. FRAME provides pallets covering balances, staking, governance and referenda, treasury, multisig, identity, and more. The runtime compiles to WebAssembly and is stored on chain, which enables forkless runtime upgrades. Consensus is pluggable: BABE for block production plus GRANDPA for finality is the standard solo-chain combination. Smart contracts run through pallet-revive (PolkaVM, ink! v6, with Solidity ABI support). Cross-chain messaging uses XCM. A Substrate chain can run as an independent solo chain or as a Polkadot parachain; this evaluation covers the solo-chain form only. Current versions checked 2026-11-06: `polkadot-sdk` 2609.0.0 released 2026-10-08, node version v1.25.0, built and tested with Rust 1.93.0. The stable2609-rc1 release notes record 87 SemVer-breaking changes. The single `polkadot-sdk` umbrella crate is now the recommended dependency entry point. Solo-chain precedents include Kulupu, Robonomics, Vara (Gear), and Geode Chain.

Option C: Purpose-built Go blockchain (CometBFT ABCI, own state machine). Write the state machine directly against CometBFT through ABCI, in Go, without the Cosmos SDK. CometBFT still supplies consensus, networking, and block production. Everything above consensus is OGO's own code: accounts, balances, signatures, staking and validator sets, governance, fees, storage, serialization, RPC and query APIs, and any interoperability. This is best understood as "CometBFT plus a hand-built application", not as "write consensus from scratch". Writing a BFT consensus engine from scratch is not evaluated here: it would be a multi-year research project with no benefit over reusing CometBFT.

### 2.3 Three-way comparison

Ratings: Strong (production-proven for this criterion, low residual work), Adequate (workable, with known gaps or costs), Weak (material gap, major build or ecosystem work, or a structural mismatch). Ratings are for OGO's stated requirements, not for the frameworks in the abstract.

| Criterion | A: Cosmos SDK + CometBFT | B: Substrate, Polkadot SDK | C: Purpose-built Go on CometBFT ABCI |
|---|---|---|---|
| Sovereignty and consensus ownership | Strong. Own validator set, own CometBFT consensus, own governance. No shared security dependency. | Strong. Own validator set, own BABE + GRANDPA consensus, own governance as a solo chain. | Strong. Own validator set and own state machine on CometBFT. Full ownership, including full responsibility. |
| Security and validator architecture | Strong. CometBFT BFT, instant finality, proven staking, slashing, and unbonding modules, sentry architecture, cosmos/kms remote signer. Hub slashing precedent: 0.01 percent plus 10 minute jail for downtime, 5 percent plus tombstoning for double signing. | Strong. BABE + GRANDPA, Nominated Proof of Stake, slashing, session and era timing. Finality is probabilistic block production plus BFT finality, typically seconds to about a minute, slower to finality than CometBFT per block. | Weak to Adequate. CometBFT consensus is proven, but staking, slashing, key rotation, unbonding, and validator-set management must all be designed, built, and audited by OGO. Every bug is OGO's bug. |
| Native asset implementation | Strong. Native denom (`uogo` proposed, 6 decimals) is a first-class bank balance, not a contract token. Fixed supply, schedule-only mint, and IBC escrow accounting are already coded and tested in Milestone B (54 tests). | Adequate. Native asset is `pallet-balances` plus a custom issuance pallet. The fixed-supply and schedule-only mint rules, purpose labels, and bypass tests would all need re-implementation in Rust and re-proof. No verified OGO scaffold exists on this stack. | Adequate in principle, Weak in practice. A native asset is trivial to represent in a hand-built store, but supply enforcement, fee handling, and burn semantics must be built and tested from zero. |
| Performance and scalability | Adequate to Strong. CometBFT gives instant finality per block, typical block times 1 to 6 seconds, vendor claims of 1,000 to 10,000 TPS per chain under optimal conditions. Cosmos Labs' 2026 roadmap targets 5,000 TPS sustained and 500 ms block times as a goal, not a measured OGO result. The binding constraint is fee demand, not raw TPS. | Adequate to Strong. Six-second slots are the common solo-chain default, with finality after GRANDPA rounds. Throughput is ample for OGO's expected volumes. Block time is coarser than a tuned CometBFT chain, which matters for user-perceived latency on payments. | Strong in principle. A lean hand-built state machine has no SDK overhead and can be tuned to OGO's exact transaction set. In practice the gain is unproven until benchmarked, and it is bought by giving up every module listed above. |
| Wallet ecosystem | Strong. Keplr and Leap are the standard interchain wallets, with Cosmostation also established, covering staking, governance voting, and IBC transfers across more than 50 integrated Cosmos chains. Ledger hardware support is established. Chain registry metadata makes a new chain addable without building a wallet from scratch. | Adequate. SubWallet, Talisman, and Polkadot.js are the standard Substrate wallets, and they support solo chains as well as parachains. The ecosystem is real and maintained, but smaller and less familiar to the general crypto user than the Cosmos wallet set. | Weak. No wallet ecosystem comes with the framework. OGO must build wallet support, signing, address formats, and hardware wallet integration itself. This is the largest hidden cost of this option. |
| Developer experience | Strong for OGO's current position. Go is the language of the verified scaffold, the team already has a working Go build, test, and reproducible-build loop, and SDK modules are documented with a large body of production code to copy. Go compile times for the full Gaia module graph are slow (first `go mod tidy` about 13 minutes, cold-cache build about 7 minutes in Milestone B), but incremental builds are fast. | Adequate, with a steeper start. Rust gives stronger compile-time guarantees, and FRAME pallets are a clean module system. But OGO has no verified Rust scaffold, no Rust build loop, and no in-house Substrate experience evidenced in the record. Rust compile times for a full Substrate node are long, and quarterly stable releases carry many breaking changes (87 in stable2609-rc1). | Adequate for a small Go team, Weak at scale. Go is familiar and the code is fully OGO's. But there is no module system, no documented keeper pattern, and no community to copy from. Every new feature is a fresh design. |
| Maintenance burden | Adequate. OGO inherits SDK, CometBFT, wasmd, and ibc-go maintenance, including security patches, and must track release families. The Gaia replace pin is a concrete example: a single dependency cannot be upgraded in isolation, the matrix moves as one tested unit. A September 2026 CosmWasm sandbox escape (CWA-2026-006, Wasmer compiler, crafted contract could mint tokens) shows why tracking upstream security patches is not optional. | Adequate to Weak. OGO inherits Polkadot SDK maintenance on a quarterly cadence, with frequent breaking changes. The umbrella crate reduces version drift, but each stable release still needs a migration pass. The ecosystem is centred on Parity and the Polkadot relay chain, so solo-chain concerns can lag parachain concerns in upstream priorities. | Weak. OGO owns all maintenance: consensus upgrades, storage migrations, cryptography updates, and every security patch must be written or backported by OGO. |
| Upgrade mechanisms | Adequate to Strong. Coordinated binary upgrades through Cosmovisor (v1.7.3 in the scaffold), with governance-gated upgrade plans, in-place store migrations, and timelocks. Upgrades require validator coordination and a halt or rolling restart. The Hub has executed expedited upgrades (for example a 72-hour vote, Proposal 1055, in September 2026). | Strong. Forkless runtime upgrades are Substrate's signature feature: new WASM runtime code is stored on chain and activated by governance, without a coordinated binary swap for logic changes. Client and consensus changes still need node upgrades. This does not remove governance risk: a forkless upgrade that changes monetary rules is still a monetary change. | Weak. OGO designs its own upgrade mechanism: versioned state, migration code, coordinated binary swaps, and rollback procedure, all from scratch. |
| Interoperability | Strong. IBC is native, light-client verified, and production-proven between Cosmos chains, with 115 plus connected chains as of October 2026. ibc-go v10 supports IBC Classic and IBC v2. IBC Eureka (March 2025) extends IBC toward Ethereum through Solidity IBC Eureka and zero-knowledge light-client proofs. The supply rule (global economic supply equals local supply minus escrow plus vouchers) is already tested at ledger level in Milestone B. No EVM bridge is inherent, and any bridge to a non-IBC chain is third-party trust, the most exploited component class in the industry (over 2.8 billion USD lost since 2022). | Adequate. XCM is native and expressive inside the Polkadot ecosystem, and solo chains can use XCM, bridges, and BEEFY plus MMR proofs for external connections. But XCM does not connect to Cosmos chains or to IBC, so OGO would lose the IBC network effect, the IBC escrow accounting already tested, and the Eureka path to Ethereum. Note: an XCM executor bug in October 2026 let an attacker reduce the recorded issuance of dotUSD on Polkadot Asset Hub without changing underlying balances (about 1.381 million USD in DOT value; the fix had been available since June 2026 but missed the stable-2606 cut-off and shipped in stable-2609). Supply accounting in any cross-chain executor needs the same rigor as on-chain minting. | Weak. No interoperability protocol comes with the framework. IBC could in principle be re-implemented on top of ABCI, but ibc-go depends on the SDK, so OGO would be writing a light-client protocol, packet handling, and escrow accounting from scratch, or relying on a trusted third-party bridge. |
| Long-term operating requirements | Adequate. Validator operations follow the researched Cosmos baseline: 4 validators minimum for liveness, at least 50 validators for the proposed sufficiently-decentralized bar, standard infrastructure at 200 USD per validator per month, sentry nodes, remote signing through cosmos/kms, monitoring, snapshots, and upgrade response. The operator and tooling pool (validators, relayers, indexers, explorers) is the largest of the three options. | Adequate. Validator operations are comparable in kind, and the same fee-revenue caution applies. The operator pool for solo Substrate chains is smaller than for Cosmos chains, and most Substrate operational knowledge is parachain-shaped (collators, relay chain, coretime), which does not map one to one onto a solo chain. | Weak. Operating requirements are the same in kind plus the permanent cost of maintaining bespoke node software, bespoke tooling, and bespoke documentation. Every operator question becomes an OGO support question. |

Per-criterion summary, in OGO terms:

- Sovereignty does not separate the options. All three give an independent validator set and independent consensus. What separates them is what sovereignty costs. The real sovereignty risk is stake concentration and validator economics (Hub Nakamoto coefficient 3, top 10 about 56 percent, Coinbase 20.56 percent), not framework choice.
- Security: Cosmos leads on evidence. The SDK's staking, slashing, distribution, and governance modules are the same code that secures the Hub, Osmosis, and dYdX. Substrate is also strong, but its relay-chain security (about 300 active validators staking over 700 million DOT, about 58 percent of supply, as of 2026) belongs to the relay chain; a solo chain does not inherit it.
- Native asset: Cosmos is the only option with a verified OGO implementation. Re-implementing the monetary rules in Rust discards 54 passing tests and the localnet proof and re-opens every bypass question in a new language.
- Performance: none of the three is throughput-limited for OGO's realistic early volumes. The binding constraint is fee demand, not TPS. Cosmos Hub collects tens of dollars a day in chain fees; Osmosis collects 372 USD a day in chain fees against 18,562 USD a day in application fees. OGO should not select a framework on vendor TPS claims.
- Wallet ecosystem: Cosmos leads clearly. Keplr and Leap are the default wallets for the interchain, and a new Cosmos chain is added through chain registry metadata. For a network whose founding principle is that people should not need an OGO account to own, transfer, or use OGO Coin, the wallet ecosystem is a first-class requirement.
- Developer experience: Cosmos leads for this team, because the team already has a working Go loop. Substrate's Rust and FRAME are arguably the better-designed module system, but they require a new language, a new build loop, and a new mental model, with no verified OGO work to build on.
- Maintenance: Cosmos and Substrate are comparable in kind and both are real commitments. Purpose-built Go is the worst, because it replaces shared upstream maintenance with sole OGO maintenance.
- Upgrades: Substrate wins outright. Forkless WASM runtime upgrades are a genuine architectural advantage. But OGO's monetary logic is designed to change almost never (fixed schedule, constitutional process only), so the advantage is real but narrow.
- Interoperability: Cosmos wins decisively. IBC is the only protocol in this comparison that is native, light-client verified, production-proven at scale, and already partially implemented in OGO's own code.
- Long-term operations: Cosmos leads on the size of the operator and tooling pool.

### 2.4 The rejected option: Polkadot parachain

A Polkadot parachain is explicitly NOT the default, and this plan rejects it as the deployment form, for one reason that outweighs its benefits. A parachain would give OGO shared security from the Polkadot relay chain, forkless upgrades, XCM interoperability, and a mature slot and coretime market. Those are real benefits.

They are rejected because OGO's stated goal is its own independent validator security and consensus. A parachain does not have its own validator set securing its own consensus: block production is by collators, finality and security come from relay-chain validators, and the chain's liveness and upgrades depend on relay-chain governance, coretime availability, and Polkadot's own roadmap. That is shared security, not sovereign security, and it contradicts the founding architecture and the decision-preparation framing. Secondary reasons reinforce the rejection but do not drive it: ongoing coretime costs, competition for relay-chain attention, inherited Polkadot governance risk, and a migration liability if OGO later decided to leave for full sovereignty. If Benross later reverses the sovereignty goal in writing, a parachain should be re-evaluated on its merits. Until then it is out of scope as a deployment form. Substrate as a solo chain, evaluated above, is a different option and is not rejected on this ground.

### 2.5 Recommendation

Recommendation: Cosmos SDK plus CometBFT, continuing from the verified Milestone B scaffold. This is a recommendation, not a decision. The framework decision stays with Benross, and the proof of concept below must pass before final selection.

Why Cosmos, in order of weight:

1. Verified work already exists, and it is the hard part. The monetary module, the fixed-supply and schedule-only mint rules, the bypass tests, the IBC escrow accounting, and the reproducible build are done and passing on this stack. Changing frameworks discards the 54 tests, the localnet proof, and the build loop, and re-opens every monetary invariant in a new language or a new hand-built store.
2. IBC is a requirement, and only Cosmos has it natively. The escrow and voucher conservation rule is already tested in OGO's own code.
3. The wallet ecosystem is a requirement, and Cosmos has the largest one. Keplr, Leap, and Cosmostation give OGO Coin a path to users without an OGO account.
4. Epoch-based periods are the Cosmos-native path. The recommendation assumes a Cosmos stack with an epochs module (Osmosis x/epochs style) or the SDK x/mint epoch path. On Substrate, the equivalent would be a custom session or era hook with no epochs-module precedent. On purpose-built Go, it would be a hand-built timer with the same timestamp and halt questions and no module to copy.
5. Validator operations and security are researched in Cosmos terms. Moving frameworks would require re-researching the equivalent facts for a new stack before any validator plan could be written.
6. The team and the toolchain are already on Go. Substrate would require a Rust hiring and training plan with no evidenced in-house capacity.

What Substrate does better, and why it does not change the recommendation: forkless runtime upgrades are genuinely better than Cosmovisor-coordinated binary upgrades, and Rust's type system and FRAME pallet model are well designed. If OGO's dominant requirement were frequent logic upgrades, or if OGO already had a Rust team, Substrate would be the recommendation. Neither is true. Substrate remains the credible fallback. If the proof of concept fails on Cosmos for a reason structural to the SDK or CometBFT (not a Milestone C wiring gap), the next step is a time-boxed Substrate spike that re-implements x/monetary as a FRAME pallet and re-runs the bypass suite, not a full rewrite. That spike is not authorized by this plan; it is named here so the fallback is concrete if it is ever needed.

What purpose-built Go does better, and why it is rejected as the default: a lean hand-built state machine could in principle be faster and smaller than a full SDK node, and it gives OGO total control. That control is the problem: total control means total responsibility for staking, slashing, governance, fees, storage, upgrades, wallets, and interoperability, with no upstream, no operator pool, and no audit precedent. Performance is not OGO's binding constraint, so the one advantage does not pay for the costs.

Conditions on the recommendation:

- Milestone C must wire the standard keepers (auth, bank, staking, gov, distribution, fees, wasmd, ibc-go) with full depinject, replace JSON transactions with signature-verified transactions, instantiate the wasm VM, and wire live IBC relaying. Until that is done, the scaffold is a monetary prototype, not a chain.
- The block-based schedule in the scaffold must migrate to epoch-based periods, with the period identifier, duration D, and nominal year fixed in genesis and changeable only through the constitutional process.
- The SDK 0.55 and CometBFT 0.40 upgrade must be evaluated as a separate decision after Milestone C passes, not folded into Milestone C.
- The monetary guards in the scaffold (no discretionary mint, no governance schedule change, no wasm mint) must be re-proven after every upgrade, with the bypass suite re-run, not assumed to survive a dependency bump.

### 2.6 Proof-of-concept specification: what must be demonstrated before final framework selection

A technical proof of concept must validate performance, security, maintenance, and operating requirements before the framework is finally selected. It runs on the Cosmos stack (the recommended framework), in staging, with no public network, no real value, and no monetary parameters selected. All schedule and supply numbers used in it are TEST-ONLY and labelled so. If any MUST criterion fails, final selection is blocked until the failure is fixed or the framework is re-evaluated. SHOULD criteria are recorded as gaps with owners and dates, but do not by themselves block selection if the MUST criteria pass.

Revision note, v1.1, 2026-11-06: the proof of concept has 21 MUST criteria (P1 to P5, S1 to S8, M1 to M5, O1, O2, O5). O3 (IBC smoke test) and O4 (Keplr or Leap wallet smoke test) were removed from the framework-selection criteria by Benross's decision (Phase 2 Execution Directive, 2026-11-06). Keplr and IBC are separated from framework-selection tests and are not prerequisites for the initial monetary prototype unless a specific technical dependency demonstrates necessity. They are evaluated separately when that dependency arises. The O3 and O4 rows below are retained as the definitions of what those tests must show when they happen, marked DEFERRED, not deleted. The earlier figure of 26 MUST criteria was a count error; 23 was correct while O3 and O4 were included; 21 is correct now.

Scope and setup:

- Base: the Milestone B scaffold, plus Milestone C keeper wiring (auth, bank, staking, gov, distribution, fees, wasmd, ibc-go) with depinject, signature-verified transactions, instantiated wasm VM, and live IBC relaying to a second local chain.
- Chain id: a fresh localnet id, not `ogo-local-1` reused with different genesis, so results are not confused with the Milestone B proof.
- Epochs: x/epochs module wired (Osmosis style) or SDK x/mint epoch path, with a short TEST-ONLY epoch duration (for example 60 seconds) so epoch behaviour can be observed in hours, not months. The production duration D remains TBD and is not selected by this proof of concept.
- Validators: multi-validator localnet, minimum 4 validators for the MUST pass, with 7 validators as the SHOULD target, each on a separate home directory and separate key, with at least 2 sentry nodes for one validator to prove the sentry pattern. Revision note, v1.1, 2026-11-06: this is the proof-of-concept network, which is a separate run from the M1 localnet proof. M1 runs 3 validators by Benross's decision (Phase 2 Execution Directive). DECIDED by Benross, 2026-11-06 (Option 1): the proof-of-concept run uses a separate 4-validator network, in addition to the unchanged 3-validator local prototype; S5 is NOT redefined as a halt-and-resume demonstration. The separate 4-validator proof of concept ran on 2026-11-06 (`ogo-network-poc-4validator/`, chain-id `ogo-poc-4val-1`): S5 liveness demonstrated (blocks continued with 1 of 4 offline, restart and sync verified), P1 partially executed (24-hour duration not completed in one session), M2 not executed (no Cosmovisor or upgrade-handler support in the scaffold), O1 partially executed (runbook and snapshot/restore executed; no sentry topology on a single machine). Full results in `ogo-network-poc-4validator/ogo-network-poc-4validator-report.md`. P1, S5, M2, and O1 as written below need 4 validators and are not claimed as passed by the 3-validator local prototype.
- Duration: the proof of concept runs until every criterion below has a recorded result, with at least one continuous 24-hour run for stability.

MUST criteria, performance:

| ID | Demonstration | Pass criterion | Fail means |
|---|---|---|---|
| P1 | Multi-validator block production | 4 validators produce blocks continuously for 24 hours, with no halt longer than 5 minutes except a planned halt test, and block time stays within 1 to 10 seconds for at least 95 percent of blocks | Chain cannot sustain liveness with a real validator set |
| P2 | Transaction throughput, simple transfers | Sustained at least 200 simple bank transfers per second for 10 minutes, measured end to end from submission to commit, with zero failed transactions attributable to the chain | Throughput is below the floor OGO needs for early volumes |
| P3 | Transaction throughput, mixed load | Sustained at least 50 transactions per second for 10 minutes under a mixed load (transfers, staking operations, governance votes, and one CosmWasm contract call per block on average), with p95 finality under 10 seconds | Mixed execution, not just transfers, is the bottleneck |
| P4 | Finality and latency | p95 time from transaction submission to final commit is under 6 seconds for simple transfers during the P2 run | User-perceived latency is unacceptable for payments |
| P5 | State growth and sync | A new node joins the 4-validator localnet at height 10,000 or later, syncs from genesis or from a snapshot, and catches up to tip within 2 hours on the documented standard hardware baseline (4 plus CPU cores, 32 GB RAM, 500 GB NVMe) | State growth or sync time makes operating a validator impractical |

Thresholds P2 to P4 are proof-of-concept floors, not production promises and not capacity claims. They are set low deliberately: if Cosmos cannot clear 200 TPS on simple transfers, something is wrong with the wiring, not with the framework.

MUST criteria, security:

| ID | Demonstration | Pass criterion | Fail means |
|---|---|---|---|
| S1 | Signature verification | A transaction naming a sender it is not signed by is rejected before execution, and the rejection is logged with the reason. The Milestone B gap (unsigned JSON transactions accepted) is closed and proven closed. | Authentication is not enforced |
| S2 | Monetary bypass suite, re-run | All 54 Milestone B tests pass on the wired chain, plus new tests for the wired keepers: no discretionary mint through bank, staking, distribution, gov, or wasm; no governance change to max supply or schedule; no wasm mint. Zero bypasses succeed, and state is unchanged after each attempt. | A mint or supply path exists outside the schedule |
| S3 | Epoch issuance, exactness | Issuance for each epoch equals exactly R_n for that epoch number, is claimed exactly once per epoch, and total issued never exceeds total scheduled. The issued-by-epoch set is keyed by epoch number, and a re-claim of the same epoch fails. Mandatory depinject hook-fires test, release-blocking for the monetary prototype: in the wired application, with depinject, advance the chain past an epoch boundary and confirm all five of the following, in order. (1) The epoch module initializes in the wired application. (2) The monetary hook is registered: the x/monetary `AfterEpochEnd` hook is present in the wired hook set, not only in a keeper unit test. (3) The hook actually fires at an epoch transition: issuance is minted and the epoch event is emitted when the boundary is crossed. (4) Exactly one scheduled issuance occurs for that epoch: the minted amount equals exactly R_n for the epoch, and the issued-by-epoch set contains the epoch exactly once. (5) A duplicate hook execution cannot issue twice: invoking the hook a second time for the same epoch mints nothing and leaves total issued unchanged. Compilation alone is not an acceptable test, and a keeper-level unit test alone is not an acceptable test: the failure this test exists to catch is a hook that is wired but never fires in the deployed application. This test exists because of the depinject hook-registration bug (issues #24770, #25425). x/epochs is a first-class Cosmos SDK module from v0.53.0, upstreamed through PR #23815. | Scheduled issuance is not exact or not once-only |
| S4 | Epoch halt and catch-up | Halt the chain for 30 minutes (no blocks), restart, and observe missed epochs ticking one per block on restart, in epoch order, with no lump mint, no skipped epoch, and no double mint. Total issued after catch-up equals the sum of R_n for all elapsed epochs. | Halt behaviour breaks the schedule or the issued set |
| S5 | Slashing and liveness | Take 1 of 4 validators offline for longer than the downtime window, observe downtime jailing and the documented slashing, and observe the chain continuing to produce blocks with the remaining 3 validators. Then unjail and rejoin. | Liveness or slashing does not work as configured |
| S6 | Double-sign protection | Attempt a double-sign on a test validator (or simulate the equivocation evidence path if live double-signing is unsafe in the harness), and observe tombstoning or the configured double-sign penalty, with the evidence recorded. | Equivocation is not punished |
| S7 | Wasm sandbox | Deploy a test CosmWasm contract that attempts a mint, a bank send above its balance, and an unknown method call. All three fail, state is unchanged, and a legitimate bank send from the contract succeeds and preserves total supply. This closes the Milestone B limitation (VM not instantiated) with live execution evidence. | Contract execution can mint or overspend |
| S8 | Key management | At least one validator runs with a remote signer (cosmos/kms v0.1.0, or the documented equivalent for the 0.53 matrix), with the consensus key never present as a plaintext file on the validator host. | Validator keys are not protected to the documented standard |

MUST criteria, maintenance and upgrades:

| ID | Demonstration | Pass criterion | Fail means |
|---|---|---|---|
| M1 | Reproducible build, re-proven | Clean-copy build from a fresh module cache exits 0, vet 0, tests pass, and the binary hash matches the documented hash under the canonical GOMODCACHE condition from Milestone B. Any hash difference is investigated and explained, not ignored. | The chain cannot be independently reconstructed |
| M2 | Coordinated upgrade via Cosmovisor | Plan a governance upgrade, run it through Cosmovisor on all 4 validators, with a store migration that adds one new state field, and observe: upgrade height reached, binaries swapped, state migration applied, chain resumes and produces blocks, balances and total supply are identical before and after, and total downtime is under 30 minutes. | Upgrades cannot be executed safely |
| M3 | Upgrade guard, re-proven | After the M2 upgrade, re-run the S2 bypass suite. A governance upgrade attempting to change max supply or schedule is still rejected, and a non-monetary upgrade is still accepted. | The monetary guard does not survive upgrades |
| M4 | Dependency upgrade dry run | Document the effort to move the matrix from the 0.53 family to the 0.55 family (SDK v0.55.0, CometBFT v0.40.0) in a branch, with build, vet, and test results recorded. Pass is a completed dry run with a written effort estimate, not a successful upgrade. | Upgrade cost is unknown |
| M5 | Rollback | After a failed upgrade attempt on one validator (wrong binary or bad migration, in staging), roll that validator back to the prior binary and state, rejoin, and catch up to tip. Record the exact steps and time taken. | There is no tested rollback path |

MUST criteria, operating requirements:

| ID | Demonstration | Pass criterion | Fail means |
|---|---|---|---|
| O1 | Multi-validator operation | 4 validators operated as if independent (separate homes, keys, and monitoring), with at least one validator behind sentries, and a written runbook for start, stop, upgrade, snapshot, and restore, each step executed at least once. | Operations are not documented and not exercised |
| O2 | Monitoring and alerting | Block height, block time, missed blocks per validator, peer count, mempool size, and total supply are visible on a dashboard or log query for the full 24-hour run, with an alert fired and recorded for at least one missed-block event. | The chain cannot be observed in operation |
| O3 | IBC or bridge smoke test | DEFERRED by Benross's decision, 2026-11-06 (Phase 2 Execution Directive). The test as written (open an IBC channel to a second local Cosmos chain, transfer a TEST-ONLY amount, and verify: escrow on the source equals the transferred amount, voucher on the destination equals the transferred amount, global economic supply (local minus escrow plus vouchers) is unchanged, and a return transfer releases the escrow exactly) is retained as the definition of what a live IBC test must show, but it is NOT a framework-selection gate and NOT a prerequisite for the initial monetary prototype. IBC is evaluated separately if a specific technical dependency demonstrates necessity. | Interoperability is not proven, even at ledger level plus live relaying |
| O4 | Wallet smoke test | DEFERRED by Benross's decision, 2026-11-06 (Phase 2 Execution Directive). The test as written (a Keplr or Leap wallet, or the chain-registry metadata for one, connects to the localnet, displays a TEST-ONLY balance, signs a transfer, and the transfer commits with the correct balances) is retained as the definition of what an external-wallet compatibility test must show, but it is NOT a framework-selection gate and NOT a prerequisite for the initial monetary prototype. Keplr is evaluated separately if a specific technical dependency demonstrates necessity. The wallet product is built in M4, where O4 is re-tested at product level. | Users cannot hold OGO Coin without an OGO account |
| O5 | Recovery from data loss | Stop one validator, delete its data directory, restore from a snapshot, and rejoin and catch up to tip within 2 hours. Record the snapshot size and time. | Validator recovery is not proven |

SHOULD criteria (recorded, not blocking): 7 validators instead of 4, with stake deliberately uneven (largest validator under one third of voting power); a 7-day continuous run; a load test with 1,000 TPS simple transfers to find the actual ceiling; a second IBC counterparty; a governance proposal executed end to end (submit, vote, pass, execute a non-monetary parameter change); a treasury multisig simulation if the treasury module exists by then; an upgrade dry run to SDK 0.55 actually completed if M4 shows the effort is small.

What the proof of concept does NOT demonstrate: no monetary parameters are selected or validated (max supply, schedule, split, reward percentages remain TBD); no fee-market assumption is validated (only the High fee scenario covers the minimum need, and no scenario replaces launch-level validator pay); no legal, securities, or money-transmission question is answered; no public testnet or mainnet readiness is claimed. Passing the proof of concept authorises a framework selection, not a launch.

Risks carried forward: the Gaia replace pin means the matrix moves as one unit, and any future wasmd or ibc-go upgrade must be tested against the whole matrix. The September 2026 CosmWasm sandbox escape (CWA-2026-006) and the February 2026 CometBFT median-time vulnerability (GHSA-c32p-wcqj-j677, fixed in CometBFT v0.38.21) show upstream security patches are a permanent operating cost; the scaffold's CometBFT v0.38.23 already includes the median-time fix. Epoch-based periods depend on the epochs module being available and maintained in the chosen SDK version; the proof of concept must confirm this (S3 and S4). Validator economics remain the binding constraint, not framework choice. No live Cosmos chain in the evidence meets all four sufficiently-decentralized thresholds at once; the framework does not solve distribution, and a mechanism for a flat stake distribution still needs to be named.

## 3. OGO Network Protocol Specification

### 3.1 Runtime stack and dependency matrix

The stack is the verified Milestone B matrix, unchanged unless a milestone acceptance criterion requires a change. Every component moves as one tested unit, and no single dependency may be upgraded in isolation.

| Component | Version | Notes |
|---|---|---|
| Go | 1.25.7 | Locked prototype toolchain. Go 1.23 was the blueprint proposal; Milestone B actually used 1.25.7 and the authorization locks it. |
| Cosmos SDK | v0.53.6 required, v0.53.4 effective through Gaia's replace | Gaia v28.3.1 replace pin, preserved because wasmd calls in SDK v0.53.6. Full Gaia go.mod hash: `72b985a58a88efea8193e2bf77be47b51f227da0055427cc57dd21877c3db872`. |
| CometBFT | v0.38.23 | Includes the median-time fix (GHSA-c32p-wcqj-j677, fixed in v0.38.21). CometBFT v1.x is retracted and is not a target. |
| CosmWasm (wasmd) | v0.60.9 | wasmvm/v2 v2.3.5, CGO path embedded in the binary (see reproducible-build condition in section 2.1). |
| IBC | ibc-go/v10 v10.7.0 | IBC Classic plus IBC v2. 08-wasm v10.5.0. |
| Cosmovisor | v1.7.3 | Separate binary, for coordinated binary upgrades. |
| Reference implementation | Gaia v28.3.1 | Source of the dependency matrix. |
| Upgrade evaluation only | SDK v0.55.0, CometBFT v0.40.0 | Ledger Security 2026.1 family. Separate evaluated upgrade after Milestone C. Not a build target for the first localnet. |

Do not adopt SDK v0.53.8 or Gaia v29 for the first localnet: multisig `--gas auto` is broken in SDK v0.53.8. The 0.53.8 issue is tracked for the upgrade decision, not worked around by skipping it silently.

### 3.2 Module layering

The chain is organised as four layers: consensus and networking (CometBFT), the application framework (Cosmos SDK), the core economic modules (standard plus custom), and the application-facing interface (RPC, LCD, indexing).

The layering principle: use the standard module wherever it satisfies the requirement, and build a custom module only where a standard module cannot express OGO's rule. This is how the design stays close to upstream, so security patches and upgrades remain applicable.

### 3.3 Standard modules, and what is used for what

| Module | Role in OGO Network | Key parameters (proposal, TBD) |
|---|---|---|
| x/auth | Accounts, transaction authentication, signature verification, sequence handling. The scaffold's unsigned JSON transactions must be replaced by signed, sequence-checked transactions through x/auth. | Standard defaults until a parameter is proposed. |
| x/bank | Native denom balances (`uogo`), sends, multi-sends, supply queries, burn. Total supply, circulating supply, and issued totals are derived from x/bank plus x/monetary state. | Denom `uogo`, 6 decimals (proposal only). |
| x/staking | Validator set, bonding, delegation, unbonding, redelegation, validator signing info. | Unbonding 21 days (Hub precedent, proposal). Commission floor 5 percent (Hub precedent, proposal). See section 5. |
| x/slashing | Downtime and equivocation evidence, jailing, tombstoning. | Downtime 0.01 percent plus 10 minute jail; double-sign 5 percent plus tombstoning (Hub precedent, proposal). |
| x/distribution | Validator and delegator reward distribution, community pool accounting. | x/monetary owns the issuance schedule; x/distribution owns the split of what is minted into validator, delegator, and community shares, per the approved split. |
| x/gov | Governance: proposals, deposits, voting, tally, execution. Includes expedited proposals if adopted. | Thresholds, deposit, voting period, quorum: all TBD (section 11). |
| x/upgrade | Upgrade plans, store migration coordination, upgrade handlers. | Upgrade plans must pass through x/gov and be executed through x/upgrade, never through a direct binary swap on a live network. |
| x/epochs | Named timers for epoch-based periods. This is the module that makes section 3.5 work. | Epoch duration D (TBD), epoch identifiers, first epoch start height. |
| x/evidence | Equivocation evidence handling, feeding x/slashing. | Standard defaults. |
| x/params | Parameter storage (where retained in SDK 0.53). | Used only through the standard module parameter stores, not as a back door for monetary rules. |
| ibc-go | IBC client, connection, channel, transfer, and the 08-wasm client. IBC v2 support included. | Standard defaults; custom escrow accounting rules live in x/monetary, not in ibc-go. |
| wasmd | CosmWasm smart contracts. | Wasm VM must be instantiated (it was not in the scaffold). Binding surface frozen at [BankSend, BankBurn, QueryBalance, QuerySupply]. No mint method ever. |

### 3.4 Custom modules: only three

Three custom modules are justified. No other custom module is created at this stage.

x/monetary (production name; extends and renames the scaffold's x/ogocoin). This module owns the supply cap, the issuance schedule, the mint operation, and supply accounting. It is the only module that can mint. It enforces: max supply is a mandatory genesis parameter with no default; the issuance schedule is a mandatory genesis parameter, fully specified; minting happens only through the schedule, only at the exact scheduled amount, only once per epoch; the schedule cannot be changed by ordinary governance; the cap is enforced on every mint attempt; dust is left permanently unissued. It exposes the schedule, the issued-by-epoch set, total issued, total supply, and total supply after all periods as queries, so anyone can verify the rules without trusting OGO Systems. The scaffold's x/ogocoin is the prototype of this module; the rename to x/monetary is a naming decision for the production module, and the scaffold code is its ancestor, not a separate implementation to maintain in parallel.

x/treasury. This module owns treasury accounting and the spending rule. It implements the R-F rule (recommended default, section 4.4), the emergency reserve floor, the mode transition from issuance to post-issuance, and treasury spending proposals through x/gov. It does not mint. It does not hold coins that are not accounted for in x/bank. It enforces that treasury spending cannot exceed the rule, and that the rule itself cannot be changed except through the constitutional process.

x/rewards. This module owns per-epoch and per-block split accounting: how the scheduled issuance for an epoch is divided between validators, delegators, and the community pool or treasury, per the approved split, and how transaction fees are distributed between the same recipients. It does not mint and does not decide the split; it executes the split. The split parameters are governance-controlled (section 4.6), but the split must always conserve: the sum of the parts equals the amount minted plus the fees collected, exactly, in integer arithmetic.

Explicitly NOT created, with reasons:

- x/governance is NOT created. x/gov plus the standard proposal types already provide governance. A separate governance module would duplicate x/gov and create a second, divergent voting path. Governance security rules (parameter separation, upgrade-handler denylist, amount-less mint entrypoint, every-block invariant, public schedule queryability) are enforced inside x/monetary and x/treasury, not in a new module.
- x/fees is NOT created. Fee collection uses the standard SDK fee mechanism (x/auth ante handler plus x/distribution for fee distribution), and fee splitting is handled by x/rewards. A separate fee module would duplicate the ante handler and create a second fee path.
- No x/bridge module. Bridges are third-party trust and are not part of the chain. If a bridge is ever used, it is an external service, not a chain module, and it must be evaluated under the same supply-accounting rules as IBC (escrow and voucher conservation).

### 3.5 Block and epoch architecture

Periods are epoch-based, not block-based. This is the recommendation carried forward from Decision Preparation v0.1 section 1, for two evidenced reasons: block-based periods drift with block time (Solana was 276 days behind its intended inflation timeline by June 2026), and pure calendar time is fragile because block timestamps are a weighted median under validator control (Tachyon, GHSA-c32p-wcqj-j677, February 2026).

Design:

- An epochs module (Osmosis x/epochs style, or the SDK x/mint epoch path if it is available and maintained in the chosen SDK version) provides named timers.
- A period is one epoch of a named timer, with a fixed duration D in seconds, fixed in genesis. The nominal year is D times the number of periods per nominal year, fixed in genesis.
- Issuance for epoch n is R_n, computed from the genesis-fixed schedule. It is claimed once, in the epoch hook, keyed by epoch number. The issued-by-epoch set is the guard: a re-claim of the same epoch fails.
- The first epoch start height and the epoch duration are genesis parameters. The final issuance period N is genesis-fixed. The transition from issuance mode to post-issuance mode is keyed to N, not to a wall-clock date and not to a governance vote.
- Halt behaviour is specified, not assumed. If the chain halts, missed epochs tick one per block on restart, in epoch order, with no lump mint, no skipped epoch, and no double mint. Total issued after catch-up equals the sum of R_n for all elapsed epochs. This is a MUST criterion of the proof of concept (S4).
- The scaffold currently implements block-based periods. Milestone C must migrate the schedule code from block-based to epoch-based periods. This is a schedule-code migration, not a rewrite of the monetary rules.

No duration D is recommended in this plan. Period length is a sustainability decision (geometric ends at period 110, which is 9.2 years monthly or 27.5 years quarterly; hybrid ends at period 40, which is 3.3 years monthly or 10.0 years quarterly). It is an open decision (section 11).

### 3.6 Accounts and address format

Proposed, not decided:

- Bech32 prefix: `ogo` (provisional). Derivation path: `44'/118'/0'/0/0` (provisional, the Cosmos standard coin type 118). Both are proposals in the wallet draft and are listed as open decisions in section 11.
- Account addresses are derived from the public key (standard Cosmos secp256k1). Validator operator addresses use the `ogovaloper` prefix, consensus addresses use `ogovalcons`, following the standard Cosmos prefix pattern with the `ogo` prefix.
- An OGO wallet address is not an OGO username. A Supabase account does not own or control a blockchain balance. A user must be able to access coins even if OGO's website is unavailable. See section 6.
- Account numbers and sequence numbers follow the standard x/auth model. Sequence handling is what prevents replay; the scaffold's unsigned transactions must be replaced by this.

### 3.7 Transaction structure

Standard Cosmos SDK transactions, once Milestone C wires x/auth:

- Messages (MsgSend, MsgDelegate, MsgUndelegate, MsgBeginRedelegate, MsgSubmitProposal, MsgVote, MsgExec, and the custom module messages) are signed by the account that owns the funds or the authority.
- Signatures use secp256k1 by default (DIRECT sign mode, with Amino JSON as the legacy fallback for hardware wallets that require it).
- Fees are paid in `uogo`, collected by the ante handler, and distributed by x/rewards per the approved fee distribution. The fee market is specified in section 5.4. No fee is ever collected in a different denomination at launch.
- Memo field is supported for payment references and is not used for monetary rules.
- Transaction validity is consensus-critical: a transaction that fails signature verification, sequence checking, or balance checking is rejected before execution, and the rejection is deterministic across nodes.

### 3.8 Consensus and finality

- CometBFT Byzantine fault tolerant consensus, Tendermint-style, with instant finality per block. A block is final when committed; there is no probabilistic confirmation depth.
- Block time target: 1 to 6 seconds typical for CometBFT chains, with the proof of concept requiring 95 percent of blocks within 1 to 10 seconds during a 24-hour run. The target is a proof-of-concept pass criterion, not a production promise.
- Validator set: the active set is the top N validators by bonded stake, where N is a genesis parameter (TBD, section 11). The minimum operational set is 4 validators. The sufficiently-decentralized bar is a Nakamoto coefficient of at least 10, top 10 under 33 percent, at least 50 validators, and no country or provider over 33 percent (section 5.2). No live Cosmos chain in the evidence meets all four thresholds at once; this is a target, not a claim.
- Liveness: the chain halts if more than one third of voting power is offline. The proof of concept must demonstrate continued block production with 3 of 4 validators (S5).
- Consensus key rotation: in-place validator consensus key rotation is available in the 0.55 family (SDK v0.55.0, CometBFT v0.40.0). In the 0.53 matrix, key rotation requires the standard unbond and re-bond path. The upgrade evaluation (M4) must record which path applies.

### 3.9 Validator requirements (summary; full specification in section 5)

- Minimum operational set: 4 validators. Sufficiently-decentralized set: at least 50 validators, Nakamoto coefficient at least 10, top 10 under 33 percent, no country or provider over 33 percent.
- Infrastructure baseline (standard tier): 4 plus CPU cores, 32 GB RAM, 500 GB NVMe, 100 Mbps, about 200 USD per month, plus a sentry node architecture and remote signing through cosmos/kms. Lean, standard, and professional tiers are specified in section 5.2.
- Slashing (Hub precedent, proposal): downtime 0.01 percent plus 10 minute jail; double-sign 5 percent plus tombstoning.
- Unbonding: 21 days (Hub precedent, proposal). Commission floor: 5 percent (Hub precedent, proposal).
- Validator economics are the binding constraint, not framework choice. The fee market analysis (section 5.4) shows that only the High fee scenario covers the minimum validator need, and only at period 38 (or period 48 at standard-cost need). This is carried forward as an open problem, not solved by this plan.

### 3.10 Staking and delegation model

- Native staking through x/staking. Delegators bond `uogo` to a validator, earn a share of the validator's rewards per the approved split, and pay the validator's commission.
- Unbonding takes 21 days (proposal). During unbonding, funds earn no rewards and cannot be transferred. Redelegation is allowed once per unbonding period per delegation, following the standard SDK rule.
- Slashing applies to bonded stake. A downtime slash reduces the validator's and its delegators' bonded stake by the slash fraction. A double-sign slash is larger and tombstones the validator permanently.
- Liquid staking is not part of the first monetary prototype. It is an open question for a later phase, and it must be evaluated against the same supply and slashing rules before it is adopted. No liquid staking token is created by this plan.
- Validator set changes are governed by bonded stake ranking, not by appointment. There is no allowlist of validators in the protocol. If a permissioned validator set is ever needed for a testnet phase, it is a testnet configuration, not a protocol rule, and it must be removed before mainnet.

### 3.11 Fee market

Full specification in section 5.4. Summary here: fees are paid in `uogo`, collected by the ante handler, and distributed by x/rewards. The fee market must be designed before genesis, because it is the post-issuance funding mechanism, and the evidence shows the fee transition has not happened on any of the four comparison networks. The fee policy is an open decision (section 11). A flat fee cannot serve both security and agent viability at the same time: agents are 45 percent of transactions but only 7.7 percent of fee revenue in the High scenario, so any fee policy must state explicitly how agent transactions are priced. Revision note, v1.1, 2026-11-06 (finding c): agent micropayment viability is an explicit protocol design requirement, not an application-layer afterthought, and agent-native batching is a core protocol consideration, not an afterthought. The fee architecture must explicitly distinguish ordinary transactions from agent micropayments. The headline case: at a 1,000 uogo fee, a 0.001 OGO payment costs 100 percent in fees, while batching 100 payments into one transaction makes it 0.10 percent per payment. The fee transition requires about one million transactions per day at the modeled fee level of 1,000 uogo to cover the standard validator need (100 validators, standard cost); at the 100 uogo fee level that keeps small payments viable, the same need requires about 8,200,000 transactions per day. Fee sponsorship is evaluated as an additional mechanism: a third party pays the transaction fee on behalf of the agent, so the fee burden leaves the micropayment entirely, but the fee itself is still paid in full and validators are still compensated; sponsorship needs caps, revocation, and abuse controls, and a treasury-funded sponsorship program is a treasury spend subject to the R-F rule. Batching, aggregation, and fee sponsorship are evaluated in the fee market specification, section 7.

### 3.12 Governance system

- Governance runs through x/gov: proposals, deposits, voting, tally, execution. Proposal types include text, parameter change, software upgrade, community pool spend, and treasury spend (through x/treasury).
- What governance CAN do: approve upgrades (through x/upgrade), change governance-controlled parameters (section 4.6), approve treasury spending within the R-F rule, and approve community pool spending.
- What governance CANNOT do: change max supply, change the issuance schedule, mint outside the schedule, change the final issuance period N, or change the treasury rule except through the constitutional process. These are enforced in code, not by convention. The governance security boundary is specified in section 10.3.
- Voting thresholds, deposit, voting period, and quorum are all TBD (section 11). The Hub's expedited proposal path (72-hour vote, Proposal 1055, September 2026) is a precedent worth evaluating for emergency upgrades, but it is not adopted by this plan.
- Hard-fork honesty: no chain can promise its community will never adopt different software. The protection is transparent rules, independent operators, and a visible upgrade process, not a claim that a hard fork is impossible.

### 3.13 Upgrade mechanism

- Coordinated binary upgrades through Cosmovisor (v1.7.3). An upgrade plan is submitted as a governance proposal, voted on, and executed at a fixed height. Cosmovisor swaps the binary at that height, and the upgrade handler applies any store migration.
- Store migrations are in-place and versioned. Each migration is tested on the proof-of-concept chain before it is proposed for a live network.
- An upgrade handler must contain a denylist gate: it rejects any upgrade that would modify monetary state (max supply, schedule, issued-by-epoch set, treasury rule) outside the constitutional process. See section 10.3.
- Rollback is a tested procedure, not an assumption. The proof of concept must demonstrate rolling back a single validator after a failed upgrade (M5).
- The SDK 0.55 upgrade is evaluated separately after Milestone C, with a documented effort estimate (M4). It is not folded into Milestone C.

### 3.14 Treasury architecture

- The treasury is a module account in x/bank, owned and governed by x/treasury, not by any single key and not by OGO Systems. Treasury funds are on-chain `uogo`, visible in supply queries.
- Funding: a share of scheduled issuance, per the approved split, plus a share of transaction fees, per the approved fee distribution. The split is TBD (section 11). The treasury inherits the issuance curve, so it is front-loaded under geometric and more even under hybrid, and it stops receiving issuance funding at period N.
- Spending: through the R-F rule (section 4.4), executed through x/gov proposals. Spending cannot exceed the rule. The rule itself cannot be changed except through the constitutional process.
- Emergency reserve: a floor F, set in genesis, below which the treasury cannot spend except through a separate emergency process that is itself documented before genesis. The value of F is TBD.
- Post-issuance funding: after period N, the treasury receives only its share of transaction fees plus any other explicitly disclosed funding mechanism named before genesis (fee share, blockspace or coretime sales, endowment yield, contributions). Treasury staking is noted as circular and is not a funding mechanism. See section 4.5.

### 3.15 IBC and interoperability

- IBC through ibc-go v10, with the 08-wasm client for contract-aware IBC. IBC v2 support is included in the matrix.
- Supply rule (Benross, 2026-11-05): IBC transfers of native OGO between chains must preserve global supply accounting. A native token escrowed on one chain and represented as a voucher on another must not be counted as two independent units of economic supply. Global economic supply equals local supply minus escrow plus vouchers, and this is enforced in x/monetary, not in ibc-go.
- Ledger-level escrow and voucher accounting is already tested in Milestone B (6 IBC tests). Live IBC relaying is not yet wired. Revision note, v1.1, 2026-11-06: the live IBC demonstration (O3) is DEFERRED by Benross's decision (Phase 2 Execution Directive). It is not a framework-selection gate and not a prerequisite for the initial monetary prototype; IBC is evaluated separately if a specific technical dependency demonstrates necessity. The ledger-level escrow and voucher accounting remains tested and required.
- No EVM bridge is part of the chain. Any bridge to a non-IBC chain is third-party trust and must be evaluated under the same supply-accounting rules. Bridges are the most exploited component class in the industry (over 2.8 billion USD lost since 2022). No bridge is adopted by this plan.

### 3.16 Smart contracts (CosmWasm)

- CosmWasm through wasmd v0.60.9, with the wasm VM instantiated (it was not in the scaffold). Contracts are evaluated for future programmable applications and agent services, but CosmWasm is NOT a dependency of the first monetary prototype (M0 to M3). See section 9.4 for the evaluation criteria and sequencing.
- Binding surface frozen at [BankSend, BankBurn, QueryBalance, QuerySupply]. No mint method ever. A contract that attempts a mint, a bank send above its balance, or an unknown method call fails, and state is unchanged. This is a MUST criterion of the proof of concept (S7).
- The September 2026 CosmWasm sandbox escape (CWA-2026-006, Wasmer compiler, a crafted contract could mint tokens) is the reason the binding surface is frozen and the wasm VM must be kept patched. The scaffold's wasmd v0.60.9 must be checked against the CWA-2026-006 patch status before Milestone C begins.

### 3.17 Wiring remaining from the scaffold (Milestone C)

The scaffold is a monetary prototype, not a chain. The following must be done before the first proof (M1) can be claimed as a chain, not just a schedule test:

1. Full SDK depinject wiring of all standard keepers (auth, bank, staking, gov, distribution, fees, wasmd, ibc-go), replacing the scaffold's blank imports.
2. Replace JSON transactions with signature-verified transactions through x/auth.
3. Instantiate the wasm VM and prove the binding surface with live contract execution (S7).
4. Wire live IBC relaying to a second local chain (O3, DEFERRED by Benross's decision, 2026-11-06; evaluated separately if a specific technical dependency demonstrates necessity, not a prerequisite for the initial monetary prototype).
5. Migrate the schedule from block-based periods to epoch-based periods (section 3.5).
6. Rename and extend x/ogocoin into x/monetary, and build x/treasury and x/rewards as specified in section 3.4.
7. Replace plaintext localnet keys with the documented key management path (cosmos/kms remote signer, S8).
8. Add the every-block supply invariant check (section 10.3, layer 4).
9. Add the upgrade-handler denylist gate (section 10.3, layer 2).

Each item has a test or a proof-of-concept criterion. None is assumed done until its criterion passes.

## 4. Native Coin Technical Specification

### 4.1 Denomination

Proposed, not decided:

- Denom: `uogo` (micro-OGO). Decimals: 6, so 1 OGO equals 1,000,000 uogo. Both are proposals from the blueprint and the wallet draft, listed as open decisions in section 11. No denom or decimal is final until Benross approves it.
- Display: the wallet and explorer display OGO with 6 decimal places. The chain stores and computes in `uogo` integers only. No floating point appears in any consensus-critical monetary code.
- The denom and decimals are genesis parameters. They cannot be changed after genesis without a hard fork, so they must be decided before genesis, not after.

### 4.2 Supply accounting: three numbers, not one

Three numbers must be distinguished at all times, because confusing them is how supply is misreported. This is the three-number accounting rule, carried forward from the Decision Preparation and Hybrid Terminal reports as a recommendation.

- S_target: the theoretical target for total scheduled issuance, the analytical number used in model comparison (1,000,000 illustrative units in the research). This is a design input, not a chain value.
- S_actual: the actual cumulative scheduled issuance, computed by exact integer summation of the floored per-period amounts. S_actual is always less than or equal to S_target, and the difference is dust.
- M: maximum supply, defined as M = G + S_actual, where G is genesis circulating supply. M is computed at genesis by exact integer summation, after the schedule is fixed. M is not chosen first and then fitted to a schedule; it is derived from G and the schedule.

Dust: integer flooring leaves dust unissued. In the research, dust is 45 to 118 uogo per 1,000,000 normalized units, depending on the model. The standing recommendation, not a decision, is to leave dust permanently unissued. Dust must never become an alternative minting allowance, a rounding buffer, or a discretionary top-up. If a future proposal ever tries to issue dust, it must go through the constitutional process, not through an ordinary governance proposal.

All three numbers are exposed as chain queries, so anyone can verify them without trusting OGO Systems: total supply, total issued, total scheduled, dust remaining, and max supply.

### 4.3 Issuance schedule: the two candidates

No model is selected. Geometric and hybrid remain the two candidates. The schedule parameters (reduction factor, floor, period count, initial amount) are all TBD. This section records what each candidate is, so the decision can be made on the evidence.

Geometric decay (primary candidate). Each period's issuance is the previous period's issuance multiplied by a reduction factor r (less than 1), floored to an integer. In the research, r equals 0.90 gives a terminal period of 110, with the last period at or above 10 percent of the initial amount at period 21, and dust of 65. r equals 0.85 gives a terminal period of 74 and dust of 45. r equals 0.95 gives a terminal period of 211 and dust of 118. A 0.05 change in r more than doubles the horizon, so r is the most sensitive parameter in the design. The end rule recommendation, not a decision, is to specify a fixed final period in genesis so the end date is predictable. Without a fixed final period, geometric decay has no fixed end date, which is the Cardano warning: a declining reserve with no landing point.

Hybrid (leading alternative). A fixed absolute base rate (the floor) is paid for a fixed number of periods N, plus a geometric bonus that decays. In the research, with N equals 40 and r equals 0.90, a base share of 40 percent (10,000 per period) gives an initial amount of 70,900, 50 percent issued by period 11, 90 percent by period 32, issuance never below 14.1 percent of the initial amount while running, and a 10 percent horizon at period 39. A base share of 0 percent is truncated geometric. A base share of 60 percent (15,000 per period) gives an initial amount of 55,600 and a first-10-periods share of 41.4 percent. The floor is a fixed absolute base rate for a fixed number of periods ending at period N. A floor as a share of remaining reserve never ends; a fixed absolute floor cannot be paid past the end date. This distinction matters, because the two definitions have opposite terminal behaviour.

Terminal behaviour, the question Benross asked to be compared before any decision: an abrupt end (base floor directly to zero) is a reward cliff. The final payment in the research is 11,000. A short linear taper is worse than abrupt (drops up to 35.7 percent). A gentle geometric taper, factor about 0.95 over the final 20 periods, halves the final payment to about 5,000, with a maximum drop of 6.5 percent. The taper never removes the 100 percent end; it shrinks it. The gentle taper is carried forward as a candidate if hybrid is chosen. It is not a decision.

Epoch-based (legibility fallback). Four epochs of 10 periods, with amounts 40,000, 30,000, 20,000, 10,000 per period, giving 50 percent issued by period 14 and 100 percent at period 40. This is the most legible as a table and is retained as a fallback if the geometric and hybrid schedules prove too hard to communicate.

Halving (philosophical reference only). 50 percent issued by period 4, 90 percent by period 14, 99 percent by period 27, with a 50 percent drop at every boundary. Retained as a reference, not a primary candidate for a proof-of-stake network, because the cliff at each boundary is too large.

### 4.4 Treasury rule: R-F (recommended default)

The R-F rule is the recommended default from Decision Preparation v0.1 section 7, not an approved rule. It is the only treasury rule in this plan that is a single formula for both modes (issuance and post-issuance), which is why it is recommended.

Definition: let B be the treasury balance at the end of the previous period, D be the desired spend for the period (from approved proposals), Trail_4 be the trailing four-period average treasury inflow, and F be the emergency reserve floor (genesis-fixed, TBD). Then:

spend = min(D, max(Trail_4, floor(0.05 x B)), max(0, B - F))

In words: spend no more than the desired amount, and no more than the larger of (the trailing inflow average, or 5 percent of the balance), and never spend below the emergency reserve floor. The transition from issuance mode to post-issuance mode is keyed to the genesis-fixed final issuance period N, not to a governance vote.

Why R-F is recommended: it is one formula for both modes, so there is no mode-switch discontinuity; it keeps spending after issuance ends (unlike trailing-inflow caps, which freeze spending to zero after issuance ends); and it is bounded by the balance and the floor, so it cannot bankrupt the treasury in a single period.

P4 combined policy (candidate, not approved): the most conservative option in the research, combining a trailing-inflow cap with a balance cap and a floor. It is retained as a candidate because it is the most conservative, but it has a known defect: it freezes spending to zero after issuance ends, which is exactly the post-issuance problem it is meant to solve. P4 remains a candidate, not an approved rule. See section 11.

The level problem, stated plainly because no spending policy fixes it: the maximum sustainable constant spend is 909 per period under geometric and 2,500 under hybrid, in the research normalisation. The 30-year desired spend is 540,000 against about 100,000 total treasury inflow. No spending policy fixes a level problem. The spend level must be decided before genesis, separately from the spending rule, and it must be set at a level the inflow can actually support. If the desired spend exceeds the sustainable spend, the answer is to lower the desired spend, not to choose a cleverer rule. Revision note, v1.1, 2026-11-06 (finding b), corrected 2026-11-06 (GitHub publication): desired spending and sustainable spending are separate numbers and must be reported separately. The low-fee finding: under Low fees, raising desired spend D from 1,500 to 2,500 per period changes actual R-F spending by 0 units over 120 periods (hybrid model): both runs spend 114,511. The rule binds, not the desire. (An earlier version of this note stated 4 units and recorded the discrepancy as an open question in the decision register; recomputation from the stated model gives exactly 0, and the decision register question is now resolved.) Treasury reporting requirement: every treasury report, scenario table, and dashboard must report six figures separately, for the same horizon: (1) desired treasury spending, (2) actual permitted spending, (3) available inflows, (4) reserve balance, (5) unfunded spending (desired minus actual permitted, per period and cumulative), and (6) number of periods with a shortfall. A report that shows only desired spending, or only actual spending, is incomplete. A treasury spending formula is not proof of treasury sustainability.

### 4.5 Post-issuance funding

After period N, issuance stops permanently. The treasury receives only its share of transaction fees plus any other explicitly disclosed funding mechanism named before genesis. The mechanisms named in the research, with their status:

- Fee share: a share of transaction fees goes to the treasury. This is the primary post-issuance mechanism. Its adequacy depends on the fee market (section 5.4). At a Cardano-like fee level, fee revenue is 469 per period against a minimum need of 14,625 per period (100 validators, illustrative), or 3.2 percent of need. At standard validator costs, the minimum need is 24,625 per period. Fee revenue must be modelled, not assumed.
- Blockspace or coretime sales: selling blockspace or execution capacity as a separate revenue line. This is a named mechanism, but no design exists yet. It must be designed and its revenue modelled before genesis if it is to be relied on.
- Endowment: investing treasury funds and spending the yield. This fails on arithmetic in the research: a 300,000 balance is needed at 0.5 percent yield per period to fund the minimum need, which exceeds the total treasury inflow. An endowment cannot be the primary post-issuance mechanism unless the treasury balance is far larger than the research normalisation implies.
- Contributions: voluntary contributions to the treasury. These are real but unpredictable and cannot be budgeted.
- Treasury staking: noted as circular. Staking treasury funds to earn rewards that fund the treasury is staking the treasury's own coins to pay itself. It is not a funding mechanism and is not relied on in this plan.

The decision order for post-issuance funding, carried forward: name the post-issuance funding mechanism and the fee share before genesis. Do not launch a fixed-supply network without a written answer to "what pays validators and the treasury after issuance ends".

### 4.6 Consensus-critical rules vs governance-controlled rules

This separation is the governance security boundary. It must be enforced in code, not by convention.

Consensus-critical rules (deterministic, identical across nodes, changeable only through the constitutional process, never by ordinary governance):

- Denomination and decimal precision.
- Issuance arithmetic: the schedule formula, the reduction factor, the floor, the period count, the per-period amounts, the integer-floor rounding rule, the deterministic period indexing (period n covers heights [start + n x L, start + (n + 1) x L - 1], 0-indexed, in the block-based scaffold; period n is epoch n in the epoch-based architecture), and deterministic cumulative calculations. No floating point in any consensus-critical monetary code.
- Supply limits: max supply M, the cap enforcement on every mint attempt, the issued-by-epoch set, and permanent issuance cessation at the cap.
- Fee calculation: the fee formula itself, if it is consensus-critical. The fee level may be governance-controlled; the formula must be deterministic.
- Validator rewards: the reward calculation for a given split, once the split is fixed. The split itself is governance-controlled; the arithmetic that executes it is consensus-critical.
- Transaction validity: signature verification, sequence checking, balance checking, and the rejection of invalid transactions.

Governance-controlled rules (explicit permissions, voting thresholds, safeguards, changeable through x/gov):

- Approved upgrades, through x/upgrade, subject to the denylist gate (section 10.3).
- Treasury spending proposals, within the R-F rule.
- Selected network parameters: the validator, delegator, community, and treasury split; the fee level and fee distribution; the governance parameters themselves (deposit, voting period, quorum, threshold); and any parameter explicitly designated as governance-controlled in genesis.
- The split, the fee level, and the governance parameters must each have an explicit permission, a voting threshold, and a safeguard (a maximum change per proposal, a timelock, or both) documented before genesis. A governance proposal that changes a parameter without a safeguard is a governance attack vector, not a feature.

### 4.7 Supply invariants and minimum test list

Eleven invariants, each with a minimum test. These are the acceptance tests for x/monetary. They must pass on the wired chain (S2), not just on the scaffold.

1. Total supply never exceeds max supply M. Test: attempt to mint above the cap; mint fails, supply unchanged.
2. Total supply equals genesis supply plus total scheduled issued minus total burned, exactly, in integer arithmetic. Test: after a transfer and a burn, query total supply and compare to the computed value.
3. Scheduled issuance for epoch n equals exactly R_n. Test: for each epoch in a test schedule, claim and compare.
4. Each epoch is claimed exactly once. Test: re-claim the same epoch; second claim fails, issued total unchanged.
5. Total issued never exceeds total scheduled. Test: claim all epochs, then attempt one more; it fails.
6. Dust is never issued. Test: after all epochs, total issued equals S_actual, and the difference S_target minus S_actual is exactly the dust, which remains unissued.
7. No discretionary mint exists. Test: attempt to mint through bank, staking, distribution, gov, and wasm; all fail, state unchanged. (Milestone B: 17 bypass tests, all passing on the scaffold; must be re-run on the wired chain.)
8. Governance cannot change max supply, the schedule, or the final issuance period. Test: submit a governance proposal to change each; each is rejected or has no effect, and the schedule query returns the original values.
9. IBC escrow and voucher accounting conserves global economic supply. Test: transfer out, verify escrow equals transferred amount and voucher equals transferred amount; transfer back, verify escrow released exactly. (Milestone B: 6 IBC tests at ledger level. Revision note, v1.1, 2026-11-06: the live re-run in the proof of concept (O3) is DEFERRED by Benross's decision; live IBC is evaluated separately if a specific technical dependency demonstrates necessity.)
10. The split conserves. Test: for a given epoch, validator plus delegator plus community plus treasury shares equal exactly the amount minted, with no remainder and no shortfall, in integer arithmetic.
11. The every-block invariant holds. Test: at every block, the invariant check (total supply equals genesis plus issued minus burned, and total issued is less than or equal to total scheduled) passes; if it fails, the chain halts or the block is rejected, per the configured behaviour, and the failure is logged.

### 4.8 Decisions that must be resolved before genesis

In decision order, carried forward from Decision Preparation v0.1 section 8, with the revision that the public cryptocurrency option and identity decision now sits after the architecture decision (which is made: L1 native coin), and that legal analysis is the longest-lead item and must be commissioned first.

1. Legal analysis, commissioned in writing, before genesis distribution: securities, money transmission, tax, international distribution, exchange admission. This is the longest-lead item. It is not yet commissioned. Nothing in this plan is legal advice.
2. Period architecture: epoch-based, recommended (section 3.5). Confirm the epochs module is available and maintained in the chosen SDK version.
3. Period length D and nominal year. A sustainability decision, not a cosmetic one.
4. Public cryptocurrency identity: the native coin is OGO Coin on OGO Network (decided). The remaining identity question is whether the same asset appears in any other form (for example, a wrapped or bridged representation), and if so, under what accounting. One issuance authority or the cap breaks. No second issuance authority is created by this plan.
5. Minimum validator set and cost basis: Definition A (4 validators, liveness floor) or Definition B (at least 50 validators, sufficiently-decentralized bar), and which cost tier (lean, standard, professional) the minimum fee need is calculated on.
6. Treasury spend level and rule: the R-F rule is recommended; the spend level D must be set at a level the inflow can support. The P4 policy remains a candidate.
7. Post-issuance funding mechanism and fee share: named before genesis, with revenue modelled, not assumed.
8. Fee policy evidence: fee level, fee distribution, and how agent transactions are priced. Only the High fee scenario covers the minimum need, and only at period 38 (or 48 at standard costs). Moderate never covers it at standard costs.
9. Model family: geometric or hybrid. Then parameters: reduction factor r, floor and base share, period count N, initial amount. Then the split: validator, delegator, community, treasury. Then fee analysis. Then max supply as M = G + S_actual.
10. Network name and chain id, wallet address format (prefix and derivation path), denom and decimals. All are open decisions in section 11.

No decision in this list is made by this plan. Each needs Benross's explicit approval, and the legal analysis needs to be commissioned in writing before genesis distribution.

## 5. Validator and Consensus Specification

### 5.1 Consensus model

CometBFT Byzantine fault tolerant consensus, Tendermint-style, with instant finality per block. A block is final when committed. The active validator set is the top N validators by bonded stake, where N is a genesis parameter (TBD). Liveness requires more than two thirds of voting power online. The chain halts if more than one third is offline. Block time target is 1 to 6 seconds typical, with the proof-of-concept criterion of 95 percent of blocks within 1 to 10 seconds over 24 hours.

Consensus key rotation: in-place rotation is available in the 0.55 family. In the 0.53 matrix, rotation uses the standard unbond and re-bond path. The upgrade evaluation (M4) records which path applies.

### 5.2 Validator definitions and costs

Two definitions, carried forward from the Decision Preparation report. They are definitions, not decisions. The choice between them is an open decision (section 11).

Definition A, minimum operational set: 4 validators. This is the liveness floor: the smallest set that can tolerate one validator offline and still produce blocks (3 of 4 is more than two thirds). Cost, illustrative, at the standard tier (200 USD per validator per month, infrastructure only): about 800 USD per month, or about 300 USD per month at the lean tier (75 USD per validator per month). Definition A is what the proof of concept runs. It is not a production decentralization claim.

Definition B, sufficiently-decentralized set: at least 50 validators, Nakamoto coefficient at least 10, top 10 under 33 percent, no country or provider over 33 percent. Cost, illustrative, infrastructure only: 10,000 to 50,000 USD per month for 50 to 100 validators (standard 200 USD to professional 500 USD per validator per month). With labour, a professional validator costs 1,500 to 4,000 USD per month, so a 50-validator professional set costs 75,000 to 200,000 USD per month in total operating cost. No live Cosmos chain in the evidence meets all four Definition B thresholds at once (Hub coefficient 3, top 10 about 56 percent, Coinbase 20.56 percent). Definition B is a target, not a claim, and a mechanism for a flat stake distribution still needs to be named.

Validator tiers (infrastructure only, per validator per month):

| Tier | Infrastructure cost | Hardware baseline | Notes |
|---|---|---|---|
| Lean | about 75 USD | 2 plus CPU cores, 8 GB RAM, 200 GB SSD, 50 Mbps | Minimum for a testnet or a small validator. Not recommended for mainnet. |
| Standard | about 200 USD | 4 plus CPU cores, 32 GB RAM, 500 GB NVMe, 100 Mbps | The baseline used for the minimum fee need calculation. |
| Professional | about 500 USD | 8 plus CPU cores, 64 GB RAM, 1 TB NVMe, 1 Gbps, plus sentry nodes and remote signing | The baseline for a validator that is expected to be reliable under load. |

Labour is separate and dominates at the professional tier: 1,500 to 4,000 USD per validator per month for monitoring, upgrades, incident response, and key management. The minimum fee need in the research (14,625 per period at lean costs, 24,625 at standard costs, for 100 validators illustrative) is infrastructure only. If labour is included, the need is higher. This is why the fee transition is the binding problem, not a detail.

### 5.3 Staking, delegation, slashing, and unbonding

Staking through x/staking. Delegators bond `uogo` to a validator, earn a share of that validator's rewards per the approved split, and pay the validator's commission. The commission floor is 5 percent (Hub precedent, proposal). The commission rate is set by each validator, subject to the floor, and can be changed only with a notice period and a maximum change per period, to prevent a validator from raising commission to 100 percent after attracting delegations. The exact notice period and maximum change are TBD.

Unbonding: 21 days (Hub precedent, proposal). During unbonding, funds earn no rewards and cannot be transferred. Redelegation is allowed once per unbonding period per delegation, following the standard SDK rule. Unbonding is the security mechanism that makes slashing meaningful: a validator cannot exit instantly after misbehaving.

Slashing (Hub precedent, proposal):

- Downtime: 0.01 percent slash plus 10 minute jail. A validator that misses too many blocks in the signed-blocks window is jailed and must unjail to rejoin. The signed-blocks window and the missed-block threshold are TBD.
- Double-sign (equivocation): 5 percent slash plus tombstoning. A tombstoned validator can never rejoin. Equivocation evidence is handled through x/evidence and x/slashing.
- Slashing applies to bonded stake, so delegators share the loss. This is the incentive for delegators to choose reliable validators, and it is also the reason validator selection matters.

Validator set changes are governed by bonded stake ranking, not by appointment. There is no allowlist of validators in the protocol. A permissioned validator set for a testnet phase is a testnet configuration, not a protocol rule, and must be removed before mainnet.

Liquid staking is not part of the first monetary prototype. It is an open question for a later phase, and it must be evaluated against the same supply and slashing rules before adoption. No liquid staking token is created by this plan.

### 5.4 Fee market

Fees are paid in `uogo`, collected by the ante handler, and distributed by x/rewards per the approved fee distribution. The fee market is specified here because it is the post-issuance funding mechanism, and the evidence shows the fee transition has not happened on any of the four comparison networks.

Fee scenarios from the research (illustrative, normalised to the research units):

- Low: fee revenue never covers the minimum validator need, under either cost basis.
- Moderate: fee revenue covers the minimum need at lean costs (14,625) at some point, but never at standard costs (24,625).
- High: fee revenue covers the minimum need at period 38 at lean costs, and at period 48 at standard costs. Only the High scenario covers the minimum need, and only late in the issuance schedule.

The fee market must answer, before genesis:

1. What is the fee level (base fee, per-byte fee, priority fee, or a combination)? The research uses a flat fee for modelling; a flat fee cannot serve both security and agent viability at the same time, because agents are 45 percent of transactions but only 7.7 percent of fee revenue in the High scenario. If agents are charged the same flat fee as human transactions, either the fee is too high for agents or too low for security. The fee policy must state explicitly how agent transactions are priced.
2. What is the fee distribution (validator, delegator, community, treasury shares)? The distribution must be fixed before genesis and must conserve exactly.
3. Is there a fee burn? Ethereum burns the base fee, but burning is not a substitute for a funding model: Ethereum's burn collapsed from 8,844 per day in 2021 to about 57 per day in 2026, offsetting only about 2.4 percent of issuance, and ETH has been net-inflationary for 28 straight months. If OGO burns any fee, the burn must be modelled as a reduction in fee revenue available to validators and the treasury, not as a funding mechanism.
4. What happens if fee revenue never covers the minimum need? This is the honest question. The research shows that under Low and Moderate scenarios it does not, and under High it does so only late. The plan must state what happens then: validator set shrinks, validator pay falls, security degrades, or an explicitly disclosed alternative funding mechanism is used. None of these is chosen by this plan, but one must be chosen before genesis.

The fee policy is an open decision (section 11). It is not solved by this plan, and it must not be assumed solved by the fee market.

### 5.5 What the fee transition means for validators

The fee transition is the eventual shift from issuance-funded security to fee-funded security. It is the central problem of a fixed-supply proof-of-stake network, and it is not solved by any of the four comparison networks. All four remain issuance-funded in November 2026. Cardano is the warning: a fixed max plus declining issuance can coexist, but a declining reserve does not automatically produce a successful fee-funded transition. Fees are 3.3 million ADA against rewards of 493.7 million ADA over 73 epochs to 1 September 2026, or 0.67 percent. The fee transition is the explicit design assumption in Cardano, and it has not happened.

For OGO, the fee transition means:

- At period 20, fees must replace 87.8 percent of period-0 validator pay under geometric (r equals 0.90), or 75.5 percent under hybrid (base 40 percent), to keep validator pay flat. Neither model eliminates the transition; hybrid makes it gentler until its end date, then ends in a 100 percent step at period N.
- The hybrid valley is the period after issuance ends when fee revenue has not yet grown to cover the minimum need. In the research, the hybrid valley is 26 periods with a shortfall of 116,065 at 5 percent fee growth, and 69 periods with a shortfall of 572,847 at 3 percent growth. The treasury covers about half the shortfall at 5 percent growth. The valley is extremely growth-sensitive, which is why the growth assumption must be stated explicitly and not hidden. Revision note, v1.1, 2026-11-06 (finding d, corrected): the 26-period result assumes fee growth begins only after issuance ends. When fee growth begins at genesis instead, the valley is 0 periods. Every scenario table that quotes a valley length must identify its fee-growth start assumption; the 26-period figure without the assumption stated is incorrect.
- Geometric has no funding valley in the research, because its issuance declines gradually and its treasury is front-loaded, but its treasury bankrupts at period 66 under a fixed 1,500 per period spend during issuance. The treasury spending rule and the spend level matter as much as the issuance model.

The validator specification therefore does not promise that fees will cover validator costs. It states the need, the scenarios, and the gap, and leaves the funding decision to Benross.

## 6. Wallet Architecture

### 6.1 Principles

Five principles govern the wallet. They are not negotiable, because they follow from the founding architecture: OGO owns its monetary infrastructure, while users own their coins and private keys.

1. Noncustodial by construction. Users retain control of private keys. OGO must not store seed phrases or unrestricted private keys. OGO must not be able to move a user's coins.
2. An address is not a username. A wallet address is a cryptographic identifier derived from a public key. An OGO username is an application-level profile. The two are linked only by an explicit, revocable user action (signature verification, section 7.2), never by default.
3. A Supabase account does not own a blockchain balance. Supabase is responsible for application authentication, profiles, and application data. The blockchain is the source of truth for native coin balances. If the Supabase account is deleted, the on-chain balance is unaffected. If the OGO website is unavailable, the user can still access coins through any compatible wallet.
4. Signing happens only in the user's wallet. The OGO application, the Cloudflare Worker, and any OGO server never see a private key and never sign on a user's behalf.
5. Agents never get unrestricted wallet access. An agent acts only through a scoped authorization with spending limits, approved recipients, expiry, and revocation. See section 6.8.

### 6.2 Wallet creation and recovery

Creation:

- A new wallet is created from a BIP39 mnemonic. Prototype: 12 words. Optional: 24 words for users who want higher entropy. The mnemonic is generated on the user's device, never on a server, and never transmitted.
- The mnemonic is displayed once, with a clear instruction to write it down on paper and store it offline. The wallet does not store the mnemonic in plaintext. The wallet does not upload the mnemonic. The wallet does not email the mnemonic.
- The user confirms the mnemonic by re-entering a subset of words, to prove they recorded it. Creation is not complete until the confirmation passes.
- An optional passphrase (BIP39 passphrase, the 13th or 25th word) is supported for users who want it. The passphrase is never stored. If the user forgets the passphrase, the wallet cannot be recovered from the mnemonic alone, and the wallet must say so plainly before the user sets one.

Recovery:

- Recovery is by mnemonic only. The user enters the mnemonic, the wallet derives the keys, and the balance is read from the chain. No account, no email, no password is needed. This is what "access coins even if OGO's website is unavailable" means in practice.
- The wallet must support recovery from a mnemonic generated by another compatible wallet (Keplr, Leap, Cosmostation), and must generate mnemonics that those wallets can import. Interoperability is tested, not assumed (section 6.9).
- There is no password reset for a noncustodial wallet. If the mnemonic is lost and no backup exists, the coins are unrecoverable. The wallet must say this plainly at creation, not after a loss.

### 6.3 Private-key security

- Keys are derived from the mnemonic using the standard Cosmos derivation path (proposed: `44'/118'/0'/0/0`, provisional, section 11). The derivation path is a proposal, not a decision, and it must be fixed before the first wallet is released, because changing it after users have funded addresses creates a recovery failure.
- Local storage: the encrypted seed is stored using Argon2id encryption, with the encryption key held in the platform keystore (iOS Keychain, Android Keystore, or the operating system keychain on desktop). The seed is never stored in plaintext, never stored in localStorage as plaintext, and never logged.
- The wallet is unlocked by a user PIN or biometric, which releases the key from the platform keystore. The PIN is not a recovery mechanism. If the PIN is forgotten, the wallet is recovered from the mnemonic, not from the PIN.
- The signing key is held in memory only while the wallet is unlocked, and is cleared on lock, on timeout, and on app backgrounding after a short interval. The timeout is a user setting with a secure default (for example, 5 minutes).
- The wallet must not store a seed phrase or unrestricted private key on any OGO server, in Supabase, in the Cloudflare Worker, or in any log. This is a hard rule, not a preference. Any code path that would transmit a seed or private key to a server is a security bug, and the wallet must be tested for it (section 10, gate G7).

### 6.4 Transaction signing

- Signing uses Cosmos DIRECT sign mode (protobuf) by default, with Amino JSON as the legacy fallback for hardware wallets that require it. Both modes are tested against the chain.
- The wallet constructs the transaction, displays the full details to the user (recipient, amount, fee, memo, chain id, sequence), and signs only after the user confirms. The user must be able to see the exact amount and recipient before signing. There is no silent signing.
- The wallet verifies the chain id before signing, to prevent a transaction intended for one chain from being replayed on another. The chain id is displayed in the signing prompt.
- The wallet verifies the sequence number and the account number against the chain before signing, and rejects a transaction if they do not match.
- The wallet must never sign a transaction that the user did not see, and must never sign a blank or partially specified transaction. This is the defence against a compromised application interface presenting one transaction and signing another.

### 6.5 Address generation

- Addresses are bech32, with the proposed prefix `ogo` (provisional). The prefix is a proposal, not a decision, and it must be fixed before the first wallet is released, because addresses are permanent once funded.
- The address is derived from the public key using the standard Cosmos address derivation (SHA256 then RIPEMD160, then bech32 encoding). The derivation is deterministic: the same mnemonic and derivation path always produce the same address.
- The wallet displays the address with a copy button and a QR code. The QR code encodes the address and the chain id, so a sender cannot accidentally send to the right address on the wrong chain.
- The wallet must validate an address before sending: correct prefix, correct length, valid bech32 checksum. An invalid address is rejected before the transaction is constructed, not after it is signed.

### 6.6 Send, receive, balance verification, and transaction history

Send:

- The user enters a recipient address and an amount. The wallet validates the address, checks the balance (including the fee), and displays the fee and the total before the user confirms.
- The wallet does not allow sending more than the available balance minus the fee. The wallet does not allow sending to the user's own address without a confirmation (to prevent accidental self-sends that still cost a fee).
- After the transaction is broadcast, the wallet displays the transaction hash and a link to the explorer, and tracks the transaction until it is committed or fails. A failed transaction is displayed with the reason, not hidden.

Receive:

- The wallet displays the user's address and QR code. Receiving requires no action from the recipient beyond sharing the address. There is no "accept" step for an incoming transfer.
- The wallet must not display a balance for an address that has not been verified against the chain. A balance shown from a cache must be labelled as cached, with the time it was last verified.

Balance verification:

- Balances are read from chain state only, through a public RPC or LCD endpoint, or through the indexer. The balance is the x/bank balance for the `uogo` denom at that address. No other source is authoritative.
- The wallet must query at least two independent RPC endpoints for a balance, or one RPC endpoint plus the indexer, and must flag a discrepancy rather than silently displaying one value. This is the defence against a single compromised RPC endpoint (section 10.2, threat 7).
- The wallet must display the total supply and the user's share of it, so the user can see their balance in context. The total supply is read from the chain, not from an OGO server.

Transaction history:

- History is read from the chain (via the indexer) and from the user's local transaction log. The chain is authoritative. The local log is a cache.
- Each entry shows the transaction hash, the direction (sent or received), the counterparty address, the amount, the fee, the block height, the timestamp, and the status (committed, failed, pending). A pending transaction that never commits is labelled as such, not as completed.
- The wallet must not display a transaction as completed until it is committed in a block. A transaction that is only in the mempool is pending.

### 6.7 External wallet compatibility

External wallets are a founding requirement, not a nice-to-have. A user must be able to hold OGO Coin without an OGO account, and without the OGO wallet.

Priority order, with reasons:

1. Keplr. The standard interchain wallet, with the largest Cosmos user base, staking and governance support, and IBC transfer support. Keplr integration requires Cosmos Chain Registry registration (chain id, RPC and LCD endpoints, denom, decimals, bech32 prefix, derivation path, gas price), standard signing and derivation, and public CORS-enabled RPC and LCD endpoints.
2. Leap. The second standard interchain wallet, with a similar integration path. Leap is the fallback if Keplr integration is blocked, and the second wallet to support in any case, so users are not dependent on a single wallet vendor.
3. Cosmostation. The third standard interchain wallet, with mobile and desktop support. Cosmostation is the third wallet to support, for the same reason.

Integration requirements, all of which must be true before an external wallet is claimed as supported:

- Chain registry metadata is registered and correct.
- The chain's RPC and LCD endpoints are public, CORS-enabled, and documented.
- The bech32 prefix and derivation path are fixed and match the registry.
- A test transfer from the external wallet to an OGO wallet, and from an OGO wallet to the external wallet, both commit with correct balances. Revision note, v1.1, 2026-11-06: this was a MUST criterion of the proof of concept (O4); O4 is DEFERRED by Benross's decision (Phase 2 Execution Directive) and is no longer a framework-selection gate. Keplr is evaluated separately if a specific technical dependency demonstrates necessity. The wallet product is built in M4, where O4 is re-tested at product level.
- Staking and governance voting work from the external wallet, or are documented as not yet supported, with a date.

If an external wallet cannot be integrated in the proof-of-concept window, the blocker is recorded and the criterion is treated as a MUST failure until it passes, because wallet access is a founding requirement. Revision note, v1.1, 2026-11-06: this sentence described O4 as a MUST failure; O4 is DEFERRED by Benross's decision (Phase 2 Execution Directive), so it no longer blocks framework selection. The blocker is still recorded when the evaluation happens.

### 6.8 Hardware wallet compatibility roadmap

Hardware wallets are staged, not promised for launch. Each stage has a clear entry criterion.

- Stage 1, software only, with an abstract signing seam. The wallet is built so that the signing operation is an interface, not inline code. The software signer implements the interface. This is the prototype. Entry criterion: the abstract signing seam exists and is tested.
- Stage 2, Ledger via the Cosmos app. Ledger support through the Cosmos Ledger app, using Amino JSON signing (which Ledger requires) and the standard Cosmos derivation path. Entry criterion: a test transfer signed on a Ledger device commits with correct balances.
- Stage 3, Trezor and Keystone. Trezor support through the Cosmos app, and Keystone (air-gapped QR signing) support. Entry criterion: a test transfer signed on each device commits with correct balances.
- Stage 4, multisig and air-gapped signing. Multisig wallets (multiple keys required to sign) and fully air-gapped signing (transaction constructed offline, signed offline, broadcast from a separate online device). Entry criterion: a multisig transaction commits, and an air-gapped transaction commits.

No stage is claimed as supported until its entry criterion passes. Hardware wallet support is not a dependency of the first monetary prototype (M0 to M3).

### 6.9 Agent authorization mechanism

Agents must never get unrestricted wallet access. An agent acts only through a scoped authorization, which is a separate mechanism from the wallet itself.

Authorization format (user-signed document, not a smart contract in the prototype):

- Spending limit: the total amount the agent may spend under this authorization, in `uogo`.
- Per-transaction limit: the maximum amount for a single transaction under this authorization.
- Approved recipients: an explicit list of recipient addresses. No wildcard. An empty list means no recipients are approved, not all recipients.
- Expiry: a mandatory expiry time. An authorization without an expiry is invalid. The maximum expiry is a user setting with a secure default (for example, 24 hours).
- Nonce: a unique nonce per authorization, to prevent replay.
- The authorization is signed by the user's wallet key. The signature covers all fields above, so no field can be changed after signing without invalidating the signature.

Enforcement:

- In the prototype, enforcement is off-chain, in the wallet. The wallet checks every agent-initiated transaction against the authorization: amount within the spending limit and the per-transaction limit, recipient on the approved list, expiry not passed, nonce not used. If any check fails, the wallet refuses to sign. The wallet logs every agent-initiated transaction, whether it was signed or refused, with the reason.
- On-chain enforcement is a later phase, via x/authz (the standard Cosmos authorization module) or via agent sub-accounts (a separate account controlled by the agent with a limited balance). On-chain enforcement is not a dependency of the first monetary prototype, but the authorization format must be designed so it can be moved on-chain without changing the fields.
- Revocation is immediate and in-wallet. The user can revoke an authorization at any time, from the wallet, without needing the agent's cooperation. Revocation takes effect immediately for all future transactions. A revoked authorization cannot be used again, even if it has not expired. The wallet must display all active authorizations and their remaining limits, so the user can see what each agent is allowed to do.

Defaults (secure, not convenient):

- No agent has any authorization by default. An agent must request one, and the user must approve it explicitly, with all fields visible.
- The default spending limit is zero until the user sets one. The default per-transaction limit is zero. The default recipient list is empty. The default expiry is 24 hours.
- An authorization for an agent that the user has not explicitly approved is invalid, even if it is correctly signed by someone else. Only the user's wallet key can create an authorization for that user's wallet.

Agent payment permissions in the OGO application are specified in section 7.5. The wallet mechanism in this section is the enforcement layer; the application mechanism in section 7.5 is the permission layer. Both must be true for an agent to spend.

### 6.10 Prototype scope vs later stages

Prototype (M0 to M3):

- Software wallet only, with the abstract signing seam.
- Mnemonic creation and recovery, Argon2id local storage, platform keystore.
- DIRECT and Amino JSON signing, bech32 `ogo` prefix (provisional), derivation path `44'/118'/0'/0/0` (provisional).
- Send, receive, balance verification from chain state, transaction history from the indexer.
- Agent authorization as a user-signed document, enforced off-chain in the wallet, with immediate revocation.
- Keplr integration, or a documented blocker. Revision note, v1.1, 2026-11-06: this was a MUST criterion (O4); O4 is DEFERRED by Benross's decision (Phase 2 Execution Directive) and is no longer a framework-selection gate or a prerequisite for the initial monetary prototype.

Later stages (M4 to M8):

- Leap and Cosmostation integration.
- Ledger (Stage 2), then Trezor and Keystone (Stage 3), then multisig and air-gapped signing (Stage 4).
- On-chain agent authorization via x/authz or agent sub-accounts.
- WalletConnect or equivalent for dApp connections, if adopted, with the same signing-prompt rules as the wallet itself.

### 6.11 Wallet verification checklist

The wallet is not claimed as working until every item passes:

1. A mnemonic created in the OGO wallet can be imported into Keplr and recovers the same address and balance, and vice versa.
2. A transaction signed in the wallet commits with the correct balances, and the fee is exactly as displayed before signing.
3. A transaction with a wrong signature, a wrong sequence, or an insufficient balance is rejected before execution.
4. The wallet displays a balance only from chain state, and flags a discrepancy between two RPC endpoints rather than silently displaying one value.
5. A transaction is not displayed as completed until it is committed in a block.
6. No seed phrase or private key is transmitted to any server, in any log, or in any error report. This is tested by inspecting network traffic and logs during wallet operations.
7. An agent authorization with a spending limit, a per-transaction limit, an approved recipient list, and an expiry is enforced: a transaction above the limit, to an unapproved recipient, or after expiry is refused, and the refusal is logged.
8. Revocation takes effect immediately: after revocation, an agent-initiated transaction is refused even if the authorization has not expired.
9. The wallet works when OGO's website is unavailable: a user can recover from a mnemonic and view a balance and send a transaction using only a public RPC endpoint and an independent wallet, with no OGO server involved.

## 7. OGO Integration Specification

### 7.1 What exists today (verified on live main, 2026-11-06)

The integration specification starts from what is actually in the repository, verified by the coordinator against live main, not from what earlier reports assumed.

Economy wallet (live main, commit `e2fd900dbcf9`, 2026-11-06T02:41:09Z, "Add account-secured OGO Coins and Credits wallet"):

- Migration `supabase/migrations/20261105000000_ogo_economy_wallet.sql` creates `ogo_wallets` (one row per auth user, integer balance plus reserved balance, timestamps) and `ogo_wallet_transactions` (append-only ledger, transaction types `purchase`, `credit`, `debit`, `hold`, `release`, `refund`; signed integer amounts; external reference fields, order id, description, status, metadata). Row Level Security allows users to read only their own wallet row and transactions. A `wallet_reconcile` RPC exists, intended for server-side reconciliation only. Writes are intended to be service-role only.
- The Worker `chat-worker.js` (7,267 lines, including 75 economy lines) provides `GET /api/economy/catalog` (5 Stripe packs: 100, 550, 1150, 2750, 6000 Coins for 1, 5, 10, 25, 50 USD), `GET /api/economy/wallet`, `POST /api/economy/checkout`, and `POST /api/economy/webhook`. The webhook revalidates pack metadata on every call and uses an idempotency key `stripe_checkout:<session id>`. No earn or spend path exists in the repository. A pack purchase is the only inflow into the Coins balance.
- Frontend: `src/OgoEconomy.tsx` (111 lines) exists on live main, but no file imports it. `src/App.tsx` (3,377 lines) renders `SubscriptionSection`, not `OgoEconomy`. `src/main.tsx` and `src/SubscriptionSection.tsx` contain no reference to `OgoEconomy`. The wallet UI component appears to have no mount point and no import path. This needs one live UI confirmation (section 11).
- There is no conversion to any external asset. The economy is a closed loop, and it is a database ledger, not a self-custody wallet.

Supabase and Worker architecture:

- 23 migrations are catalogued in the repository. Two gaps matter for integration: the Worker queries `forge_symbols` and calls `forge_search`, and the Builder Engine writes to `user_apps`, but no repository migration defines those tables. They are likely dashboard SQL. Any integration that depends on Forge or on the Builder Engine's `user_apps` table must confirm the live schema in the dashboard before relying on it.
- The local working copy at `~/workspace/ogo/` is stale versus live main by exactly the economy wallet feature: it lacks `src/OgoEconomy.tsx` and the economy migration, and its `worker/chat-worker.js` differs by only the 75 economy lines. It also contains local extras not on live main (`dist/`, `node_modules/`, `supabase/migrations/20261028_connect_phase0.sql`, `governance/__pycache__`). The working copy is not a git repository and has no `.git` directory. Any M6 implementation work must start from a fresh clone, not from this working copy.

What does NOT exist in the repository:

- No OGO Network material at all: no `docs/network/` or `docs/blockchain/`, zero `.go` files, zero genesis configs. The word "network" in the repository means the agent network, not a blockchain.
- No OGO Network code or prototype in benross-lab/OGO. The Milestone B scaffold lives in staging under `~/workspace/goals/ogo-ai-builder-and-network/files/ogo-network-milestone-b/chain/`, not in the repository. Nothing was copied into the working repo.

### 7.2 Integration principles

1. Three monies, separate, always. OGO Coins (account-based, in Supabase), OGO Credits (account-based, in Supabase), and native OGO Coin (on-chain, in x/bank). No combined balance is ever shown. No automatic conversion exists. Any future conversion is a separate decision requiring legal review and explicit approval.
2. The application talks to the network only through public RPC and API services. There is no private chain connection, no direct database access to chain state, and no Worker-side chain write path.
3. The chain is the source of truth for native balances. Supabase is the source of truth for application authentication, profiles, and application data. If the two disagree, the chain wins for native balances, and the discrepancy is logged and surfaced, not silently resolved.
4. The Worker may read the chain, verify on-chain state, and prepare unsigned transactions. The Worker must never hold keys, sign for users, write chain state, broadcast transactions, trust client claims about chain state, write to the platform ledger as a side effect of a chain operation, act as a validator, or move treasury funds. The full may and must-never lists are in section 7.3.
5. Blockchain consensus and validator responsibilities are never merged into the Cloudflare Worker. The Worker is an application server. It is not a node, not a validator, and not a relayer.
6. A wallet address is not an OGO username. Linking is an explicit, revocable user action, verified by signature, not by a form field.

### 7.3 The Worker: what it may and must never do

The Cloudflare Worker (`chat-worker.js`) is the application server. Its role in network integration is narrow and explicit.

The Worker MAY:

- Read chain state through public RPC and LCD endpoints, on behalf of an authenticated user, for balances, transaction history, and validator information.
- Issue a nonce for wallet linking, and verify a signature against that nonce (section 7.4).
- Prepare an unsigned transaction for a user to sign in their own wallet (for example, a marketplace settlement transaction), and return it to the client for signing. The Worker does not sign it and does not broadcast it.
- Verify on-chain state before fulfilling an application action (for example, verify that a payment transaction is committed in a block, with the correct amount and recipient, before marking a marketplace order as paid).
- Record a fulfillment in Supabase with an idempotency key, after on-chain verification, so a duplicate webhook or a duplicate request cannot fulfil twice.
- Cache chain reads (balances, validator lists, supply) with a short time-to-live, and label cached values with the time they were last verified.

The Worker MUST NEVER:

- Hold a private key, a seed phrase, or a signing key for any user, validator, or treasury account.
- Sign a transaction on behalf of a user, a validator, or the treasury.
- Write chain state directly, or broadcast a transaction to the chain.
- Trust a client claim about chain state (a balance, a transaction hash, a payment) without verifying it against a public RPC endpoint.
- Write to the platform ledger (Supabase `ogo_wallets` or `ogo_wallet_transactions`) as a side effect of a chain operation, or vice versa. The two ledgers are separate, and a write to one never implies a write to the other.
- Act as a validator, a sentry, or a relayer, or hold validator keys.
- Move treasury funds, or initiate a treasury spending proposal, on its own authority.

Any code that violates a must-never item is a security bug, and the Worker must be tested for it (section 10, gate G7).

### 7.4 Wallet linking and signature verification

Linking an OGO wallet address to an OGO account is an explicit, revocable user action, verified by signature.

Flow:

1. The user, signed in to the OGO application, requests to link a wallet. The Worker issues a nonce: a random value, stored in Supabase with the user id, the nonce, and an expiry (for example, 10 minutes). The nonce is single-use.
2. The user's wallet signs a message containing the nonce, the OGO account identifier, the wallet address, and the chain id. The message format is fixed and documented, so the user can see exactly what they are signing.
3. The user submits the signature to the Worker. The Worker verifies the signature against the nonce, the account, the address, and the chain id, using a public RPC endpoint or a local verification library. If verification passes and the nonce has not expired and has not been used, the Worker records the link in Supabase: user id, wallet address, chain id, linked timestamp, and the nonce id. The nonce is marked as used.
4. The link is displayed in the user's settings, with the wallet address, the chain id, and the linked date. The user can unlink at any time. Unlinking removes the link from Supabase and does not affect the on-chain balance. A linked wallet can be re-linked after unlinking, with a fresh nonce.
5. The Worker must verify the signature itself. It must not trust a client claim that a wallet is linked. A link that was not verified by signature is not a link.

What linking does and does not do:

- Linking allows the OGO application to display the native balance for that address, to show transaction history, and to prepare unsigned transactions for that address.
- Linking does not give the OGO application control over the wallet, the ability to sign for the user, or the ability to move funds.
- Linking does not merge the Supabase account with the blockchain balance. The two remain separate. If the Supabase account is deleted, the on-chain balance is unaffected, and the link is deleted with the account.

### 7.5 Economy dashboard: three sections, never combined

The economy dashboard displays three separate sections, one for each money. No combined balance is ever shown.

- OGO Coins (Supabase): the account-based balance, the transaction history from `ogo_wallet_transactions`, the Stripe pack purchase flow, and any future earn or spend path (none exists today). This is the existing economy, unchanged by network integration.
- OGO Credits (Supabase): the account-based balance, the transaction history, and the credit spend path for AI services. This is the existing credit system, unchanged.
- Native OGO Coin (on-chain): the native balance for each linked wallet address, read from chain state through a public RPC endpoint, the transaction history from the indexer, and a link to the explorer for each transaction. This section displays chain data only. It does not display a Supabase balance, and it does not allow spending Supabase Coins as if they were native OGO Coin.

Each section is labelled with its source ("Supabase ledger" or "OGO Network, on-chain") and the time the balance was last verified. A cached balance is labelled as cached.

### 7.6 Live balance and transaction history

Balances and history for native OGO Coin are read from the chain, not from Supabase, not from the Worker, and not from a client claim.

- Balance: read from x/bank through a public LCD endpoint (`/cosmos/bank/v1beta1/balances/<address>`), or through the indexer. The application must query at least two independent RPC endpoints, or one RPC endpoint plus the indexer, and must flag a discrepancy rather than silently displaying one value.
- History: read from the indexer, which indexes blocks, transactions, and events from the chain. The indexer is the source for history, because the chain itself does not provide a paginated history API. The indexer must be reconciled against the chain (section 10, gate G9): if the indexer and the chain disagree, the chain wins, and the discrepancy is logged.
- Caching: balances and history may be cached in the Worker or in the application, with a short time-to-live (for example, 30 seconds for a balance, 5 minutes for history). Cached values are labelled with the time they were last verified. A cached value is never presented as live.
- Refresh: the user can refresh a balance or history manually, which bypasses the cache and queries the chain directly. A refresh that fails is displayed as a failure, not as a stale value.

### 7.7 Payment requests

A payment request is a request for a user to pay a specific amount to a specific address, for a specific purpose. It is not a charge, not a deduction, and not an automatic payment.

Flow:

1. A requester (a user, an agent, a developer, or a marketplace seller) creates a payment request: recipient address, amount in `uogo`, a description, and an expiry. The request is stored in Supabase, with a unique request id.
2. The payer sees the request in the OGO application, with the recipient, the amount, the description, and the expiry. The payer can approve or decline. Approving does not move funds. Approving prepares an unsigned transaction for the payer to sign in their own wallet.
3. The payer signs the transaction in their wallet, after seeing the exact amount and recipient. The wallet broadcasts the transaction (or returns it to the application to broadcast, at the payer's explicit instruction). The application tracks the transaction hash until it is committed.
4. The application verifies the transaction is committed, with the correct amount and recipient, before marking the request as paid. Verification is against a public RPC endpoint, not against a client claim.
5. A payment request that expires is marked as expired, not as paid. A payment request that is declined is marked as declined. A payment request that is paid is marked as paid, with the transaction hash and the block height recorded.

Payment requests are the mechanism for marketplace settlement (section 7.8) and for agent payments (section 7.5 of the wallet architecture, and the agent permission layer below). They are not a subscription, not a recurring charge, and not a direct debit.

### 7.8 Agent payment permissions (application layer)

The wallet authorization mechanism (section 6.8) is the enforcement layer. The application permission layer is the user-facing control that creates and manages those authorizations.

- An agent must request a payment permission before it can spend. The request specifies: the spending limit, the per-transaction limit, the approved recipient list, and the expiry. The user sees all fields and approves or declines.
- The user can view all active agent permissions, with their remaining limits, in the application settings. The user can revoke a permission at any time, from the application, without needing the agent's cooperation. Revocation takes effect immediately in the wallet (section 6.8) and is reflected in the application on the next refresh.
- The application must not grant an agent permission that the user did not explicitly approve. A permission created by an agent, by the Worker, or by a default is invalid.
- The application logs every agent-initiated payment request, whether it was approved, declined, signed, or refused, with the reason. The log is visible to the user.

### 7.9 Developer payment APIs

Developers can accept native OGO Coin payments in their applications, through a public API that is separate from the OGO application.

- The API provides: a payment request creation endpoint (recipient, amount, description, expiry, returning a request id and a payment URL), a payment status endpoint (request id, returning pending, paid, expired, or declined, with the transaction hash and block height if paid), and a webhook that notifies the developer when a payment is verified as committed on-chain.
- The API does not hold keys, does not sign, and does not broadcast. The developer's application prepares the payment request; the payer signs in their own wallet; the API verifies the transaction on-chain before notifying the developer.
- The API is rate-limited and authenticated, so a developer cannot spam payment requests or enumerate other developers' requests.
- The API documentation is published in OGO-Network-Docs, with example code, and is versioned. A breaking change to the API requires a new version, not a silent change to the existing version.

### 7.10 Marketplace settlement

Marketplace settlement is gated on verified on-chain payment. A marketplace order is not marked as paid until the payment is verified as committed on-chain.

Flow:

1. A buyer and a seller agree on a price in native OGO Coin. The marketplace creates a payment request (section 7.7) with the seller's address, the amount, and an expiry.
2. The buyer signs the payment in their wallet. The application tracks the transaction hash until it is committed.
3. The application verifies the transaction against a public RPC endpoint: correct amount, correct recipient, committed in a block, with the block height recorded. Only after verification is the order marked as paid, and only then is the seller notified and the item released.
4. The settlement is recorded in Supabase with an idempotency key (the transaction hash, or the payment request id), so a duplicate webhook or a duplicate request cannot fulfil twice. The idempotency key is the same pattern as the Stripe webhook (`stripe_checkout:<session id>`), applied to on-chain payments.
5. If the payment fails, expires, or is for the wrong amount, the order is not marked as paid, and the buyer and seller are both notified with the reason. There is no partial settlement.

Marketplace settlement does not use Supabase Coins or Credits. It uses native OGO Coin only. A marketplace order priced in Supabase Coins is a separate flow, in the existing economy, and must not be confused with a native coin settlement.

### 7.11 Network activity and public explorer

- Network activity (recent blocks, recent transactions, validator set, total supply, issuance progress) is displayed in the OGO application from the public RPC and indexer APIs, labelled with the source and the time last verified. It is a read-only view. It does not allow the application to change chain state.
- The public explorer (OGO-Explorer repository) is a separate application that displays blocks, transactions, accounts, validators, and supply, from the chain and the indexer. The explorer is the canonical place for a user to verify a transaction, a balance, or total supply, without an OGO account. The explorer must be usable without signing in, without an OGO account, and without any OGO server beyond the public RPC and indexer endpoints.
- The OGO application links to the explorer for every transaction hash, balance, and block height it displays, so a user can verify any claim independently.

## 8. Repository Structure

### 8.1 Five repositories

No repository is created by this plan. Creation needs explicit approval for each repository. The structure below is the plan for when creation is approved.

| Repository | Purpose | Language and stack |
|---|---|---|
| OGO (existing, benross-lab/OGO) | Main application and Economy interface. The application talks to the network through public RPC and API services only. | Existing stack (TypeScript, React, Supabase, Cloudflare Worker). |
| OGO-Network | Blockchain node, runtime and modules, genesis, protocol. The chain. | Go 1.25.7, Cosmos SDK v0.53.6 required with v0.53.4 effective, CometBFT v0.38.23, wasmd v0.60.9, ibc-go/v10 v10.7.0, Cosmovisor v1.7.3. |
| OGO-Wallet | Wallet interface and key management. The wallet. | To be decided in M4. The wallet must support Keplr, Leap, and Cosmostation integration, and must implement the signing seam for hardware wallets. |
| OGO-Explorer | Blocks, transactions, accounts, validators, supply. The explorer. | To be decided in M6. The explorer reads from the chain and the indexer, and must be usable without an OGO account. |
| OGO-Network-Docs | Protocol documentation, validator guides, API specifications. The docs. | Markdown, published as a static site. The docs are the canonical reference for the protocol, and must be kept in sync with the code. |

### 8.2 OGO-Network layout

Refined from the integration draft, with the corrections in the verification note at the top of this plan: only three custom modules (x/monetary, x/treasury, x/rewards), no x/governance and no x/fees, and x/monetary as the production name for the scaffold's x/ogocoin.

```
OGO-Network/
  app/
    app.go
    app_setup.go
    modules.go
  cmd/
    ogod/
      main.go
  x/
    monetary/
      module.go
      keeper/
        keeper.go
        schedule.go
        genesis.go
        invariants.go
      types/
        params.go
        schedule.go
        errors.go
      client/
        cli/
          query.go
          tx.go
    treasury/
      module.go
      keeper/
        keeper.go
        spending.go
        genesis.go
      types/
        params.go
        errors.go
    rewards/
      module.go
      keeper/
        keeper.go
        split.go
        genesis.go
      types/
        params.go
        errors.go
  proto/
    ogo/
      monetary/
        v1/
          monetary.proto
      treasury/
        v1/
          treasury.proto
      rewards/
        v1/
          rewards.proto
  genesis/
    mainnet.example.json
    testnet.example.json
    localnet.json
  config/
    config.toml
    app.toml
  scripts/
    build.sh
    test.sh
    localnet.sh
    upgrade.sh
  docs/
    ARCHITECTURE.md
    GENESIS_SPEC.md
    TOKENOMICS.md
    GOVERNANCE.md
    SECURITY.md
    REPRODUCIBLE_BUILD.md
  tests/
    integration/
      localnet_test.go
      ibc_test.go
      upgrade_test.go
  Makefile
  go.mod
  go.sum
  README.md
```

Notes on the layout:

- `x/monetary/` is the production module. It extends and renames the scaffold's `x/ogocoin`. The scaffold code is its ancestor. The schedule logic, the cap enforcement, the issued-by-epoch set, and the bypass tests move into this module, with the block-based periods migrated to epoch-based periods (section 3.5).
- `x/treasury/` and `x/rewards/` are new modules, built in Milestone C. They do not exist in the scaffold.
- `genesis/` contains example genesis files, labelled with their purpose. The localnet genesis uses TEST-ONLY parameters, labelled so. The testnet and mainnet example genesis files contain no real parameters until Benross approves them. No genesis file in the repository contains a real max supply, a real schedule, or a real allocation until those decisions are made.
- `docs/` in OGO-Network is a mirror of the canonical docs in OGO-Network-Docs, kept in sync. The canonical source is OGO-Network-Docs. If the two disagree, OGO-Network-Docs wins, and the discrepancy is fixed in OGO-Network.

### 8.3 OGO application module (in benross-lab/OGO)

The OGO application gains a network module, in the existing repository, when integration work is approved. It does not gain chain code, validator code, or consensus code.

- `src/OgoNetwork/` (or the equivalent module directory, following the existing application structure): the network integration module, containing the economy dashboard's native coin section, the wallet linking flow, the payment request flow, the agent permission UI, and the network activity view. Each component reads from public RPC and indexer APIs only.
- The Worker gains the network read and verification endpoints described in section 7.3, in `worker/chat-worker.js` or in a separate Worker module, following the existing Worker structure. The Worker does not gain signing, key management, or chain write capability.
- Supabase gains tables for wallet links, payment requests, agent permissions, and settlement records, each with Row Level Security and an idempotency key where applicable. Migrations are applied only with explicit approval, and only after the live schema is confirmed in the dashboard (see the Forge and `user_apps` gap in section 7.1).

### 8.4 OGO-Wallet layout (proposed)

```
OGO-Wallet/
  src/
    wallet/
      create.ts
      recover.ts
      keys.ts
      signing.ts
    agent/
      authorization.ts
      revocation.ts
    ui/
      dashboard.tsx
      send.tsx
      receive.tsx
      settings.tsx
  tests/
    signing.test.ts
    recovery.test.ts
    agent.test.ts
  docs/
    WALLET_ARCHITECTURE.md
    SECURITY.md
  README.md
```

The layout is a proposal. The language and framework for the wallet are decided in M4, not in this plan. The layout is included so the scope is visible, not so the technology is chosen.

### 8.5 OGO-Explorer layout (proposed)

```
OGO-Explorer/
  src/
    pages/
      blocks.tsx
      transactions.tsx
      accounts.tsx
      validators.tsx
      supply.tsx
    api/
      rpc.ts
      indexer.ts
  indexer/
    (see section 9, M6)
  docs/
    EXPLORER.md
  README.md
```

The explorer reads from the chain and the indexer. It must be usable without an OGO account. The indexer is built in M6, and the explorer depends on it for transaction history.

### 8.6 OGO-Network-Docs layout (proposed)

```
OGO-Network-Docs/
  docs/
    network/
      FOUNDING_CHARTER.md
      TECHNICAL_BLUEPRINT.md
      TOKENOMICS.md
      GOVERNANCE.md
      SECURITY.md
    blockchain/
      ARCHITECTURE.md
      DEPENDENCIES.md
      GENESIS_SPEC.md
    wallet/
      WALLET_ARCHITECTURE.md
    api/
      RPC_API.md
      INDEXER_API.md
      DEVELOPER_PAYMENT_API.md
    validators/
      VALIDATOR_GUIDE.md
      OPERATIONS_RUNBOOK.md
  README.md
```

The docs in OGO-Network-Docs are the canonical protocol documentation. The proposed docs from the earlier milestone (in `ogo-network-proposed-docs/docs/network/` and `docs/blockchain/`) are the starting point, and they must be revised to match the decisions in this plan (L1 native coin, separate repositories, x/monetary, epoch-based periods, R-F treasury rule) before they are copied into OGO-Network-Docs. Copying them without revision would reintroduce the superseded Base recommendation and the old module list.

## 9. Phased Implementation Backlog

### 9.1 Phase to milestone map

| Phase | Scope | Milestones |
|---|---|---|
| Phase 0 | Protocol specification | M0 Architecture and Decisions, M1 Localnet Proof (planning and first proof) |
| Phase 1 | Native monetary prototype | M1 Localnet Proof, M2 Monetary Prototype |
| Phase 2 | Validator and consensus prototype | M3 Validator Prototype |
| Phase 3 | Wallet and user access | M4 Wallet Prototype |
| Phase 4 | Application integration | M5 App Integration |
| Phase 5 | Infrastructure and explorer | M6 RPC and Indexer, M7 Testnet |
| Phase 6 | Launch readiness | M8 Mainnet Readiness |

Related continuation work, commissioned 2026-11-06 and recorded in memory: 11 standalone specification documents were delivered 2026-11-06 as a continuation of this plan, in `ogo-network-architecture/` (ARCHITECTURE, CONSENSUS_AND_EPOCH_SPEC, MONETARY_ENGINE_SPEC, VALIDATOR_AND_STAKING_SPEC, TREASURY_AND_SUSTAINABILITY, NATIVE_WALLET_SPEC, FEE_MARKET_SPEC, NETWORK_API_SPEC, SECURITY_MODEL, IMPLEMENTATION_ROADMAP, DECISION_REGISTER). Revision note, v1.1, 2026-11-06: v1.0 said these documents were being prepared and the directory did not yet exist; they have now been delivered, and this section is aligned with that state. Those documents expand this plan's sections into standalone specs. Where a continuation spec disagrees with this plan, the disagreement must be resolved explicitly and recorded, not silently overwritten. The continuation confirms Option B as locked and carries the same restrictions: no deployment, no token, no testnet, no production change, no spending, and no monetary parameter selection is authorized by it.

### 9.2 Milestones M0 to M8: deliverables and acceptance criteria

M0, Architecture and Decisions. Deliverables: this plan (12 deliverables), the open decisions register (section 11), the proof-of-concept specification (section 2.6). Acceptance criteria: the framework recommendation and the proof-of-concept specification are approved by Benross; all open decisions are listed with their status; no monetary parameter is selected; the two Benross decisions (L1 native coin, separate repositories) are recorded. M0 is this plan. It is complete when Benross approves it, not when it is written.

M1, Localnet Proof. This is the immediate focus, with M0. First proof, stated exactly: three independent local validator nodes, genesis accounts, test OGO Coin transferred between wallets, and the same balances verified from all three nodes. Revision note, v1.1, 2026-11-06: validator count THREE, decided by Benross (Phase 2 Execution Directive). The reconciliation pass had resolved this to 4 with BFT reasoning (a 3-validator network tolerates zero faulty validators and halts if any one goes down); Benross's decision supersedes that resolution for the local prototype. Distinction: 3 is the decided local-prototype execution requirement; 4 is the minimum operational set (smallest set tolerating one faulty validator); 50 or more is the decentralization target for a public network. Deliverables: the Milestone C wiring from section 3.17 (depinject, signature-verified transactions, epoch-based periods, x/monetary, x/treasury, x/rewards, remote signer, every-block invariant, upgrade denylist), a 3-validator localnet, genesis accounts, and a proof log. Acceptance criteria: all three nodes produce blocks; a test transfer commits with the correct balances; the same balances are verified from all three nodes by independent queries; total supply matches the expected value exactly; the proof log records the block height, the transaction hash, and the balances from each node. All numbers are TEST-ONLY and labelled so.

M2, Monetary Prototype. Deliverables: x/monetary with epoch-based periods, schedule-only mint, cap enforcement, the issued-by-epoch set, and the three-number accounting queries; the bypass suite re-run on the wired chain. Acceptance criteria: all 54 Milestone B tests pass on the wired chain, plus new tests for the wired keepers; proof-of-concept criteria S2, S3, and S4 pass; the 11 supply invariants in section 4.7 pass; no discretionary mint exists through any path.

M3, Validator Prototype. Deliverables: a 4-validator localnet with staking, delegation, unbonding, slashing, and remote signing; a Cosmovisor upgrade executed on the localnet; a rollback executed on one validator. Acceptance criteria: proof-of-concept criteria P1, S5, S6, S8, M2, M3, and M5 pass; a validator can be jailed for downtime and unjailed; a double-sign is punished; an upgrade completes with identical balances and supply before and after.

M4, Wallet Prototype. Deliverables: the software wallet (section 6.10 prototype scope), with mnemonic creation and recovery, Argon2id local storage, DIRECT and Amino JSON signing, send, receive, balance verification, history, and off-chain agent authorization with revocation; Keplr integration or a documented blocker. Acceptance criteria: the wallet verification checklist in section 6.11 passes in full; external-wallet compatibility (O4) passes at product level; no seed or private key is transmitted to any server. Revision note, v1.1, 2026-11-06: O4 is no longer a framework-selection gate (DEFERRED by Benross's decision); it is re-tested here at product level, in M4.

M5, App Integration. Deliverables: the economy dashboard with three separate sections, wallet linking with signature verification, payment requests, agent payment permissions, developer payment APIs, and marketplace settlement gated on verified on-chain payment. Acceptance criteria: an end-to-end payment request flow is verified on the localnet (request, sign, commit, verify, mark paid); the Worker must-never list (section 7.3) is tested and no violation exists; a duplicate webhook or duplicate request cannot fulfil twice (idempotency tested).

M6, RPC and Indexer. Deliverables: public RPC and LCD endpoints, an indexer, and the explorer (OGO-Explorer). Acceptance criteria: the indexer is reconciled against the chain (gate G9); the explorer is usable without an OGO account; a balance and a transaction history displayed in the explorer match the chain exactly; the explorer links to the chain for every value it displays.

M7, Testnet. Deliverables: a public testnet with independent validators, a faucet, documented genesis, and a testnet runbook. Acceptance criteria: independent validators (not operated by OGO Systems) join and produce blocks; a governance upgrade is executed on the testnet; an incident (a validator outage, or a halt) is handled using the runbook and the handling is recorded; the testnet uses TEST-ONLY parameters, labelled so, with no real value.

M8, Mainnet Readiness. Deliverables: all five security audits complete (section 10.5), the legal analysis commissioned in writing and received, genesis parameters approved by Benross, runbooks, monitoring, and an incident response plan. Acceptance criteria: all testing gates G1 to G11 pass; audits have no unresolved critical or high findings; the legal analysis is on file; Benross approves the genesis parameters and the launch in writing. M8 is readiness, not a launch. Launch itself needs a separate approval.

### 9.3 Test and acceptance criteria, per milestone (summary)

Each milestone's acceptance criteria are stated in section 9.2. The proof-of-concept criteria in section 2.6 (P1 to P5, S1 to S8, M1 to M5, O1 to O5) are the acceptance criteria for the framework decision and for M1 to M3. The wallet checklist in section 6.11 is the acceptance criteria for M4. The security gates G1 to G11 in section 10.4 are the acceptance criteria for M8, and are exercised progressively from M1 onward. No milestone is claimed as complete on the basis of code written; it is complete when its acceptance criteria pass with recorded evidence.

### 9.4 CosmWasm evaluation

Decision: CosmWasm is evaluated for future programmable applications and agent services, but it is NOT a dependency of the first monetary prototype (M0 to M3). The evaluation happens after M3, when the chain itself is proven, so a contract problem cannot be confused with a chain problem.

What CosmWasm adds, if adopted: programmable applications on the chain (for example, an on-chain agent authorization contract, an escrow contract for marketplace settlement, or a custom distribution contract), without changing the core modules. It does not add monetary capability: the binding surface is frozen, and no contract can mint.

Evaluation criteria (all eight must pass before CosmWasm is adopted for any production use):

1. The wasm VM instantiates and executes a test contract (proof-of-concept criterion S7).
2. The binding surface is frozen at [BankSend, BankBurn, QueryBalance, QuerySupply], with no mint method, and a contract attempting a mint fails with state unchanged.
3. A contract can call bank send within its balance, and the send preserves total supply.
4. Contract state is isolated: one contract cannot read or write another contract's state.
5. Gas metering prevents infinite loops: a contract that loops indefinitely is halted by gas exhaustion, with state rolled back.
6. The contract upgrade path is documented (how a contract is migrated, who can migrate it, and what happens to its state).
7. The wasmd version in use is patched against CWA-2026-006 (the September 2026 sandbox escape) and any later CosmWasm security advisory, verified against the advisory list at evaluation time.
8. Performance: contract calls do not degrade the proof-of-concept throughput criteria (P2, P3) below their thresholds.

If any criterion fails, CosmWasm is not adopted, and the failure is recorded with the reason. The chain does not depend on CosmWasm, so a failed evaluation does not block M0 to M3.

Sequencing: M0 to M3, no CosmWasm dependency. After M3, evaluate against the eight criteria in a staging environment. If the evaluation passes, CosmWasm is enabled for testnet (M7) with a documented contract review process. If it fails, the evaluation is repeated after the blocking issue is fixed, or CosmWasm is deferred to a later phase. No contract is deployed to a public network without a security review (section 10.5, audit 3).

## 10. Security and Threat Model

### 10.1 Threat model: nine threats

Each threat lists the attack path, the impact, the mitigations (design and operations), and the residual risk. Residual risk is stated honestly: no threat is claimed as eliminated.

Threat 1, single validator key compromise. Attack path: an attacker obtains a validator's consensus private key (through a host compromise, a backup leak, or an insider) and uses it to sign conflicting blocks or to impersonate the validator. Impact: the validator can be slashed for equivocation, its delegators lose stake, and if the compromised key is used to sign a malicious fork, the chain can halt or split. Likelihood drivers: plaintext keys on the validator host (the scaffold's localnet keys are plaintext, which is why they are localnet only), weak host security, and unencrypted backups. Mitigations: remote signing through cosmos/kms, with the consensus key never present as a plaintext file on the validator host (proof-of-concept criterion S8); hardware security module or cloud KMS for key storage; no key backups in plaintext; key rotation through the in-place rotation path (0.55 family) or unbond and re-bond (0.53 matrix); monitoring for unexpected signing behaviour. Residual risk: medium. A compromised remote signer is still a compromise, and the mitigation is detection speed, not prevention.

Threat 2, validator cartel. Attack path: a group of validators controlling more than one third of voting power colludes to halt the chain, censor transactions, or extract value; a group controlling more than two thirds can finalise arbitrary blocks. Impact: the chain halts, transactions are censored, or the ledger is rewritten. This is the single largest residual risk, and it is high at launch, because a new chain has few validators and stake is concentrated. Likelihood drivers: low validator count, concentrated stake, and a small number of entities operating multiple validators. Mitigations: the Definition B decentralization targets (at least 50 validators, Nakamoto coefficient at least 10, top 10 under 33 percent, no country or provider over 33 percent); a mechanism for a flat stake distribution, still to be named; transparent validator identities and voting power, visible in the explorer; a minimum validator set that grows with the network; and governance that can reduce a cartel's power only if the cartel does not already control governance. Residual risk: high at launch, decreasing only as the validator set grows and stake disperses. No framework choice and no code change eliminates this risk. It is a distribution problem, not a software problem, and it must be stated as such to Benross.

Threat 3, malicious governance proposal. Attack path: an attacker submits a governance proposal that changes a parameter, approves a treasury spend, or executes an upgrade that weakens a monetary guard, and the proposal passes because turnout is low, because the attacker controls enough stake, or because voters do not understand what the proposal does. Impact: the schedule is changed, the cap is raised, treasury funds are drained, or a backdoor is introduced through an upgrade. Mitigations: the governance security boundary in section 10.3 (parameter separation, upgrade-handler denylist, amount-less mint entrypoint, every-block invariant, public schedule queryability); voting thresholds, deposit, and quorum set before genesis (TBD); a timelock between a proposal passing and its execution, so the community can react; and a requirement that any proposal touching monetary state is rejected in code, not just by convention. Residual risk: medium. The denylist and the invariant are code, but a hard fork can still replace the code. The protection is transparency and independent operators, not a guarantee.

Threat 4, user wallet key theft. Attack path: an attacker obtains a user's mnemonic or private key (through phishing, malware, a compromised wallet, or a backup leak) and drains the wallet. Impact: the user loses their coins. There is no recovery, because the wallet is noncustodial and there is no password reset. Mitigations: the wallet security rules in section 6.3 (Argon2id, platform keystore, no seed transmitted to any server); the signing prompt rules in section 6.4 (the user sees the exact amount and recipient before signing); user education at wallet creation (the mnemonic is displayed once, with a plain warning that losing it means losing the coins); and the agent authorization limits in section 6.8, which bound what an agent can spend even if the agent is compromised. Residual risk: medium to high. Phishing and malware are the dominant loss vectors in practice, and no wallet design eliminates them. The wallet can only make theft harder and limit the damage.

Threat 5, validator key theft (operational keys). Attack path: distinct from threat 1 (consensus key), this is theft of a validator's operator key, withdrawal key, or the keys used to manage the validator (unbonding, commission changes, governance votes). Impact: the attacker can unbond and withdraw the validator's stake, change its commission to 100 percent, or vote its stake in governance. Mitigations: operator keys held in a separate keystore from consensus keys, with multisig for high-value operations (unbonding, commission changes above a threshold); commission change limits (notice period and maximum change per period, TBD); and monitoring for unexpected operator transactions. Residual risk: medium. Multisig reduces the risk but adds operational complexity, and the multisig itself must be managed.

Threat 6, treasury key theft. Attack path: an attacker obtains control of the treasury module account or the keys that authorise treasury spending, and drains the treasury. Impact: the treasury is emptied, and the network loses its funding for development, grants, and validator support. Mitigations: the treasury is a module account, not a single key; spending goes through x/gov proposals within the R-F rule, so no single key can spend; the R-F rule bounds spending per period; the emergency reserve floor F prevents spending below the floor except through a separate emergency process; and all treasury transactions are public and visible in the explorer. Residual risk: low to medium. The module-account design removes the single-key failure, but a governance compromise (threat 3) can still approve a malicious treasury spend, which is why the R-F rule and the floor matter.

Threat 7, RPC manipulation. Attack path: an attacker compromises or impersonates a public RPC or LCD endpoint, and returns false balances, false transaction statuses, or false supply figures to a wallet or to the OGO application. Impact: a user sees a wrong balance, a payment is marked as paid when it was not, or a marketplace settlement is triggered by a false verification. Mitigations: the wallet queries at least two independent RPC endpoints, or one RPC endpoint plus the indexer, and flags a discrepancy (section 6.6); the Worker verifies on-chain state against a public RPC endpoint and never trusts a client claim (section 7.3); the application labels cached values with the time they were last verified; and the explorer links to the chain for every value it displays. Residual risk: low to medium. The mitigations require the wallet and the application to actually implement multi-source verification, and a user who ignores a discrepancy warning is still at risk.

Threat 8, network partitioning. Attack path: a network partition (a cloud outage, a BGP hijack, a targeted denial of service, or a geographic split) divides the validator set so that no side has more than two thirds of voting power. Impact: the chain halts until the partition heals. If a side has more than two thirds, it continues, and the other side must catch up when the partition heals, with no double mint and no skipped epoch (the halt and catch-up rule, section 3.5). Mitigations: validators distributed across countries, providers, and networks (the Definition B target of no country or provider over 33 percent); sentry nodes that obscure the validator's network location; and the halt and catch-up rule, tested in the proof of concept (S4). Residual risk: medium. A partition that lasts long enough to halt the chain is a liveness failure, not a safety failure, and the chain recovers when the partition heals, but the downtime is real and must be communicated to users.

Threat 9, software supply-chain attack. Attack path: a compromised dependency (a Go module, a CosmWasm library, a build tool, or a CI action) introduces malicious code into the chain binary, the wallet, or the explorer. Impact: the malicious code can mint, steal keys, or exfiltrate data, depending on where it lands. This is a high residual risk, because the chain depends on a large module graph (the Gaia matrix) and on upstream projects that have had vulnerabilities (CWA-2026-006 in CosmWasm, GHSA-c32p-wcqj-j677 in CometBFT). Mitigations: pinned dependencies with hashes (go.sum), verified against the Gaia reference hash; reproducible builds, re-proven after every dependency change (M1); a dependency audit before each upgrade, with a written record of what changed and why; minimal dependencies (no dependency is added without a reason); and the audit plan in section 10.5, which includes a build and supply-chain audit. Residual risk: high. A supply-chain compromise is the hardest threat to detect, because the malicious code is inside a trusted dependency. The mitigations are detection and recovery, not prevention.

### 10.2 Monetary bypass surface

The monetary bypass surface is the set of paths by which the fixed-supply and schedule-only mint rules could be circumvented. Milestone B tested 17 bypass attempts on the scaffold, all failing. The bypass surface must be re-tested on the wired chain (S2), because wiring the keepers adds new paths.

Paths, and the test for each:

- Module accounts: a module account (rewards pool, treasury, distribution, or a custom module account) must not be able to mint, or to receive funds that were not minted through the schedule. Test: attempt to mint from each module account; attempt to transfer from a module account an amount greater than its balance; both fail, state unchanged.
- Governance upgrades: a governance proposal must not be able to change max supply, the schedule, or the final issuance period, either directly (a parameter change) or indirectly (an upgrade that modifies monetary state). Test: submit a proposal to change each; each is rejected or has no effect, and the schedule query returns the original values. Test an upgrade handler that attempts to modify monetary state; the denylist gate rejects it.
- Smart-contract interfaces: a CosmWasm contract must not be able to mint, or to call a bank send above its balance, or to call an unknown method. Test: deploy a test contract that attempts each; all fail, state unchanged (S7). The binding surface is frozen at [BankSend, BankBurn, QueryBalance, QuerySupply].
- Founder mint: there is no founder allocation, no reserved allocation, and no founder mint path. Test: attempt to mint to a founder address, or to create a genesis account with a founder purpose label; both fail.
- Off-schedule mint: minting must happen only at the scheduled amount, only once per epoch, only through the epoch hook. Test: attempt to mint an off-schedule amount; attempt to mint twice in the same epoch; both fail, issued total unchanged.
- IBC escrow and voucher double-counting: a native token escrowed on one chain and represented as a voucher on another must not be counted as two independent units of economic supply. Test: transfer out, verify escrow equals transferred amount and voucher equals transferred amount, and global economic supply is unchanged; transfer back, verify escrow released exactly (O3, DEFERRED by Benross's decision, 2026-11-06; the ledger-level accounting is tested in Milestone B, the live relaying is evaluated separately if a specific technical dependency demonstrates necessity).
- Dust as a minting allowance: dust (the difference between S_target and S_actual) must never be issued as a top-up, a rounding buffer, or a discretionary mint. Test: after all epochs, attempt to mint the dust amount; it fails, and the dust remains unissued.

### 10.3 Governance security boundary: five code layers

The governance security boundary is enforced in five code layers. Each layer is a separate defence, so a failure in one layer does not by itself compromise the monetary rules.

Layer 1, parameter separation: monetary parameters (max supply, schedule, final issuance period, treasury rule) are stored in x/monetary and x/treasury state, not in a generic parameter store that x/gov can change. A governance proposal that targets a monetary parameter has no parameter to change, because the parameter is not in the store it can reach.

Layer 2, upgrade-handler denylist gate: every upgrade handler contains a denylist that rejects any upgrade attempting to modify monetary state (max supply, schedule, issued-by-epoch set, treasury rule) outside the constitutional process. The denylist is checked before the migration runs, and a rejected upgrade does not apply any part of its migration. The denylist is tested by submitting an upgrade that attempts a monetary change and observing the rejection.

Layer 3, amount-less MintScheduled entrypoint: the mint function takes no amount argument. It takes only the epoch number, and it mints exactly R_n for that epoch, as computed from the genesis-fixed schedule. There is no code path that mints an arbitrary amount, so a caller cannot request a larger mint even if it can call the function. This is the same design as the scaffold's MintScheduled(height), migrated to epochs.

Layer 4, every-block invariant: at every block, the chain checks the supply invariants (total supply equals genesis plus issued minus burned; total issued is less than or equal to total scheduled; total supply is less than or equal to max supply). If the invariant fails, the block is rejected or the chain halts, per the configured behaviour, and the failure is logged. The invariant is a separate module (x/monetary invariants), not inline in the mint code, so a bug in the mint code is caught by the invariant.

Layer 5, public schedule queryability: the schedule, the issued-by-epoch set, total issued, total scheduled, dust, and max supply are all public queries. Anyone can verify the rules without trusting OGO Systems, and any deviation is visible in the explorer. Transparency is the last layer, and it is the one that catches a failure in the other four layers after the fact.

Hard-fork caveat, stated honestly: no chain can promise its community will never adopt different software. A hard fork that replaces the code can remove all five layers. The protection is transparent rules, independent operators, and a visible upgrade process, not a claim that a hard fork is impossible.

### 10.4 Testing gates G1 to G11

Each gate has a concrete pass evidence. A gate is not passed by a claim; it is passed by the evidence listed.

G1, independent state validation. Pass evidence: a third party (not OGO Systems, not the team that built the chain) obtains the source, builds the binary, runs the tests, and verifies the supply rules independently, with a written report. The reproducible build (M1) is the prerequisite. The third party's report is filed, not just asserted.

G2, invalid signature and double-spend rejection. Pass evidence: a transaction with an invalid signature is rejected before execution, with the reason logged; a transaction that spends the same funds twice (a double-spend attempt, or a replayed transaction with the same sequence) is rejected, with the reason logged; the chain state is unchanged after each attempt. This closes the Milestone B gap (unsigned transactions accepted).

G3, supply-cap enforcement. Pass evidence: an attempt to mint above the cap fails; an attempt to mint after the final issuance period fails; total supply never exceeds M in any test run, including the halt and catch-up test (S4) and the upgrade test (M2).

G4, deterministic arithmetic. Pass evidence: the same schedule, computed on two different machines and in two different orders, gives the same per-period amounts and the same total, exactly, in integer arithmetic; no floating point appears in any consensus-critical monetary code, verified by a code search; the test vectors in section 4.7 pass.

G5, validator failure handling. Pass evidence: a validator taken offline for longer than the downtime window is jailed and slashed, the chain continues with the remaining validators, and the validator can unjail and rejoin (S5); a double-sign is punished (S6); a validator restored from a snapshot rejoins and catches up (O5).

G6, upgrade safety. Pass evidence: a governance upgrade is executed through Cosmovisor, with a store migration, and balances and supply are identical before and after (M2); the bypass suite is re-run after the upgrade and still passes (M3); a failed upgrade on one validator is rolled back and the validator rejoins (M5); an upgrade attempting a monetary change is rejected by the denylist (section 10.2).

G7, wallet recovery. Pass evidence: a wallet is recovered from a mnemonic on a different device, with the same address and balance; a wallet recovered from a mnemonic generated by Keplr recovers the same address and balance, and vice versa; no seed or private key is transmitted to any server, verified by inspecting network traffic and logs during wallet operations (section 6.11, items 1, 6, and 9).

G8, treasury bypass resistance. Pass evidence: an attempt to spend from the treasury above the R-F rule fails; an attempt to spend below the emergency reserve floor F fails (except through the documented emergency process); an attempt to mint directly into the treasury fails; a governance proposal to change the treasury rule outside the constitutional process is rejected or has no effect.

G9, indexer reconciliation. Pass evidence: the indexer is run against the chain for a full test period, and every block, transaction, and balance in the indexer matches the chain exactly; any discrepancy is logged, investigated, and explained; if the indexer and the chain disagree, the chain wins, and the discrepancy is surfaced, not silently resolved.

G10, operation during node outages. Pass evidence: the chain continues to produce blocks with 3 of 4 validators while one is offline; the application continues to display balances and history from a surviving RPC endpoint while one RPC endpoint is down; the wallet continues to work when OGO's website is unavailable, using only a public RPC endpoint and an independent wallet.

G11, audits. Pass evidence: all five audits in section 10.5 are complete, with written reports; there are no unresolved critical or high findings; medium and low findings have owners and dates, or are documented as accepted risks with reasons.

### 10.5 Audit plan

Five audits, in order, each with a scope and an independence rule. Audits are ordered so the monetary core is audited first, because it is the hardest to fix after launch.

Audit 1, monetary core. Scope: x/monetary, x/treasury, x/rewards, the issuance schedule, the cap enforcement, the issued-by-epoch set, the three-number accounting, the R-F rule, and the every-block invariant. Independence: the auditor has not written any of the monetary code. Findings register: every finding is recorded with severity, status, and owner, and no critical or high finding is left unresolved before mainnet.

Audit 2, consensus, staking, and governance. Scope: the CometBFT configuration, x/staking, x/slashing, x/distribution, x/gov, x/upgrade, the upgrade-handler denylist, and the governance security boundary (five layers). Independence: the auditor has not written any of the consensus or governance code.

Audit 3, contracts and IBC. Scope: the CosmWasm binding surface, the wasm VM configuration, the IBC escrow and voucher accounting, and any contract deployed on the testnet. Independence: the auditor has not written any of the contract or IBC code. This audit is required before any contract is deployed to a public network.

Audit 4, wallet, keys, and RPC. Scope: the wallet (key generation, storage, signing, recovery), the agent authorization mechanism, the RPC and LCD endpoints, and the indexer. Independence: the auditor has not written any of the wallet or RPC code. This audit includes a network-traffic inspection to verify no seed or private key is transmitted.

Audit 5, build and supply chain. Scope: the reproducible build, the dependency matrix, the CI pipeline, and the release process. Independence: the auditor has not configured the build or the CI. This audit verifies that the binary a user builds from source matches the binary that is released, and that no dependency has been substituted.

Findings register: a single register, with every finding from every audit, its severity (critical, high, medium, low), its status (open, fixed, accepted risk), its owner, and the date it was resolved or accepted. An accepted risk must have a written reason, and the reason must be approved by Benross. A critical or high finding is never an accepted risk before mainnet.

### 10.6 Residual risks

Ten residual risks, in order of severity. Each is a risk that remains after the mitigations in this plan, and each must be communicated to Benross, not hidden.

1. Stake concentration and validator cartel (threat 2). High at launch. No code change eliminates it. It decreases only as the validator set grows and stake disperses.
2. Post-issuance fee funding (section 5.4). The fee transition has not happened on any comparison network. Under Low and Moderate fee scenarios, fee revenue never covers the minimum need. Under High, it does so only late. If fees never cover the need, the validator set shrinks or security degrades.
3. Software supply-chain compromise (threat 9). High. A compromised dependency is the hardest threat to detect. Mitigations are detection and recovery, not prevention.
4. Upgrade removing the checks. A hard fork or a malicious upgrade can remove the monetary guards. The denylist and the invariant are code, and code can be replaced. The protection is transparency and independent operators.
5. Governance capture. If an attacker controls enough stake, or if turnout is low, a malicious governance proposal can pass. The thresholds, deposit, and quorum are TBD, and they must be set to make capture expensive, not just possible.
6. Wallet key theft through phishing or malware (threat 4). The dominant loss vector in practice. The wallet can only make theft harder and limit the damage.
7. IBC and bridge accounting errors. The escrow and voucher accounting is tested at ledger level, but live IBC relaying is not yet wired, and any bridge is third-party trust. An accounting error in a bridge can create unbacked vouchers.
8. Epoch and timestamp edge cases. The epoch-based period architecture depends on the epochs module and on the halt and catch-up rule. An edge case in the epochs module, or a long halt, could cause a missed or double epoch. The proof of concept tests the 30-minute halt (S4), but longer halts and repeated halts are not yet tested.
9. Fee market mispricing for agents. A flat fee cannot serve both security and agent viability. If the fee is set too high, small agent payments become impractical. If it is set too low, validators are underpaid. The fee policy must state explicitly how agent transactions are priced, and it has not yet done so.
10. Legal and regulatory change. The legal analysis is not yet commissioned. A change in securities, money-transmission, or tax law could affect the network, the token, or the distribution. This is an external risk, not a code risk, and it must be tracked as one.

## 11. Open Decisions Register

This register uses the categories from the continuation commission: Benross decisions (decided), technical recommendations (recommended, awaiting Benross approval), open decisions (not yet decided), parameters awaiting approval (numbers not yet selected), and items needing external legal or security review. Nothing in the open-decision or parameter categories is decided by this plan.

### 11.1 Benross decisions (decided, recorded)

| Decision | Date | Notes |
|---|---|---|
| OGO Network is an independent Layer 1 blockchain; OGO Coin is its native currency, not an ERC-20 or SPL token | 2026-11-06 | Feasibility report option B. The conditional Base recommendation is superseded for the native coin. Guiding principle: OGO owns its monetary infrastructure, while users own their coins and private keys. |
| Option B confirmed and locked | 2026-11-06 | Do not reopen Base, Ethereum, or Solana token-first alternatives unless new evidence creates a compelling reason. |
| Blockchain lives in separate repositories: OGO-Network, OGO-Wallet, OGO-Explorer, OGO-Network-Docs | 2026-11-06 | Resolves the repo tension. OGO-Network created 2026-11-06 (public, benross-lab/OGO-Network) by Benross's explicit authorization. OGO-Wallet, OGO-Explorer, and OGO-Network-Docs are not created; each creation still needs explicit approval. |
| Fixed maximum coin supply, community governance | 2026-11-05 | Founding charter. "OGO Systems creates the initial technology. The OGO community ultimately governs the network." |
| No OGO account needed to own, transfer, or use OGO Coin | 2026-11-05 | Usable via independent wallets, exchanges, developers, and third-party services. |
| OGO Coins and Credits remain separate from OGO Coin; no automatic conversion | 2026-11-05 | Accounting must never be merged. |
| Maximum supply is TBD, but is a mandatory genesis parameter with no default | 2026-11-05 | "I would not choose a number just to move development forward." Build the monetary rules first; decide the monetary quantity second. |
| Zero reserved founder allocation; zero reserved OGO Systems allocation; no preferential private presale | 2026-11-05 | Founders may acquire OGO later under the same publicly available conditions as everyone else. OGO Systems is the founding technology developer, not the owner of the currency. |
| IBC transfers must preserve global supply accounting | 2026-11-05 | A native token escrowed on one chain and represented as a voucher on another must not be counted as two independent units of economic supply. |
| Fixed-supply protections cannot be bypassed through privileged module accounts, governance upgrades, or smart-contract interfaces | 2026-11-05 | The build and test report must include actual failures, not just a successful-looking scaffold. |
| Independent reproducibility: someone outside OGO Systems must be able to obtain the source, build the same software, run the tests, and independently verify the supply rules | 2026-11-05 | Monetary rules must be enforced by the protocol, not by trusting OGO Systems. |
| Naming: Network OGO Network, native coin OGO, ticker OGO, Wallet OGO Wallet, Explorer OGO Scan | 2026-11-05 | Locked. |
| Go 1.25.7 locked as prototype build toolchain; Gaia v28.3.1 locked as dependency reference | 2026-11-05 | Entire matrix as one tested unit. Go 1.23 kept as documented original minimum only, not a verified build target. |
| A fixed-supply network must design its eventual end of issuance and its treasury economics together | 2026-11-06 | Benross's reframing of the Decision Report. Cardano is the most important warning; Polkadot shows monetary changes affect the entire ecosystem; Solana shows the appeal of a floor, with OGO's variation making the floor temporary with a predetermined end; Ethereum shows burning is not a substitute for a funding model. |
| No em dashes in OGO copy and chat | Standing | Use commas, colons, periods, and bullets. Hyphens inside words are kept. |

### 11.2 Technical recommendations (recommended, awaiting Benross approval)

| Recommendation | Section | Status |
|---|---|---|
| Cosmos SDK plus CometBFT, continuing from the verified Milestone B scaffold | 2.5 | Recommended, not final. Final selection requires the proof of concept (section 2.6) to pass. |
| Substrate as the credible fallback, not the default | 2.5 | If the proof of concept fails for a structural reason, the next step is a time-boxed Substrate spike, not a full rewrite. Not authorized by this plan. |
| Epoch-based periods via an epochs module | 3.5 | Recommended. Period length D is not recommended here. |
| Three custom modules only: x/monetary, x/treasury, x/rewards. No x/governance, no x/fees | 3.4 | Recommended. x/monetary is the production name for the scaffold's x/ogocoin. |
| Three-number accounting: M = G + S_actual, computed at genesis by exact integer summation; dust left permanently unissued | 4.2 | Recommended, model-independent. |
| Gentle geometric taper (factor about 0.95 over the final 20 periods) as a hybrid candidate | 4.3 | Candidate if hybrid is chosen. Short linear tapers are rejected as worse than abrupt. |
| R-F treasury rule: spend = min(D, max(Trail_4, floor(0.05 x B)), max(0, B - F)) | 4.4 | Recommended default. B is the treasury balance at the end of the previous period. Transition keyed to the genesis-fixed final issuance period N. |
| Fixed final period in genesis for geometric, so the end date is predictable | 4.3 | Recommended, not a decision. |
| CosmWasm evaluated after M3, not a dependency of the first monetary prototype | 9.4 | Recommended. Eight pass/fail criteria must pass before adoption. |
| Validator tiers: lean 75 USD, standard 200 USD, professional 500 USD per validator per month, infrastructure only | 5.2 | Recommended as the cost basis for planning. Labour is separate (1,500 to 4,000 USD per validator per month at the professional tier). |
| Hub precedents as proposals: 21-day unbonding, 5 percent commission floor, downtime 0.01 percent plus 10 minute jail, double-sign 5 percent plus tombstoning | 5.3 | Proposals, not decisions. Each needs approval. |

### 11.3 Open decisions (not yet decided)

| Decision | Blocks | Notes |
|---|---|---|
| Issuance model family: geometric or hybrid | Parameters, split, fee analysis, max supply | Geometric and hybrid remain the two candidates. Epoch-based is the legibility fallback. Halving is a reference only. |
| Period length D and nominal year | Issuance schedule, treasury projections, fee transition timing | A sustainability decision. Geometric ends at period 110 (9.2 years monthly, 27.5 years quarterly); hybrid ends at period 40 (3.3 years monthly, 10.0 years quarterly). |
| Validator minimum: Definition A (4 validators) or Definition B (at least 50, Nakamoto at least 10, top 10 under 33 percent, no country or provider over 33 percent) | Validator set, fee need, decentralization claims | Definition B targets are targets, not launch requirements. A mechanism for a flat stake distribution still needs to be named. |
| Fee policy: fee level, fee distribution, and how agent transactions are priced | Post-issuance funding, validator pay | Only the High fee scenario covers the minimum need, and only at period 38 (or 48 at standard costs). A flat fee cannot serve both security and agent viability. The fee market must not make small AI-agent payments impractical. |
| Treasury spending level D and emergency reserve floor F | Treasury sustainability | No spending policy fixes a level problem. Maximum sustainable constant spend is 909 per period geometric, 2,500 hybrid. The 30-year desired spend is 540,000 against about 100,000 total treasury inflow. |
| Treasury policy: R-F (recommended) or P4 (candidate) | Treasury spending rule | P4 remains a candidate, not an approved rule. P4 has a known post-issuance freeze defect. |
| Governance model: voting thresholds, deposit, voting period, quorum, expedited proposal path | Governance security | All TBD. The Hub's 72-hour expedited path is a precedent worth evaluating, not adopted. |
| Network name and chain id | Genesis, chain registry, wallet integration | The network name is OGO Network (locked). The chain id is not yet chosen. |
| Denomination and decimal precision | Genesis, wallet, explorer, chain registry | Proposed: `uogo`, 6 decimals. Proposals, not decisions. Cannot be changed after genesis without a hard fork. |
| Wallet address format: bech32 prefix and derivation path | Wallet, chain registry, external wallet integration | Proposed: prefix `ogo`, path `44'/118'/0'/0/0`. Both provisional. Must be fixed before the first wallet is released. |
| Commission floor, notice period, and maximum change per period | Validator economics, delegator protection | Proposed floor 5 percent (Hub precedent). Notice period and maximum change are TBD. |
| Unbonding period | Staking, slashing effectiveness | Proposed 21 days (Hub precedent). |
| Slashing parameters: downtime window, missed-block threshold, slash fractions | Validator reliability, delegator risk | Proposed: Hub precedent. Window and threshold are TBD. |
| Wallet technology stack and framework | M4 | Decided in M4, not in this plan. Must support Keplr, Leap, and Cosmostation integration and the hardware signing seam. |
| Explorer technology stack | M6 | Decided in M6, not in this plan. Must be usable without an OGO account. |
| Post-issuance funding mechanism and fee share | Treasury and validator funding after period N | Must be named before genesis, with revenue modelled, not assumed. Mechanisms: fee share, blockspace or coretime sales, endowment, contributions. Treasury staking is circular and is not a funding mechanism. |
| Protocol revenue vs OGO company revenue | Accounting, treasury, fee distribution | Recorded principle from the continuation commission: protocol revenue is kept separate from OGO company revenue. The separation must be documented in the treasury and fee specifications before genesis, with the accounts, the flows, and the audit trail stated explicitly. |
| Liquid staking | Later phase | Not part of the first monetary prototype. Must be evaluated against the same supply and slashing rules before adoption. No liquid staking token is created by this plan. |
| Bridge to a non-IBC chain | Interoperability | No bridge is adopted by this plan. Any bridge is third-party trust and must be evaluated under the same supply-accounting rules as IBC. |

### 11.4 Parameters awaiting approval (numbers not yet selected)

No parameter in this table is selected by this plan. Each needs Benross's explicit approval, in the decision order in section 4.8.

Maximum supply M (derived as G + S_actual, not chosen directly); genesis circulating supply G; issuance duration and period count N; reduction factor r; floor and base share (hybrid); initial per-period amount; validator, delegator, community, and treasury split; fee level and fee distribution; treasury spending level D and emergency reserve floor F; minimum validator count and active set size N; governance deposit, voting period, quorum, and threshold; epoch duration D and nominal year; denom and decimals; bech32 prefix and derivation path; commission floor and change limits; unbonding period; slashing fractions and windows; chain id.

### 11.5 Items needing external legal or security review

| Item | Status | Notes |
|---|---|---|
| Legal analysis: securities, money transmission, tax, international distribution, exchange admission | Not yet commissioned | Must be commissioned in writing before genesis distribution. Longest-lead item. Nothing in this plan is legal advice. |
| Security audits 1 to 5 (monetary core, consensus and governance, contracts and IBC, wallet and RPC, build and supply chain) | Not yet commissioned | Ordered in section 10.5. No unresolved critical or high findings before mainnet. |
| Economic review of the fee market | Not yet commissioned | The fee transition is the binding problem. An independent economic review of the fee scenarios, the agent pricing question, and the post-issuance funding gap is needed before genesis. |
| Fee-market evidence for agent transactions | Open | Agents are 45 percent of transactions but 7.7 percent of fee revenue in the High scenario. The fee policy must state explicitly how agent transactions are priced. |

### 11.6 Open questions from the repository inspection

| Question | Status | Notes |
|---|---|---|
| `src/OgoEconomy.tsx` exists on live main but no file imports it; the wallet UI appears unmounted | Open | Confirmed by the coordinator against live main. Needs one live UI confirmation: does the wallet UI render anywhere in the application, or is the component dead code? |
| Worker queries `forge_symbols` and calls `forge_search`, and the Builder Engine writes to `user_apps`, but no repository migration defines those tables | Open | Likely dashboard SQL. Any integration depending on Forge or `user_apps` must confirm the live schema in the dashboard before relying on it. |
| Local working copy contains `supabase/migrations/20261028_connect_phase0.sql`, not on live main | Open | The working copy is stale and is not a git repository. Any M6 implementation work must start from a fresh clone. |
| Naming collision: OGO Token vs OGO Coin vs OGO Coins | Open | The native coin is OGO, ticker OGO (locked). OGO Coins is the existing Supabase account-based balance. The feasibility report flagged the collision; it remains unresolved as a naming and communication question, not as a technical one. |

### 11.7 What this plan does not decide

For the avoidance of doubt: this plan does not select a monetary model, does not select any monetary parameter, does not approve a repository creation, does not approve a commit or a push, does not approve a publication or a deployment, does not approve a testnet or a mainnet, does not approve any spending, does not commission legal analysis or an audit, and does not authorise token issuance or distribution. Each of those needs a separate, explicit approval from Benross.

## 12. Estimated engineering resources and infrastructure requirements

All figures in this section are estimates for planning. No spending is authorized by this plan. No hiring is authorized by this plan. Costs exclude audits, legal analysis, incentives, and salaries unless stated.

### 12.1 Team shape

Nine roles, not all full-time, not all needed from day one. The shape is an estimate, and the actual team depends on who is available and on Benross's hiring decisions.

| Role | Focus | Needed from |
|---|---|---|
| Protocol engineer (Cosmos SDK, Go) | Chain, modules, genesis, upgrades | M0 to M1 |
| Monetary engineer (or protocol engineer with monetary focus) | x/monetary, x/treasury, x/rewards, issuance arithmetic, test vectors | M1 to M2 |
| Validator operations engineer | Localnet, testnet, runbooks, monitoring, incident response | M1, then M3 and M7 |
| Wallet engineer | OGO Wallet, key management, signing, external wallet integration | M4 |
| Application engineer (OGO app) | Economy dashboard, wallet linking, payment requests, agent permissions, developer APIs | M5 |
| Indexer and explorer engineer | Indexer, public API, explorer | M6 |
| Security engineer (or external auditor) | Threat model, bypass tests, audit coordination, findings register | M0 onward, audits at M8 |
| DevOps and infrastructure engineer | CI, reproducible builds, deployment, monitoring, snapshots | M0 onward |
| Technical writer | Protocol docs, validator guides, API specifications, OGO-Network-Docs | M0 onward |

A single engineer can cover more than one role in the early milestones (for example, protocol and monetary engineering in M1 to M2), but the roles are listed separately so the work is visible and so no role is silently dropped.

### 12.2 Validator cost matrix

Infrastructure only, per month, in USD. Labour is separate. Lean is 75 USD, standard is 200 USD, professional is 500 USD per validator per month. The matrix is the continuation commission's requested shape (4, 25, 50, and 100 validators), computed from the tier rates in section 5.2.

| Validators | Lean (75 USD each) | Standard (200 USD each) | Professional (500 USD each) |
|---|---|---|---|
| 4 (Definition A, liveness floor) | 300 | 800 | 2,000 |
| 25 | 1,875 | 5,000 | 12,500 |
| 50 (Definition B, minimum) | 3,750 | 10,000 | 25,000 |
| 100 | 7,500 | 20,000 | 50,000 |

Labour, per validator per month, is separate and dominates at the professional tier: 1,500 to 4,000 USD. A 50-validator professional set therefore costs 75,000 to 200,000 USD per month in labour alone, plus 25,000 USD per month in infrastructure. The minimum fee need in the research (14,625 per period at lean costs, 24,625 at standard costs, for 100 validators illustrative) is infrastructure only. If labour is included, the need is higher, which is why the fee transition is the binding problem.

Testnet infrastructure: 775 to 2,350 USD per month, for a small validator set, RPC endpoints, an indexer, and an explorer, in the integration draft's estimate. Mainnet OGO-operated infrastructure (the validators, RPC endpoints, indexer, explorer, and monitoring that OGO Systems itself operates, not the whole network): 2,600 to 5,700 USD per month. Both figures exclude audits, legal analysis, incentives, and salaries.

### 12.3 Infrastructure requirements, by milestone

M0 to M1 (localnet): a development machine with Go 1.25.7, the verified dependency matrix, and the ability to run 3 local validator nodes. Revision note, v1.1, 2026-11-06: v1.0 said 3 to 4 local validator nodes; the count is now 3, decided by Benross. No cloud infrastructure. No public endpoints.

M2 to M3 (monetary and validator prototypes): the same localnet, plus a remote signer (cosmos/kms) for at least one validator (S8). Revision note, v1.1, 2026-11-06: a second local chain for the IBC smoke test (O3) was listed here in v1.0; O3 is DEFERRED by Benross's decision. Still no public infrastructure.

M4 (wallet prototype): a localnet plus public CORS-enabled RPC and LCD endpoints on a staging host, so Keplr and Leap can connect. The endpoints are staging only, with no real value.

M5 (app integration): a staging OGO application connected to the staging chain through public RPC endpoints, plus a staging indexer. No production application change.

M6 (RPC and indexer): public RPC and LCD endpoints, an indexer with a database, and the explorer, on staging infrastructure. Monitoring and alerting for block height, block time, missed blocks, peer count, mempool size, and total supply.

M7 (testnet): a public testnet with independent validators, a faucet, documented genesis, and a testnet runbook. Public RPC, LCD, indexer, and explorer endpoints. Incident response exercised.

M8 (mainnet readiness): the production infrastructure plan, with OGO-operated validators, sentry nodes, remote signing, monitoring, alerting, snapshots, and a rollback procedure, all documented and tested. Mainnet itself needs a separate approval and is not part of M8.

### 12.4 Monitoring and deployment architecture

Monitoring: block height, block time, missed blocks per validator, peer count, mempool size, total supply, issued total, treasury balance, and governance proposal status, visible on a dashboard or through log queries, with alerts for missed blocks, halted chain, supply invariant failure, and upgrade height approaching. The proof of concept requires monitoring and alerting to be demonstrated (O2).

Deployment: reproducible builds from source (M1), with the binary hash recorded and verified; coordinated upgrades through Cosmovisor (M2); store migrations tested on the proof-of-concept chain before they are proposed; rollback tested on a single validator (M5); snapshots taken regularly and tested by restoring a validator from a snapshot (O5). No deployment to a public network happens without a separate approval.

### 12.5 What is excluded from these estimates

Audits (five, section 10.5); legal analysis (section 11.5); validator incentives or staking rewards (the issuance schedule, TBD); exchange listing or market-making costs; community grants or treasury spending (the spend level D, TBD); salaries; and any cost of a Substrate fallback spike (not authorized). Each excluded item needs its own estimate and its own approval before it is incurred.

---

## Sources

Prior OGO Network reports, all in `~/workspace/goals/ogo-ai-builder-and-network/files/`:

- `ogo-network-blueprint-v0.1.md` (2026-11-05)
- `ogo-network-implementation-report-v0.1.md` (2026-11-05, with a 2026-11-06 addendum)
- `ogo-network-monetary-constitution-v0.1.md` (2026-11-05/06)
- `ogo-network-milestone-b/ogo-network-milestone-b-build-report.md` and `ogo-network-milestone-b/REPRODUCIBLE_BUILD.md` (2026-11-06)
- `ogo-network-issuance-mathematics-v0.1.md` (2026-11-06)
- `ogo-network-geometric-vs-hybrid-deep-dive-v0.1.md` (2026-11-06)
- `ogo-network-monetary-design-decision-report-v0.1.md` (2026-11-06)
- `ogo-network-hybrid-terminal-and-treasury-v0.1.md` (2026-11-06)
- `ogo-network-treasury-sustainability-v0.2.md` (2026-11-06)
- `ogo-network-monetary-architecture-decision-preparation-v0.1.md` (2026-11-06)
- `ogo-public-cryptocurrency-feasibility-v0.1.md` (2026-11-06)
- `ogo-network-milestone-b-authorization-20261105.md` and `ogo-network-milestone-b-review-checklist.md` (2026-11-05/06)

Workstream drafts consolidated into this plan, in `/tmp/ogo-native-plan-drafts/`:

- `01-repo-inspection.md` (187 lines, 31,828 bytes)
- `02-framework-evaluation.md` (243 lines, 55,250 bytes)
- `03-protocol-monetary.md` (594 lines, 96,560 bytes)
- `04-wallet.md` (328 lines, 46,620 bytes)
- `05-integration-plan.md` (721 lines, 63,770 bytes)
- `06-security-threat-model.md` (360 lines, 86,069 bytes)

External sources, checked 2026-11-06, as cited in the workstream drafts:

- Cosmos SDK and CometBFT release reporting (SDK v0.55.0, 2026-07-27; CometBFT v0.40.0; Ledger Security 2026.1, 2026-10-08; Gaia v28.3.1 go.mod hash).
- Polkadot SDK stable2609 release reporting (node v1.25.0, 2026-10-08; 87 breaking changes in the stable2609-rc1 notes; Rust 1.93.0; Ankr changelog, 2026-10-26).
- Cosmos Hub governance and slashing precedents (Proposal 1055, September 2026; 0.01 percent downtime plus 10 minute jail; 5 percent double-sign plus tombstoning; 21-day unbonding; 5 percent commission floor).
- CosmWasm security advisory CWA-2026-006 (September 2026, Wasmer compiler sandbox escape) and CometBFT advisory GHSA-c32p-wcqj-j677 (February 2026, median-time manipulation, fixed in CometBFT v0.38.21).
- IBC ecosystem reporting (115 plus connected chains, October 2026; IBC Eureka, March 2025) and bridge exploit totals (over 2.8 billion USD lost since 2022).
- Solana, Cardano, Polkadot, and Ethereum monetary facts, as recorded in the Decision Report section 9 and the comparative research notes in `~/workspace/research_notes/ogo-network-comparative-monetary-researc-20261106-0538/`.

All illustrative numbers in this plan are comparison inputs only and are labelled as such where they appear. No number in this plan is a selected parameter, a proposed supply, or a commitment. The decisions stay with Benross.

## Companion specifications (index)

This plan is the overview and index for eleven specification documents, delivered 2026-11-06 in `ogo-network-architecture/`. Each expands one part of this plan into a standalone specification. Where a specification disagrees with this plan, the disagreement must be resolved explicitly and recorded, not silently overwritten.

1. `OGO_NETWORK_ARCHITECTURE.md`: architecture diagram, component responsibility map, proposed repository structure, dependency and module map, security boundaries, recommended development stack. Includes the x/epochs investigation: x/epochs is a first-class Cosmos SDK module from v0.53.0 (upstreamed from Osmosis, PR #23815), appropriate for OGO's monetary schedule subject to four conditions, including a hook-fires test for the depinject hook-registration bug (issues #24770, #25425).
2. `OGO_CONSENSUS_AND_EPOCH_SPEC.md`: epoch-based monetary architecture: epoch lifecycle, transition recognition, exactly-once issuance, duplicate prevention, halt recovery (missed epochs processed sequentially, no lump-sum mint, no silent skipping), epoch ordering, upgrades, parameter protection. Compares block-based and epoch-based periods. Epoch duration not selected.
3. `OGO_MONETARY_ENGINE_SPEC.md`: the monetary engine: all monetary choices as genesis parameters, a common issuance interface so the model is selected through genesis configuration, both models in the simulator, exact integer arithmetic, three-number accounting, 11 supply invariants. No model selected.
4. `OGO_VALIDATOR_AND_STAKING_SPEC.md`: staged validator strategy (local, private multi-validator testnet, public incentivized testnet, mainnet), onboarding, staking, slashing, unbonding, monitoring, distribution, remote signing. Decentralization targets as research targets. Cost model for 4, 25, 50, and 100 validators at lean, standard, and professional tiers, and what funding validator participation requires without assuming an OGO/USD price.
5. `OGO_TREASURY_AND_SUSTAINABILITY.md`: R-F treasury model evaluated under low, moderate, and high fee revenue, with a computed scenario model (script `/tmp/treasury_v03.py`): under Low fees the treasury ends at the floor under both models; under Moderate and High fees the full desired spend is paid; P4 is the most conservative rule but has a post-issuance freeze defect. Protocol treasury income kept separate from OGO company revenue.
6. `OGO_NATIVE_WALLET_SPEC.md`: noncustodial wallet: balances, sending, receiving, history, staking, fees, addresses, signing, recovery, external wallets, signature-verified ownership, scoped agent authorization. Designed so OGO Coin is never confused with OGO platform credits; the Supabase wallet is a separate accounting system, never merged or silently converted.
7. `OGO_FEE_MARKET_SPEC.md`: fee architecture for ordinary and AI-agent micropayments: base fees, gas pricing, dynamic adjustment options, distribution, validator compensation, treasury allocation, spam prevention, batching. Computed test: at a 1,000 uogo fee, a 0.001 OGO payment costs 100 percent in fees; batching 100 payments makes it 0.10 percent. Application fees are not protocol revenue.
8. `OGO_NETWORK_API_SPEC.md`: integration interfaces for accounts, projects, agents, developer tools, marketplace, external wallets, external chains, and (longer term) OGO Oasis services. Public by default; the network is usable by external developers through documented APIs and SDKs, not exclusively through OGO's interface.
9. `OGO_NETWORK_SECURITY_MODEL.md`: the threat model (nine threats) and the testing gates G1 to G11, the monetary bypass surface, the five-layer governance security boundary, the five-audit plan, and ten residual risks.
10. `OGO_NETWORK_IMPLEMENTATION_ROADMAP.md`: Phase 1 Architecture, Phase 2 Local prototype (including initializing a separate OGO Network repository, local dev chain, native denomination, basic transfers, deterministic monetary-engine interfaces, configurable issuance models, automated accounting tests), Phase 3 Multi-validator testnet, Phase 4 Public testnet preparation, Phase 5 Mainnet readiness. Security, legal, and genesis gates not to be skipped.
11. `OGO_NETWORK_DECISION_REGISTER.md`: decisions explicitly made by Benross, technical recommendations, open decisions, parameters awaiting approval, and items requiring external legal or security review, plus superseded items and the decision order.
