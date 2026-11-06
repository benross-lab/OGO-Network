# OGO Network API Specification

Status: Technical specification. Planning only. No APIs deployed, no SDKs published, no production change authorized.
Date: 2026-11-06
Companion document: `ogo-network-native-coin-architecture-implementation-plan-v1.0.md` (section 7)
Architecture direction: CONFIRMED AND LOCKED, Option B. OGO Network is an independent Layer 1 blockchain with native OGO Coin.

## 1. Purpose

This document specifies the integration interfaces for OGO Network: how OGO accounts, projects, and applications, AI agents, developer tools, marketplace transactions, external wallets, external chains, and (in the longer term) OGO Oasis services connect to the network. The network must be usable by external developers and services through documented APIs and SDKs, not exclusively through OGO's own interface. The blockchain is kept independent from the OGO application backend: the application talks to the network only through public RPC and API services.

## 2. Principles

1. **Public by default.** Every interface the OGO application uses is public and documented. There is no private chain connection, no private API, and no direct database access to chain state. If the OGO application can do it, an external developer can do it.
2. **The chain is the source of truth for native balances.** Supabase is the source of truth for application authentication, profiles, and application data. If the two disagree, the chain wins for native balances, and the discrepancy is logged and surfaced, not silently resolved.
3. **Read is open, write is signed.** Reading chain state (balances, supply, blocks, validators) requires no authentication. Writing chain state (a transaction) requires a signature from the account that owns the funds or the authority. There is no API key that can move funds.
4. **The Worker never signs.** The Cloudflare Worker may read the chain, verify on-chain state, and prepare unsigned transactions. It must never hold keys, sign for users, write chain state, broadcast transactions, trust client claims, write to the platform ledger as a side effect of a chain operation, act as a validator, or move treasury funds.
5. **Versioned and stable.** Public APIs are versioned. A breaking change requires a new version, not a silent change to the existing version. Deprecated versions are announced with a removal date.
6. **Documented for outsiders.** API documentation is published in OGO-Network-Docs, with example code, and is written for a developer who has never used OGO, not for the OGO team.

## 3. Core chain interfaces

### 3.1 RPC (CometBFT and Tendermint RPC)

Standard CometBFT RPC, public, CORS-enabled, documented. Endpoints include: block height and block by height; transaction by hash; broadcast transaction (returns the hash; the caller tracks commitment); validators and voting power; consensus state; and network info. The RPC is the interface for node operators, indexers, and advanced integrators. It is not the interface for ordinary users; they use the LCD or the wallet.

### 3.2 LCD (REST API, Cosmos SDK)

Standard Cosmos SDK REST API, public, CORS-enabled, documented. Endpoints include: `/cosmos/bank/v1beta1/balances/<address>` (balances); `/cosmos/bank/v1beta1/supply` (total supply); `/cosmos/staking/v1beta1/validators` (validator set); `/cosmos/staking/v1beta1/delegations/<address>` (delegations); `/cosmos/gov/v1/proposals` (governance proposals); `/cosmos/tx/v1beta1/txs/<hash>` (transaction by hash); and the x/monetary, x/treasury, and x/rewards queries (schedule, issued-by-epoch set, total issued, total scheduled, dust, max supply, treasury balance, spending rule state). The LCD is the interface for wallets, the OGO application, and developer tools.

### 3.3 Indexer API

The indexer indexes blocks, transactions, and events from the chain, and provides a paginated history API that the chain itself does not provide. Endpoints include: transaction history by address (with direction, counterparty, amount, fee, block height, timestamp, status); transaction history by block; validator history (uptime, missed blocks, commission changes); and supply history (total supply over time). The indexer is reconciled against the chain (gate G9): if the indexer and the chain disagree, the chain wins, and the discrepancy is logged and surfaced. The indexer API is public and documented.

### 3.4 WebSocket subscriptions (optional, later phase)

A WebSocket interface for subscribing to new blocks, new transactions for an address, and governance events. This is a later-phase addition, not required for the first monetary prototype. When added, it must be documented with the same versioning rules as the REST APIs.

## 4. OGO accounts and wallet linking

