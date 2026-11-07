# OGO Consensus and Epoch Specification

Status: Technical specification. Planning only. No blockchain deployed, no token issued, no monetary parameter selected, no production change authorized.
Date: 2026-11-06
Companion document: `ogo-network-native-coin-architecture-implementation-plan-v1.0.md` (sections 2.6, 3.5, 3.8, 4.7)
Architecture direction: CONFIRMED AND LOCKED, Option B. OGO Network is an independent Layer 1 blockchain with native OGO Coin.
Epoch duration: NOT SELECTED. No duration D is chosen in this document.

## 1. Purpose

This document specifies how OGO Network reaches consensus, how time is divided into epochs, and how the monetary schedule executes on top of epochs. It answers, in order: how an epoch begins and ends, how the network recognizes an epoch transition, how issuance executes exactly once per epoch, how duplicate issuance is prevented, how missed epochs are handled after a chain halt, how epoch ordering is preserved, how upgrades affect the schedule, and how epoch parameters are protected from arbitrary modification.

## 2. Consensus model

CometBFT Byzantine fault tolerant consensus, Tendermint-style, with instant finality per block. A block is final when committed. There is no probabilistic confirmation depth.

- Block time target: 1 to 6 seconds typical for CometBFT chains. Proof-of-concept criterion: 95 percent of blocks within 1 to 10 seconds over a 24-hour run.
- Active validator set: the top N validators by bonded stake, where N is a genesis parameter (TBD). Minimum operational set: 4 validators. Sufficiently-decentralized target: at least 50 validators, Nakamoto coefficient at least 10, top 10 under 33 percent, no country or provider over 33 percent (research targets, not launch requirements).
- Liveness: the chain halts if more than one third of voting power is offline. The proof of concept must demonstrate continued block production with 3 of 4 validators.
- Consensus key rotation: in-place rotation is available in the 0.55 family (SDK v0.55.0, CometBFT v0.40.0). In the 0.53 matrix, rotation uses the standard unbond and re-bond path.
- Block timestamps: the block time is the weighted median of validator timestamps. This matters for epochs (section 4), because epoch ticks are driven by block time.

## 3. Why epochs, not blocks

A period is one epoch of a named timer, with a fixed duration D in seconds, fixed in genesis. The nominal year is D times the number of periods per nominal year, fixed in genesis. The final issuance period N is genesis-fixed.

Three ways to divide time were evaluated:

| Approach | How it works | Failure mode | Verdict |
|---|---|---|---|
| Block-based periods | A period is a fixed number of blocks (for example, every 100,000 blocks) | Block time drifts. If blocks are 10 percent slower than assumed, a geometric schedule ends 11 months to 2.8 years late, depending on the horizon. Solana was 276 days behind its intended inflation timeline by June 2026 for this reason. | Rejected for production. The scaffold uses block-based periods today; Milestone C migrates to epochs. |
| Pure calendar time | A period ends at a fixed wall-clock time, read from block timestamps | Block timestamps are a weighted median under validator control. A validator set can shift the median. Tachyon, GHSA-c32p-wcqj-j677, February 2026, demonstrated median-time manipulation in CometBFT. Pure calendar time trusts the median directly. | Rejected. |
| Epoch-based periods (x/epochs) | A period is one epoch of a named timer. The timer ticks at the first block whose block time exceeds the epoch end time. The schedule is keyed by epoch number, not by wall-clock time. | Epoch boundaries still depend on block time for the tick, but the monetary rules depend only on the epoch number and the fixed duration D, never on the exact wall-clock time of a tick. | Recommended. |

The epoch approach does not eliminate timestamp trust. It confines it: a manipulated median can shift when a tick happens by a block or two, but it cannot change which epoch number is current, how much is issued for that epoch, or whether an epoch has already been claimed. The issued-by-epoch set (section 6) is the guard that makes this true.

## 4. Epoch lifecycle: how an epoch begins and ends

