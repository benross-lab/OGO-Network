# OGO Fee Market Specification

Status: Technical specification and scenario model. Planning only. No fee policy selected, no monetary parameter selected, no production change authorized.
Date: 2026-11-06
Companion documents: `ogo-network-native-coin-architecture-implementation-plan-v1.0.md` (sections 3.11, 5.4), `OGO_TREASURY_AND_SUSTAINABILITY.md` (sections 7, 11)
Architecture direction: CONFIRMED AND LOCKED, Option B. OGO Network is an independent Layer 1 blockchain with native OGO Coin.
All figures in this document are ILLUSTRATIVE comparison inputs, not selected parameters. No OGO/USD price is assumed anywhere. Fees are expressed in uogo and in OGO, and as a share of the payment, which is price-independent.

## 1. Purpose

This document specifies the fee architecture for ordinary transactions and AI-agent micropayments: base transaction fees, gas pricing, dynamic fee adjustment, fee distribution, validator compensation, treasury allocation, spam prevention, high-frequency agent transactions, batch transactions, and other micropayment approaches. The fee model must not make a small AI-agent payment economically impractical. The document models different transaction sizes and fee structures, and it does not assume application fees automatically become protocol revenue.

## 2. Requirements

1. Fees are paid in `uogo`, the native denomination. No fee is ever collected in a different denomination at launch.
2. Fees are collected by the ante handler and distributed by x/rewards, per the approved fee distribution. The distribution conserves exactly.
3. The fee model must not make a small AI-agent payment economically impractical. This is a requirement, not a preference, and section 5 tests it.
4. Fees must fund validators after issuance ends, or the fee policy must state explicitly what happens if they do not (section 8).
5. Fees must prevent spam: a transaction that costs nothing can be replayed without limit, and a network with free transactions will be filled with them.
6. Fees must be predictable enough for agents to budget: an agent that cannot predict what a transaction will cost cannot decide whether to make it.
7. Application fees are not protocol revenue. See section 9.
8. Agent micropayment viability is an explicit protocol design requirement, not an application-layer afterthought. The fee architecture must explicitly distinguish ordinary transactions from agent micropayments: they have different sizes, different frequencies, and different fee tolerance, and a single flat fee cannot serve both (section 5). Agent-native batching is a core protocol consideration, evaluated in section 7 alongside aggregation, fee sponsorship, and other agent transaction mechanisms, and the fee policy must state which mechanisms the protocol supports before genesis.

Revision note, 2026-11-06: requirement 8 is new. The earlier version of this document treated agent micropayments as a test the fee model had to pass (requirement 3) while leaving the mechanisms that make them viable, batching in particular, as application conveniences. That framing is corrected here: viability is a protocol design requirement, ordinary transactions and agent micropayments are distinguished in the architecture, and batching is a core protocol consideration. The 1,000 uogo fee example and the one-million-transactions-per-day finding that motivate this requirement are recorded in sections 5 and 6.

## 3. Fee architecture

### 3.1 Components

- **Base fee:** a minimum fee per transaction, in uogo, set as a network parameter. Every transaction pays at least the base fee.
- **Gas price:** a price per unit of gas, in uogo, set as a network parameter (minimum gas price). The fee for a transaction is gas used times gas price, with a minimum of the base fee. Gas is a measure of computation and storage, following the standard Cosmos model.
- **Gas limit:** the maximum gas a transaction may consume, set by the sender. A transaction that exceeds its gas limit fails, and the gas consumed up to the failure is still charged, so a failing transaction is not free.
- **Priority fee:** an optional additional fee the sender can pay to prioritize their transaction when blocks are full. Priority fees go to the block proposer (the validator), following the standard model.
- **Memo:** a transaction memo is supported for payment references. The memo does not affect the fee, except through gas (a longer memo uses more gas).

### 3.2 Gas pricing

Gas pricing follows the standard Cosmos SDK model: each message type has a gas cost, storage operations have a gas cost, and the total gas for a transaction is the sum of its operations. The gas costs are set in the application configuration, not in consensus-critical monetary code, but the fee calculation (gas times gas price) is deterministic and identical across nodes.

