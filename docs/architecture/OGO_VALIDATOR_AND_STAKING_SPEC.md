# OGO Validator and Staking Specification

Status: Technical specification. Planning only. No validators operated, no staking live, no testnet launched, no spending authorized, no production change authorized.
Date: 2026-11-06
Companion document: `ogo-network-native-coin-architecture-implementation-plan-v1.0.md` (sections 3.8, 3.9, 3.10, 5)
Architecture direction: CONFIRMED AND LOCKED, Option B. OGO Network is an independent Layer 1 blockchain with native OGO Coin.
Decentralization targets in this document are RESEARCH TARGETS, not approved launch requirements.

## 1. Purpose

This document specifies the validator strategy in stages, from local development to public mainnet, and the staking, delegation, slashing, and unbonding rules that govern it. It states the decentralization targets as research targets, models infrastructure costs for 4, 25, 50, and 100 validators under lean, standard, and professional assumptions, and identifies what OGO would need to fund validator participation without assuming an OGO/USD market price.

## 2. Staged validator strategy

### Phase 1: Local development validators (M0 to M1)

- 3 independent local validator nodes, each on a separate home directory and a separate key, plus genesis accounts. First proof: test OGO Coin transferred between wallets, with the same balances verified from all three nodes. Validator count decided by Benross, 2026-11-06 (Phase 2 Execution Directive): THREE for the authorized local prototype. Correction note, 2026-11-06: the reconciliation pass had resolved this to 4 with BFT reasoning (3f+1; a 3-validator network tolerates zero faulty validators). Benross's decision supersedes that resolution for the local prototype. The distinction to keep: 3 is the decided local-prototype execution requirement; 4 is the minimum operational set (Definition A), the smallest set tolerating one faulty validator; 50 or more is the decentralization target (Definition B) for a public network. 3 is not a decentralization claim.
- Keys are localnet-only (plaintext keys under the home directory are acceptable in this phase and nowhere else).
- No sentry nodes, no remote signing, no monitoring beyond log inspection. The purpose is to prove the chain produces blocks and transfers work, not to prove operations.
- Acceptance: blocks produced, transfer commits, balances identical across all three nodes, supply query matches the expected value exactly. All numbers TEST-ONLY.

### Phase 2: Private multi-validator testnet (M2 to M3)

- 4 validators minimum, 7 as the target, each operated as if independent (separate homes, keys, and monitoring), with at least 2 sentry nodes for one validator to prove the sentry pattern.
- Staking, delegation, unbonding, slashing, and remote signing are wired and tested. At least one validator runs with a remote signer (cosmos/kms v0.1.0), with the consensus key never present as a plaintext file on the validator host.
- A Cosmovisor upgrade is executed on all validators, with a store migration, and balances and supply are identical before and after. A failed upgrade on one validator is rolled back and the validator rejoins.
- Duration: at least one continuous 24-hour run for stability. The proof-of-concept criteria P1, S5, S6, S8, M2, M3, M5, O1, O2, and O5 are the acceptance tests for this phase.
- Validators in this phase are operated by OGO Systems and by invited operators under a testnet agreement. The set is permissioned by configuration in this phase only. A permissioned set is a testnet configuration, not a protocol rule, and it must be removed before mainnet.

### Phase 3: Public incentivized testnet (M7)

- Public testnet with independent validators (not operated by OGO Systems), a faucet, documented genesis, and a testnet runbook.
- Validators join by bonding testnet tokens (TEST-ONLY, no real value) and by meeting the onboarding requirements (section 3). The set is open: any operator meeting the requirements can join.
- A governance upgrade is executed on the testnet. An incident (a validator outage, or a halt) is handled using the runbook, and the handling is recorded.
- Incentives in this phase are testnet-only (recognition, testnet tokens, documentation credit). No real OGO Coin is distributed on a testnet, and no incentive is promised that would require a monetary decision. Any mainnet incentive design is a separate decision.
- Acceptance: independent validators join and produce blocks; an upgrade executes; an incident is handled from the runbook; the testnet uses TEST-ONLY parameters, labelled so, with no real value.

