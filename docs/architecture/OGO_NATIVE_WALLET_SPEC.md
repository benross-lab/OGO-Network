# OGO Native Wallet Specification

Status: Technical specification. Planning only. No wallet released, no keys generated, no production change authorized.
Date: 2026-11-06
Companion document: `ogo-network-native-coin-architecture-implementation-plan-v1.0.md` (section 6)
Architecture direction: CONFIRMED AND LOCKED, Option B. OGO Network is an independent Layer 1 blockchain with native OGO Coin.
Address format: PROPOSED, not decided. Bech32 prefix `ogo` and derivation path `44'/118'/0'/0/0` are proposals.

## 1. Purpose

This document specifies the native OGO wallet: a noncustodial wallet for OGO Coin that lets users hold, send, receive, stake, and delegate native OGO Coin, verify balances and history from chain state, recover from a mnemonic, and authorize agents within strict limits. It is designed so OGO users can interact with the native network without confusing OGO Coin with existing OGO platform credits.

## 2. Principles

Five principles govern the wallet. They follow from the founding architecture: OGO owns its monetary infrastructure, while users own their coins and private keys.

1. **Noncustodial by construction.** Users retain control of private keys. OGO must not store seed phrases or unrestricted private keys on OGO servers, in Supabase, in the Cloudflare Worker, or in any log. OGO must not be able to move a user's coins.
2. **An address is not a username.** A wallet address is a cryptographic identifier derived from a public key. An OGO username is an application-level profile. The two are linked only by an explicit, revocable user action (signature verification), never by default.
3. **A Supabase account does not own a blockchain balance.** Supabase is responsible for application authentication, profiles, and application data. The blockchain is the source of truth for native coin balances. If the Supabase account is deleted, the on-chain balance is unaffected. If the OGO website is unavailable, the user can still access coins through any compatible wallet.
4. **Signing happens only in the user's wallet.** The OGO application, the Cloudflare Worker, and any OGO server never see a private key and never sign on a user's behalf.
5. **Agents never get unrestricted wallet access.** An agent acts only through a scoped authorization with spending limits, approved recipients, expiry, and revocation.

## 3. Native OGO Coin balances

- Balances are read from chain state only, through a public RPC or LCD endpoint, or through the indexer. The balance is the x/bank balance for the `uogo` denom at that address. No other source is authoritative.
- The wallet queries at least two independent RPC endpoints, or one RPC endpoint plus the indexer, and flags a discrepancy rather than silently displaying one value. This is the defence against a single compromised RPC endpoint.
- The wallet displays the total supply and the user's share of it, read from the chain, so the user can see their balance in context.
- A balance shown from a cache is labelled as cached, with the time it was last verified. A cached value is never presented as live.
- The wallet displays native OGO Coin in its own section, labelled with its source ("OGO Network, on-chain"). It never displays a combined balance with OGO Coins or OGO Credits. It never displays a Supabase balance as if it were a native balance.

## 4. Sending and receiving

### 4.1 Send

1. The user enters a recipient address and an amount. The wallet validates the address (correct prefix, correct length, valid bech32 checksum) before the transaction is constructed, not after it is signed.
2. The wallet checks the balance, including the fee, and displays the fee and the total before the user confirms.
3. The wallet does not allow sending more than the available balance minus the fee.
4. The wallet does not allow sending to the user's own address without a confirmation (to prevent accidental self-sends that still cost a fee).
5. After broadcast, the wallet displays the transaction hash and a link to the explorer, and tracks the transaction until it is committed or fails. A failed transaction is displayed with the reason, not hidden.

### 4.2 Receive

- The wallet displays the user's address with a copy button and a QR code. The QR code encodes the address and the chain id, so a sender cannot accidentally send to the right address on the wrong chain.
- Receiving requires no action from the recipient beyond sharing the address. There is no "accept" step for an incoming transfer.
- The wallet must not display a balance for an address that has not been verified against the chain.

## 5. Transaction history

- History is read from the chain (via the indexer) and from the user's local transaction log. The chain is authoritative. The local log is a cache.
- Each entry shows the transaction hash, the direction (sent or received), the counterparty address, the amount, the fee, the block height, the timestamp, and the status (committed, failed, pending).
- A transaction is not displayed as completed until it is committed in a block. A transaction that is only in the mempool is pending. A pending transaction that never commits is labelled as such, not as completed.
- The wallet links every history entry to the explorer, so the user can verify it independently.

## 6. Staking and delegation

