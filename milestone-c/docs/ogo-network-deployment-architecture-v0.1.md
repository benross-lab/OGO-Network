# OGO Network Deployment Architecture v0.1

Date: 2026-11-06
Status: Documentation only. No infrastructure purchased, nothing deployed, no testnet or mainnet launched.
Diagram: [ogo-network-deployment-architecture-diagram.svg](ogo-network-deployment-architecture-diagram.svg)
Sources: the 11 architecture specifications in `ogo-network-architecture/`, the implementation plan v1.1, the wallet API contract v0.1, the Milestone C plan v0.1, the three-validator prototype build report, and the four-validator PoC report. Every decided answer cites its source file and section.

## How to read this document

Each answer is labelled:

* **DECIDED:** Benross decided it, or it is recorded as decided in the decision register.
* **RECOMMENDED:** a recommendation with reasoning. The decision is still Benross's.
* **OPEN:** needs Benross's decision, or more evidence, before it can be treated as settled.

Where the specifications are silent, this document says so and gives ranges, not invented exact figures. Prices from web research are snapshots dated 2026-11-06 and must be rechecked before any purchase. Nothing in this document authorizes spending, a public testnet, or a mainnet launch.

## Summary: the 10 questions, one line each

| # | Question | Short answer | Status |
|---|----------|--------------|--------|
| 1 | Independent Layer 1? | Yes. Own genesis, native OGO Coin, own consensus, own validator network. Option B, confirmed and locked. | DECIDED |
| 2 | Consensus mechanism? | CometBFT Byzantine fault tolerant consensus, Tendermint style, instant finality, greater than two thirds quorum. | DECIDED |
| 3 | Node software? | The `ogod` binary: Go 1.25.7, Cosmos SDK 0.53.6 (effective 0.53.4), CometBFT 0.38.23, wasmd 0.60.9, Cosmovisor 1.7.3. | DECIDED |
| 4 | Hardware specs? | Standard tier per validator: 4+ CPU cores, 32 GB RAM, 500 GB NVMe, 100 Mbps. Lean and Professional tiers defined in the validator spec. | DECIDED (tiers), RECOMMENDED (mapping) |
| 5 | Cloud dev network and testnet? | A private cloud devnet is technically feasible. A public testnet needs separate explicit approval and is not authorized. | RECOMMENDED / OPEN |
| 6 | Provider? | Zero cost local development first. For paid hosting later: Hetzner recommended, DigitalOcean runner-up. Oracle Always Free is the only free-tier remote option, with limits. | RECOMMENDED |
| 7 | Validator count for the initial network? | Proposal: 7 validators for the initial private testnet network. Launch minimum still open (4 vs 50+). | RECOMMENDED / OPEN |
| 8 | Supabase vs blockchain? | Chain is the source of truth for native balances. Supabase is the source of truth for auth, profiles, app data, and the Coins/Credits ledger. Never merged. | DECIDED |
| 9 | Wallet RPC endpoints? | CometBFT RPC for broadcast, LCD/REST for balances and queries, indexer for history. LCD and indexer do not exist yet; Milestone C delivers them. | DECIDED (design), OPEN (live endpoints) |
| 10 | What is needed for mainnet? | Five audits, written legal analysis, genesis parameters approved in writing, operations readiness, incident exercise, and a separate written launch approval. Most of it not started. | OPEN |

## 1. Is OGO being built as an independent Layer 1 blockchain?

**DECIDED.** Yes. OGO Network is an independent, open-source Layer 1 blockchain with its own genesis, its own native coin, its own consensus, and its own validator network.

* "OGO Network is an independent, open-source Layer 1 blockchain. Its native asset is OGO Coin." (OGO_NETWORK_ARCHITECTURE.md, Sec. 2)
* "Architecture direction: CONFIRMED AND LOCKED, Option B. OGO Network launches with its own native blockchain and native OGO Coin. Do not reopen Base, Ethereum, or Solana token-first alternatives unless new evidence creates a compelling reason." (OGO_NETWORK_ARCHITECTURE.md, header; OGO_NETWORK_DECISION_REGISTER.md, rows 24-25, 2026-11-06)
* "The OGO application is a client of the network, not its owner. The network must operate independently of OGO Systems, and must be usable without an OGO account, through independent wallets, exchanges, developers, and third-party services." (OGO_NETWORK_ARCHITECTURE.md, Sec. 2)
* "OGO owns its monetary infrastructure, while users own their coins and private keys." (OGO_NETWORK_DECISION_REGISTER.md, row 24)

What this means in practice:

* OGO Coin is not an ERC-20 token, not an SPL token, and not a token on Base, Ethereum, or Solana. The feasibility report's conditional Base recommendation for the native coin is superseded.
* The network has its own genesis file, its own chain ID, its own validator set, and its own consensus. The OGO application talks to it only through public RPC and API services.
* Three money systems exist and never merge: native OGO Coin (on-chain, in x/bank), OGO Coins (account-based, in Supabase), and OGO Credits (account-based, in Supabase). No combined balance is ever shown, and no automatic conversion exists. (OGO_NETWORK_ARCHITECTURE.md, Sec. 2)

What is not yet decided:

