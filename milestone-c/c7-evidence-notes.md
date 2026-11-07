# Milestone C: C7 Evidence Notes (gas, fee market, load)

Local staging only. Chain: 4-validator wired chain, ogo-local-1,
post-upgrade v2 binary, min gas price 0.001uogo, epoch 60s.

## Gas metering and fee market (B9)

Evidence collected live 2026-11-06:

- Signed transfer faucet to alice, 1,000,000uogo: committed height
  109, code 0, gas_used 92,777 (real gas metering, no hardcoded
  GasWanted anywhere in the new app; grep confirms only comments
  mention GasWanted).
- Unjail tx from val4: committed height 209, code 0, gas_used
  68,978.
- Below-minimum-fee tx rejected live: fee 1uogo with gas limit
  200,000, code 13, "insufficient fees; got: 1uogo required:
  200uogo". The required fee is exactly gas limit times the
  0.001uogo minimum gas price.
- Adequate-fee tx accepted live: fee 1,000uogo, code 0, txhash
  4CC49141A6E9070813C78CFEBEDBECB2AA416C5049E11EBBA8339E857D46F87E.
- Fee market conclusion: the chain enforces a minimum gas price
  and rejects underpriced transactions before execution; fees are
  metered per transaction. A gas-priced fee market exists in the
  minimal sense (min gas price plus per-tx gas metering). There is
  no priority fee auction or dynamic base fee; that was not in the
  plan scope.

## Load test calibration (60 seconds, 50 TPS target)

Ran 2026-11-06 ~22:51 with chain/loadgen (20 senders, 60s, 50 TPS
target, recipient alice). Results in loadgen-calib.jsonl and
loadgen-calib-summary.json:

- Submitted 2,957, accepted 2,934, failed 23. Achieved 48.9 TPS.
- Finality (sampled 200/200 committed): p50 0.018s, p95 1.079s,
  max 1.918s. p95 is under the P2 6 second threshold.
- The 23 failures were all one worker (worker 6) with account
  sequence mismatch ("expected 120, got 125"): five of its txs
  were accepted by CheckTx but never landed on chain, so its
  local sequence ran ahead. The generator now resyncs its sequence
  from the chain on a code 32 rejection instead of failing
  forever. Root cause of the five vanished txs was not isolated
  (mempool eviction under load is the likely cause); it is a
  load-test observation, not a chain safety issue, because no
  double-spend or state corruption resulted.
- Note on the p50 finality of 0.018s: it is suspiciously low for
  a ~1.8s block time. The workers' tickers are synchronised, so
  submissions arrive in bursts that can align just before a block
  proposal. The P2 run adds a clock-independent cross-check
  (commit height minus submit height) to validate the time-based
  finality numbers.

## P2 load test (200 TPS target, 10 minutes)

Ran 2026-11-06 22:53 to 23:03 UTC with chain/loadgen (100 senders,
10 minutes, 200 TPS target). Full results in loadgen-p2.jsonl and
loadgen-p2-summary.json.

- Submitted 45,858, accepted 45,857, failed 1. Achieved 76.4 TPS.
- Finality (sampled 180, all 180 committed): p50 2.71s, p95
  10.00s, max 28.40s.
- Height lag (commit height minus submit height): p50 2 blocks,
  p95 8 blocks, max 14 blocks. This is likely inflated by RPC
  staleness under load (the loadgen's per-tx /status calls hit the
  same saturated RPC endpoint), so the time-based finality is the
  primary measure.
- Verdict: P2 FAILS both thresholds. Throughput 76.4 TPS against a
  200 TPS target; p95 finality 10.0s against a sub-6s target.
  Under load the chain produced blocks of about 670 to 1,069
  transactions at roughly 9 seconds per block (height 1850 to
  1892 in the first 6 minutes), down from ~1.8 seconds per block
  unloaded. The mempool held 1,719 unconfirmed transactions at
  one check. The bottleneck was not isolated to a single component
  in this run; the likely contributors are block execution of
  1,000-transaction blocks, consensus rounds with a saturated RPC
  endpoint on the proposer, and mempool pressure.