- The wallet supports staking through x/staking: delegate to a validator, undelegate, redelegate, and claim rewards (through x/distribution).
- The wallet displays, for each validator: commission rate, voting power, uptime, and the delegator's share of rewards. The wallet does not recommend a validator. A validator's commission, uptime, and voting power are facts, not endorsements.
- The wallet warns the user before delegation: bonded stake can be slashed (downtime and double-sign), unbonding takes 21 days (proposal), and during unbonding funds earn no rewards and cannot be transferred.
- The wallet displays the unbonding queue: what is unbonding, how much, and when it completes.
- Liquid staking is not supported in the first wallet. It is an open question for a later phase.

## 7. Network fees

- The wallet displays the fee for a transaction before the user confirms, in OGO and in uogo, with the fee rate used.
- The wallet explains the fee in plain language: what it pays for (validator compensation and treasury allocation, per the approved fee distribution), and that it is paid in OGO Coin.
- The wallet does not hide the fee, bundle it invisibly into the amount, or present a transaction as free when it is not.
- The fee model itself is specified in `OGO_FEE_MARKET_SPEC.md`. The wallet implements the fee display; it does not decide the fee policy.

## 8. Address management

- Addresses are bech32, with the proposed prefix `ogo` (provisional) and the proposed derivation path `44'/118'/0'/0/0` (provisional, the Cosmos standard coin type 118). Both must be fixed before the first wallet is released, because addresses are permanent once funded and the derivation path determines recovery.
- The address is derived from the public key using the standard Cosmos address derivation (SHA256 then RIPEMD160, then bech32 encoding). The derivation is deterministic: the same mnemonic and derivation path always produce the same address.
- The wallet supports multiple accounts from one mnemonic (different derivation indexes), each with its own address, and lets the user label them.
- Validator operator addresses use the `ogovaloper` prefix and consensus addresses use `ogovalcons`, following the standard Cosmos prefix pattern with the `ogo` prefix.
- The wallet validates any address the user enters, and rejects an invalid address before constructing a transaction.

## 9. Transaction signing

- Signing uses Cosmos DIRECT sign mode (protobuf) by default, with Amino JSON as the legacy fallback for hardware wallets that require it. Both modes are tested against the chain.
- The wallet constructs the transaction, displays the full details to the user (recipient, amount, fee, memo, chain id, sequence), and signs only after the user confirms. The user must be able to see the exact amount and recipient before signing. There is no silent signing.
- The wallet verifies the chain id before signing, to prevent a transaction intended for one chain from being replayed on another. The chain id is displayed in the signing prompt.
- The wallet verifies the sequence number and the account number against the chain before signing, and rejects a transaction if they do not match.
- The wallet must never sign a transaction the user did not see, and must never sign a blank or partially specified transaction. This is the defence against a compromised application interface presenting one transaction and signing another.

## 10. Wallet creation and recovery

### 10.1 Creation

- A new wallet is created from a BIP39 mnemonic. Prototype: 12 words. Optional: 24 words for users who want higher entropy. The mnemonic is generated on the user's device, never on a server, and never transmitted.
- The mnemonic is displayed once, with a clear instruction to write it down on paper and store it offline. The wallet does not store the mnemonic in plaintext, does not upload it, and does not email it.
- The user confirms the mnemonic by re-entering a subset of words, to prove they recorded it. Creation is not complete until the confirmation passes.
- An optional passphrase (BIP39 passphrase, the 13th or 25th word) is supported. The passphrase is never stored. If the user forgets the passphrase, the wallet cannot be recovered from the mnemonic alone, and the wallet must say so plainly before the user sets one.

### 10.2 Recovery

- Recovery is by mnemonic only. The user enters the mnemonic, the wallet derives the keys, and the balance is read from the chain. No account, no email, no password is needed. This is what "access coins even if OGO's website is unavailable" means in practice.
- The wallet supports recovery from a mnemonic generated by another compatible wallet (Keplr, Leap, Cosmostation), and generates mnemonics those wallets can import. Interoperability is tested, not assumed.
- There is no password reset for a noncustodial wallet. If the mnemonic is lost and no backup exists, the coins are unrecoverable. The wallet must say this plainly at creation, not after a loss.

## 11. Private-key security