* The framework (Cosmos SDK plus CometBFT) is the leading candidate, not a finalized selection. Final selection requires the proof of concept to pass its 21 MUST criteria. (OGO_NETWORK_ARCHITECTURE.md, Sec. 9)
* Genesis exists only as example files with no real parameters until Benross approves them. Open items include network name and chain ID, denomination and decimals, bech32 prefix and derivation path. (OGO_NETWORK_ARCHITECTURE.md, Sec. 4 and Sec. 9)
* No repository has been created on GitHub, no blockchain is deployed, no token or smart contract is deployed, and no monetary parameter is selected. (OGO_NETWORK_ARCHITECTURE.md, header)

## 2. What consensus mechanism is being implemented?

**DECIDED** as the design; **OPEN** as a finalized framework selection, which depends on the proof of concept (M2 is the deciding blocking criterion).

* "CometBFT Byzantine fault tolerant consensus, Tendermint-style, with instant finality per block. A block is final when committed. There is no probabilistic confirmation depth." (OGO_CONSENSUS_AND_EPOCH_SPEC.md, Sec. 2)
* Quorum: more than two thirds of voting power must sign. The chain halts if more than one third of voting power is offline. The proof of concept must demonstrate continued block production with 3 of 4 validators. (OGO_CONSENSUS_AND_EPOCH_SPEC.md, Sec. 2)
* Block time target: 1 to 6 seconds typical for CometBFT chains. Proof of concept criterion: 95 percent of blocks within 1 to 10 seconds over a 24-hour run. (OGO_CONSENSUS_AND_EPOCH_SPEC.md, Sec. 2)
* Validator set: the top N validators by bonded stake, where N is a genesis parameter (TBD). (OGO_CONSENSUS_AND_EPOCH_SPEC.md, Sec. 2; OGO_VALIDATOR_AND_STAKING_SPEC.md, Sec. 4)
* Block timestamps use the weighted median of validator timestamps. (OGO_CONSENSUS_AND_EPOCH_SPEC.md, Sec. 2)
* Epochs use x/epochs, a first-class Cosmos SDK module from v0.53.0, upstreamed from Osmosis in PR #23815. Epoch duration is not selected. (OGO_CONSENSUS_AND_EPOCH_SPEC.md, header and Sec. 4)

Measured evidence so far:

* Three-validator prototype: blocks produced and verified at heights 5, 15, 16, 17, and 18 with identical app hash, balances, supply, and issuance on all three nodes. Stopping one validator halted the chain, as expected: 2 of 3 is exactly two thirds and is not a quorum. Restart resumed blocks. (ogo-network-prototype-build-report.md, Sec. 3.1)
* Four-validator PoC: with one validator down, the remaining 30 of 40 voting power (3/4) stayed above quorum and blocks continued (height 18 to 30 over 30 seconds). The stopped validator restarted, caught up to height 31, and all four nodes agreed on app hash, balances, and supply. S5 liveness is accepted as passed and recorded as established. (ogo-network-poc-4validator-report.md, Sec. 4; OGO_NETWORK_ACCEPTANCE_CRITERIA.md, Sec. 3.6)

Slashing parameters are proposals, not decisions (Hub precedents): downtime 0.01 percent slash plus 10 minute jail, with the signed-blocks window and missed-block threshold TBD; double-sign 5 percent slash plus tombstoning, and a tombstoned validator can never rejoin; unbonding 21 days. Slashing applies to bonded stake, so delegators share the loss. (OGO_VALIDATOR_AND_STAKING_SPEC.md, Sec. 5)

## 3. What software runs the blockchain nodes?

**DECIDED** as the locked prototype and Milestone C matrix. The node binary is `ogod`, built from `cmd/ogod/main.go`. (OGO_NETWORK_ARCHITECTURE.md, Sec. 4)

| Component | Locked version | Note |
|-----------|---------------|------|
| Go toolchain | 1.25.7 | Go 1.23 was the blueprint proposal; Milestone B actually used 1.25.7 |
| Cosmos SDK | v0.53.6 required, v0.53.4 effective | Effective version comes through Gaia's replace directive, preserved deliberately |
| CometBFT | v0.38.23 | Includes the median-time fix GHSA-c32p-wcqj-j677, fixed in v0.38.21. CometBFT v1.x is retracted |
| CosmWasm wasmd | v0.60.9 | wasmvm/v2 v2.3.5 |
| IBC | ibc-go/v10 v10.7.0 | Live IBC relaying is not wired yet; escrow accounting is ledger-level only |
| Cosmovisor | v1.7.3 | Separate binary, not a Go module dependency; upgrade mechanism, not yet executed |
| Dependency reference | Gaia v28.3.1 | The whole matrix is one tested unit; do not upgrade individual dependencies independently |

Standard modules: x/auth, x/bank, x/staking, x/slashing, x/distribution, x/gov, x/upgrade, x/epochs, x/evidence, ibc-go, wasmd. (OGO_NETWORK_ARCHITECTURE.md, Sec. 5.3)

Custom modules, only three: x/monetary (the only module that can mint), x/treasury, x/rewards. Explicitly not created: x/governance and x/fees; no x/bridge module. (OGO_NETWORK_ARCHITECTURE.md, Sec. 5.3)