- **Wallet linking:** the Worker issues a nonce (random, single-use, 10-minute expiry, stored in Supabase with the user id). The user's wallet signs a message containing the nonce, the OGO account identifier, the wallet address, and the chain id. The Worker verifies the signature. If verification passes, the link is recorded in Supabase (user id, wallet address, chain id, linked timestamp, nonce id). The user can unlink at any time. Unlinking does not affect the on-chain balance.
- **What linking does:** it allows the OGO application to display the native balance for that address, to show transaction history, and to prepare unsigned transactions for that address.
- **What linking does not do:** it does not give the OGO application control over the wallet, the ability to sign for the user, or the ability to move funds. It does not merge the Supabase account with the blockchain balance.
- **Multiple wallets:** a user can link multiple wallet addresses to one OGO account. Each link is a separate record, verified by its own signature.

## 5. Projects and applications

- **Project registration:** a project (an OGO application, a developer's application, or an agent service) registers with the OGO application, providing a name, a description, and a wallet address for receiving payments. Registration is an application-level record in Supabase, not a chain transaction.
- **Project payments:** a project receives native OGO Coin payments through the payment request flow (section 8). The project does not need a chain-side registration to receive payments: any wallet address can receive.
- **Project identity on-chain:** there is no on-chain project registry in the first monetary prototype. A project's on-chain identity is its wallet address. An on-chain registry is a later-phase evaluation, and it must not be confused with the application-level registration.

## 6. AI agents

- **Agent authorization:** an agent acts only through a scoped authorization, signed by the user's wallet key, with a spending limit, a per-transaction limit, an explicit approved-recipient list (no wildcard), a mandatory expiry, and a nonce. The authorization format is specified in `OGO_NATIVE_WALLET_SPEC.md` section 14. Enforcement is off-chain in the wallet in the prototype, and on-chain via x/authz or agent sub-accounts in a later phase.
- **Agent payment requests:** an agent creates a payment request (section 8) for a user to approve. The user sees the recipient, the amount, the description, and the expiry, and approves or declines. Approving does not move funds; it prepares an unsigned transaction for the user to sign in their own wallet.
- **Agent transaction API:** an agent with a valid authorization can submit a signed transaction through the public RPC broadcast endpoint, like any user. There is no separate agent API for broadcasting. The authorization is checked by the wallet before signing, not by the chain in the prototype.
- **Agent discovery:** agents are discoverable through the OGO application (profiles, capabilities, reputation), not through the chain. The chain records transactions; it does not record agent identities.

## 7. Developer tools and SDKs

- **SDKs:** client libraries for the chain (transaction construction, signing, broadcasting, querying) in the languages developers use. The first SDKs are TypeScript (for the OGO application and web developers) and Go (for backend developers and validators). Other languages follow based on demand. SDKs are published with version numbers, changelogs, and example code.
- **Chain registry metadata:** the chain is registered in the Cosmos Chain Registry (chain id, RPC and LCD endpoints, denom, decimals, bech32 prefix, derivation path, gas price), so Keplr, Leap, and Cosmostation can add it without custom code. Registry registration is required before external wallet integration is claimed as supported.
- **Testnet faucet:** a public testnet faucet dispenses TEST-ONLY tokens for development. The faucet is rate-limited, and its tokens have no value. The faucet is a testnet service, not a mainnet service.
- **Documentation:** API documentation is published in OGO-Network-Docs, with a getting-started guide, authentication (none for reads, signatures for writes), rate limits, error codes, and example code for common operations (check a balance, send a transaction, delegate, vote, create a payment request). Documentation is versioned with the APIs.

## 8. Marketplace transactions

- **Payment requests:** a payment request is a request for a user to pay a specific amount to a specific address, for a specific purpose. It is stored in Supabase with a unique request id, a recipient address, an amount in uogo, a description, and an expiry. The payer approves or declines. Approving prepares an unsigned transaction for the payer to sign in their own wallet. The application verifies the transaction is committed, with the correct amount and recipient, before marking the request as paid. Verification is against a public RPC endpoint, not against a client claim.
- **Developer payment API:** an API for developers to accept native OGO Coin payments in their applications. Endpoints: create a payment request (recipient, amount, description, expiry; returns a request id and a payment URL); payment status (request id; returns pending, paid, expired, or declined, with the transaction hash and block height if paid); and a webhook that notifies the developer when a payment is verified as committed on-chain. The API does not hold keys, does not sign, and does not broadcast. It is rate-limited and authenticated, so a developer cannot spam payment requests or enumerate other developers' requests.
- **Settlement:** marketplace settlement is gated on verified on-chain payment. An order is not marked as paid until the payment is verified as committed on-chain. The settlement is recorded in Supabase with an idempotency key (the transaction hash or the payment request id), so a duplicate webhook or a duplicate request cannot fulfil twice. If the payment fails, expires, or is for the wrong amount, the order is not marked as paid, and the buyer and seller are both notified with the reason. There is no partial settlement.
- **Marketplace settlement uses native OGO Coin only.** A marketplace order priced in Supabase Coins is a separate flow, in the existing economy, and must not be confused with a native coin settlement.

## 9. External wallets

- External wallets (Keplr, Leap, Cosmostation, in that priority order) connect through the public RPC and LCD endpoints and the chain registry metadata. No custom integration code is required on the chain side beyond the registry entry and the CORS-enabled endpoints.
- The chain supports standard Cosmos signing (DIRECT and Amino JSON) and standard derivation, so external wallets work without OGO-specific code.
- A test transfer from the external wallet to an OGO wallet, and from an OGO wallet to the external wallet, both committing with correct balances, is the acceptance test (proof-of-concept criterion O4). If wallet integration cannot be completed, the blocker is recorded and the criterion is treated as a MUST failure until it passes. Clarification, 2026-11-06: O4 is a framework-selection compatibility check, not the wallet product; the full wallet, including Keplr, Leap, and Cosmostation integration, is built in M4 (see `OGO_NETWORK_IMPLEMENTATION_ROADMAP.md` section 3). Correction, 2026-11-06: O4 is DEFERRED by Benross's decision (Phase 2 Execution Directive). It is no longer a framework-selection gate or a MUST criterion, and its failure does not block framework selection. Keplr is evaluated separately if a specific technical dependency demonstrates necessity.

## 10. External chains

- **IBC:** the primary cross-chain interface, through ibc-go v10, with the 08-wasm client for contract-aware IBC. IBC v2 support is included. Supply rule: a native token escrowed on one chain and represented as a voucher on another must not be counted as two independent units of economic supply. Global economic supply equals local supply minus escrow plus vouchers, enforced in x/monetary. Live IBC relaying must be demonstrated in the proof of concept (criterion O3). Clarification, 2026-11-06: O3 is a framework-selection compatibility check (live relaying between two local chains, with exact escrow release). Live IBC to real external chains is later work, from the M7 testnet onward, not M0 to M3 product scope. Correction, 2026-11-06: O3 is DEFERRED by Benross's decision (Phase 2 Execution Directive). It is no longer a framework-selection gate or a prerequisite for the initial monetary prototype. IBC is evaluated separately if a specific technical dependency demonstrates necessity. The supply rule above is unchanged.
- **Relayers:** IBC relaying is performed by independent relayers, not by OGO Systems alone. Relayer operation is documented, and the relayer set is open.
- **Non-IBC chains:** no EVM bridge is part of the chain. Any bridge to a non-IBC chain is third-party trust and must be evaluated under the same supply-accounting rules as IBC. Bridges are the most exploited component class in the industry (over 2.8 billion USD lost since 2022). No bridge is adopted by these specifications.

## 11. OGO Oasis services (longer term)

OGO Oasis (the longer-term Africa telecom direction: OasisNet, OasisTV, OasisPhone) is not part of the first monetary prototype, and no Oasis service is specified here. When Oasis services are designed, they will connect to OGO Network through the same public APIs as any other external service: payment requests for billing, wallet linking for user accounts, and the indexer API for history. Oasis services must not receive a private chain connection, a private API, or a privileged module account. If an Oasis service needs a chain-side feature (for example, a telecom-specific module), that feature is specified, reviewed, and approved as a protocol change, not added as a back door.

## 12. Rate limits, errors, and versioning

- **Rate limits:** public read APIs are rate-limited per IP and per API key (where keys are used), with limits published in the documentation. Rate limiting protects the endpoints; it does not gate access to chain state, which is public.
- **Errors:** errors use standard HTTP status codes and a documented error body (code, message, and, where applicable, the transaction hash or block height). A failed transaction returns the reason, not a generic error.
- **Versioning:** APIs are versioned (for example, `/v1/`). A breaking change requires a new version. Deprecated versions are announced with a removal date at least 90 days in the future. The chain's RPC and LCD follow the Cosmos SDK versioning, and the OGO APIs (payment requests, developer payment API, indexer API) follow semantic versioning.

## 13. What this document does not decide

The wallet technology stack and framework (decided in M4); the indexer technology (decided in M6); the explorer technology (decided in M6); the SDK languages beyond TypeScript and Go; the rate limit values; and the Oasis service designs. All are listed in `OGO_NETWORK_DECISION_REGISTER.md` where they are decisions, or deferred to the milestone where they are decided.
