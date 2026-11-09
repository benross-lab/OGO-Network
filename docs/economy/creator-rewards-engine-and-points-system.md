# OGO Economy: Creator Rewards Engine and Points System

**Status:** Proposed product and protocol specification  
**Date:** 2026-11-09  
**Launch phase:** OGO Points only  
**Not authorized by this document:** cryptocurrency issuance, token deployment, cash redemption, treasury spending, liquidity provisioning, chain migration, public testnet/mainnet deployment, or production financial activation.

## 1. Product objective

OGO is an AI-powered application-building, discovery, connection, and digital-commerce ecosystem. Users create applications and games, discover products made by others, collaborate on projects, and buy, sell, improve, and transfer eligible digital products.

The OGO Creator Economy should let application developers configure in-app rewards and purchases while OGO centrally governs the issuance of platform-wide OGO Points. Developers should be rewarded for genuine use and contribution, not merely for generating raw page views.

The first release must work without a live cryptocurrency, funded liquidity pool, guaranteed cash-out, or native OGO Coin activation.

## 2. Four distinct asset and balance classes

Keep these classes separate in product design, storage, APIs, accounting, and UI:

1. **App-specific points/currency:** Created by an individual app developer for that app. It is isolated to that app and follows that app's rules. It is not automatically convertible to OGO Points.
2. **OGO Points:** Platform-wide, off-chain rewards issued through OGO's centrally governed rewards ledger. Initially non-transferable and not cryptocurrency. They may be spent only on eligible OGO products or experiences. They have no promised dollar value or guaranteed redemption.
3. **OGO Credits:** Purchased with fiat through supported in-app purchases or Stripe for AI usage, builder operations, premium features, and designated services. Credits are not OGO Points, OGO Token, or OGO Coin.
4. **OGO Token and native OGO Coin:** Future cryptocurrency assets. Show as planned/disabled until separately approved, implemented, audited, and activated. Do not display a live balance or imply current issuance where none exists.

## 3. OGO Points launch model

### 3.1 Earning sources

The Rewards Engine may issue OGO Points for approved, measurable events including:

- Daily login.
- Qualifying app participation.
- Game achievements, milestones, and competitions.
- Verified project contributions and collaboration.
- App creation, publishing, and qualifying developer activity.
- Community challenges and approved campaigns.
- Other explicitly configured reward programs.

Every reward program must have an identifiable event definition, eligibility criteria, amount, budget, frequency limit, effective dates, and owner/approver.

### 3.2 Daily login progression

Implement the following initial daily login schedule:

| Consecutive qualifying login day | OGO Points |
|---|---:|
| Day 1 | 1 |
| Day 2 | 1 |
| Day 3 | 1 |
| Day 4 | 1 |
| Day 5 | 1 |
| Day 6 | 1 |
| Day 7 | 1 |
| Day 8 onward | Up to 5 per day |

Rules:
- The first seven qualifying daily logins award 1 point each.
- Beginning with qualifying login day 8, the daily reward may increase to a maximum of 5 points per day.
- The exact progression from 1 to 5 after day 7 must be configurable and documented before activation; do not invent a stepped schedule beyond the stated maximum.
- A qualifying login is at most one reward event per account per platform-defined reward day.
- Define the reward-day timezone consistently (prefer UTC for ledger computation; display local dates to users).
- Prevent repeat claims by refresh, multiple devices, API replay, or duplicate requests using an idempotent claim key.
- Specify streak behavior for missed days before launch. Until approved, do not silently assume whether a missed day resets the streak or pauses it.
- Daily login points remain subject to platform-wide issuance limits and the configured daily reward budget.
- Display the current streak/day, today's eligible reward, next reward status, and claim history.
- Do not reward background refreshes or automated requests as logins.

### 3.3 App participation rewards

OGO may reward qualifying users for entering or using an app. A basic example is one point for each unique qualifying user in a defined reward period. A developer dashboard can therefore show 1,000 qualifying users and 1,000 player points issued when the configured rule is one point per qualifying user.

However, a raw page load must not be treated as a qualifying event by default. Each app's policy must define a verified unique user, qualifying session/event, per-user frequency, reward period, and fraud checks. Repeated visits, reloads, bots, scripts, fake accounts, and event replay must not generate duplicate points.