Remote signing: cosmos/kms v0.1.0. "Do not build new validator operations on TMKMS." TMKMS is deprecated and will be removed from the Cosmos Bug Bounty Program at the start of Q1 2027. (OGO_NETWORK_ARCHITECTURE.md, Sec. 8; OGO_VALIDATOR_AND_STAKING_SPEC.md, Sec. 7)

What is still to be built in Milestone C: the standard keepers are currently pinned and compiled only, not wired (x/bank, x/auth, x/staking, x/slashing, x/distribution, x/gov, x/upgrade). Real x/epochs wiring, signature verification, gas metering, and the wasm VM are not implemented yet. (ogo-network-prototype-build-report.md, Sec. 6; OGO_NETWORK_PROTOTYPE_BLOCKERS.md, B1, B2, B6, B7, B9)

Evaluation tracks only, not adopted: SDK 0.55 / CometBFT 0.40, and CosmWasm 3.0 / wasmd 0.61. Do not adopt SDK v0.53.8 or Gaia v29 for the first localnet. (OGO_NETWORK_ARCHITECTURE.md, Sec. 5.1; OGO_NETWORK_PROTOTYPE_BLOCKERS.md)

## 4. What hardware and server specifications are required?

**DECIDED:** the validator spec defines three infrastructure tiers per validator per month, in USD, infrastructure only, labelled illustrative planning assumptions. Labour is separate. (OGO_VALIDATOR_AND_STAKING_SPEC.md, Sec. 8)

| Tier | Cost per validator per month | CPU | RAM | Disk | Network | Use |
|------|------|-----|-----|------|---------|-----|
| Lean | About $75 | 2+ cores | 8 GB | 200 GB SSD | 50 Mbps | Minimum for a testnet or small validator. Not recommended for mainnet |
| Standard | About $200 | 4+ cores | 32 GB | 500 GB NVMe | 100 Mbps | Baseline used for the minimum fee need calculation |
| Professional | About $500 | 8+ cores | 64 GB | 1 TB NVMe | 1 Gbps | Plus sentry nodes and remote signing. Baseline for a validator expected to be reliable under load |

Notes and cross-checks:

* The Standard tier matches the current Cosmos Hub sizing guidance for a pruned validator (4 or more cores, 32 GB RAM, 500 GB NVMe, local NVMe or equivalent low-latency SSD) and the P5 baseline (a new node catches up to tip within 2 hours on 4+ CPU cores, 32 GB RAM, 500 GB NVMe). (docs.cosmos.network system requirements, checked 2026-10-15; OGO_NETWORK_ACCEPTANCE_CRITERIA.md, Sec. 3.1, P5)
* The Lean tier sits below the Cosmos Hub floor of 32 GB RAM. The validator spec itself says Lean is not recommended for mainnet. Treat Lean as testnet-only.
* No GPU is needed. Block proposal and signing (ed25519) is pure CPU work.
* Disk grows every day. 500 GB is a starting point with pruning; an archive node (pruning disabled) needs 1 TB or more and grows without bound.
* Sentry nodes: the Cosmos Hub guidance allows 16 GB RAM plus swap on a sentry. Mainnet requires at least one sentry node between the validator and the public network, so the validator's network location is not publicly exposed. (OGO_VALIDATOR_AND_STAKING_SPEC.md, Sec. 3.2)
* Labour at the Professional tier is $1,500 to $4,000 per validator per month for monitoring, upgrades, incident response, and key management. A 50-validator Professional set costs $25,000 per month in infrastructure plus $75,000 to $200,000 per month in labour. (OGO_VALIDATOR_AND_STAKING_SPEC.md, Sec. 8)

Monthly infrastructure totals from the spec (validators only, before sentries and labour):

| Validators | Lean | Standard | Professional |
|------------|------|----------|--------------|
| 4 | $300 | $800 | $2,000 |
| 7 | $525 | $1,400 | $3,500 |
| 50 | $3,750 | $10,000 | $25,000 |
| 100 | $7,500 | $20,000 | $50,000 |

What the local prototype actually ran on: a single local machine (this VM), all validators as processes on localhost with separate home directories, keys, and ports. No CPU model, RAM total, disk size, or host specification is recorded in the prototype or PoC reports, and no cloud server was used. The PoC states explicitly that a sentry topology cannot be demonstrated on a single-machine localnet. (ogo-network-poc-4validator-report.md, Sec. 7 and Sec. 11)

## 5. Can we initially run a development network and testnet using cloud infrastructure?

**RECOMMENDED:** yes for a private cloud development network; **OPEN** for any public testnet, which needs Benross's separate explicit approval and is not authorized by this document or by any current directive.

Three different networks must not be confused:

| Network | What it is | Status |
|---------|-----------|--------|
| Local devnet (localnet) | Validators as processes on one local machine, TEST-ONLY tokens, no public access | Exists today. The prototype and PoC run this way. Cost is zero (see Section 11) |
| Private cloud devnet | The same chain on separate cloud servers, still private, still TEST-ONLY tokens, access limited to the team | Technically feasible. Not provisioned. Needs a spending decision first, because even cheap servers cost money |
| Public testnet | Open to independent validators not operated by OGO Systems, with a faucet, documented genesis, and a testnet runbook | Not authorized. Roadmap Phase 4. Requires Benross's separate explicit approval |

