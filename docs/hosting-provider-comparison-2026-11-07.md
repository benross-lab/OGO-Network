# OGO Network hosting provider comparison

**Status:** planning only. No provider account, server, snapshot, tunnel, or paid service was activated.

## Requirements

The four-validator development network needs separate validator homes and keys, persistent SSD/NVMe storage, private P2P connectivity, snapshots, restore testing, monitoring, and restart/recovery. A separate full node is required for public RPC later. The repository’s Standard planning baseline is 4+ CPU cores, 32 GB RAM, 500 GB NVMe, and 100 Mbps per validator.

## Provider comparison

### Hetzner — recommended first choice

Hetzner provides private networks, stateful firewalls, API/CLI automation, multiple European and US locations, and dedicated-vCPU cloud options. Its current public page lists dedicated-vCPU general-purpose cloud plans starting at €43.49/month excluding VAT, but the exact 32 GB / 500 GB NVMe combination is location- and plan-dependent and should be quoted before approval.[1]

The repository’s own planning model recommends Hetzner for future paid hosting and estimates roughly $200/month per Standard validator, before sentries, backups, and labor. Use three validators in one Hetzner region/provider and place the fourth validator or the full node with a second provider to avoid a single-provider failure domain.

**Lean planning option:** four small testnet-grade instances at approximately $75 each, plus one similar full node: approximately $375/month before backups and monitoring.  
**Standard planning option:** four Standard validators at approximately $200 each, plus one Standard full node: approximately $1,000/month before backups and monitoring. These are repository planning figures, not a quote.[2]

### DigitalOcean — strongest operational alternative

DigitalOcean publishes clear configurations and includes private networking, public IPv4 by default, and snapshot/backup options. Its current pricing page lists a memory-optimized 4 vCPU / 32 GiB / 100 GiB SSD Droplet at $168/month and a storage-optimized 4 vCPU / 32 GiB / 600 GiB SSD Droplet at $262/month.[3] The storage-optimized shape better matches the OGO 500 GB disk target.

Using the $262 storage-optimized configuration for four validators and one full node gives $1,310/month before backups. DigitalOcean lists percentage backups at 20% weekly or 30% daily, so the corresponding planning totals are $1,572/month weekly-backup or $1,703/month daily-backup, before additional volumes, monitoring, and taxes. Droplet snapshots are listed at $0.06/GB-month.[3]

DigitalOcean is a good choice when transparent billing, private networking, and simpler multi-region operations matter more than lowest price. Provider-level isolation still requires placing at least one validator outside DigitalOcean.

### Vultr — viable secondary provider

Vultr’s current pricing page lists a regular 8 vCPU / 32 GB / 512 GB configuration at $192/month, but that class uses shared vCPUs. Its dedicated general-purpose 8 vCPU / 32 GB plan is listed at $240/month with 160 GB storage, so it would need extra storage to meet the OGO 500 GB target.[4] Vultr lists automatic backups at 20% above the base price and stored snapshots at $0.05/GB-month.[4]

Vultr is useful as a second provider or sentry location. For the primary validator set, the storage mismatch and shared/dedicated plan distinction make it less straightforward than Hetzner or DigitalOcean.

## Recommended deployment specification

### Lean development network

- Four validators: 2+ cores, 8 GB RAM, 200 GB SSD minimum each.
- One full node: at least the same CPU/RAM, preferably 500 GB NVMe if it will retain history.
- Two optional lean sentries: 2 cores, 8 GB RAM, 200 GB SSD.
- Separate provider or region for at least one validator and the public full node.
- Off-host encrypted snapshots; never copy validator private keys into backup archives.
- Estimated infrastructure floor from the repository: $375/month for four validators plus one full node, before sentries and backup storage.

### Standard development network

- Four validators: 4+ dedicated cores, 32 GB RAM, 500 GB NVMe, 100 Mbps each.
- One public full node: 4+ dedicated cores, 32 GB RAM, 500 GB NVMe or larger; pruning policy must be decided before launch.
- Two sentries if private P2P topology is exercised.
- At least two providers and two geographic regions.
- External monitoring for block height, block time, peer count, mempool, validator signing, disk, process health, and supply invariants.
- Estimated repository planning floor: $1,000/month for four validators plus one full node, before sentries, backups, monitoring, and labor.

## Recommendation for approval

Approve a **private Standard development network**, not a public testnet, with three validators on Hetzner, one validator and the full node on DigitalOcean or Vultr, and two lean sentries only if the team wants to exercise the sentry runbook immediately. Use Cloudflare Tunnel or an equivalent private-origin design for the full node after the hosting provider is selected. Do not create DNS, tunnels, or paid resources until this topology and budget are approved.

## References

[1]: https://www.hetzner.com/cloud/ "Hetzner Cloud pricing and features"
[2]: https://raw.githubusercontent.com/benross-lab/OGO-Network/main/milestone-c/docs/ogo-network-deployment-architecture-v0.1.md "OGO Network deployment architecture and planning tiers"
[3]: https://www.digitalocean.com/pricing/droplets "DigitalOcean Droplet pricing"
[4]: https://www.vultr.com/pricing/ "Vultr pricing"
