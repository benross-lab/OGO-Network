# Milestone C transaction authentication: security note pointer

The chain security note for this milestone lives outside this
repository, at:

`../ogo-network-milestone-c-security-note.md`

(workspace path: `~/workspace/goals/ogo-ai-builder-and-network/files/ogo-network-milestone-c-security-note.md`)

It covers the authentication architecture, signature verification,
signer and public key validation, chain ID and account number
validation, sequence and replay protection, CheckTx and
block execution behaviour, monetary bypass protection, the test names
and actual results, and the remaining risks and unverified
assumptions. It states plainly that no independent security audit has
occurred. Implementation tests are not an independent audit.

Tests added in this commit: `chain/app/auth_test.go` (15 new tests).
No application code was changed: the existing standard SDK ante
handler already enforced every tested property.
