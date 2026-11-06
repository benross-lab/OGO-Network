# OGO Network Security Policy

## Scope

This policy covers the OGO Network source code in this repository, including the
chain implementation, validator-related code, protocol specifications, and
proof-of-concept code.

The code currently published here is a development prototype. It is not a
production blockchain, public testnet, wallet, or place to store or transfer
real value. Do not test against systems you do not own or have explicit
permission to assess.

## Reporting a vulnerability

Please do not publish an exploitable vulnerability in a public issue or pull
request.

If GitHub's private vulnerability reporting is enabled for this repository,
use the **Security** tab and **Report a vulnerability** to submit a private
report. If that option is unavailable, contact the repository owner through a
verified private channel before sharing technical details. Do not include
private keys, seed phrases, credentials, or personal data in a report.

A useful report includes:
- A concise description and affected commit, release, or file.
- The impact and realistic attack conditions.
- Reproduction steps or a minimal proof of concept.
- Any suggested mitigation.

Please allow maintainers time to investigate and coordinate a fix before
public disclosure. There is no guaranteed response or remediation SLA at this
prototype stage.

## Safe harbor

Good-faith security research is welcome when it is limited to systems you own
or are explicitly authorized to test, avoids privacy violations and service
disruption, and follows the reporting guidance above. Stop testing and report
promptly if you encounter user data or credentials.

## Current security boundary

The prototype documentation identifies important limitations, including
transaction authentication that is not yet implemented in the scaffold.
The prototype must not be exposed to adversarial submitters or used to hold
real value. A passing test suite or automated scan is not a security audit
and does not establish mainnet readiness.