What Phase 4 requires before a public testnet exists: independent validators not operated by OGO Systems; a faucet; documented genesis; a testnet runbook; validators joining by bonding TEST-ONLY tokens with no real value; governance exercised on testnet; an incident handled from the runbook; the halt and catch-up rule exercised on testnet; the eight CosmWasm adoption criteria tested in staging; documentation tested by an operator who did not write it; and independent state validation (G1) exercised, with a third party obtaining the source, building the binary, running the tests, and verifying the supply rules independently, with a written report. (OGO_NETWORK_IMPLEMENTATION_ROADMAP.md, Sec. 5)

A private cloud devnet is the natural next step after the localnet, because it is the only way to prove two things the localnet cannot prove: true validator independence (separate hosts, separate operators, separate monitoring) and sentry topology. The Milestone C plan records this as an open question: approve multi-host staging infrastructure when Milestone C reaches O1, or record O1 independence as single-host only with the limitation stated. Spending or purchasing infrastructure is not authorized (approval blocker A6). (OGO_NETWORK_PROTOTYPE_BLOCKERS.md, A6; ogo-network-milestone-c-plan-v0.1.md, Sec. 6)

## 6. What infrastructure provider do you recommend?

**RECOMMENDED.** Three separate answers, because development, testnet, and mainnet have different constraints.

### 6.1 Development: zero cost, local machine (RECOMMENDED, and Benross's constraint)

The development network runs locally at zero cost. No purchased hardware, no rented servers, no paid hosting, no paid RPC services, no subscriptions. See Section 11 for the full zero-cost setup. This is the default and the only option in use today.

### 6.2 Paid hosting for a future private testnet or mainnet: Hetzner (RECOMMENDED), DigitalOcean runner-up

Recommendation: **Hetzner Cloud (dedicated-vCPU CCX line) or Hetzner dedicated root servers (AX line)** for validator nodes, when paid hosting is approved.

Reasoning:

* Dedicated vCPUs, so no noisy-neighbor effect on block signing latency. Shared-vCPU plans are cheaper but unsuitable for a validator that must sign every block on time.
* Local NVMe storage on every plan, 20 to 60 TB of included traffic, and regions in Germany, Finland, the United States, and Singapore, which supports the geographic distribution target (no country or provider over 33 percent).
* Price. Post-June 2026 list prices, net, excluding VAT and public IPv4, verified from Hetzner's own price adjustment documentation (June 15, 2026) via third-party price trackers checked October 2026: CCX23, 4 dedicated vCPU, 16 GB RAM, 160 GB NVMe, about 86 euros per month; CCX33, 8 dedicated vCPU, 32 GB RAM, 240 GB NVMe, about 138 euros per month; AX42 dedicated server, 8-core Ryzen, 64 GB RAM, 2x 1.92 TB NVMe, about 97 euros per month plus a 49 euro setup fee. Prices rose sharply in June 2026 (CCX23 rose about 173 percent), so recheck at hetzner.com before any purchase. Older articles quoting 15 to 30 euros for these plans are out of date.
* Limitation: Hetzner is a single provider. The decentralization target caps any one provider at 33 percent, so a production network must spread validators across at least three providers or add bare-metal operators. Hetzner alone can never host the whole mainnet validator set.

Runner-up: **DigitalOcean**. General Purpose droplets, 4 dedicated vCPU, 16 GB RAM, 50 GB SSD, $126 per month; CPU-Optimized, 4 vCPU, 8 GB RAM, 50 GB SSD, $84 per month; Basic shared, 4 vCPU, 8 GB RAM, 160 GB SSD, $48 per month (shared CPU, not recommended for validators). More regions than Hetzner, including the United States, Canada, Europe, India, Singapore, and Australia, and simpler operations, but roughly 45 percent more expensive than Hetzner for the same dedicated shape, and the 50 GB disk on the General Purpose line is too small for a validator without added block storage. (digitalocean.com droplet pricing, accessed 2026-11-06)

Not recommended for validators: AWS and Google Cloud for raw validator compute (two to three times the cost of Hetzner for the same shape, though their managed services are useful elsewhere); any shared-CPU plan for a signing validator; and any provider without local NVMe or equivalent low-latency SSD.

Cloudflare does not run validator nodes. It serves the OGO application (Pages frontend and Workers API) and cannot host an `ogod` process. Supabase does not run validator nodes either. Both are covered in Section 8 and in the diagram.

### 6.3 Free-tier remote hosting, only if remote access is necessary (RECOMMENDED as an option, nothing provisioned)

If the team needs a remote node without spending, the only free tier that can run a usable always-on server is **Oracle Cloud Always Free**:

* Current allowance, after Oracle halved it on June 15, 2026: 1,500 OCPU-hours and 9,000 GB-hours per month on Ampere A1 ARM, which Oracle states equals 2 OCPUs and 12 GB of RAM running always-on, plus 200 GB of block storage and 10 TB of outbound traffic per month, with no expiry. Older articles still quote the pre-cut 4 OCPU and 24 GB figure; that figure is out of date.
* Limits: ARM only (the `ogod` binary builds for linux/arm64, so this is workable, but it must be tested, not assumed); Ampere capacity is scarce in some regions and instance creation can fail with "out of host capacity"; signup requires a card, with a $1 authorization; staying inside the free limits is the operator's responsibility, and exceeding them bills at standard rates.
* The other free tiers cannot host a validator: Google Cloud gives one e2-micro (2 shared vCPU burst, 1 GB RAM, 30 GB disk, 1 GB egress per month, three US regions only), which is too small for a validator and usable at most for a tiny RPC or indexer probe. AWS new accounts get credits only (up to $200 for up to 6 months), with no permanent free VM. Azure gives 12 months of small burstable VMs plus a $200 credit, then paid.

