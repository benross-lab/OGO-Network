# OGO Network API v1 contract

Status: isolated development implementation. No public endpoint exists.

## Principles

The API reads native OGO state from the chain or a chain-reconciled indexer. It never reads native balances from the application Coins/Credits ledger. Reads are unauthenticated in development; writes require a client-signed transaction. The API does not hold keys, sign, mint, allocate, or alter monetary parameters.

## Endpoints

| Method | Path | Source | Development behavior |
| --- | --- | --- | --- |
| GET | `/v1/network` | CometBFT `/status`, `/net_info` | Implemented |
| GET | `/v1/balances/{address}` | Cosmos LCD bank query | Implemented when `OGO_API_LCD_URL` is configured; otherwise explicit 501 |
| GET | `/v1/transactions/{hash}` | CometBFT `/tx` | Implemented |
| GET | `/v1/addresses/{address}/transactions` | Separate indexer | Implemented when `OGO_API_INDEXER_URL` is configured; otherwise explicit 501 |
| GET | `/v1/validators` | CometBFT `/validators` | Implemented |
| GET | `/v1/epochs` | Configured LCD epoch path | Implemented when LCD is configured; otherwise explicit 501 |
| POST | `/v1/transactions/prepare` | Chain status + client inputs | Returns unsigned signing request; no signature is generated |
| POST | `/v1/transactions/broadcast` | CometBFT `/broadcast_tx_sync` | Requires client-signed base64 bytes, idempotency key, and explicit development enablement |

## Settlement verification

A caller must query `/v1/transactions/{hash}` after broadcast and verify the committed transaction, code, recipient, amount, denom, and block height before marking an application payment successful. A client claim or broadcast response alone is not settlement proof.

## Wallet linking and agent permissions

Wallet linking is a separate application flow: the server issues a short-lived nonce, the wallet signs a challenge containing the nonce, account identifier, wallet address, and chain ID, and the server verifies the signature. Private keys and seed phrases never cross the client boundary.

Agent spending permissions remain wallet-side in the prototype and must include a spending limit, per-transaction limit, explicit recipients, expiry, and nonce. This API does not grant an agent authority to sign or move funds.