### Phase 4: Public mainnet (M8, readiness only)

- Mainnet readiness, not a launch. Launch itself needs a separate approval from Benross.
- The validator set is open and ranked by bonded stake. There is no allowlist. The minimum operational set is 4 validators. The sufficiently-decentralized target is Definition B (section 6).
- All testing gates G1 to G11 pass. All five audits are complete with no unresolved critical or high findings. The legal analysis is commissioned in writing and received. Genesis parameters are approved by Benross in writing.
- Mainnet validators operate under the requirements in section 3, the monitoring in section 5, and the key security in section 7.

## 3. Validator onboarding and requirements

### 3.1 Onboarding

1. The operator reads the validator guide (in OGO-Network-Docs) and the operations runbook.
2. The operator provisions infrastructure to at least the standard tier (section 8), configures sentry nodes, and configures remote signing.
3. The operator generates a consensus key through the remote signer (never as a plaintext file on the validator host), an operator key, and a withdrawal key, held in separate keystores.
4. The operator creates the validator (MsgCreateValidator) with a commission rate (subject to the floor), a description, and a self-delegation. The self-delegation minimum is TBD.
5. The operator registers monitoring and alerting, and confirms the validator is visible in the explorer with the correct voting power.
6. The operator joins the validator set when its bonded stake ranks in the top N.

There is no application, no approval, and no allowlist in the protocol. If a testnet phase uses a permissioned set, that is a testnet configuration, and it is documented as such, with the date it will be removed.

### 3.2 Requirements

- Infrastructure: at least the standard tier (4 plus CPU cores, 32 GB RAM, 500 GB NVMe, 100 Mbps). Lean tier is acceptable for a testnet, not for mainnet.
- Sentry nodes: at least one sentry node between the validator and the public network. The validator's network location is not publicly exposed.
- Remote signing: the consensus key is held in a remote signer (cosmos/kms v0.1.0) or a hardware security module. It is never a plaintext file on the validator host.
- Monitoring: block height, block time, missed blocks, peer count, mempool size, and voting power are monitored, with alerts for missed blocks and halted chain.
- Upgrades: the operator can execute a Cosmovisor upgrade within the upgrade window, and can roll back a single validator after a failed upgrade.
- Key management: operator, withdrawal, and consensus keys are separate. High-value operations (unbonding, commission changes above a threshold) use multisig where the operator supports it.
- Geographic and provider distribution: the operator discloses its country and infrastructure provider (section 6). The disclosure is public, so the network's distribution can be measured.

## 4. Staking and delegation

- Native staking through x/staking. Delegators bond `uogo` to a validator, earn a share of that validator's rewards per the approved split, and pay the validator's commission.
- The active validator set is the top N validators by bonded stake, where N is a genesis parameter (TBD).
- Commission: each validator sets its own commission rate, subject to a floor. Proposed floor: 5 percent (Cosmos Hub precedent, proposal, not a decision). A validator can change its commission only with a notice period and a maximum change per period, to prevent raising commission to 100 percent after attracting delegations. The notice period and maximum change are TBD.
- Rewards: per-epoch issuance is split by x/rewards between validators, delegators, and the community pool or treasury, per the approved split. The split conserves exactly. Validator rewards are distributed through x/distribution.
- Redelegation: allowed once per unbonding period per delegation, following the standard SDK rule.
- Liquid staking is not part of the first monetary prototype. It is an open question for a later phase, and it must be evaluated against the same supply and slashing rules before adoption. No liquid staking token is created by these specifications.

## 5. Slashing and unbonding

Slashing parameters are Hub precedents, proposed, not decided. Each needs approval.

- Downtime: 0.01 percent slash plus 10 minute jail. A validator that misses too many blocks in the signed-blocks window is jailed and must unjail to rejoin. The signed-blocks window and the missed-block threshold are TBD.
- Double-sign (equivocation): 5 percent slash plus tombstoning. A tombstoned validator can never rejoin. Equivocation evidence is handled through x/evidence and x/slashing.
- Slashing applies to bonded stake, so delegators share the loss. This is the incentive for delegators to choose reliable validators, and the reason validator selection matters.
- Unbonding: 21 days (Hub precedent, proposal). During unbonding, funds earn no rewards and cannot be transferred. Unbonding is the security mechanism that makes slashing meaningful: a validator cannot exit instantly after misbehaving.
- Validator set changes are governed by bonded stake ranking, not by appointment.