No cloud resource is provisioned by this document. Free-tier hosting is research only, and any use of it is a private devnet, not a public testnet.

## 7. How many validator nodes are required for the initial network?

**DECIDED** distinctions, **RECOMMENDED** proposal for the initial network, **OPEN** launch minimum.

The decided distinctions (OGO_NETWORK_DECISION_REGISTER.md; OGO_VALIDATOR_AND_STAKING_SPEC.md, Sec. 2 and Sec. 6.3):

| Count | Purpose | Status |
|-------|---------|--------|
| 3 | Approved Phase 2 local prototype | Decided by Benross, 2026-11-06. Preserved unchanged as the stable reference |
| 4 | Minimum operational set | Smallest set that tolerates one faulty validator (3 of 4 is more than two thirds). Proven by the PoC |
| 7 | Private testnet target | Plan SHOULD target, with stake deliberately uneven (largest validator under one third of voting power) |
| 50+ | Decentralization target | At least 50 validators, Nakamoto coefficient at least 10, top 10 under 33 percent, no country or provider over 33 percent. Research targets, not launch requirements, and must not be presented as a launch promise |

Proposal for the initial network (RECOMMENDED): **7 validators for the initial private testnet network**, each on a separate host, with at least one validator behind at least 2 sentry nodes to prove the sentry pattern (O1), and with stake deliberately uneven so the largest validator holds under one third of voting power.

Reasoning: 4 is the minimum that survives one fault, but it leaves no headroom: a second fault halts the chain. 7 tolerates two faulty validators (5 of 7 is more than two thirds) and is the plan's SHOULD target. 7 also matches the spec's private testnet phase (4 minimum, 7 as the target, each operated as if independent). Cost at the Standard tier is about $1,400 per month for 7 validators before sentries and labour (see Section 4).

The launch minimum for mainnet is **OPEN**: Definition A (4 validators) versus Definition B (at least 50 validators). That choice is recorded as decision blocker D5 and is Benross's call. It must not be treated as decided by this document. (OGO_NETWORK_PROTOTYPE_BLOCKERS.md, D5)

Also open: the active set size N is a genesis parameter, TBD; there is no application, approval, or allowlist in the protocol for mainnet, and a permissioned set is a testnet configuration only and must be removed before mainnet. (OGO_VALIDATOR_AND_STAKING_SPEC.md, Sec. 3.1 and Sec. 4)

## 8. What is the role of Supabase versus the blockchain?

**DECIDED.** The chain and Supabase are two separate systems with a hard boundary between them. (OGO_NETWORK_API_SPEC.md, Sec. 2, Principle 2; OGO_NETWORK_ARCHITECTURE.md, Sec. 2)

| System | Source of truth for | Must never |
|--------|-------------------|------------|
| OGO Network blockchain | Native OGO Coin balances, transfers, staking, governance, issuance, supply | Trust a client claim about a balance |
| Supabase | Application authentication, profiles, application data, OGO Coins ledger, OGO Credits ledger | Own or control a blockchain balance; be the source of truth for native balances |
| Cloudflare Worker | Application server: chain reads, nonce issue and signature verification, unsigned transaction preparation, on-chain verification before fulfillment | Hold keys, sign, write chain state, broadcast, trust client claims, write to the platform ledger as a side effect of a chain operation, act as a validator, move treasury funds |

Rules that follow:

* Native balances are read from chain state only, through a public RPC or LCD endpoint or through the indexer, as the x/bank balance for denom `uogo`. No other source is authoritative. The wallet queries at least two independent RPC endpoints, or one RPC endpoint plus the indexer, and flags a discrepancy. If the indexer and chain disagree, the chain wins and the discrepancy is logged and surfaced. (OGO_NATIVE_WALLET_SPEC.md, Sec. 3; OGO_NETWORK_API_SPEC.md, Sec. 3.3)
* A Supabase account does not own a blockchain balance. If the Supabase account is deleted, the on-chain balance is unaffected. If the OGO website is unavailable, the user can still access coins through any compatible wallet. (OGO_NATIVE_WALLET_SPEC.md, Sec. 2, Principle 3)
* Wallet links are recorded in Supabase (user id, wallet address, chain id, linked timestamp, nonce id). Unlinking removes the Supabase link and does not affect the on-chain balance. (OGO_NETWORK_API_SPEC.md, Sec. 4)
* Project registration is an application-level record in Supabase, not a chain transaction. There is no on-chain project registry in the first monetary prototype; a project's on-chain identity is its wallet address. (OGO_NETWORK_API_SPEC.md, Sec. 5)
* Payment requests are stored in Supabase with a unique request id, recipient address, amount in uogo, description, and expiry. Settlement is recorded in Supabase with an idempotency key and the transaction hash, and only after on-chain payment is verified as committed. An order is not marked as paid until the payment is verified as committed on-chain against a public RPC endpoint, never on a client claim. (OGO_NETWORK_API_SPEC.md, Sec. 8; OGO_NETWORK_ACCEPTANCE_CRITERIA.md, Sec. 7, M5-7)