The platform must distinguish:
- App opens / visits (analytics metric).
- Qualifying participation (reward-eligible event).
- Points authorized.
- Points actually posted to the ledger.

### 3.4 Player and developer reward allocation

Support configurable allocation between player rewards and developer incentives, within platform-defined limits.

Example only: an activity budget of 1,000 points could allocate 700 to eligible players and 300 to the developer. This 70/30 split is illustrative and must not be hard-coded as the default.

Developer rewards should be based on verified qualifying use and/or other approved criteria such as retention, purchases, quality, successful app publication, and contribution. Do not make developers' earnings a simple unbounded multiplier of raw visits.

## 4. Creator-configurable app economy

When a developer publishes an app/game, provide configuration for:

- App-specific currency name and rules.
- App-specific reward events and amounts.
- Achievements, levels, challenges, and competitions.
- Daily login or participation rewards inside the app.
- In-app purchases and premium upgrades.
- Digital items, collectibles, passes, or subscriptions where supported.
- A request to participate in OGO-wide Points reward programs.

Important separation:
- Developers may configure app-specific rewards independently.
- Only OGO's central Rewards Engine can authorize issuance of platform-wide OGO Points.
- App-specific points must never silently become OGO Points.
- A developer's app budget is an issuance ceiling, not a cash reserve or promise of monetary value.
- Developers cannot mint OGO Points or change platform-wide issuance rules themselves.

## 5. Central OGO Rewards Engine

Build the Rewards Engine as a central policy, authorization, and accounting service.

Required controls:
- Per-app reward budgets.
- Per-event reward limits.
- Daily, weekly, and monthly issuance ceilings.
- Per-user earning caps.
- Platform-wide issuance cap for each configured period.
- Approval workflow for creating or changing reward programs.
- Effective dates and versioned policy records.
- Pause/disable controls for a program or app.
- Idempotent event processing.
- Duplicate-event prevention.
- Fraud/bot screening and anomaly flags.
- Manual review and adjustment workflow with reason codes.
- Separation of duties for policy creation, approval, and ledger adjustment.
- Complete audit logs.

If a budget is exhausted, the system must stop issuing that program's rewards and clearly report that the allocation is exhausted. It must not silently over-issue or create a debt.

## 6. Developer dashboard

Provide a dedicated Creator Economy / Developer Rewards dashboard displaying, at minimum:

- Total app visitors and unique users (analytics).
- Qualifying unique users.
- Qualifying events by type and period.
- Player OGO Points authorized and posted.
- Developer OGO Points authorized and posted.
- Current app reward budget.
- Budget consumed and budget remaining.
- Rejected, duplicate, and flagged events.
- In-app purchases, gross transaction amounts, refunds, and net revenue where payment integration exists.
- Creator earnings and marketplace transaction records where supported.
- Full points and reward transaction history.

All dashboard metrics must clearly distinguish actual posted ledger balances from estimates, pending events, and analytics counts.

## 7. User-facing experience

Users should be able to:
- View OGO Points balance.
- View daily login streak and daily reward status.
- View points earned by app, game, achievement, or other event.
- View points spent and current available balance.
- See pending, approved, rejected, and reversed reward events where applicable.
- Spend points on eligible OGO products or experiences.

Use clear labels: **OGO Points are platform rewards, not cryptocurrency and not cash.** Do not display fiat-equivalent valuations.

## 8. Points ledger and accounting

Use an auditable append-only transaction/event ledger. At minimum record:
- Unique ledger/event ID.
- Account/user ID.
- App/project ID where applicable.
- Event type and source event ID.
- Reward policy ID and policy version.
- Debit/credit or issuance/spend direction.
- Points amount.
- Status (pending, approved, posted, rejected, reversed).
- Idempotency key.
- Timestamp and reward period.
- Approval identity or automated policy decision.
- Reason code and related reversal reference.

Never edit historical ledger entries in place. Correct mistakes with compensating adjustment entries that retain the original event reference and a reason.

Enforce non-negative spendable balances, atomic spending, and consistency between balances and ledger entries. Distinguish allocated budgets from awarded balances and spent balances.

