# OGO Network: Monetary Constitution v0.1

Status: Founding draft, for review. Not implemented. Not authorized for token creation, issuance, genesis allocation, distribution, listings, testnet/mainnet, spending, publication, or deployment.
Date: 2026-11-05
Source: Benross, OGO Network Monetary Constitution v0.1

## Founding decisions (locked as direction)

1. Scheduled issuance with a fixed maximum supply.
2. Zero reserved founder allocation.

OGO Systems remains the founding technology developer, not the owner of the currency. Founders may acquire OGO later under the same publicly available conditions as everyone else, on equal terms. No special founder reserve, privileged issuance, or undisclosed advantage.

## 1. Monetary model: REVISED

Fixed supply. Scheduled issuance.

- Maximum supply: 100,000,000,000 OGO (approved 2026-11-08), mandatory genesis parameter, no default.
- Reserved founder coins: 0.
- Discretionary inflation: 0%.
- Publicly specified issuance rules: 100%.

REVISION: This replaces the earlier rule of zero inflation and minting disabled after genesis. Scheduled issuance requires a narrowly defined protocol-level issuance mechanism. What remains prohibited is discretionary or unauthorized minting.

How it works (conceptual): network launches with genesis config, validators secure it, protocol-defined rewards release new OGO on a predetermined schedule, issuance gradually decreases, eventually max supply is reached and issuance stops permanently. Bitcoin-like philosophy on predictable issuance, Proof of Stake instead of Proof of Work.

## 2. Distribution without founder allocation

- Founder: no reserved allocation.
- OGO Systems: no reserved allocation.
- Private investors: no preferential presale.
- Validators: eligible for protocol-defined rewards.
- Delegators: eligible for protocol-defined rewards.
- Community contributors: eligible for approved community programs.
- General public: can acquire OGO under public market conditions.

Transparent channels to investigate (no percentages invented):
- Validator and delegator rewards, governed by the published issuance schedule.
- Community development grants, funded through an explicitly disclosed protocol allocation.
- Public participation programs with published eligibility and anti-abuse rules.

The community treasury needs a defined funding source, not an invisible reserve.

## 3. Technical implications

Do not adopt Cosmos SDK mint module defaults (adjustable inflation, governance-controlled parameters). Custom OGO issuance module, proposed enforcement:

- Maximum supply: mandatory genesis parameter.
- Issuance schedule: deterministic and predefined.
- Issuance authority: protocol module only.
- Founder minting: prohibited.
- Administrative minting: prohibited.
- Governance minting: prohibited.
- Supply cap: enforced every block.
- Schedule modification: not available through ordinary governance.
- Cap reached: issuance permanently stops.
- Burns: reduce circulating and total outstanding supply.
- IBC: escrow and voucher accounting reconciled.

Scheduled issuance is not unrestricted minting. The protocol may create coins only when the predetermined schedule says so. No person gets a discretionary mint button.

Hard-fork note: a future coordinated hard fork could change consensus rules. No chain can honestly promise its community will never adopt different software. Protection is transparent rules, independent operators, and a publicly visible upgrade process.

## 4. Legal and regulatory implications: professional review required

Flagged by the constitution as needing deliberate review, including US exchange-listing ambition. No legal conclusion is recorded here as verified fact.

- Securities law: classification depends on actual design, distribution, promotion, and surrounding arrangements, not decentralization alone. A community-first launch with no founder reserve does not automatically exempt OGO.
- Money transmission: custodial wallets, exchanges, payment services, or transmitting funds on behalf of customers may carry money-transmitter obligations depending on activities and exceptions.
- Tax: validator rewards, staking income, acquisitions, and sales can have different tax consequences. US tax professional to design reporting/accounting before rewards/distributions begin.
- International distribution (including Nigeria, Europe): jurisdiction-specific review required.
- Exchange admission (e.g. Crypto.com, Webull): each exchange conducts its own legal, compliance, technical, and commercial review. Decentralized design does not guarantee listing.

Recommendation recorded: commission a written legal analysis before genesis distribution, not after exchange applications begin.

## 5. Articles