- Only 1 broadcast failure in the run (the sequence resync fixed
  the desynced worker after a single code 32). The finality sample
  (180/180 committed) cannot rule out a small number of
  CheckTx-accepted txs that never committed; the calibration run
  proved that phenomenon exists (5 txs), so it is recorded as an
  open observation, not as a safety issue: no double-spend or
  state corruption resulted in either run.

## P2 side effect: val1 jailed for downtime during the load test

At height 1885 (about 22:57, during P2), val1 was jailed for
downtime and slashed 1 percent (100,000,000 to 99,000,000 tokens,
unbonding until 2026-11-27). Cause: all loadgen traffic (broadcasts
plus one /status call per transaction) went to val1's RPC endpoint
(27657), saturating it and starving val1's consensus participation
until it missed more than half of the last 50 blocks
(signed_blocks_window 50, min_signed_per_window 0.5). Val1 was
unjailed after P2 (tx D99872F81E58632DB774E43170EB389E4ECA682078CED891991E3990762D9C75,
code 0) and is BONDED again at 99,000,000 tokens; the slash is
retained, as designed. Lesson for future load tests: spread the
load across all four RPC endpoints (or use a dedicated non-validator
RPC node) so the generator cannot jail the validator it talks to.

## O5 snapshot recovery (val4, state sync)

Executed live 2026-11-06, twice. First (O5): val4 stopped, data
wiped, state sync from the height-2000 snapshot (trust hash
4DB8B5831D3E0312E734C89874447E386D14013F24210FC68CDC2B9D238A8723),
restored, caught up to 2038, all four nodes agreeing, val4 in the
active validator set at voting power 100. Second (M5 recovery):
same procedure after the rollback failure, caught up to 2216,
all four agreeing. Findings: Cosmovisor requires the data
directory to exist (recreate the empty dir after wiping);
priv_validator_state.json must be recreated (the node will not
start without it) and should be set to the validator's last
signed height from the backup, not the snapshot height (setting
it to 2000 when val4 had signed to 2024 produced "found
conflicting vote from ourselves" warnings at height 2033; no
slash or evidence resulted, and val4 is bonded and validating).

## M5 rollback (val4)

Executed live 2026-11-06. The rollback command ran ("Rolled back
state to height 2114"), but the node could NOT restart: it
panicked at load with "failed to load latest version: version of
store monetary-meta mismatch root store's version; expected 2114
got 0; new stores should be added using StoreUpgrades". The
monetary-meta store was added by the v2 store upgrade, and the
SDK rollback does not handle a store added mid-chain. M5 FAILS
for a node that has undergone a store upgrade. The chain itself
was unaffected (the other three validators kept producing
blocks). Val4 was recovered by the O5 state-sync procedure.
Do not use ogod rollback on a node that has applied a store
upgrade; use snapshot restore instead.

## O2 monitoring

- Prometheus metrics are live on val4 at 127.0.0.1:27690/metrics
  (enabled 2026-11-06, 273 cometbft_ metrics including
  cometbft_consensus_height and cometbft_mempool_size). The other
  three nodes still have prometheus disabled.
- The P1 observer (scripts/p1_observe.py, started 2026-11-06
  22:32:52 UTC) polls all four RPC endpoints every 60 seconds
  and logs height and app-hash agreement plus hourly block-rate
  summaries to p1-24h.log.
- Alerting is NOT demonstrated: no Prometheus server, alert
  rules, or notification channel are configured. That is the
  remaining O2 gap.

## Epoch hook errors after the schedule ends

Since epoch 5 (about 21:40), every epoch end logs "Error after
epoch end with identifier monetary epoch number N" from x/epochs.
This is expected: the TEST-ONLY monetary schedule has 4 periods
and was fully minted by epoch 4 (pool 3,439, supply
1,800,003,439, both unchanged since). The hook refuses to mint
beyond the schedule, and the x/epochs keeper logs the hook error
and advances the epoch, which is the upstream behaviour. No
over-minting occurred. The log line does not include the error
text; the refusal is proven by the unchanged pool and supply.