## 9. Future conversion architecture — disabled at launch

Design extension points for a later:
- OGO Points-to-OGO Token reward conversion or allocation.
- OGO Token transfer/trading on an approved external chain.
- OGO Token-to-native-OGO-Coin conversion or migration.
- Native OGO Coin usage on OGO Network.

These are future proposals only. Before enabling any conversion, OGO must publish eligibility, snapshot rules, conversion ratio or pricing method, supply/allocation limits, funding/liquidity source, terms, and applicable compliance requirements. No fixed conversion ratio, cash value, redemption right, or future entitlement is promised by holding OGO Points.

Do not activate token minting, swaps, deposits, withdrawals, bridges, or cash-out features as part of the points launch.

Keep OGO Coin's previously approved 100,000,000,000 maximum-supply direction separate from OGO Points and any future OGO Token supply. This maximum-supply direction is not itself an instruction to mint or activate supply.

## 10. Security, abuse prevention, and fairness

- Authenticate reward claims and validate app/user eligibility server-side.
- Use idempotency keys and unique constraints to prevent duplicate claims.
- Rate-limit event ingestion and claims.
- Detect bot traffic, scripted sessions, account farms, collusion, and suspicious event patterns.
- Do not trust client-provided user counts or reward amounts.
- Require server-authorized reward events.
- Add per-account, per-app, device/session, and global abuse controls as appropriate.
- Maintain an appeal/review path for disputed adjustments.
- Avoid reward designs that encourage meaningless clicks or deceptive engagement.
- Publish program terms and changes clearly.

## 11. Suggested delivery phases

### Phase 1 — Specification and data model
- Finalize daily login streak reset/pause behavior and post-day-7 progression.
- Define qualifying app participation.
- Define developer allocation controls.
- Define reward-day timezone and ledger invariants.
- Review existing OGO app and backend architecture before selecting implementation points.

### Phase 2 — OGO Points MVP
- Points ledger and balances.
- Daily login claims.
- Per-app budgets and platform issuance rules.
- User points history.
- Basic creator reward configuration.
- Developer dashboard.

### Phase 3 — Creator Economy integration
- App-specific currencies.
- In-app rewards and purchase configuration.
- Achievements and qualifying participation.
- Developer incentives.
- Marketplace and game integration.

### Phase 4 — abuse testing and economic validation
- Duplicate/replay tests.
- Bot/fake-account simulations.
- Budget exhaustion tests.
- Concurrent claim and spending tests.
- Ledger reconciliation.
- Refund/reversal tests.
- Daily login boundary and timezone tests.

### Phase 5 — future token architecture
- Separate technical and economic proposal for external OGO Token.
- Liquidity/market-maker partnership research.
- Native OGO Coin integration plan.
- Security, legal, and operational review.

## 12. Acceptance criteria

The OGO Points MVP is not complete until:
1. A user can claim at most one daily login reward per reward day.
2. Qualifying daily login days 1–7 each award exactly 1 point.
3. Day 8 onward follows an explicitly configured progression capped at 5 points/day.
4. Missed-day streak behavior is explicitly specified and tested.
5. A duplicate request cannot issue a second reward.
6. App-specific currencies cannot be credited as OGO Points.
7. Only an authorized Rewards Engine policy can issue platform OGO Points.
8. Each app's budget and platform limits are enforced under concurrent activity.
9. Developer and player allocations reconcile to the approved budget.
10. Every grant, spend, rejection, and reversal is auditable.
11. Dashboard metrics distinguish visits from qualifying participation and posted points.
12. No cryptocurrency minting, conversion, cash-out, or financial promise is active in the MVP.

## 13. Authorization boundary

This specification authorizes documentation and implementation planning for the OGO Points rewards system. It does not authorize cryptocurrency issuance, changes to native OGO Coin supply, spending, purchases, treasury transfers, liquidity provisioning, chain migration, production financial activation, or deployment. Those require separate explicit approval.

**Product principle:** OGO lets creators build applications and their own app-level economies, while the OGO Rewards Engine provides a shared, transparent, capped platform-wide points economy that rewards genuine participation and creator contribution.