The epochs module is `x/epochs`, a first-class Cosmos SDK module from v0.53.0 (upstreamed from Osmosis, PR #23815). See `OGO_NETWORK_ARCHITECTURE.md` section 6 for the investigation findings. OGO uses one epoch identifier for the monetary schedule (name TBD in genesis, for example `monetary`). Other identifiers may be added later for other purposes (for example, a daily identifier for incentive distribution), but the monetary identifier is the only one x/monetary listens to.

Revision note, 2026-11-06: the first-class status of x/epochs from SDK v0.53.0, upstreamed through PR #23815, is confirmed and retained here. No vendored Osmosis copy and no custom timer is needed. The mandatory depinject hook-fires test that proves the module works in the wired application is specified in section 13; it is release-blocking for the monetary prototype.

### 4.1 EpochInfo

The module keeps a single `EpochInfo` per identifier, with these fields:

- `identifier`: the timer name (for example, `monetary`).
- `start_time`: the start time of the current epoch.
- `duration`: D, the fixed duration in seconds, set in genesis.
- `current_epoch`: the epoch number, starting at 0 or 1 as fixed in genesis (0-indexed is used in this specification: period n is epoch n).
- `current_epoch_start_height`: the block height at which the current epoch started.
- `current_epoch_start_time`: the block time at which the current epoch started.
- `epoch_counting_started`: whether the timer has started.

EpochInfos are initialized as part of genesis initialization or upgrade logic. Adding x/epochs to an application that was built without it requires a StoreUpgrade.

### 4.2 How an epoch begins

An epoch begins when the previous epoch ends (section 4.3). At the start, the module:

1. Sets `current_epoch_start_time` to the prior epoch's end time (not to the current block time). This is what keeps the epoch grid fixed: the grid is anchored at genesis, and each epoch starts where the previous one was scheduled to end, even if the ticking block arrives late.
2. Sets `current_epoch_start_height` to the current block height.
3. Increments `current_epoch` by 1.
4. Emits an `epoch_start` event carrying the new epoch number.
5. Calls the `BeforeEpochStart` hook on every registered module, with the identifier and the new epoch number. x/monetary filters on the monetary identifier and ignores all others.

The first epoch is special: its start time is the genesis-fixed first epoch start, and its start height is the genesis height or the height at which the epochs module is first initialized. The first epoch start height and the epoch duration are genesis parameters.

### 4.3 How an epoch ends

An epoch ends in the BeginBlocker of the first block whose block time is greater than the epoch end time, where end time equals start time plus duration. At the end, the module:

1. Calls the `AfterEpochEnd` hook on every registered module, with the identifier and the ending epoch number. x/monetary executes issuance for that epoch number in this hook (section 5).
2. Emits an `epoch_end` event carrying the ending epoch number.
3. Immediately begins the next epoch (section 4.2), so there is no gap between epochs.

The tick is therefore driven by block time, but the epoch grid is fixed. If blocks are slow, an epoch ends late (at the first block after its end time). If blocks are fast, an epoch still ends only when a block time passes its end time; it cannot end early. An epoch can never end twice, because the end time advances with each tick.

### 4.4 How the network recognizes an epoch transition

Every node recognizes the same transition, because the transition is a deterministic function of chain state:

- The `EpochInfo` (identifier, start time, duration, current epoch number) is on-chain state, replicated to every node.
- The BeginBlocker logic is deterministic: given the same block time and the same `EpochInfo`, every node takes the same branch.
- The `epoch_start` and `epoch_end` events are emitted in the block results, so an indexer, an explorer, or an external observer can see the transition without trusting a single node.
- The epoch number is queryable (`<appd> query epochs epoch-infos`), so anyone can check which epoch the chain believes it is in.

A node that disagrees about the current epoch number disagrees about chain state, which means it has a different state machine or a different history. That is a consensus failure, not an epoch ambiguity, and it is detected the same way any state divergence is detected.

## 5. Issuance execution: exactly once per epoch

Issuance for epoch n is R_n, computed from the genesis-fixed schedule. It is claimed once, in the `AfterEpochEnd` hook, keyed by epoch number.

### 5.1 The execution path

1. x/epochs calls `AfterEpochEnd(ctx, identifier, epochNumber)` on x/monetary at the end of epoch n.
2. x/monetary filters on the identifier. If the identifier is not the monetary identifier, it does nothing and returns.
3. x/monetary checks the issued-by-epoch set. If epoch n is already in the set, it returns without minting. This is the duplicate guard (section 6).
4. x/monetary computes R_n from the schedule. The computation is deterministic integer arithmetic: no floating point, one integer-floor rounding rule, deterministic period indexing (period n is epoch n, 0-indexed).
5. x/monetary mints exactly R_n into the rewards pool module account, through the amount-less entrypoint `MintScheduled(epochNumber)`. The function takes no amount argument, so no caller can request a larger mint.
6. x/monetary records epoch n in the issued-by-epoch set, with the amount minted and the block height.
7. x/rewards splits the minted amount between validators, delegators, and the community pool or treasury, per the approved split. The split conserves exactly: the parts sum to R_n, in integer arithmetic.
8. The every-block invariant checks that total issued is less than or equal to total scheduled, and that total supply is less than or equal to max supply M.

### 5.2 What "exactly once" means

Exactly once means: for each epoch number n from 0 to N minus 1, R_n is minted one time, and only one time. It does not mean the mint happens at a particular wall-clock time, or in a particular block. It means the issued-by-epoch set contains n exactly once, and the total issued equals the sum of R_n over all epochs in the set, exactly.

A re-claim of the same epoch fails. A claim for an epoch number greater than the current epoch fails. A claim for an epoch number less than or equal to the last claimed epoch, if it is already in the set, does nothing. There is no code path that mints an arbitrary amount, and no code path that mints outside the hook.

## 6. Duplicate issuance prevention

Four layers, each independent, so a failure in one does not by itself allow a duplicate mint:

1. **Issued-by-epoch set.** The set is on-chain state, keyed by epoch number. The hook checks membership before minting and records membership after minting, in the same block, in the same transaction context. A second call for the same epoch number sees the set entry and returns without minting. The set is queryable, so anyone can verify which epochs have been claimed.
2. **Amount-less entrypoint.** `MintScheduled(epochNumber)` takes no amount. It computes R_n from the schedule. A caller cannot mint twice the amount by passing a larger number, because there is no number to pass.
3. **Deterministic hook ordering.** Hooks run in a fixed order in the BeginBlocker, and x/monetary's hook runs once per epoch end. The hook is not callable from a transaction, not callable from a contract, and not callable from governance. The only caller is x/epochs, at the end of an epoch.
4. **Every-block invariant.** At every block, the chain checks that total issued is less than or equal to total scheduled, and that total supply equals genesis supply plus total issued minus total burned. If a duplicate mint somehow occurred, the invariant would fail at the next block, and the block would be rejected or the chain would halt, per the configured behaviour, with the failure logged.

Test: claim epoch n, then attempt to claim epoch n again in the same block and in a later block. Both attempts fail, and total issued is unchanged. Test: attempt to call `MintScheduled` directly from a transaction or a contract. It fails, because the entrypoint is not exposed as a message or a binding.

## 7. Halt recovery: missed epochs processed sequentially

Halt-recovery principle, locked: missed epochs are processed sequentially, without a lump-sum mint and without silently skipping scheduled issuance.

### 7.1 What happens during a halt

While the chain is halted (no blocks), no epoch ends, because epoch ends happen in the BeginBlocker of a block. Epoch end times pass in wall-clock time, but no tick occurs. The `EpochInfo` still says the current epoch is the one that was running when the chain halted. No issuance happens during the halt, because issuance happens in the `AfterEpochEnd` hook, which only runs in a block.

### 7.2 What happens on restart

On restart, the chain produces blocks again. For each block, the BeginBlocker checks whether the block time is greater than the current epoch end time. If it is, the current epoch ends (with issuance for that epoch number), and the next epoch begins. The next epoch's end time is start time plus duration, where start time is the prior epoch's end time, not the current block time. So if the chain was halted for longer than one epoch duration, the next epoch's end time is already in the past, and it will end at the next block. Missed epochs therefore tick one per block on restart, in epoch order, until the chain catches up to the current epoch.

Concretely: if the chain halts for 30 minutes and the epoch duration is 60 seconds, about 30 epochs are missed. On restart, the first block ends the oldest missed epoch and mints its R_n. The second block ends the next missed epoch and mints its R_n. This continues, one epoch per block, until all missed epochs are processed. At no point is more than one epoch's R_n minted in a single block. At no point is an epoch skipped. Total issued after catch-up equals the sum of R_n for all elapsed epochs, exactly.

### 7.3 What must be verified, not assumed

Whether the pinned x/epochs BeginBlocker advances one epoch per block or attempts to catch up multiple epochs in a single block must be read from the actual SDK source at the pinned version during Milestone C. If the pinned implementation advances more than one epoch per block, x/monetary must enforce the one-per-block catch-up itself, by claiming at most one epoch per block and leaving the rest for subsequent blocks. The acceptance test is proof-of-concept criterion S4: halt the chain for 30 minutes, restart, and observe missed epochs ticking one per block on restart, in epoch order, with no lump mint, no skipped epoch, and no double mint. Total issued after catch-up must equal the sum of R_n for all elapsed epochs.

### 7.4 Long halts and repeated halts

A halt longer than the full issuance schedule (all N epochs) is an edge case: on restart, the chain would need N blocks to catch up, minting one epoch per block. This is correct behaviour, not a problem, but it means the catch-up takes N blocks, and during those blocks the chain is issuing historical epochs, not the current one. The specification does not cap the catch-up. If a halt is so long that catching up is impractical, that is a governance decision (through the constitutional process), not an automatic skip. The chain never silently skips scheduled issuance.

Repeated halts are handled the same way: each halt leaves a backlog of missed epochs, and each restart processes them one per block. There is no state that a halt can corrupt, because the epoch state is just the `EpochInfo` and the issued-by-epoch set, both of which are ordinary on-chain state.

## 8. Epoch ordering

Epoch ordering is preserved by construction:

- Epoch numbers are sequential integers, starting at the genesis-fixed first epoch number, incrementing by 1 at each tick.
- The issued-by-epoch set is keyed by epoch number, so issuance for epoch n cannot be recorded as issuance for epoch m.
- The `AfterEpochEnd` hook receives the epoch number as an argument, so x/monetary always knows which epoch it is issuing for.
- The `epoch_start` and `epoch_end` events carry the epoch number, so the ordering is visible in the block results.
- A claim for an epoch number that is not the next unclaimed epoch, in the catch-up case, is handled by the one-per-block rule: the oldest unclaimed epoch is claimed first, then the next, in order. The chain never claims epoch n plus 2 before epoch n plus 1.

## 9. Upgrades and the schedule

Upgrades affect the schedule only through the constitutional process, never through an ordinary governance proposal or an ordinary upgrade.

- A software upgrade (through x/upgrade and Cosmovisor) must not modify monetary state. The upgrade handler contains a denylist gate that rejects any upgrade attempting to modify max supply, the schedule, the issued-by-epoch set, or the treasury rule outside the constitutional process. The denylist is checked before the migration runs.
- Adding x/epochs to an application built without it requires a StoreUpgrade. That is a state migration, not a schedule change: it adds the epochs store and initializes the `EpochInfo`, with the duration D and the first epoch start taken from genesis. The schedule itself is unchanged.
- The SDK 0.55 upgrade changes the `x/epochs` NewAppModule signature (keeper by pointer instead of value). That is a wiring change, not a schedule change, and it must be tested with the mandatory depinject hook-fires test (section 13), not just a compile check.
- After any upgrade, the bypass suite is re-run (proof-of-concept criterion M3). A governance upgrade attempting to change max supply or the schedule is still rejected, and a non-monetary upgrade is still accepted.
- A hard fork can replace the code, including the schedule. No chain can promise its community will never adopt different software. The protection is transparent rules, independent operators, and a visible upgrade process, not a claim that a hard fork is impossible.

## 10. Epoch parameter protection

Epoch parameters are protected from arbitrary modification by the same five-layer governance security boundary that protects the monetary rules (see `OGO_NETWORK_SECURITY_MODEL.md`):

1. **Parameter separation.** The epoch duration D, the first epoch start, and the final issuance period N are stored in x/monetary and x/epochs genesis state, not in a generic parameter store that x/gov can change. A governance proposal that targets an epoch parameter has no parameter to change.
2. **Upgrade-handler denylist.** An upgrade that attempts to modify epoch parameters outside the constitutional process is rejected before its migration runs.
3. **Amount-less mint.** The mint entrypoint takes only the epoch number. It cannot be used to change the schedule, only to execute it.
4. **Every-block invariant.** If the schedule were somehow changed, the invariant (total issued less than or equal to total scheduled, total supply less than or equal to max supply) would fail at the next block.
5. **Public queryability.** The epoch duration, the current epoch number, the issued-by-epoch set, total issued, total scheduled, and max supply are all public queries. Any change is visible in the explorer.

What governance CAN change: governance-controlled parameters listed in the plan (the validator, delegator, community, and treasury split; the fee level and fee distribution; governance parameters themselves), each with an explicit permission, a voting threshold, and a safeguard. The epoch duration is not among them.

## 11. Comparison: block-based periods vs epoch-based periods

| Property | Block-based periods | Epoch-based periods (x/epochs) |
|---|---|---|
| Period definition | Fixed number of blocks | Fixed duration D in seconds, one epoch |
| Drift | Drifts with block time. 10 percent slower blocks slip a geometric end 11 months to 2.8 years. Solana 276 days behind by June 2026. | No drift in the grid: each epoch starts where the previous one was scheduled to end, even if the ticking block arrives late. |
| Timestamp trust | Block height is trusted, not time. | Block time drives the tick, but monetary rules depend only on the epoch number and D, never on exact wall-clock time. |
| Halt behaviour | A halt delays periods by the halt length, and the schedule resumes from the halted height. No catch-up is needed, but the schedule end date slips by the halt length. | A halt creates a backlog of missed epochs, processed one per block on restart, with no lump mint and no skip. The schedule end date does not slip, but catch-up takes one block per missed epoch. |
| Determinism | Deterministic in block height. | Deterministic in epoch number. Two nodes with the same state agree on the current epoch number. |
| Implementation | Hand-built counter in x/monetary (the scaffold's current approach). | SDK module, maintained upstream, with hooks, events, and queries. Requires a StoreUpgrade to add. |
| Known risks | Drift (evidenced). No upstream maintenance (OGO owns the counter). | Depinject hook-registration bug in the 0.53 line (issues #24770, #25425; fixed by the v0.54 NewAppModule change). Halt catch-up behaviour must be verified against the pinned source. |
| Verdict | Rejected for production. Retained in the scaffold only until Milestone C migrates it. | Recommended. |

## 12. What this document does not decide

Epoch duration D is not selected. The nominal year is not selected. The first epoch start is not selected. The issuance model, its parameters, the split, the fee policy, the treasury spend level, and the emergency reserve floor are not selected. The epoch identifier name is not selected (it is fixed in genesis, and this document uses `monetary` as a working name). All are listed in `OGO_NETWORK_DECISION_REGISTER.md`.

## 13. Acceptance tests

- S3 (epoch issuance, exactness): issuance for each epoch equals exactly R_n, is claimed exactly once per epoch, and total issued never exceeds total scheduled. The issued-by-epoch set is keyed by epoch number, and a re-claim of the same epoch fails.
- S4 (epoch halt and catch-up): halt the chain for 30 minutes, restart, observe missed epochs ticking one per block on restart, in epoch order, with no lump mint, no skipped epoch, and no double mint. Total issued after catch-up equals the sum of R_n for all elapsed epochs.
- Hook-fires test (mandatory depinject test, release-blocking for the monetary prototype): in the wired application, with depinject, advance the chain past an epoch boundary and confirm all five of the following, in order. (1) The epoch module initializes in the wired application. (2) The monetary hook is registered: the x/monetary `AfterEpochEnd` hook is present in the wired hook set, not only in a keeper unit test. (3) The hook actually fires at an epoch transition: issuance is minted and the epoch event is emitted when the boundary is crossed. (4) Exactly one scheduled issuance occurs for that epoch: the minted amount equals exactly R_n for the epoch, and the issued-by-epoch set contains the epoch exactly once. (5) A duplicate hook execution cannot issue twice: invoking the hook a second time for the same epoch mints nothing and leaves total issued unchanged. Compilation alone is not an acceptable test, and a keeper-level unit test alone is not an acceptable test: the failure this test exists to catch is a hook that is wired but never fires in the deployed application, which looks identical to a schedule working correctly with zero issuance. This test exists because of the depinject hook-registration bug (issues #24770, #25425).

Revision note, 2026-11-06: the hook-fires test above replaces the earlier, looser wording, which only asked to "observe the hook executing" without naming what must be confirmed. The five confirmations are now mandatory acceptance criteria, and the test is release-blocking for the monetary prototype: the prototype does not pass while any one of the five is unproven.
- StoreUpgrade test: add x/epochs to the scaffold through a store upgrade, with the migration tested on the proof-of-concept chain before it is used for any real schedule.
- Duplicate-claim test: claim epoch n, then attempt to claim epoch n again in the same block and in a later block. Both fail, total issued unchanged.
- Upgrade test: after a Cosmovisor upgrade, re-run the bypass suite (M3). A governance upgrade attempting to change max supply or the schedule is still rejected.