- Keys are derived from the mnemonic using the standard Cosmos derivation path (proposed `44'/118'/0'/0/0`, provisional). The derivation path must be fixed before the first wallet is released.
- Local storage: the encrypted seed is stored using Argon2id encryption, with the encryption key held in the platform keystore (iOS Keychain, Android Keystore, or the operating system keychain on desktop). The seed is never stored in plaintext, never stored in localStorage as plaintext, and never logged.
- The wallet is unlocked by a user PIN or biometric, which releases the key from the platform keystore. The PIN is not a recovery mechanism. If the PIN is forgotten, the wallet is recovered from the mnemonic, not from the PIN.
- The signing key is held in memory only while the wallet is unlocked, and is cleared on lock, on timeout, and on app backgrounding after a short interval. The timeout is a user setting with a secure default (for example, 5 minutes).
- Any code path that would transmit a seed or private key to a server is a security bug. The wallet is tested for it by inspecting network traffic and logs during wallet operations (gate G7).

## 12. External wallet integrations

External wallets are a founding requirement, not a nice-to-have. A user must be able to hold OGO Coin without an OGO account and without the OGO wallet.

Priority order: (1) Keplr, the standard interchain wallet, with the largest Cosmos user base, staking and governance support, and IBC transfer support. (2) Leap, the second standard interchain wallet, the fallback if Keplr integration is blocked, and the second wallet to support in any case, so users are not dependent on a single wallet vendor. (3) Cosmostation, the third standard interchain wallet, with mobile and desktop support.

Integration requirements, all of which must be true before an external wallet is claimed as supported: chain registry metadata is registered and correct; the chain's RPC and LCD endpoints are public, CORS-enabled, and documented; the bech32 prefix and derivation path are fixed and match the registry; a test transfer from the external wallet to an OGO wallet, and from an OGO wallet to the external wallet, both commit with correct balances (proof-of-concept criterion O4); staking and governance voting work from the external wallet, or are documented as not yet supported, with a date.

Hardware wallet roadmap, staged: Stage 1, software only, with an abstract signing seam. Stage 2, Ledger via the Cosmos app. Stage 3, Trezor and Keystone. Stage 4, multisig and air-gapped signing. No stage is claimed as supported until its entry criterion passes (a test transfer signed on the device commits with correct balances). Hardware wallet support is not a dependency of the first monetary prototype.

## 13. Verifying wallet ownership through cryptographic signatures

Wallet ownership is verified by signature, never by a form field and never by a client claim.

- **Wallet linking:** to link a wallet address to an OGO account, the Worker issues a nonce (random, single-use, 10-minute expiry, stored in Supabase with the user id). The user's wallet signs a message containing the nonce, the OGO account identifier, the wallet address, and the chain id. The Worker verifies the signature against the nonce, the account, the address, and the chain id. If verification passes, the link is recorded. If it fails, no link is recorded. A link that was not verified by signature is not a link.
- **Payment requests:** a payment request is verified by checking the on-chain transaction (correct amount, correct recipient, committed in a block) against a public RPC endpoint, not by trusting a client claim that a payment was made.
- **Agent authorization:** an agent's authority is verified by checking the user-signed authorization document (section 14) against the user's wallet key.
- **Unlinking:** the user can unlink at any time. Unlinking removes the link from Supabase and does not affect the on-chain balance.

## 14. Agent-authorized transactions (future)

Agents must never get unrestricted wallet access. An agent acts only through a scoped authorization, which is a separate mechanism from the wallet itself.

Authorization format (user-signed document, not a smart contract in the prototype):

- Spending limit: the total amount the agent may spend under this authorization, in uogo.
- Per-transaction limit: the maximum amount for a single transaction under this authorization.
- Approved recipients: an explicit list of recipient addresses. No wildcard. An empty list means no recipients are approved, not all recipients.
- Expiry: a mandatory expiry time. An authorization without an expiry is invalid. The maximum expiry is a user setting with a secure default (for example, 24 hours).
- Nonce: a unique nonce per authorization, to prevent replay.
- The authorization is signed by the user's wallet key. The signature covers all fields, so no field can be changed after signing without invalidating the signature.

Enforcement: in the prototype, enforcement is off-chain, in the wallet. The wallet checks every agent-initiated transaction against the authorization (amount within the spending limit and the per-transaction limit, recipient on the approved list, expiry not passed, nonce not used). If any check fails, the wallet refuses to sign. The wallet logs every agent-initiated transaction, whether signed or refused, with the reason. On-chain enforcement is a later phase, via x/authz or agent sub-accounts. The authorization format is designed so it can be moved on-chain without changing the fields.

Revocation is immediate and in-wallet. The user can revoke an authorization at any time, without needing the agent's cooperation. Revocation takes effect immediately for all future transactions. A revoked authorization cannot be used again, even if it has not expired. The wallet displays all active authorizations and their remaining limits.