The minimum gas price is a governance-controlled parameter. It can be changed through x/gov, with an explicit permission, a voting threshold, and a safeguard (a maximum change per proposal, a timelock, or both). The fee formula itself is consensus-critical; the fee level is governance-controlled.

### 3.3 Dynamic fee adjustment

A static fee is simple and predictable, but it cannot respond to congestion: when blocks are full, a static fee gives no way to prioritize, and when blocks are empty, a static fee overcharges. Dynamic fee adjustment is therefore specified as an option, with the choice left open:

- **Option A, static fee:** the base fee and gas price are fixed network parameters, changed only by governance. Predictable, simple, and the easiest to reason about. Risk: no congestion response, and the fee can become wrong (too high or too low) between governance changes.
- **Option B, congestion-adjusted fee:** the base fee adjusts per block based on the previous block's fullness (for example, the EIP-1559 style: the base fee rises when blocks are more than half full and falls when they are less than half full, bounded by a maximum change per block). Responsive to congestion, and it gives a price signal. Risk: less predictable for agents, and the adjustment parameters (target fullness, maximum change per block, minimum and maximum fee) must be chosen and tested.
- **Option C, hybrid:** a static base fee plus a congestion-adjusted priority fee. The base fee is predictable; the priority fee is the congestion response. This is the recommended starting point for evaluation, not a decision: it gives agents a predictable floor and gives the network a congestion signal, without making the base fee unpredictable.

No option is selected in this document. The choice is an open decision, and it must be tested in the proof of concept under load (criteria P2 and P3) before it is chosen.

## 4. Fee distribution

Fees are distributed by x/rewards, per the approved fee distribution. The distribution shares (validator, delegator, community, treasury) are TBD. The distribution must:

- Sum to exactly 1, in integer arithmetic. The parts sum to the fees collected, with no remainder and no shortfall.
- Be fixed before genesis, and changeable only through x/gov with an explicit permission, a voting threshold, and a safeguard.
- Be visible: the fee distribution for each block is queryable, and the explorer displays it.

Fee burn: if any fee is burned, the burn is a reduction in fee revenue available to validators and the treasury, not a funding mechanism. Ethereum burns the base fee, but burning is not a substitute for a funding model: Ethereum's burn collapsed from 8,844 per day in 2021 to about 57 per day in 2026, offsetting only about 2.4 percent of issuance, and ETH has been net-inflationary for 28 straight months. If OGO burns any fee, the burn amount, the burn destination, and the effect on validator and treasury revenue must be modelled before genesis. No burn is adopted by these specifications.

## 5. The agent micropayment test: fee as a share of payment

The requirement is that the fee model must not make a small AI-agent payment economically impractical. The test is price-independent: the fee as a percentage of the payment, in OGO, with no USD conversion and no price assumption.

Fee as a share of payment, at three illustrative fee levels (100 uogo = 0.0001 OGO, 1,000 uogo = 0.001 OGO, 10,000 uogo = 0.01 OGO per transaction):

| Payment (OGO) | Fee 100 uogo | Fee 1,000 uogo | Fee 10,000 uogo | Fee 100 uogo, batched 100 payments per transaction |
|---|---|---|---|---|
| 0.001 | 10.00 percent | 100.00 percent | 1,000.00 percent | 0.10 percent |
| 0.01 | 1.00 percent | 10.00 percent | 100.00 percent | 0.01 percent |
| 0.1 | 0.10 percent | 1.00 percent | 10.00 percent | 0.001 percent |
| 1.0 | 0.01 percent | 0.10 percent | 1.00 percent | 0.0001 percent |
| 10.0 | 0.00 percent | 0.01 percent | 0.10 percent | 0.0000 percent |

Reading:

