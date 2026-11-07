# OGO Explorer and indexer plan

**Status:** design only. No explorer or indexer is deployed.

## Data boundary

The chain is authoritative. The indexer is a derived read model and must be reconciled against CometBFT and LCD state. If indexed data disagrees with the chain, the API must surface the discrepancy and prefer chain state. The explorer must not invent balances, transactions, validators, epochs, or supply.

## Pipeline

1. A dedicated indexer process connects to the non-validator full node over CometBFT RPC/WebSocket.
2. It records block headers, transactions, events, validator-set changes, epoch events, and supply-query samples.
3. It stores an append-only ingestion cursor keyed by chain ID, block height, and app hash.
4. It retries failed heights and detects gaps before advancing the published cursor.
5. A reconciliation worker compares indexed transaction results, balances, validators, and supply against independent RPC/LCD queries.
6. The API publishes only indexed history whose cursor is healthy; current balances and transaction verification can query live chain state directly.

## Initial read model

- `blocks`: height, hash, time, proposer, app hash.
- `transactions`: hash, height, index, code, gas, fee, memo, raw events, sender/recipient when decoded.
- `address_transactions`: address, transaction hash, direction, counterparty, denom, amount, fee, height, time, status.
- `validators`: operator address, consensus address, voting power, status, commission, first/last observed height.
- `epochs`: epoch number, start/end height or time, issuance/event metadata.
- `supply_snapshots`: height, time, total supply, total issued, treasury-related values when exposed by the chain.
- `indexer_health`: chain ID, last ingested height, last reconciled height, app hash, gap status, last error.

## Public API dependency map

- Explorer block and transaction pages require indexed blocks and transactions.
- Address balance pages should read live bank state, with indexed history beside them.
- Validator pages require live validator data plus historical samples.
- Epoch pages require a real epoch query/event source; until that exists the UI must say unavailable.
- Supply charts require sampled chain queries; they must not interpolate missing data as if it were observed.
- Search by transaction hash can query live RPC first and use the indexer for history.

## Failure handling

The explorer should display stale-data state when the indexer cursor is behind the chain. It should display “not available” when a query is not supported by the current node/API. It should never display a simulated balance or success status. A transaction is settled only after live chain verification returns a committed success code and expected recipient/amount/denom.

## Implementation order

1. Define the decoded event schema against a real localnet transaction.
2. Build a single-process development indexer with SQLite or another explicitly development-only store.
3. Add restart-safe cursoring and re-ingestion tests.
4. Add reconciliation against two RPC endpoints or RPC plus LCD.
5. Expose indexer history through the isolated API origin.
6. Build the explorer frontend only after the indexer has real block and transaction fixtures.
7. Move the indexer and explorer to separate persistent services for any public testnet.

## Current blockers

No live LCD, public full node, or indexer is present in this workspace. The current isolated API therefore returns an explicit unavailable response for address history and epoch queries instead of generating approximations. The explorer remains a plan until real chain data and a persistent indexer are available.
