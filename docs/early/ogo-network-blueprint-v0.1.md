# OGO Network: Milestone A Blueprint v0.1

Status: Design phase only. Not deployed. No spending, publication, token issuance, testnet, or mainnet authorized.
Date: 2026-11-05
Source: OGO Network Founding Architecture v0.1 / Founding Technical Charter v0.1 (Benross, 2026-11-05)
Goal: OGO AI builder and network (goal_4c003a399d37)

## 1. What this is

OGO Network is a new division of OGO Systems: open, decentralized financial and blockchain infrastructure powering the wider OGO ecosystem. OGO Systems creates the initial technology. The community ultimately governs the network.

Open to the world: people should not need an OGO account to own, transfer, or use OGO Coin. Independent wallets, exchanges, developers, and third-party services must be able to use it outside the OGO app.

OGO Credits remain separate from OGO Coin. Credits pay for internal AI services. OGO Coin is the native asset of the independent network.

## 2. Dependency assessment: checked 2026-11-05

Cosmos Labs now ships release families, not LTS. Only two families were supported at check time:

- 2025.1: Cosmos SDK 0.53.x + CometBFT 0.38.x
- 2026.1: Cosmos SDK 0.55.x + CometBFT 0.40.x

Everything below SDK 0.53 / CometBFT 0.38 is end of life. CometBFT v1.x was retracted and is not supported.

Live reference: Cosmos Hub mainnet gaia v28.3.1 (2026-10-01) runs SDK v0.53.4 (go.mod requires v0.53.6) + CometBFT v0.38.23.

CosmWasm: v3.0 shipped with wasmd 0.61, with IBCv2 support. Contracts built on 1.x / 2.x / 3.0 remain executable on a 3.0 chain unless they use new IBCv2 features. wasmd 0.60.0 upgraded to SDK v0.53.0. wasmd 0.60.1 / 0.55.1 were security patches (June 5 disclosure).

Sources checked:
- Cosmos Docs, Versions: gaia v28.3.1, SDK v0.53.4, CometBFT v0.38.23
- Cosmos Docs, Release Families: 2025.1 (0.53 / 0.38) and 2026.1 (0.55 / 0.40) only supported families
- Cosmos SDK CHANGELOG: v0.55.0 released 2026-07-27
- Cosmos Labs Ledger Security 2026.1 (2026-10-08): SDK v0.55.0 + CometBFT v0.40.0 + PoA module v1.1.0 + cosmos/kms v0.1.0, ML-DSA post-quantum signatures
- CosmWasm 3.0 announcement: IBCv2, wasmd 0.61

### Recommendation for the prototype

Start Milestone B on the proven stack, evaluate the newest family as a separate upgrade decision:

- Cosmos SDK v0.53.x
- CometBFT v0.38.x
- Go 1.23
- wasmd v0.60.x / CosmWasm 2.x for the first localnet, with CosmWasm 3.0 / wasmd 0.61 as the evaluated upgrade path
- IBC via ibc-go matching the SDK 0.53 family (pin exact version in go.mod at scaffold time, do not guess it in this doc)
- Cosmovisor v1.7.3 for upgrades (independent of SDK version)

Why not start on 0.55 / 0.40: it is the newest family, it brings a second consensus-layer change plus ML-DSA key rotation, and the Hub itself still runs 0.53 / 0.38. For a first local prototype, proven tooling and docs matter more than newest features. Revisit 0.55 once the 0.53 prototype passes monetary invariant tests.

Do not pin to SDK 0.50 or lower, CometBFT 0.37 or lower, or CometBFT v1.x. All are out of support.

## 3. Protocol specification (founding)

- Network type: Independent Layer 1, sovereign Cosmos-based chain
- Consensus: CometBFT, Byzantine fault-tolerant Proof of Stake (not Proof of Work)
- Interoperability: IBC
- Smart contracts: CosmWasm
- Native asset: OGO Coin, ticker OGO
- Maximum supply: Fixed, exact number TBD, see open decisions
- Inflation after genesis: 0%
- Additional minting after genesis: Disabled
- Transaction fees: Paid in OGO
- Validator security: Staking from existing supply
- Governance: Community, public proposals, voting periods, transparent results, timelocks for upgrades, emergency security procedures, treasury records, independent code review
- Custody: User-controlled wallets, no OGO account required to own / transfer / use OGO
- Ledger: Public transaction ledger, open-source protocol

### Denomination (proposed, not final)

- Display: OGO
- Base denom (proposed): `uogo`, 6 decimal places, following the ATOM / uatom pattern
- Chain ID (proposed for localnet only): `ogo-local-1`; mainnet chain ID TBD and must not be claimed until genesis spec is approved
- This is a naming proposal for code. It does not create a coin, a network, or a value.

### Fixed-supply consequence

If supply is truly fixed and inflation is 0%, validators cannot be paid from new minting. They must be paid from transaction fees, an explicitly allocated reserve, or a combination that is written down before genesis. No perpetual newly minted staking reward can be promised alongside zero inflation. This must be specified in TOKENOMICS.md before any monetary module is written.

### Modules