- **At a 1,000 uogo fee, a 0.001 OGO payment costs 100 percent in fees.** The payment is entirely consumed by the fee. That fee level makes the smallest agent payments impossible.
- **At a 100 uogo fee, a 0.001 OGO payment costs 10 percent in fees.** That is expensive but not impossible. A 0.01 OGO payment costs 1 percent, which is comparable to a payment processing fee.
- **At a 10,000 uogo fee, even a 0.1 OGO payment costs 10 percent, and a 0.01 OGO payment costs 100 percent.** That fee level makes all small agent payments impossible.
- **Batching changes the answer completely.** If 100 payments of 0.001 OGO are batched into one transaction paying a single 100 uogo fee, the fee is 0.10 percent of each payment. Batching is the difference between a viable micropayment system and an impossible one.

Revision note, 2026-11-06: the 1,000 uogo fee example in this section is confirmed and retained as the headline case: a 0.001 OGO payment carries a 100 percent fee burden at a 1,000 uogo fee, so the payment is entirely consumed by the fee; the same 100 payments batched into one transaction carry a 0.10 percent burden each at a 100 uogo fee. This example is what makes agent micropayment viability an explicit protocol design requirement (requirement 8) and agent-native batching a core protocol consideration (section 7), rather than an application convenience.

The fee policy must therefore state, explicitly, how agent transactions are priced. A flat fee cannot serve both security and agent viability at the same time: agents are 45 percent of transactions but only 7.7 percent of fee revenue in the High fee scenario in the research, because their payments are small and their fee contribution per transaction is the same as a large payment's. If agents are charged the same flat fee as human transactions, either the fee is too high for agents or too low for security. The policy options are: a lower fee for small payments (a tiered fee); batching (section 7); a separate agent transaction type with its own fee; or an explicit statement that small agent payments are not a supported use case (which would contradict the network's purpose, and must not be chosen silently). No option is selected in this document.

## 6. Fee revenue vs validator need

Fee revenue per period, assuming 1 period equals 1 month (30 days), at three fee levels and five transaction volumes. Minimum need is 24,625 normalized units per period (100 validators, standard cost, infrastructure only).

| Transactions per day | Fee 100 uogo | Fee 1,000 uogo | Fee 10,000 uogo | Need coverage at 1,000 uogo |
|---|---|---|---|---|
| 1,000 | 3 | 30 | 300 | 0.1 percent |
| 10,000 | 30 | 300 | 3,000 | 1.2 percent |
| 100,000 | 300 | 3,000 | 30,000 | 12.2 percent |
| 1,000,000 | 3,000 | 30,000 | 300,000 | 121.8 percent |
| 5,000,000 | 15,000 | 150,000 | 1,500,000 | 609.1 percent |

Reading:

- **Below 100,000 transactions per day, fee revenue is negligible against the need at any fee level that keeps small payments viable.** At 100,000 transactions per day and a 1,000 uogo fee, fee revenue covers 12.2 percent of the need.
- **At 1,000,000 transactions per day and a 1,000 uogo fee, fee revenue covers 121.8 percent of the need.** The fee transition requires about one million transactions per day at that fee level, for 100 validators at standard cost.

Revision note, 2026-11-06: the one-million-transactions-per-day finding is confirmed and retained: approximately one million transactions per day are required to cover the standard validator need (100 validators, standard cost, 24,625 normalized units per period) at the modeled fee level of 1,000 uogo per transaction. At the fee level that keeps small payments viable (100 uogo), the same need requires about 8,200,000 transactions per day. No comparison network has made this transition (section 6, final paragraph), so the volume assumption behind any fee policy must be stated with its evidence, not assumed.
- **The tension is direct:** the fee level that makes small payments viable (100 uogo) produces one tenth the revenue of the 1,000 uogo fee, so it requires ten times the transaction volume to cover the same need. At 100 uogo, covering the need requires about 8,200,000 transactions per day (24,625 divided by 0.0001 divided by 30). The fee policy must state the transaction volume it assumes, and what evidence supports that assumption.
- **Validator count matters:** the need figures are for 100 validators. For 4 validators at standard cost, the need is 985 per period, which is covered at 100,000 transactions per day and a 100 uogo fee (300 per period covers 30 percent; 1,000,000 per day covers it). The fee transition is easier with fewer validators, which is a reason to be explicit about the validator count the fee policy is designed for.

