# Repository Security Baseline

Status: Initial hardening proposal. Apply repository-level settings separately;
workflow files alone do not enable every GitHub security control.

## Source and access boundaries

- Keep protocol source and public technical specifications reviewable.
- Never commit validator private keys, node keys, wallet seeds, signing
  credentials, API tokens, production environment files, or live validator
  state.
- Keep wallet services, user-account data, treasury operations, production
  credentials, and private business systems in appropriately restricted
  repositories and infrastructure.
- Use short-lived, least-privilege credentials for automation. Do not put
  secrets in workflow YAML, logs, issues, pull requests, or artifacts.
- Treat localnet keys and test balances as disposable test material only.

## GitHub repository settings to enable

These are owner-level settings and are not enabled merely by adding files:

1. Enable private vulnerability reporting.
2. Enable secret scanning and push protection where available.
3. Enable Dependabot security updates and alerts.
4. Enable CodeQL code scanning and review its first results.
5. Protect `main`: require pull requests, require the Go CI and CodeQL checks
   to pass, require conversation resolution, dismiss stale approvals after
   new commits, and block force pushes and branch deletion.
6. Restrict who can bypass the branch rules. Keep emergency bypass exceptional
   and auditable.
7. Require signed commits if the maintainers can consistently sign them.
8. Review GitHub Actions permissions; use read-only defaults and grant write
   permissions only to jobs that need them.
9. Pin third-party Actions to verified full commit SHAs before treating CI as
   a hardened production control.

## Change-control rules

- Work on topic branches and submit pull requests.
- Do not merge protocol, consensus, cryptography, monetary, or validator
  changes without focused review and tests.
- Never treat a successful CI run as an independent security audit.
- Do not publish a public testnet or enable real-value transfers until
  transaction authentication, validator operations, upgrade safety, and
  independent security review meet the approved launch criteria.

## Current prototype warning

The current ABCI scaffold documents that JSON transactions are not
signature-verified. This repository is a local development prototype and
must not be exposed to adversarial transaction submitters or used to hold
real value. See the security model and prototype blocker register for the
full list of limitations.
