# OGO Network free-first infrastructure strategy

**Scope:** development only. No server, subscription, paid Cloudflare feature, DNS record, tunnel, or production deployment was activated.

## $0 setup available now

The current sandbox is enough for local development: 6 CPUs, 23 GiB RAM, and 34 GiB free disk were available during this assessment. The isolated four-validator localnet has already run successfully on one machine, reached live block production, agreed on the same app hash, and served live data through the new OGO Network API.

This setup is not independent hosting. All four validators share one kernel, filesystem, process supervisor, network interface, and failure domain. It cannot prove provider isolation, 24-hour availability, public RPC, or production recovery.

The authenticated Cloudflare account has a Free Website plan for `ogosystems.com`. Existing `ogo` and `ogo-chat` Workers remain in place and were not changed. No Cloudflare Tunnel exists, R2 is not enabled, and no compute service is attached to Cloudflare. Free Workers/Pages resources can host an edge proxy or static developer UI, but they cannot run persistent CometBFT validators or provide durable blockchain storage. The existing Cloudflare resources must remain isolated from `ogo-network-api` until separately approved.

## Exact local instructions

From the repository branch:

```sh
cd /home/ubuntu/work/ogo-network-dev/repo
export PATH=/home/ubuntu/work/ogo-network-dev/tools/go/bin:$PATH
export GOTOOLCHAIN=local

# Build, initialize, start, and smoke-test a separate four-validator devnet.
bash dev/ogo-local-dev.sh test

# Inspect live local chain/API state.
bash dev/ogo-local-dev.sh status
bash dev/ogo-local-dev.sh verify

# Stop nodes and the API.
bash dev/ogo-local-dev.sh stop

# Create a filesystem snapshot while the nodes are stopped.
bash dev/ogo-local-dev.sh snapshot

# Restore a snapshot and verify the nodes rejoin locally.
bash dev/ogo-local-dev.sh restore /home/ubuntu/work/ogo-network-dev/localnet4-snapshots/localnet4-<timestamp>.tar.gz

# Remove only the isolated development data.
bash dev/ogo-local-dev.sh clean
```

The helper defaults to `/home/ubuntu/work/ogo-network-dev/localnet4-dev` and never uses the preserved `poc-4validator` or prior evidence directories. It keeps API broadcast disabled by default. To test a client-signed transaction against the local chain, explicitly set `OGO_API_ALLOW_BROADCAST=true`; no private key or seed phrase is accepted by the API.

## Cheapest viable remote options

Oracle Cloud’s Always Free Ampere A1 allocation is the only researched option that can plausibly provide a persistent remote development node at $0/month. The free allowance is 2 Arm OCPUs, 12 GiB total RAM, and up to 200 GB combined boot/block storage in the home region, with private VCN networking and a published outbound allowance. It can support one credible development node, or a very small two-node lab if resources are split. It is not a production network: capacity is not guaranteed, Arm compatibility must be tested, idle resources may be reclaimed, and free-tier capacity can be unavailable.

The cheapest conventional paid options are approximately:

| Provider | One plausible remote dev node | Small 3-validator devnet | Four validators + full node | Development caveat |
| --- | ---: | ---: | ---: | --- |
| Oracle Cloud Always Free A1 | $0 | $0 for a constrained 2-node lab; not a full 4-validator set | Not viable at $0 | Arm, capacity limits, reclaim risk, one-region tenancy |
| Hetzner CX23 | €6.49/month with IPv4, excluding VAT | about €19.47/month with three nodes | about €32.45/month before optional backups, excluding VAT | Shared CPU, 4 GiB RAM/40 GB NVMe; current page indicated possible ordering unavailability |
| Akamai/Linode Nanode | $5/month | about $15/month | about $25/month | Shared CPU and only 1 GiB RAM/25 GB SSD; 2 GiB nodes are safer at about $36/month for three |
| Vultr Regular Performance | $10/month | about $30/month | about $50/month | Shared CPU, 1 vCPU/2 GiB/55 GB SSD; backups add 20% |
| AWS Lightsail | about $7.80/month with an 8 GB attached disk | about $23.40/month | about $39/month | 1 GiB RAM is tight; transfer is bundled and snapshots cost extra |
| DigitalOcean Basic | $12/month | about $36/month | about $60/month | Shared CPU, 2 GiB RAM/50 GB SSD; backups add 20–30% |