These are the same Low, Moderate, and High fee scenarios as the treasury document, expressed in transaction volume: Low is below 100,000 transactions per day, Moderate is around 100,000 to 1,000,000, High is above 1,000,000. The fee transition has not happened on any of the four comparison networks. Cardano's fees are 0.67 percent of rewards. The fee policy must not assume the transition will happen; it must state the volume and fee level at which it happens, and what happens if it does not.

## 7. High-frequency agent transactions and batching

Agents make many small transactions. Six approaches, in order of preference for evaluation. Batching is a core protocol consideration, not an application convenience: the protocol must define batch semantics (section 7, item 1) before genesis, because the viability of agent micropayments depends on it (requirement 8).

1. **Batch transactions.** Multiple payments are combined into one on-chain transaction, paying one fee. The batch is constructed by the agent or by an aggregator, signed once, and executed as a multi-send. The fee per payment is the transaction fee divided by the number of payments. At 100 payments per transaction and a 100 uogo fee, the fee per 0.001 OGO payment is 0.10 percent (section 5). Batching is the single most effective micropayment approach, and it requires no protocol change: multi-send already exists in x/bank. The open questions are who constructs the batch, who signs it, and what happens if one payment in the batch fails (the whole batch fails, or the failing payment is skipped). Those questions must be answered in the developer documentation before batching is recommended.
2. **Off-chain aggregation with on-chain settlement.** Payments are aggregated off-chain (in the OGO application, or in an agent service) and settled on-chain periodically, in a single transaction. The fee per payment is the settlement fee divided by the number of payments. This is more efficient than batching, but it introduces a trusted aggregator: the aggregator holds the payments between settlements, and if it fails, the payments are at risk. The aggregator must be disclosed as a trust assumption, and the settlement interval must be short enough that the risk is bounded. This approach is not part of the first monetary prototype.
3. **Payment channels.** Two parties open a channel, transact off-chain, and settle on-chain when the channel closes. The fee per payment is near zero, but channels require both parties to be online, require locked funds, and are complex to operate. Payment channels are a later-phase evaluation, not a first-prototype feature.
4. **Fee sponsorship.** A third party pays the transaction fee on behalf of the agent, so the fee burden leaves the micropayment entirely. The sponsor may be the application the agent works for, the agent operator, or a treasury-funded sponsorship program approved by governance. Sponsorship keeps the fee, it changes who pays it: the transaction still pays the full protocol fee in uogo, and validators are still compensated. That is what distinguishes sponsorship from a fee exemption, which would remove the fee and with it the spam defence and part of validator revenue. Sponsorship needs its own rules before it can be relied on: who may sponsor and under what authorization, a per-sponsor cap and a per-agent cap, revocation, and accounting that shows sponsored fees separately from ordinary fees. The abuse case is direct: a sponsor that pays for unlimited agent transactions pays for unlimited spam unless the sponsorship is rate-limited and revocable. A treasury-funded sponsorship program is a treasury spend, subject to the R-F rule and the reporting requirement in `OGO_TREASURY_AND_SUSTAINABILITY.md` section 2.1, not a separate money source.
5. **Meta-transactions and relayers.** The agent signs a payment intent off-chain and a relayer submits it on-chain, paying the fee and being reimbursed by the agent or by a sponsor. This separates signing from fee payment, which lets an agent with no fee balance transact, but it introduces a relayer as a trusted intermediary: the relayer can delay, reorder, or drop intents, and must be compensated. Relayers are an application-layer mechanism, not a protocol feature, and any relayer the OGO application operates must be disclosed as a trust assumption with its reimbursement terms.
6. **Prepaid fee credits and fee abstraction.** The agent prepays a fee balance and each payment draws its fee from that balance, or the fee is abstracted into a subscription the agent operator pays. This smooths fee payment and removes per-transaction fee friction, but it does not reduce the fee: viability still depends on the fee level and on batching. Prepaid credits are accounting, not a fee reduction, and must not be presented as one.

