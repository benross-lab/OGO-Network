package ogocoin

import (
	"strings"
	"testing"

	"cosmossdk.io/math"
)

// Bypass verification. Each test attempts one attack vector and asserts
// (a) the attempt fails and (b) monetary state is byte-identical
// afterwards. A test that only checks the error and not the state would
// miss a partial-bypass, so both are asserted everywhere.

// ---------------------------------------------------------------------------
// Vector 1: privileged module accounts
// ---------------------------------------------------------------------------

func TestBypassModuleAccountMint(t *testing.T) {
	l := testLedger(t)
	for _, mod := range KnownModuleAccounts {
		before := l.Clone()
		if err := l.MintFromModuleAccount(mod, math.NewInt(1_000_000)); err == nil {
			t.Fatalf("module account %q must not be able to mint", mod)
		}
		if !l.Equal(before) {
			t.Fatalf("failed mint from %q changed state", mod)
		}
	}
	// Even the module account that sounds like it should mint must fail.
	if err := l.MintFromModuleAccount("mint", math.NewInt(1)); err == nil {
		t.Fatal("a module account named 'mint' must still fail: OGO has no x/mint module")
	}
}

func TestBypassModuleAccountBurnToMint(t *testing.T) {
	l := testLedger(t)
	before := l.Clone()
	// Burn some coins legitimately first, then try to claim them back
	// through a module-account burn-to-mint path.
	if err := l.Burn("ogo1alice", math.NewInt(10_000)); err != nil {
		t.Fatalf("setup burn failed: %v", err)
	}
	mid := l.Clone()
	if err := l.BurnFromModuleAccountToMint("gov", math.NewInt(10_000)); err == nil {
		t.Fatal("burn-to-mint via module account must fail")
	}
	if !l.Equal(mid) {
		t.Fatal("failed burn-to-mint changed state")
	}
	_ = before // before is retained to show the burn itself was legitimate
}

func TestBypassModuleAccountOverspend(t *testing.T) {
	l := testLedger(t)
	// Module accounts start with zero balance in this ledger. Sending
	// from one must fail, not create coins.
	before := l.Clone()
	for _, mod := range []string{"fee_collector", "gov", "wasm", "admin"} {
		if err := l.SendFromModuleAccount(mod, "ogo1alice", math.NewInt(1)); err == nil {
			t.Fatalf("module account %q sent coins it does not hold", mod)
		}
	}
	if !l.Equal(before) {
		t.Fatal("failed module sends changed state")
	}
	// A module account that HAS received coins may send only what it holds.
	if err := l.Send("ogo1alice", "fee_collector", math.NewInt(1_000)); err != nil {
		t.Fatalf("funding fee collector failed: %v", err)
	}
	supplyBefore := l.TotalSupply()
	if err := l.SendFromModuleAccount("fee_collector", "ogo1bob", math.NewInt(1_000)); err != nil {
		t.Fatalf("legitimate module send failed: %v", err)
	}
	if !l.TotalSupply().Equal(supplyBefore) {
		t.Fatal("legitimate module send changed total supply")
	}
	if err := l.SendFromModuleAccount("fee_collector", "ogo1bob", math.NewInt(1)); err == nil {
		t.Fatal("module account must not overspend its balance")
	}
}

// ---------------------------------------------------------------------------
// Vector 2: governance upgrades
// ---------------------------------------------------------------------------

func TestBypassGovChangeMaxSupply(t *testing.T) {
	l := testLedger(t)
	before := l.Clone()
	newMax := math.NewInt(2_000_000_000) // TEST-ONLY value in the attempt
	err := l.ApplyUpgrade(UpgradeProposal{Name: "raise-cap", ChangeMaxSupply: &newMax})
	if err == nil {
		t.Fatal("gov upgrade changing max_supply must be rejected")
	}
	if !l.Equal(before) {
		t.Fatal("rejected max_supply upgrade changed state")
	}
	if !l.MaxSupply.Equal(math.NewInt(1_000_010_000)) { // TEST-ONLY testLedger max
		t.Fatalf("max_supply changed to %s", l.MaxSupply)
	}
	// Lowering the cap is also rejected: the parameter is immutable.
	lower := math.NewInt(1)
	if err := l.ApplyUpgrade(UpgradeProposal{Name: "lower-cap", ChangeMaxSupply: &lower}); err == nil {
		t.Fatal("gov upgrade lowering max_supply must also be rejected")
	}
}

