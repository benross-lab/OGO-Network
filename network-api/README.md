# OGO Network API (development origin)

This package is an isolated development API origin for the OGO Network. It proxies real chain state from a configured CometBFT RPC, Cosmos SDK LCD, and optional indexer. It never fabricates balances, transaction history, validator data, or epoch data.

It does not replace `ogo-chat`, does not contain wallet keys, and does not issue OGO Coin. It is not deployed and is not a public-network service.

## Run

From this directory:

```sh
export OGO_API_RPC_URL=http://127.0.0.1:26657
# Set this only when a real LCD is running; otherwise balance/epoch calls return 501.
export OGO_API_LCD_URL=http://127.0.0.1:1317
# Set this only when a real indexer is running; otherwise history returns 501.
export OGO_API_INDEXER_URL=http://127.0.0.1:8788
export OGO_API_LISTEN=127.0.0.1:8787
# Keep false unless testing client-signed broadcast against a local development node.
export OGO_API_ALLOW_BROADCAST=false
/usr/local/go/bin/go run .
```

## Contract

- `GET /healthz`
- `GET /v1/network`
- `GET /v1/balances/{address}`
- `GET /v1/transactions/{hash}`
- `GET /v1/addresses/{address}/transactions`
- `GET /v1/validators`
- `GET /v1/epochs`
- `POST /v1/transactions/prepare`
- `POST /v1/transactions/broadcast`

`POST /v1/transactions/prepare` returns an unsigned signing request only. A wallet constructs and signs the chain transaction client-side. `POST /v1/transactions/broadcast` accepts a base64-encoded transaction that was already signed by the client. It requires an `Idempotency-Key`, rejects conflicting reuse, and is disabled unless explicitly enabled in development.

The service does not persist keys or seed phrases. The in-memory idempotency map is intentionally development-only; a production deployment needs a durable idempotency store with a documented retention policy.

## Tests

```sh
/usr/local/go/bin/go test ./...
```

The tests use a fake upstream only to verify proxy behavior. They do not claim that a public network exists.