The existing economy wallet is a separate thing from the blockchain wallet. It is an account-based database ledger, not a self-custody wallet: migration `20261105000000_ogo_economy_wallet.sql` creates `ogo_wallets` (one row per user, `coins` and `credits` as bigint, both non-negative) and `ogo_wallet_transactions` (unique idempotency key). All writes go through the `ogo_wallet_apply` SECURITY DEFINER function, executable only by the service role. The Cloudflare Worker exposes `/api/economy/wallet` for reads and a Stripe purchase flow. There are no private keys, no seed phrases, no on-chain balances, and no user-signed transactions anywhere in that implementation. (ogo-public-cryptocurrency-feasibility-v0.1.md, Sec. 1A)

## 9. What RPC endpoints will OGO Wallet use?

**DECIDED** as a design (Network API spec and wallet contract); **OPEN** as live endpoints, which do not exist yet.

| Endpoint | Used for | Used by | Status today |
|----------|----------|---------|--------------|
| CometBFT RPC, public and CORS-enabled | Block height and block by height, transaction by hash, broadcast transaction, validators and voting power, consensus state, network info. Broadcast returns the hash and the caller tracks commitment | Node operators, indexers, advanced integrators, wallet broadcast | Exists locally in the prototype and PoC (local ports only, e.g. RPC 26657, 26667, 26677 in the prototype; 26757, 26767, 26777, 26787 in the PoC). No public endpoint yet |
| LCD, Cosmos SDK REST, public and CORS-enabled | `/cosmos/bank/v1beta1/balances/<address>`, `/cosmos/bank/v1beta1/supply`, `/cosmos/staking/v1beta1/validators`, `/cosmos/staking/v1beta1/delegations/<address>`, `/cosmos/gov/v1/proposals`, `/cosmos/tx/v1beta1/txs/<hash>`, plus x/monetary, x/treasury, and x/rewards queries | Wallets, the OGO application, developer tools | Does not exist yet. Milestone C delivers it |
| Indexer API | Paginated history the chain itself does not provide: transaction history by address with direction, counterparty, amount, fee, block height, timestamp, and status; history by block; validator history; supply history | Wallet history, explorer | Does not exist yet. Milestone C and M4/M6 deliver it |
| gRPC | Not specified in the specs | Not specified | Not specified. Standard Cosmos defaults (port 9090) are conventions, not OGO decisions |
| WebSocket subscriptions | New blocks, new transactions for an address, governance events | Optional later phase only | Not required for the first monetary prototype |

The wallet contract mapping (ogo-network-wallet-api-contract-v0.1.md, Sec. 1, 4.1, 4.2, 4.6, 4.8, 6): the live wallet client uses LCD for balance and supply, public RPC broadcast for signed transactions, and the indexer for history. The v0.1 contract is mock-only, with the live adapter arriving with Milestone C. The frontend calls this contract only and never calls chain RPC, LCD, or indexer endpoints directly, and never calls platform economy endpoints for native coin data.

Other endpoint facts:

* The specs state no port numbers for RPC, LCD, or indexer endpoints. Standard Cosmos defaults are RPC 26657, P2P 26656, LCD 1317, gRPC 9090. Those are conventions, not OGO decisions, and the local prototype deliberately used non-default ports.
* External wallets (Keplr, Leap, and Cosmostation, in that priority order) connect through the public RPC and LCD endpoints and chain registry metadata, with no custom chain-side integration code beyond the registry entry and CORS-enabled endpoints. Keplr and live IBC were removed from the framework-selection criteria by Benross's decision and are evaluated separately if a specific technical dependency demonstrates necessity. (OGO_NETWORK_API_SPEC.md, Sec. 9)
* The wallet verifies chain id, sequence, and account number against the chain before signing, uses Cosmos DIRECT sign mode by default (Amino JSON as legacy fallback for hardware wallets), and never signs a transaction the user did not see. Signing happens only in the user's wallet; the OGO application, Worker, and any OGO server never see a private key. (OGO_NATIVE_WALLET_SPEC.md, Sec. 2, Principle 4, and Sec. 9)
* Amounts are exact integer `uogo` values carried as strings, never floating point. Display conversion is 1 OGO = 1,000,000 uogo, with 6 decimals proposed, not decided. (ogo-network-wallet-api-contract-v0.1.md, Sec. 2 and 3)

## 10. What is required to launch an independent production mainnet?

**OPEN.** Most of it is not started. This section is a checklist drawn from the roadmap Phase 5, the security model, and the acceptance criteria. Nothing here is a launch approval.