Evaluation summary: batching is the only mechanism that reduces the fee burden per payment without a third party, a trust assumption, or a subsidy, which is why it is the core protocol consideration. Aggregation reduces it further but adds a trusted aggregator. Sponsorship removes the burden from the agent but moves it to a sponsor, so it needs a funding source and abuse controls, and it is the only mechanism that can make a single unbatched micropayment viable without lowering the protocol fee. Payment channels and relayers are later-phase or application-layer options. No combination is selected in this document; the fee policy must state which mechanisms the protocol supports, and a mechanism that depends on a sponsor or an aggregator must name that party and its funding before genesis.

Revision note, 2026-11-06: items 4, 5, and 6 above are new. The earlier version of this section evaluated only batching, off-chain aggregation, and payment channels, and treated batching as an application convenience built on multi-send. Fee sponsorship was not evaluated at all. That gap is corrected here: sponsorship, meta-transactions and relayers, and prepaid fee credits are now evaluated, and batching is now a core protocol consideration whose semantics the protocol must define before genesis.

Spam prevention for high-frequency transactions: a per-transaction fee is the primary spam defence. Rate limiting by the application (not by the chain) is the secondary defence. The chain must not rate-limit by sender, because that would censor legitimate high-frequency agents. The chain's spam defence is the fee, and the fee must be high enough to make spam expensive and low enough to make legitimate agent payments viable. That is the same tension as section 5, and it is resolved by the same mechanisms (batching, tiered fees, or an explicit agent transaction type).

## 8. Validator compensation and treasury allocation

- Validators are compensated from two sources: a share of scheduled issuance (per the approved split, through x/rewards and x/distribution) and a share of transaction fees (per the approved fee distribution).
- The treasury receives a share of scheduled issuance and a share of transaction fees, per the approved split and fee distribution.
- The fee transition is the eventual shift from issuance-funded security to fee-funded security. At period 20, fees must replace 87.8 percent of period-0 validator pay under geometric (r = 0.90), or 75.5 percent under hybrid (base 40 percent), to keep validator pay flat. Neither model eliminates the transition; hybrid makes it gentler until its end date, then ends in a 100 percent step at period N.
- If fee revenue never covers the minimum need, the plan must state what happens: the validator set shrinks, validator pay falls, security degrades, or an explicitly disclosed alternative funding mechanism is used. None of these is chosen by these specifications, but one must be chosen before genesis. The fee policy must not assume the problem away.

## 9. Application fees are not protocol revenue

Application fees (OGO Pro subscriptions, AI inference charges, hosted apps revenue, marketplace application fees) are OGO company revenue. They are not protocol revenue, and they do not become protocol revenue by being collected in the OGO application.

- A fee charged by the OGO application for a service is company revenue, even if the service involves a blockchain transaction. The blockchain transaction's fee is protocol revenue. The application's fee is not.
- A fee charged by a third-party application built on OGO Network is that application's revenue, not protocol revenue, unless the application explicitly routes a portion to the protocol through a disclosed on-chain mechanism.
- No document, dashboard, or projection may add application fees to protocol fee revenue without naming the mechanism that routes them, how it is implemented, how it is verified on-chain, and who authorizes it. Until such a mechanism exists, is implemented, and is disclosed, application fees contribute zero to protocol revenue, to the treasury, and to validator funding.
- The same separation applies to the fee scenarios in section 6: the transaction volumes and fee levels there are protocol transactions and protocol fees only. Application revenue is not counted.

## 10. What this document does not decide

Fee level (base fee and gas price); fee structure (static, congestion-adjusted, or hybrid); fee distribution shares; whether any fee is burned; how agent transactions are priced (tiered fee, batching, separate transaction type, fee sponsorship, or another mechanism); and the transaction volume assumption the fee policy relies on. All are listed in `OGO_NETWORK_DECISION_REGISTER.md`. The fee policy must be decided before genesis, because it is the post-issuance funding mechanism, and the evidence shows the fee transition has not happened on any comparison network.
