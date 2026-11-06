## Summary

Describe the change and the reason for it.

## Scope

- [ ] Protocol or consensus
- [ ] Monetary rules
- [ ] Transaction authentication
- [ ] Validator operations
- [ ] Documentation
- [ ] Build, test, or tooling

## Verification

- [ ] Tests added or updated where behavior changes
- [ ] `go vet ./...` passes for each affected Go module
- [ ] `go test -count=1 ./...` passes for each affected Go module
- [ ] No production credentials, private keys, mnemonics, or validator state included
- [ ] No production monetary parameters or launch claims changed without an explicit decision
- [ ] Security implications and remaining limitations are documented

## Compatibility and risk

Describe state migrations, consensus changes, monetary effects, upgrade requirements,
and any operational risks. If none apply, say why.