## 6. Monitoring, distribution, and decentralization targets

### 6.1 Monitoring

Every validator monitors: block height, block time, missed blocks, peer count, mempool size, voting power, and upgrade height. The network monitors: total supply, issued total, treasury balance, governance proposal status, and validator set composition. Alerts fire for: missed blocks, halted chain, supply invariant failure, upgrade height approaching, and a validator falling out of the active set. Monitoring and alerting must be demonstrated in the proof of concept (criterion O2) before they are claimed as working.

### 6.2 Geographic and cloud-provider distribution

Validators disclose their country (or jurisdiction) and their infrastructure provider. The disclosure is public. The network measures: number of active validators, Nakamoto coefficient, top-10 stake share, and the largest single country, jurisdiction, and provider shares. The measurements are published in the explorer, so anyone can verify the distribution without trusting OGO Systems.

### 6.3 Decentralization targets (research targets, not launch requirements)

Definition B, sufficiently-decentralized set, carried forward from the Decision Preparation report:

- At least 50 active validators.
- Nakamoto coefficient at least 10.
- Top-10 stake share below 33 percent.
- No single country, jurisdiction, or infrastructure provider above 33 percent of voting power.

These are targets, not launch requirements. No live Cosmos chain in the evidence meets all four thresholds at once (Hub coefficient 3, top 10 about 56 percent, Coinbase 20.56 percent). A mechanism for a flat stake distribution still needs to be named. The targets must not be presented as a launch promise, a decentralization claim, or a guarantee. They are the bar the network works toward, measured publicly.

Definition A, minimum operational set: 4 validators. This is the liveness floor: the smallest set that can tolerate one validator offline and still produce blocks (3 of 4 is more than two thirds). Definition A is what the proof of concept runs. It is not a decentralization claim.

## 7. Remote signing and key security

- Consensus keys are held in a remote signer (cosmos/kms v0.1.0) or a hardware security module. The consensus key is never present as a plaintext file on the validator host. This is a proof-of-concept MUST criterion (S8).
- TMKMS is deprecated: it will be removed from the Cosmos Bug Bounty Program at the start of Q1 2027, and operators have about three months from the Ledger Security 2026.1 release (2026-10-08) to migrate. Do not build new validator operations on TMKMS.
- Operator keys, withdrawal keys, and consensus keys are separate keystores. A compromise of one does not compromise the others.
- Key rotation: in-place validator consensus key rotation is available in the 0.55 family (SDK v0.55.0, CometBFT v0.40.0). In the 0.53 matrix, rotation uses the standard unbond and re-bond path. The upgrade evaluation (M4) records which path applies.
- Key backups are encrypted, never plaintext, and never stored on the validator host. A plaintext backup is a compromise, not a backup.
- Multisig is recommended for high-value operator actions (unbonding, commission changes above a threshold, governance votes with large stake). Multisig adds operational complexity, and the multisig itself must be managed; it is recommended, not required, in the first mainnet phase.

## 8. Infrastructure cost model

Infrastructure only, per validator per month, in USD. Labour is separate. The tiers are planning assumptions, labelled illustrative.

| Tier | Infrastructure cost | Hardware baseline | Notes |
|---|---|---|---|
| Lean | about 75 USD | 2 plus CPU cores, 8 GB RAM, 200 GB SSD, 50 Mbps | Minimum for a testnet or a small validator. Not recommended for mainnet. |
| Standard | about 200 USD | 4 plus CPU cores, 32 GB RAM, 500 GB NVMe, 100 Mbps | Baseline used for the minimum fee need calculation. |
| Professional | about 500 USD | 8 plus CPU cores, 64 GB RAM, 1 TB NVMe, 1 Gbps, plus sentry nodes and remote signing | Baseline for a validator expected to be reliable under load. |

