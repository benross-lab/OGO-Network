# Module Boundaries: OGO Network Local Prototype (Milestones 1-2)

Status: local prototype. No deployment, no token issuance, no
production change. All values are TEST-ONLY and labelled so.

This file states, for every module in the prototype, whether it is
implemented, pinned-compiled only, or planned, and what each one owns.
It is the Milestone 1 record; update it when a status changes, and
never claim a pinned-compiled module as wired.

## How the prototype runs

The running binary is a CometBFT ABCI 2.0 application (`app/app.go`)
whose state machine is `x/ogocoin`. It produces real blocks on a
3-validator localnet, with real transfers, real scheduled issuance
through the epoch hook, and real supply, schedule, and epoch queries
over RPC. Three validators is the decided count for the local
prototype; it is distinct from the minimum operational set of 4
validators (the smallest set that tolerates one faulty validator)
and from the decentralization target of 50 or more validators for a
public network. It runs entirely on its own: the OGO application is
a client of the network, not its authority. Nothing in this prototype
reads from or writes to Supabase, the OGO Cloudflare Worker, or
benross-lab/OGO. The network operates independently of the OGO
application by construction: there is no code path from the chain to
the application at all. Keplr and live IBC are not prerequisites for
this prototype, and the monetary prototype is independent of smart
contracts; CosmWasm is evaluated separately before adoption.

## Implemented

### x/ogocoin (this repo, local module)

The only fully implemented module. It owns the monetary rules:

- Fixed maximum supply (mandatory in genesis, no default).
- Genesis validation: max supply, denom, decimals, discretionary
  minting disabled, supply cap enforced, mandatory issuance
  schedule, purpose labels on allocations, founder and reserved
  purposes rejected, genesis supply plus total scheduled issuance
  must fit under max.
- The supply ledger: balances, transfers, burns, fees (paid from
  existing supply, never minted), IBC escrow and voucher accounting
  (escrowed native OGO plus vouchers must not double-count).
- Scheduled issuance only: `MintScheduled` (block-based, legacy
  path) and `MintScheduledForEpoch` (epoch-based path). Both share
  one issued-by-period set, so a period can never be claimed twice
  no matter which path claims it first. Every other mint fails.
- The epoch timer mirror (`x/ogocoin/epochs.go`): one timer per
  identifier, fixed grid anchored at genesis, 1-indexed epochs,
  at most one epoch tick per block, `AfterEpochEnd` /
  `BeforeEpochStart` hooks, `epoch_end` / `epoch_start` /
  `epoch_issuance` events. The monetary hook mints exactly the
  schedule amount for the ending epoch, once, into the rewards pool.
  See `EPOCHS_GAP.md`: this is a faithful mirror of the pinned SDK
  x/epochs semantics, NOT the SDK module itself.
- Bypass surfaces (tested): module-account minting, governance
  changes to max supply / minting / schedule, CosmWasm minting and
  schedule changes. 54 Milestone B tests plus 13 new epoch tests
  (8 timer and mint tests in `x/ogocoin`, 5 wired-app tests in
  `app`, including the hook-fires test).

What x/ogocoin does NOT own: accounts and signatures (x/auth),
validator bonding and staking (x/staking), governance proposals
(x/gov), reward distribution (x/distribution, x/rewards), treasury
spending (x/treasury). Those are separate modules, listed below.

## Pinned-compiled only

These standard modules are pinned to the locked Gaia v28.3.1 matrix
and COMPILED into the binary through blank imports in
`app/modules.go`, proving the matrix builds as one unit. Their
keepers are NOT wired: no depinject app, no BaseApp, no stores, no
messages, no queries. Nothing in this section is a working feature;
compilation is not wiring.