Defaults (secure, not convenient): no agent has any authorization by default; the default spending limit is zero until the user sets one; the default per-transaction limit is zero; the default recipient list is empty; the default expiry is 24 hours; an authorization for an agent the user has not explicitly approved is invalid, even if correctly signed by someone else.

## 15. OGO Coin vs OGO platform credits: no confusion by design

The wallet is designed so a user cannot confuse native OGO Coin with OGO Coins or OGO Credits. The rules:

- **Separate sections, separate labels.** The OGO application displays three separate sections: OGO Coins (Supabase), OGO Credits (Supabase), and native OGO Coin (on-chain). Each is labelled with its source. No combined balance is ever shown.
- **Separate actions.** Sending native OGO Coin and spending platform credits are different actions, in different screens, with different confirmations. A user cannot spend platform credits as if they were native OGO Coin, or send native OGO Coin as if it were a platform credit.
- **Separate ledgers, no conversion.** The Supabase OGO Coins/Credits wallet is a separate platform accounting system. It is not merged into the blockchain ledger, and its balances are not silently converted into native OGO Coin. Any future conversion mechanism is a separate decision requiring legal review and explicit approval. Until then, there is no conversion path in the wallet, in the application, or in the Worker.
- **Plain language.** The wallet and the application explain the difference in plain language at the point where a user first sees both: OGO Coins and Credits are platform balances for OGO services, held in OGO's database; OGO Coin is a native blockchain asset, held in the user's wallet, controlled by the user's keys, and usable without an OGO account.
- **Recovery difference.** If a user loses their platform account password, the account can be recovered. If a user loses their wallet mnemonic, the coins cannot be recovered. The wallet states this difference at creation, not after a loss.

## 16. Prototype scope vs later stages

Prototype (M0 to M3): software wallet only, with the abstract signing seam; mnemonic creation and recovery, Argon2id local storage, platform keystore; DIRECT and Amino JSON signing, bech32 `ogo` prefix (provisional), derivation path `44'/118'/0'/0/0` (provisional); send, receive, balance verification from chain state, transaction history from the indexer; agent authorization as a user-signed document, enforced off-chain in the wallet, with immediate revocation; Keplr integration, or a documented blocker, as a MUST criterion (O4). Clarification, 2026-11-06: O4 in M0 to M3 is a framework-selection compatibility smoke test only (chain-registry metadata plus one signed transfer from Keplr or Leap that commits with correct balances). It verifies the stack supports external wallets; it is not the wallet product, and its failure blocks framework selection, not the monetary prototype. The full wallet product, including Keplr, Leap, and Cosmostation integration, is built in M4. Correction, 2026-11-06: O4 is DEFERRED by Benross's decision (Phase 2 Execution Directive). It is no longer a framework-selection gate or a MUST criterion for M0 to M3, and its failure does not block framework selection. Keplr is evaluated separately if a specific technical dependency demonstrates necessity. The full wallet product is built in M4, where O4 is re-tested at product level.

Later stages (M4 to M8): Leap and Cosmostation integration; Ledger (Stage 2), then Trezor and Keystone (Stage 3), then multisig and air-gapped signing (Stage 4); on-chain agent authorization via x/authz or agent sub-accounts; WalletConnect or equivalent for dApp connections, if adopted, with the same signing-prompt rules as the wallet itself.

## 17. Wallet verification checklist

The wallet is not claimed as working until every item passes:

1. A mnemonic created in the OGO wallet can be imported into Keplr and recovers the same address and balance, and vice versa.
2. A transaction signed in the wallet commits with the correct balances, and the fee is exactly as displayed before signing.
3. A transaction with a wrong signature, a wrong sequence, or an insufficient balance is rejected before execution.
4. The wallet displays a balance only from chain state, and flags a discrepancy between two RPC endpoints rather than silently displaying one value.
5. A transaction is not displayed as completed until it is committed in a block.
6. No seed phrase or private key is transmitted to any server, in any log, or in any error report. Tested by inspecting network traffic and logs during wallet operations.
7. An agent authorization with a spending limit, a per-transaction limit, an approved recipient list, and an expiry is enforced: a transaction above the limit, to an unapproved recipient, or after expiry is refused, and the refusal is logged.
8. Revocation takes effect immediately: after revocation, an agent-initiated transaction is refused even if the authorization has not expired.
9. The wallet works when OGO's website is unavailable: a user can recover from a mnemonic and view a balance and send a transaction using only a public RPC endpoint and an independent wallet, with no OGO server involved.
