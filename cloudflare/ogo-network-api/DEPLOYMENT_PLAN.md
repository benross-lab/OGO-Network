# `ogo-network-api` Cloudflare Worker deployment plan

This directory is preparation only. It contains no production account ID, route, DNS record, tunnel, secret, or deployment.

## Local development

Run the Go origin first:

```sh
cd network-api
OGO_API_RPC_URL=http://127.0.0.1:26657 \
OGO_API_ALLOW_BROADCAST=false \
/usr/local/go/bin/go run .
```

Run the Worker locally only after setting `ORIGIN_BASE_URL` to the local origin. The example config is intentionally named `wrangler.example.toml` and should not be copied into a production Cloudflare project without approval.

## Production approval gates

Before any Cloudflare action, obtain approval for all of the following:

- selected persistent hosting provider and origin firewall design;
- full-node RPC method allowlist and public/private boundary;
- API contract, rate limits, CORS origins, request size limits, and logging retention;
- tunnel versus another private-origin connection;
- DNS records for `api.ogosystems.com`, `rpc.ogosystems.com`, and later `explorer.ogosystems.com`;
- paid Workers, Tunnel, R2, Load Balancing, or monitoring features;
- production deployment and rollback procedure.

## Required production checks

- The Worker is a new script named `ogo-network-api` and does not reuse `ogo-chat`.
- No existing OGO Pages routes, Worker bindings, or secrets are edited.
- `ORIGIN_BASE_URL` points to an approved API origin, never a validator RPC or admin endpoint.
- Broadcast accepts only client-signed bytes and requires an idempotency key.
- Private keys and seed phrases are not included in Worker variables, logs, requests, or responses.
- The origin is reachable only through the approved private-origin path and firewall policy.
- Health checks test chain status and API origin health without exposing credentials.
- Deployment is first made to an isolated development environment, then reviewed before production publication.