ARTICLE I: Monetary independence: OGO exists as an independent digital currency. Ownership, storage, transfer, and use do not require an OGO Systems account. OGO Systems has no unilateral ownership or control of the monetary network.

ARTICLE II: Fixed maximum supply: predetermined max; numerical max approved as 100,000,000,000 OGO on 2026-11-08; mandatory in genesis; no silent default; issuance stops at max; all issuance/supply data independently verifiable.

ARTICLE III: Scheduled issuance: predetermined schedule; protocol-defined mechanisms; rewards follow publicly documented consensus rules; no individual/company/founder/administrator has discretionary minting authority; ordinary governance cannot authorize additional discretionary issuance; no further issuance once max is reached.

ARTICLE IV: Founder allocation: no reserved founder allocation; no reserved OGO Systems allocation; no undisclosed insider allocation; no preferential private presale; founders may participate under publicly available rules on equal terms; all genesis distributions publicly disclosed.

ARTICLE V: Community distribution: investigate validator rewards, delegator rewards, community development programs, open participation initiatives, publicly accessible acquisition mechanisms. All reward mechanisms, eligibility, and treasury funding sources must be disclosed before genesis.

ARTICLE VI: Governance: community participation central; transparent proposal/review; validators independently participate in security; treasury transactions publicly auditable; no undisclosed unilateral control over community assets; monetary schedule changes require a separately documented constitutional process, not ordinary admin parameter changes.

ARTICLE VII: Security and supply integrity: supply cap enforced at protocol level; no hidden admin minting; no privileged founder minting; smart-contract minting restrictions; governance upgrade protections; IBC escrow/voucher reconciliation; independent supply verification; publicly reproducible builds; independent security audits.

ARTICLE VIII: Independence: network must operate independently of OGO Systems. Third-party developers, independent validators, external wallets, exchanges, and applications can participate without mandatory reliance on OGO Systems.

ARTICLE IX: Outstanding monetary parameters (require separate approval): genesis circulating supply; issuance duration; reward schedule and reduction intervals; validator/delegator reward allocation; community distribution mechanism; community treasury funding; minimum validator requirements; staking/unbonding parameters; transaction fee policy. The hard maximum supply (100,000,000,000 OGO) was approved on 2026-11-08 and is no longer outstanding.

ARTICLE X: Authorization: this constitution establishes a proposed monetary design only. It does not authorize token creation/issuance, genesis allocation, public distribution, exchange listings, public testnet/mainnet, spending/financial commitments, publication, or deployment. All implementation/launch decisions remain subject to explicit founder approval and applicable community governance.

Founding principle: OGO Coin exists for its participating network, not as a private monetary instrument of OGO Systems or its founders.

## 6. What remains to be designed

- Maximum supply: 100,000,000,000 OGO (approved 2026-11-08).
- Issuance duration and reduction schedule: TBD.
- Validator/delegator reward percentages: TBD.
- Community treasury funding: TBD.
- Genesis circulating supply: TBD.
- Legal classification and distribution review: pending professional advice.

Next stage recommendation: design the mathematical issuance schedule before selecting the actual coin supply, so different economic models can be compared without prematurely committing to an arbitrary number.

## Supersession note

Supersedes, for monetary design only:
- Implementation report v0.1 §3 and blueprint v0.1 §2 monetary spec where they state inflation 0% after genesis and minting disabled after genesis.
- GENESIS_SPEC.md and TOKENOMICS.md drafts where they state minting_enabled false after genesis.

What is NOT superseded:
- Fixed maximum supply, mandatory in genesis, no default.
- Supply cap enforced every block.
- No discretionary/admin/founder/governance minting.
- Burn is one-way, IBC escrow/voucher no double-count.
- Bypass tests, reproducible build, Acceptance Standard v0.1.
- Milestone B boundaries: local staging only, no repo changes, no commit/push/deploy/spend without separate approval.

Milestone B build must be revised to the scheduled-issuance framework (see revision notice sent to the build). The old "mint disabled after genesis" invariant tests must be updated to "mint only exactly as the predetermined schedule dictates, never otherwise", with schedule parameters TBD and no invented schedule numbers.