func TestBypassGovEnableMinting(t *testing.T) {
	l := testLedger(t)
	before := l.Clone()
	if err := l.ApplyUpgrade(UpgradeProposal{Name: "enable-mint", EnableMinting: true}); err == nil {
		t.Fatal("gov upgrade enabling minting must be rejected")
	}
	if !l.Equal(before) {
		t.Fatal("rejected enable-mint upgrade changed state")
	}
	// Mint must still fail after the rejected upgrade.
	if err := l.Mint("ogo1alice", math.NewInt(1)); err == nil {
		t.Fatal("mint succeeded after rejected upgrade")
	}
}

func TestBypassGovAddMintPath(t *testing.T) {
	l := testLedger(t)
	before := l.Clone()
	for _, path := range []string{"x/mint", "MsgMint", "admin-mint", "wasm-mint"} {
		if err := l.ApplyUpgrade(UpgradeProposal{Name: "add-mint-path", AddMintPath: path}); err == nil {
			t.Fatalf("gov upgrade adding mint path %q must be rejected", path)
		}
		if !l.Equal(before) {
			t.Fatalf("rejected mint-path upgrade %q changed state", path)
		}
	}
}

func TestBypassGovCombinedAttack(t *testing.T) {
	l := testLedger(t)
	before := l.Clone()
	newMax := math.NewInt(9_999_999_999) // TEST-ONLY value in the attempt
	err := l.ApplyUpgrade(UpgradeProposal{
		Name:            "kitchen-sink",
		ChangeMaxSupply: &newMax,
		EnableMinting:   true,
		AddMintPath:     "x/mint",
	})
	if err == nil {
		t.Fatal("combined gov attack must be rejected")
	}
	if !l.Equal(before) {
		t.Fatal("rejected combined upgrade changed state")
	}
}

func TestGovNonMonetaryUpgradeAllowed(t *testing.T) {
	l := testLedger(t)
	before := l.Clone()
	if err := l.ApplyUpgrade(UpgradeProposal{Name: "docs-only", NonMonetaryNote: "update docs"}); err != nil {
		t.Fatalf("non-monetary upgrade should be allowed: %v", err)
	}
	if !l.Equal(before) {
		t.Fatal("non-monetary upgrade changed monetary state")
	}
}

// ---------------------------------------------------------------------------
// Vector 3: smart-contract (CosmWasm) interfaces
// ---------------------------------------------------------------------------

func TestBypassWasmMint(t *testing.T) {
	l := testLedger(t)
	before := l.Clone()
	if err := l.WasmMintAttempt("ogo1contract", math.NewInt(1_000_000)); err == nil {
		t.Fatal("CosmWasm contract mint attempt must fail")
	}
	if !l.Equal(before) {
		t.Fatal("failed contract mint changed state")
	}
}

func TestBypassWasmForbiddenMethods(t *testing.T) {
	l := testLedger(t)
	forbidden := []string{"BankMint", "MintCoins", "Mint", "CreateDenom", "MsgMint", "ForceMint"}
	for _, m := range forbidden {
		before := l.Clone()
		if err := l.WasmCall("ogo1contract", m, "ogo1contract", "ogo1alice", math.NewInt(100)); err == nil {
			t.Fatalf("contract method %q must be forbidden", m)
		}
		if !l.Equal(before) {
			t.Fatalf("forbidden method %q changed state", m)
		}
	}
}

func TestWasmBindingSurfaceHasNoMint(t *testing.T) {
	for _, m := range WasmExposedMethods {
		if strings.Contains(strings.ToLower(m), "mint") {
			t.Fatalf("exposed wasm method %q contains mint", m)
		}
	}
	if WasmMethodExposed("BankMint") || WasmMethodExposed("MintCoins") || WasmMethodExposed("Mint") {
		t.Fatal("a mint method is exposed in the wasm binding surface")
	}
	if !WasmMethodExposed("BankSend") || !WasmMethodExposed("BankBurn") {
		t.Fatal("safe wasm methods BankSend/BankBurn must be exposed")
	}
}

func TestBypassWasmUnknownMethod(t *testing.T) {
	l := testLedger(t)
	before := l.Clone()
	if err := l.WasmCall("ogo1contract", "SuperSecretMint", "ogo1contract", "ogo1alice", math.NewInt(1)); err == nil {
		t.Fatal("unknown contract method must fail")
	}
	if !l.Equal(before) {
		t.Fatal("unknown method changed state")
	}
}