| Module | Version | Status | Notes |
|---|---|---|---|
| x/auth | SDK v0.53.4 effective (v0.53.6 required) | pinned-compiled | Accounts, signatures, sequences. The prototype's transactions are JSON and NOT signature-verified; anyone with RPC access can submit a tx naming any sender. Real auth arrives with full SDK depinject wiring (Milestone C). |
| x/bank | SDK v0.53.4 effective | pinned-compiled | Denom balances, sends, supply. In this scaffold the ogocoin ledger implements bank-like balances directly; the x/bank keeper is not wired. The `uogo` denom is proposed, 6 decimals. |
| x/staking | SDK v0.53.4 effective | pinned-compiled | Validator set, bonding, delegation, unbonding. The localnet validator set is fixed at genesis (3 validators, power 10 each); no staking transactions exist yet. Validator rewards come from existing supply only; rates TBD. |
| x/gov | SDK v0.53.4 effective | pinned-compiled | Proposals, voting, tally. The upgrade-rejection rule (no monetary change through ordinary governance) is enforced in x/ogocoin and tested; on-chain gov proposals do not exist yet. |
| x/distribution | SDK v0.53.4 effective | pinned-compiled | Validator and delegator reward distribution, community pool. Not wired; issuance lands in the rewards pool account and is not yet distributed. |
| x/fees | (no standalone SDK module) | pinned-compiled (concept only) | Fee collection in the SDK lives in the ante handler plus x/distribution. In this scaffold, fees are the ledger's `DeductFee` to the fee-collector account: a fixed fee named on the transaction, paid from existing supply. There is no gas metering (CheckTx reports a hardcoded GasWanted of 1000) and no fee market. Real fee handling arrives with the full SDK app. |
| x/wasm (wasmd) | v0.60.9, wasmvm/v2 v2.3.5 | pinned-compiled | The VM is not instantiated; live contract execution is pending. The binding surface (BankSend, BankBurn, QueryBalance, QuerySupply; no mint) is defined in x/ogocoin/bypass.go and tested. |
| ibc-go core + transfer | v10.7.0 | pinned-compiled | Escrow and voucher accounting is implemented and tested at the ledger level. Live IBC relaying is not wired. |
| ibc-go 08-wasm | v10.5.0 | pinned-compiled | Light client package compiled; live use pending. |
| x/epochs (SDK) | SDK v0.53.4 effective | NOT wired (mirror instead) | Real x/epochs needs the full depinject SDK app. The prototype runs the epoch timer mirror in x/ogocoin/epochs.go instead. See EPOCHS_GAP.md for the exact gap and the StoreUpgrade path. |

Also not wired (not even pinned): x/slashing, x/evidence,
x/upgrade, x/mint (deliberately unused: x/mint is inflation-based
and cannot express a fixed-supply scheduled issuance).

## Planned custom modules (not built yet)

These are the three custom modules from the architecture. None of
them exists in this prototype; the ownership below is the plan, so a
later milestone does not have to re-decide it. The production names
are x/monetary, x/treasury, x/rewards; the scaffold's x/ogocoin is
the prototype that x/monetary grows out of.

### x/monetary (planned; grows out of x/ogocoin)

Owns: supply cap, issuance schedule, the mint operation, supply
accounting, the issued-by-epoch set. It is the ONLY module that can
mint. It mints only exactly as the schedule dictates, through the
amount-less entrypoint (`MintScheduledForEpoch`): the function takes
an epoch number, never an amount, so no caller can request a larger
mint. It never accepts a schedule change through ordinary
governance. What remains before it is x/monetary rather than the
prototype: the epoch hook must be the real SDK x/epochs hook (see
EPOCHS_GAP.md), and the hook must call the model-aware mint
(`MintScheduledModel` with the genesis-selected model) instead of
the legacy geometric schedule, with `init` CLI flags to express a
hybrid or tapered-hybrid genesis. The hybrid and tapered-hybrid
models already exist as code (`x/ogocoin/models.go`, parallel
Milestone 3 work): the remaining work is the hook rewire, the CLI
flags, and reconciling the two arithmetic definitions (the legacy
schedule floors at every multiplication step; the models use the
exact floor(R0 * r^n) formula, which gives different amounts from
period 7 onward), plus unifying the two epoch mint entrypoints
(see EPOCHS_GAP.md). Until that is done, the app's epoch path
mints the legacy geometric schedule only, which is correct for
every genesis the app can produce today.

### x/treasury (planned; not built)

Owns: treasury accounting, the R-F spending rule, the emergency
reserve floor, and mode transition. It does NOT mint, and it does
not hold coins that are not accounted for in x/bank. It spends only
under the rule, never above it. In the prototype there is no
treasury account and no spending rule; issuance lands in the rewards
pool and stays there.

### x/rewards (planned; not built)

Owns: per-epoch and per-block split accounting for issuance and
fees (validators, delegators, community pool or treasury). It does
NOT mint and it does NOT decide the split; it executes the split
the approved parameters dictate, conserving exactly (the parts sum
to the minted amount, in integer arithmetic). In the prototype
there is no split: the full epoch amount lands in the rewards pool.

## Explicitly not created

- x/governance: x/gov already provides governance; a second
  governance module would create a divergent voting path.
- x/fees as a custom module: fee collection uses the standard ante
  handler plus x/distribution, and fee splitting is handled by
  x/rewards.
- x/bridge: bridges are third-party trust and are not part of the
  chain.
- x/mint: inflation-based issuance cannot express the fixed-supply
  schedule, so it is not used.