| Requirement | Source | Status |
|-------------|--------|--------|
| All five audits complete, with written reports and a findings register; no unresolved critical or high findings | Roadmap Sec. 6; Security Model Sec. 7 | Not started. Audits in order: (1) monetary core, (2) consensus, staking, and governance, (3) contracts and IBC, (4) wallet, keys, and RPC, (5) build and supply chain. Each auditor must not have written or configured the code or system being audited. Audit 3 is required before any contract is deployed to a public network |
| Written legal analysis commissioned and received | Roadmap Sec. 6 and Sec. 7 | Not started. Not yet commissioned. The roadmap stops if it is not commissioned by Phase 4. This is the longest-lead item |
| Genesis parameters approved by Benross in writing | Roadmap Sec. 6 (genesis review gate) | Not started. Parameters: max supply (derived as M = G + S_actual), genesis supply, issuance model and parameters, period length, split, fee policy, treasury spend level and rule, emergency reserve floor, validator minimum, governance thresholds, network name and chain ID, denomination and decimals, bech32 prefix and derivation path. Any parameter without an approval is a blocker |
| Operations readiness: runbooks, monitoring, alerting, snapshots, rollback procedure documented and tested | Roadmap Sec. 6 | Not started. Monitoring must cover block height, block time, missed blocks per validator, peer count, mempool size, and total supply, with an alert fired and recorded for at least one missed-block event (O2) |
| OGO-operated validators, sentry nodes, remote signing, and monitoring in place | Roadmap Sec. 6 | Not started. Consensus keys held in a remote signer (cosmos/kms) or HSM, never as a plaintext file on the validator host |
| Incident response exercise completed | Roadmap Sec. 6 | Not started. An incident is handled from the runbook and the handling is recorded; the halt and catch-up rule is exercised on the testnet, not just on the localnet |
| Independent state validation (G1) exercised | Roadmap Sec. 5 and Sec. 7 | Not started. A third party obtains the source, builds the binary, runs the tests, and verifies the supply rules independently, with a written report. G1 blocks Phase 5 |
| Framework selection closed (21 MUST criteria, M2 deciding) | Acceptance Criteria Sec. 3.6 | Open. S5 liveness passed and established; P1 partially executed; M2 not executed and framework-selection blocking; O1 partially executed |
| Monetary model and parameters selected | Decision register, open decisions | Not started. Model family, parameters, split, then max supply as G + S_actual, in that order |
| Launch approval, separate from readiness | Roadmap Sec. 6 | Not started. M8 is readiness, not a launch. Benross approves the launch in writing, separately from approving readiness |

Additional mainnet facts:

* The five-layer governance security boundary (parameter separation, upgrade-handler denylist gate, amount-less MintScheduled entrypoint, every-block invariant checking, public schedule queryability) must be implemented and tested on the wired chain. A hard fork that replaces the code can remove all five layers; protection is transparent rules, independent operators, and a visible upgrade process, not a claim that a hard fork is impossible. (OGO_NETWORK_SECURITY_MODEL.md, Sec. 5)
* Residual risks, in severity order: stake concentration and validator cartel (high at launch), post-issuance fee funding, software supply-chain compromise (high), upgrade removing the checks, governance capture, wallet key theft through phishing or malware, IBC and bridge accounting errors, epoch and timestamp edge cases, fee market mispricing for agents, and legal and regulatory change. No threat is claimed as eliminated. (OGO_NETWORK_SECURITY_MODEL.md, Sec. 3 and Sec. 8)
* Gates that must not be skipped: architecture approval blocks Phase 2; monetary bypass re-test (S2, S3, S4) blocks Phase 3; wallet and indexer gates (G7, G9) block Phase 4; independent state validation (G1) blocks Phase 5; the audit gate (G11) blocks launch; the genesis review gate blocks launch; launch approval is separate and after Phase 5. (OGO_NETWORK_IMPLEMENTATION_ROADMAP.md, Sec. 7)

## 11. Zero-cost development deployment

**DECIDED constraint (Benross, 2026-11-06).** The development setup is completely free: no purchased hardware, no rented servers, no paid hosting, no paid RPC services, no subscriptions. The development network and testnet in this document are LOCAL (the existing localnet approach). A public testnet is not authorized and needs Benross's separate explicit approval. The private local devnet and the public testnet are kept clearly distinct throughout this document.

### 11.1 What runs, and what it costs

| Component | How it runs in development | Cost |
|-----------|---------------------------|------|
| OGO Network localnet, 3 validators (prototype) or 4 validators (PoC) | `ogod` processes on one local machine, from the existing scaffold, built with the already installed Go 1.25.7 toolchain | $0 |
| OGO Wallet frontend | `npm run dev` / `npm run build` in `~/workspace/ogo`, against the mock adapter now and the real adapter with Milestone C | $0 |
| OGO application (existing) | Already served by Cloudflare Pages and Workers on existing accounts | $0 additional |
| Supabase | Existing project `ogo`, Free plan, us-east-1, for auth, profiles, app data, and the Coins/Credits ledger | $0 additional |
| GitHub | Existing repositories and local git only, no paid plan needed | $0 |
| RPC and indexer for development | Local CometBFT RPC on localhost ports; LCD and indexer arrive with Milestone C, run locally | $0 |

Total development cost: $0 per month.

### 11.2 Start, stop, and reproduction (summary)

The implementation tracks own the detailed runbooks. The approach is:

