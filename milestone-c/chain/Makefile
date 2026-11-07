# OGO Network local prototype: Makefile (Milestones 1-2)
#
# All Go commands run with TMPDIR and GOCACHE pointed at the workspace
# (the /tmp tmpfs is only 512 MB and a wasmd CGO build fills it).
# Toolchain: ~/workspace/tools/go-1.25.7/bin/go (Go 1.25.7, locked).
# Do NOT edit go.mod or go.sum here: the dependency matrix is LOCKED
# (Cosmos SDK v0.53.6 required with Gaia's replace to v0.53.4,
# CometBFT v0.38.23, wasmd v0.60.9, wasmvm/v2 v2.3.5, ibc-go/v10
# v10.7.0). No single dependency may be upgraded in isolation.

GO      := $(HOME)/workspace/tools/go-1.25.7/bin/go
export TMPDIR  := $(HOME)/workspace/.go-tmp
export GOCACHE := $(HOME)/workspace/.go-build-cache

.PHONY: build vet test test-verbose localnet localnet-verify localnet-stop localnet-liveness demo clean help

help: ## Show this help.
	@grep -E '^[a-z-]+:.*?## ' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-18s %s\n", $$1, $$2}'

build: ## Build the ogod binary into chain/ogod.
	@mkdir -p $(TMPDIR) $(GOCACHE)
	$(GO) build -o ogod ./cmd/ogod

vet: ## Run go vet over the whole module.
	@mkdir -p $(TMPDIR) $(GOCACHE)
	$(GO) vet ./...

test: ## Run all tests once (no cache).
	@mkdir -p $(TMPDIR) $(GOCACHE)
	$(GO) test -count=1 ./...

test-verbose: ## Run all tests with verbose output.
	@mkdir -p $(TMPDIR) $(GOCACHE)
	$(GO) test -count=1 -v ./...

localnet: ## Build, init, and start the 4-validator localnet, then verify it.
	./scripts/localnet.sh up

localnet-verify: ## Verify the same ledger state on all 4 running nodes.
	./scripts/localnet.sh verify

localnet-stop: ## Stop the 4 localnet nodes.
	./scripts/localnet.sh stop

localnet-liveness: ## Halt-and-resume: stop val3, confirm halt, restart it, verify.
	./scripts/localnet.sh liveness

demo: ## Run the end-to-end demonstration (prototype root demo.sh).
	../demo.sh

clean: ## Stop the localnet and delete its data and the built binary.
	./scripts/localnet.sh clean || true
	rm -f ogod