These are planning estimates from current official pricing pages and are not quotes. Taxes, regional availability, public IPv4, overage, and account-specific credits can change the totals. Free promotional credits are not a permanent free tier and should not be treated as guaranteed budget.

For one remote development node, the practical shortlist is Oracle Always Free first, then Hetzner CX23 if available, then Linode Nanode only for a very small/light node. A Cosmos SDK node with indexing, builds, or monitoring should use at least 2 GiB RAM even when the absolute price floor is lower.

## Gradual deployment model

### Stage A — $0 local development

Use the current local four-validator network and API. This enables chain/API contract work, wallet signing tests, transaction verification, local snapshots, restore drills, and explorer/indexer development against real local state. It does not enable independent failure domains, public connectivity, persistent uptime, or a public testnet.

### Stage B — one inexpensive remote node

Use Oracle Always Free A1 if capacity is available. Otherwise choose one small Hetzner, Linode, Vultr, Lightsail, or DigitalOcean VM after approval. Run one non-validator full/development node and the API origin, with private networking where available. This enables remote connectivity, API-origin testing, firewall rules, and state-sync experiments. It does not prove validator quorum, provider isolation, or public-network readiness.

### Stage C — small multi-host development network

Use three separate low-cost VMs, one validator per VM, in one provider region and private network. A $0 Oracle design is constrained to a two-node lab and should not be represented as a four-validator network. The cheapest conventional three-node estimates are approximately €19.47/month on Hetzner, $15/month on Linode Nanodes, $30/month on Vultr, $23.40/month on Lightsail, or $36/month on DigitalOcean. This stage enables multi-process failure and recovery tests, but still has provider- and region-level correlated failure risk.

### Stage D — four independent validators plus full node

Only when the operational gates require it, use four separate validator hosts and one separate public full node. Place at least one validator and the full node outside the primary provider or region. Add tested backups, monitoring, key isolation, sentry topology, upgrade/rollback procedures, and an incident runbook. This is the first stage that can be shaped toward independent infrastructure, but it still requires explicit approval and must not be called production-ready until the acceptance matrix passes.

## Cloudflare position

Continue preparing the dedicated `ogo-network-api` Worker as an isolated, non-deployed artifact. The free Cloudflare Website plan and existing Workers can support edge routing or static assets within their limits, but they do not replace a persistent chain node. Do not create `api.ogosystems.com`, `rpc.ogosystems.com`, tunnels, R2 buckets, paid Workers, Load Balancing, or production routes without explicit approval.

## Sources

[1]: https://docs.oracle.com/iaas/Content/FreeTier/freetier_topic-Always_Free_Resources.htm "Oracle Cloud Always Free resources"
[2]: https://www.oracle.com/cloud/free/ "Oracle Cloud Free Tier"
[3]: https://www.hetzner.com/cloud/cost-optimized/ "Hetzner cost-optimized cloud pricing"
[4]: https://www.akamai.com/cloud/pricing/cpu "Akamai Cloud CPU pricing"
[5]: https://www.vultr.com/pricing/ "Vultr pricing"
[6]: https://aws.amazon.com/lightsail/pricing/ "AWS Lightsail pricing"
[7]: https://www.digitalocean.com/pricing/droplets "DigitalOcean Droplet pricing"
[8]: https://docs.oracle.com/iaas/Content/Network/Concepts/overview.htm "Oracle Cloud networking overview"
[9]: https://docs.hetzner.com/cloud/networks/overview/ "Hetzner Cloud Networks"
[10]: https://docs.digitalocean.com/products/networking/vpc/ "DigitalOcean VPC networking"