Labour is separate and dominates at the professional tier: 1,500 to 4,000 USD per validator per month for monitoring, upgrades, incident response, and key management.

Monthly infrastructure cost by validator count (USD, infrastructure only):

| Validators | Lean (75 USD each) | Standard (200 USD each) | Professional (500 USD each) |
|---|---|---|---|
| 4 (Definition A) | 300 | 800 | 2,000 |
| 25 | 1,875 | 5,000 | 12,500 |
| 50 (Definition B minimum) | 3,750 | 10,000 | 25,000 |
| 100 | 7,500 | 20,000 | 50,000 |

With labour, a 50-validator professional set costs 75,000 to 200,000 USD per month in labour alone, plus 25,000 USD per month in infrastructure. Testnet infrastructure is estimated at 775 to 2,350 USD per month. Mainnet OGO-operated infrastructure (the validators, RPC endpoints, indexer, explorer, and monitoring that OGO Systems itself operates, not the whole network) is estimated at 2,600 to 5,700 USD per month. Both figures exclude audits, legal analysis, incentives, and salaries.

Minimum validator need, in normalized OGO units per period, for comparison with fee revenue (from the Decision Preparation report and the scenario model; illustrative, 1 period equals 1 month):

| Validators | Lean | Standard | Professional (derived) |
|---|---|---|---|
| 4 | 585 | 985 | 2,462 |
| 25 | 3,656 | 6,156 | 15,390 |
| 50 | 7,312 | 12,312 | 30,781 |
| 100 | 14,625 | 24,625 | 61,562 |

Published figures are 100 validators at lean (14,625) and standard (24,625). Other rows scale linearly by validator count, which is an illustrative assumption. The professional 100-validator figure (61,562) is derived as standard times 500/200, illustrative, not published. The need figures are infrastructure only. If labour is included, the need is higher.

## 9. What OGO would need to fund validator participation, without a price assumption

No OGO/USD market price is assumed anywhere in these specifications, because no price exists and none is projected. Validator participation funding must therefore be expressed in two separate currencies, never blended:

- **Infrastructure costs, in USD.** These are real, fiat-denominated, and known: 75, 200, or 500 USD per validator per month for infrastructure, plus 1,500 to 4,000 USD per validator per month for labour at the professional tier. If OGO funds validator participation, the infrastructure portion is budgeted in USD, from OGO company funds or from a disclosed grant program, and the budget is stated in USD. It is not stated in OGO, because the OGO value of a USD cost is unknown until a market price exists.
- **Stake-based incentives, in OGO units.** If validators are incentivized with OGO (for example, a share of issuance, or a delegation of stake), the incentive is stated in OGO units, with the explicit label that its USD value is unknown and no price is assumed. A delegation of stake by OGO Systems would require OGO Systems to hold stake acquired on the same publicly available conditions as everyone else: there is no reserved founder allocation and no reserved OGO Systems allocation, so OGO Systems cannot fund participation from a reserve. It would have to acquire the stake it delegates, on public terms, or fund participation in USD.
- **What must be disclosed before mainnet:** who funds validator participation (OGO Systems, a foundation, a grant program, or no one); in what currency (USD, OGO, or both); how much (in the currency it is paid in); for how long; and what happens when the funding ends. An undisclosed subsidy is a centralization risk: validators who depend on it will leave when it ends, and the network will not know until they do.
- **What OGO must not do:** promise a USD value for an OGO-denominated incentive; fund participation from a reserved allocation (none exists); or count OGO company revenue (subscriptions, AI inference, hosted apps, marketplace revenue) as protocol revenue. Protocol revenue and company revenue are separate (see `OGO_TREASURY_AND_SUSTAINABILITY.md`).

## 10. What this document does not decide

Validator minimum (Definition A vs Definition B); commission floor, notice period, and maximum change; unbonding period; slashing fractions and windows; active set size N; self-delegation minimum; incentive design; and the funding source for validator participation. All are listed in `OGO_NETWORK_DECISION_REGISTER.md`. The fee policy, which determines whether validators can be paid from fees after issuance ends, is specified in `OGO_FEE_MARKET_SPEC.md` and remains an open decision.