Standard Cosmos SDK modules: auth (accounts), bank, staking, gov, distribution, fees.
Custom OGO modules (proposed): x/ogocoin (fixed-supply enforcement), x/governance (OGO governance rules on top of x/gov), x/treasury (multisig + public records), x/fees.
Each custom module needs monetary invariant tests: total supply never exceeds max, mint disabled after genesis, no hidden mint path, treasury moves require multisig + public record.

## 4. How OGO connects to the existing AI ecosystem

No rebuild of OGO. The existing app becomes one of the first applications using OGO Network, through an adapter:

- Users pay for AI agents, developers receive payments, marketplace sellers receive OGO
- AI agents transact only through permissioned wallet interfaces: user authorizes a limit, agent requests a transaction, wallet verifies permission + balance, transaction is signed per that authorization, user can independently inspect it on chain
- Agents never receive unrestricted access to private keys
- OGO Credits stay internal and separate

The chain must not depend on OGO's Cloudflare Worker, Supabase database, or central app server to validate transactions. If the OGO app disappears, the network should continue operating.

## 5. Repository layout (proposed, not created)

Proposed layout from the charter, kept as a blueprint only:

```
ogo-network/
  chain/ app/ cmd/ x/ogocoin/ x/governance/ x/treasury/ x/fees/ proto/ tests/
  wallet/ web/ mobile/ sdk/
  explorer/
  developer-sdk/
  documentation/ WHITEPAPER.md TOKENOMICS.md GOVERNANCE.md SECURITY.md ARCHITECTURE.md
  infrastructure/ localnet/ monitoring/
```

Conflict to resolve before creating anything: OGO repo-first direction (2026-10-28) says all OGO collaboration work goes in the existing repo benross-lab/OGO only, no separate prototype, no duplicated app. This blueprint does not create a new repo. It lives as a doc under the existing OGO goal until Benross explicitly approves a separate ogo-network repo or a folder inside benross-lab/OGO.

No repo, branch, publish, or deploy is authorized by this blueprint.

## 6. Local development plan: Milestone A (research) then B (prototype)

Milestone A: Architecture and research (this doc):
- [x] Evaluate maintained Cosmos SDK versions (see section 2)
- [x] Pin compatible CometBFT family
- [ ] Pin exact ibc-go + wasmd versions in go.mod at scaffold time
- [ ] Define chain identity + denomination (proposed above, needs approval)
- [ ] Specify fixed-supply monetary rules (needs max supply decision)
- [ ] Write genesis allocation specification (needs allocation decision)
- [ ] Design validator + governance security (needs validator entry decision)

Milestone B: Local blockchain prototype (not started, needs explicit go-ahead):
- Initialize Go chain app from the pinned family
- Configure local CometBFT dev nodes
- Implement native OGO denomination, transfers, balances
- Enforce fixed supply in code + automated monetary invariant tests
- No public network, no real value, localnet only

Milestone C: Wallet and developer infrastructure:
- Wallet prototype, signing, explorer prototype, developer API + SDK, OGO app adapter

Milestone D: Community test infrastructure:
- Simulated independent validators, governance proposal tests, failure / recovery tests, independent security review, public testnet proposal

Later: Public mainnet only after audits, reproducible builds, multiple independent validators, governance attack simulations, treasury multisig, and legal review. None authorized now.

## 7. Security and independence requirements (before any public network)

- Independent security audits
- Publicly reproducible builds
- Multiple independent validator operators, protection against concentration
- Hardware-backed key protection where appropriate
- Wallet recovery design
- Outage and recovery testing
- Governance attack simulations
- Treasury multisignature controls + public records
- Legal review of issuance, distribution, exchange availability

Non-negotiables from the charter: no founder-controlled secret minting, no undisclosed insider allocations, no unilateral treasury withdrawals, public ledger, independent validators, open-source protocol, community-visible governance, user-controlled wallets. Community oversight needs technical enforcement, not just a promise.

## 8. Open decisions: in order

1. Exact maximum OGO Coin supply: TBD. Do not pick 21 million just because Bitcoin uses it. This blocks the monetary module, not the blueprint.
2. Genesis allocation and distribution: percentages, vesting, recipients, treasury controls must be publicly specified before genesis. Illustrative buckets only: community distribution, ecosystem development, validator / security incentives, founding development allocation.
3. Validator entry requirements + community voting mechanics: minimum stake, unbonding, slashing, delegation, voting concentration limits, protections against treasury capture. Token-weighted voting alone does not guarantee democratic governance.

Team roles for later (not hired, not authorized): protocol architect, Go blockchain engineer, security engineer, wallet engineer, infrastructure engineer, developer experience engineer. AI can assist with implementation / testing / docs; independent human security review is essential.

## 9. Authorization restrictions

This blueprint authorizes architectural planning only. It does not authorize spending funds, purchasing infrastructure, deploying nodes, publishing repos or docs, issuing or selling tokens, creating a public testnet, launching a mainnet, or entering exchange agreements. Each needs explicit approval from the founder.

Founding principle: OGO Network belongs to its participating ecosystem, not to a single company's private database or unilateral monetary authority.
