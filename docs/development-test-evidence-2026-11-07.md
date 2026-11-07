# OGO Network development test evidence

All work was performed in the local clone at `/home/ubuntu/work/ogo-network-dev/repo` on branch `manus/dev-network-api`. No changes were pushed or merged.

## Chain suite

The repository-pinned Go toolchain was installed outside the repository at `/home/ubuntu/work/ogo-network-dev/tools/go`.

```text
go version go1.25.7 linux/amd64
go build ./...
go vet ./...
go test -count=1 ./...
ok   github.com/benross-lab/ogo-network/app       2.365s
ok   github.com/benross-lab/ogo-network/wasmcheck 0.039s
ok   github.com/benross-lab/ogo-network/x/monetary 0.062s
```

An initial attempt with the sandbox’s Go 1.27.3 failed during dependency compilation in `github.com/bytedance/sonic`; it was not a repository test failure. The pinned-toolchain rerun passed.

## API tests

```text
cd network-api
/usr/local/go/bin/go test ./...
ok github.com/benross-lab/ogo-network/network-api 0.006s
```

The tests cover live-state proxy shape, explicit 501 responses when LCD/indexer data is unavailable, base64 client-signed broadcast, required idempotency keys, and conflicting idempotency-key rejection.

## Real localnet smoke

A fresh four-validator localnet was initialized in `/home/ubuntu/work/ogo-network-dev/localnet4-smoke`, outside the repository. The existing `localnet4.sh` script reached height 10/11 and reported the same app hash `BF1DC06FCF1A9916` on all four nodes. The network was stopped after the smoke test.

The new API origin was pointed at `http://127.0.0.1:27657` and returned live `ogo-local-1` data for `/healthz`, `/v1/network`, and `/v1/validators`; the validator endpoint returned four validators. This proves local RPC integration only. It does not prove public-network operation, persistent hosting, or production security.

## Worker check

```text
node --check cloudflare/ogo-network-api/src/index.js
```

The check passed. The Worker was not deployed, and no DNS, route, tunnel, binding, secret, or paid Cloudflare resource was changed.

## Not tested

A 24-hour observation, independent-host failure recovery, genuine Cosmovisor upgrade and rollback, off-host snapshot restore, production monitoring/alerting, live indexer, public full node, and OGO Systems app integration remain NOT TESTED. The current workspace does not contain the OGO Systems application source, so the TypeScript wallet client is a framework-neutral contract rather than an application merge.