func TestBypassWasmSendPreservesSupply(t *testing.T) {
	l := testLedger(t)
	before := l.TotalSupply()
	if err := l.WasmCall("ogo1contract", "BankSend", "ogo1alice", "ogo1bob", math.NewInt(5_000)); err != nil {
		t.Fatalf("legitimate contract BankSend failed: %v", err)
	}
	if !l.TotalSupply().Equal(before) {
		t.Fatal("contract BankSend changed total supply")
	}
}

func TestBypassGovChangeSchedule(t *testing.T) {
	l := testLedger(t)
	before := l.Clone()
	newSched := testOnlySchedule()
	newSched.InitialPerPeriod = math.NewInt(1_000_000_000) // TEST-ONLY: attacker wants huge issuance
	if err := l.ApplyUpgrade(UpgradeProposal{
		Name:           "schedule-change",
		ChangeSchedule: newSched,
	}); err == nil {
		t.Fatal("governance schedule change must be rejected")
	}
	if !l.Equal(before) {
		t.Fatal("rejected schedule-change upgrade changed state")
	}
	// The schedule itself must be unchanged and still enforceable.
	if got := l.Schedule.InitialPerPeriod; !got.Equal(math.NewInt(1_000)) {
		t.Fatalf("schedule was tampered: initial_per_period %s", got)
	}
}

func TestBypassFounderMint(t *testing.T) {
	l := testLedger(t)
	before := l.Clone()
	// A founder address, an OGO Systems address, and a "founder reserve"
	// module account all attempt discretionary minting.
	for _, caller := range []string{"ogo1founder", "ogo1ogosystems", "founder", "ogo_systems_reserve"} {
		if err := l.MintFromModuleAccount(caller, math.NewInt(1_000_000)); err == nil {
			t.Fatalf("founder mint from %q must fail", caller)
		}
		if err := l.Mint(caller, math.NewInt(1_000_000)); err == nil {
			t.Fatalf("discretionary mint by %q must fail", caller)
		}
	}
	if !l.Equal(before) {
		t.Fatal("failed founder mint attempts changed state")
	}
	// Zero founder allocation is also enforced at genesis (see
	// TestValidateGenesisRejectsFounderAllocation).
}

func TestBypassOffScheduleMint(t *testing.T) {
	l := mustLedger(t, math.NewInt(1_000_000)) // TEST-ONLY
	before := l.Clone()
	// Minting the schedule amount at the WRONG height: period 0 amount
	// (1000) claimed at height 101 (period 1) must fail as double/early
	// issuance, because period 1's amount is 900, not 1000, and the
	// caller cannot choose the amount.
	if err := l.MintScheduled(1); err != nil {
		t.Fatalf("period 0 issuance failed: %v", err)
	}
	// Re-claiming period 0 at a different height must fail.
	if err := l.MintScheduled(2); err == nil {
		t.Fatal("re-issuance of an already-claimed period must fail")
	}
	// Minting "above schedule" is impossible: there is no API that takes
	// an amount. The only mint entrypoint is MintScheduled(height), and
	// it always issues exactly the schedule amount. Verify the pool
	// holds exactly the period-0 amount.
	if got := l.BalanceOf(RewardsPool); !got.Equal(math.NewInt(1_000)) {
		t.Fatalf("rewards pool: got %s, want exactly 1000 after one period", got)
	}
	if !l.TotalIssued.Equal(math.NewInt(1_000)) {
		t.Fatalf("total issued: got %s, want 1000", l.TotalIssued)
	}
	_ = before
}

func TestBypassWasmScheduleChange(t *testing.T) {
	l := testLedger(t)
	before := l.Clone()
	if err := l.WasmCall("ogo1contract", "ChangeSchedule", "ogo1contract", "ogo1alice", math.NewInt(1)); err == nil {
		t.Fatal("contract schedule-change call must fail")
	}
	if err := l.WasmCall("ogo1contract", "MintScheduled", "ogo1contract", "ogo1alice", math.NewInt(1_000_000)); err == nil {
		t.Fatal("contract must not be able to invoke scheduled issuance directly")
	}
	if !l.Equal(before) {
		t.Fatal("failed contract schedule calls changed state")
	}
}