1. Build the chain binary from the existing scaffold with Go 1.25.7 (`~/workspace/tools/go-1.25.7/bin/go`). No download beyond the already vendored module cache is needed for a rebuild.
2. Start the localnet with the demo script: `./demo.sh` for the three-validator prototype, `./run-poc.sh` for the four-validator PoC. Each script starts the validators from one shared genesis, runs the demonstration steps (blocks, transfer, epoch issuance, verification, halt and resume or fault tolerance), and prints measured results.
3. Run the wallet app in `~/workspace/ogo` and exercise it against the mock adapter (now) or the real local adapter (with Milestone C).
4. Stop: the scripts stop the nodes when they finish, or on failure, so they are safe to re-run. A run can also be stopped with Ctrl-C. Nothing is left running, no ports stay open, and no state leaves the machine.
5. Reproduce: a second run of the same script produces the same steps and the same verification checks. The prototype demo was run twice with identical results; the PoC coordinator re-ran `run-poc.sh` end to end with matching S5 numbers.

Test accounts and test OGO Coins only. Localnet keys are plaintext and localnet only, and must never be used for real value.

### 11.3 What cannot run within the free environment

Everything needed for development can run free locally. The following cannot, and each names what it would require:

* **True validator independence and sentry topology (O1).** A single local machine cannot demonstrate validators on separate hosts behind sentries. This needs separate hosts. The only zero-cost remote option is the Oracle Always Free tier (Section 6.3), which is ARM-only, capacity constrained in some regions, requires a card on file, and would still be a private devnet, not a public testnet. Nothing is provisioned.
* **24/7 uptime and public RPC availability.** A local machine restarts and sleeps. Milestone B recorded a VM service restart killing a node run at height 33. A public RPC endpoint that stays up needs hosted infrastructure with monitoring and alerting.
* **Production-grade monitoring, alerting, and incident response.** Log inspection is free; a monitored, alerted, on-call operation is not.
* **The five security audits.** Independent auditors, paid, with written reports. Not started.
* **The written legal analysis.** Commissioned legal work, paid. Not yet commissioned, and the longest-lead item.
* **Hardware security modules or cloud KMS for validator keys.** Remote signing through cosmos/kms needs a KMS or HSM; a local plaintext key file is a development-only shortcut and is forbidden for production.
* **A public explorer and indexer at scale.** A local indexer is free; a public, always-on explorer with history for all users needs hosted infrastructure.
* **Mainnet itself.** Paid validator infrastructure, audits, legal analysis, and operations. Mainnet is not authorized and not costed as free anywhere in this document.

## 12. Open decisions this document does not settle

| Decision | Status | Owner |
|----------|--------|-------|
| Framework selection (Cosmos SDK plus CometBFT as the final framework) | Open, M2 deciding | Benross, after the proof of concept |
| Monetary model family (geometric vs hybrid) | Open | Benross, after model selection |
| All monetary parameters (max supply, genesis supply, period length, split, fee policy, treasury spend, reserve floor) | Open | Benross, in the decided order: model, parameters, split, max supply as G + S_actual |
| Launch validator minimum (Definition A, 4 validators, vs Definition B, 50+ validators) | Open, blocker D5 | Benross |
| Infrastructure provider for paid hosting | Recommended (Hetzner), not decided | Benross |
| Multi-host staging infrastructure for O1 independence | Open (approve, or record single-host limitation) | Benross |
| Public testnet | Not authorized, needs separate explicit approval | Benross |
| Mainnet launch | Not authorized, needs separate written approval after readiness | Benross |
| Written legal analysis | Not yet commissioned | Benross |
| Cloudflare vs other hosting for the application | Existing stack (Cloudflare Pages and Workers) is already in use; no change proposed | None needed |

## Sources

* OGO_NETWORK_ARCHITECTURE.md, OGO_CONSENSUS_AND_EPOCH_SPEC.md, OGO_VALIDATOR_AND_STAKING_SPEC.md, OGO_NETWORK_API_SPEC.md, OGO_NATIVE_WALLET_SPEC.md, OGO_FEE_MARKET_SPEC.md, OGO_NETWORK_SECURITY_MODEL.md, OGO_NETWORK_IMPLEMENTATION_ROADMAP.md, OGO_NETWORK_ACCEPTANCE_CRITERIA.md, OGO_NETWORK_DECISION_REGISTER.md, OGO_NETWORK_PROTOTYPE_BLOCKERS.md, OGO_NETWORK_RECONCILIATION_REPORT.md (all in `ogo-network-architecture/`)
* ogo-network-native-coin-architecture-implementation-plan-v1.1.md (plan v1.0 untouched)
* ogo-network-wallet-api-contract-v0.1.md
* ogo-network-milestone-c-plan-v0.1.md
* ogo-network-prototype/ogo-network-prototype-build-report.md
* ogo-network-poc-4validator/ogo-network-poc-4validator-report.md
* ogo-public-cryptocurrency-feasibility-v0.1.md
* Web research, 2026-11-06: Cosmos Hub system requirements (docs.cosmos.network, checked 2026-10-15); Hetzner Cloud pricing after the June 15, 2026 price adjustment (hetzner.com via price trackers checked October 2026); DigitalOcean droplet pricing (digitalocean.com, accessed 2026-11-06); Oracle Cloud Always Free limits after the June 15, 2026 allowance cut, Google Cloud free tier, AWS free plan, Azure free account (provider documentation via comparisons checked October and November 2026). Prices are snapshots, not quotes.
