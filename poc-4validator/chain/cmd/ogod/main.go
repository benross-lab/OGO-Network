// Command ogod is the OGO Network local prototype binary (Milestones 1-2).
//
// Localnet only. Chain ID ogo-local-1. No real value, no public network.
//
// SECURITY LIMITATION (read before use): transactions in this prototype
// are JSON and are NOT signature-verified. The ledger enforces balances
// and the fixed supply, but anyone with RPC access can submit a tx naming
// any sender. Real auth (x/auth signatures) arrives with the full SDK
// depinject wiring in Milestone C. Keys stored under <home>/keys are
// plaintext and LOCALNET ONLY.
//
// Epoch mode (Milestone 2): pass --epoch-duration-seconds to init to
// enable the epoch timer mirror (see x/ogocoin/epochs.go and
// chain/EPOCHS_GAP.md). Issuance then runs only through the monetary
// AfterEpochEnd hook, once per epoch. Without the flag, issuance uses
// the legacy block-based path from Milestone B.
package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"cosmossdk.io/math"
	"github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/node"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/privval"
	"github.com/cometbft/cometbft/proxy"
	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	ctypes "github.com/cometbft/cometbft/rpc/core/types"
	"github.com/cometbft/cometbft/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/benross-lab/ogo-network/app"
	ogocoin "github.com/benross-lab/ogo-network/x/ogocoin"
)

const (
	defaultRPC = "tcp://127.0.0.1:26657"
	version    = "0.2.0-milestones-1-2"
)

func main() {
	sdk.GetConfig().SetBech32PrefixForAccount("ogo", "ogopub")
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}
	var err error
	switch os.Args[1] {
	case "version":
		fmt.Printf("ogod %s\n", version)
		fmt.Println("OGO Network local prototype, chain ogo-local-1")
		fmt.Println("Go 1.25.7 build toolchain; Go 1.23 documented minimum only")
		fmt.Println("Gaia v28.3.1 matrix: SDK v0.53.6 required / v0.53.4 effective, CometBFT v0.38.23, wasmd v0.60.9, ibc-go v10.7.0")
	case "init":
		err = cmdInit(os.Args[2:])
	case "show-node-id":
		err = cmdShowNodeID(os.Args[2:])
	case "show-validator":
		err = cmdShowValidator(os.Args[2:])
	case "set-genesis-validators":
		err = cmdSetGenesisValidators(os.Args[2:])
	case "keys":
		err = cmdKeys(os.Args[2:])
	case "add-genesis-account":
		err = cmdAddGenesisAccount(os.Args[2:])
	case "validate-genesis":
		err = cmdValidateGenesis(os.Args[2:])
	case "start":
		err = cmdStart(os.Args[2:])
	case "tx":
		err = cmdTx(os.Args[2:])
	case "query":
		err = cmdQuery(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n", os.Args[1])
		usage()
		os.Exit(1)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Println(`usage: ogod <command> [flags]

commands:
  version
  init --home <dir> --chain-id ogo-local-1 --max-supply <uogo> [--denom uogo] \
    --schedule-period-length <blocks> --schedule-periods <n> --schedule-initial-per-period <uogo> \
    [--schedule-start <height>] [--schedule-reduction NUM/DEN] \
    [--epoch-duration-seconds <secs>] [--epoch-identifier monetary]
  show-node-id [--home <dir>]
  show-validator [--home <dir>]
  set-genesis-validators --home <dir> --homes <dir1,dir2,...>
  keys add <name> [--home <dir>]
  keys list [--home <dir>]
  add-genesis-account <address> <amount> [--purpose label] [--home <dir>]
  validate-genesis [--home <dir>]
  start [--home <dir>] [--moniker <name>] [--rpc-laddr tcp://127.0.0.1:26657] \
    [--p2p-laddr tcp://127.0.0.1:26656] [--persistent-peers <id@host:port,...>]
  tx send --from-key <name> --to <addr> --amount <uogo> [--fee <uogo>] [--home <dir>] [--node tcp://127.0.0.1:26657]
  tx burn --from-key <name> --amount <uogo> [--home <dir>] [--node tcp://127.0.0.1:26657]
  query balance <address> [--node tcp://127.0.0.1:26657]
  query supply [--node tcp://127.0.0.1:26657]
  query params [--node tcp://127.0.0.1:26657]
  query schedule [--node tcp://127.0.0.1:26657]
  query epochs [--node tcp://127.0.0.1:26657]
  query global-supply <channel> [--node tcp://127.0.0.1:26657]

--max-supply and the issuance schedule are mandatory for init and have
no defaults. Values used in localnet testing are TEST-ONLY and are not
the production supply or schedule (both TBD). --epoch-duration-seconds
enables epoch mode (TEST-ONLY durations for localnet); without it the
app uses the legacy block-based issuance path. Multi-validator localnet
setup lives in chain/scripts/localnet.sh (4 validators).`)
}

func defaultHome() string {
	if h := os.Getenv("OGOD_HOME"); h != "" {
		return h
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ".ogod"
	}
	return filepath.Join(home, ".ogod")
}

func homeFlag(fs *flag.FlagSet) *string {
	return fs.String("home", defaultHome(), "ogod home directory")
}

func mustParseInt(s, what string) (math.Int, error) {
	v, ok := math.NewIntFromString(s)
	if !ok {
		return math.Int{}, fmt.Errorf("bad %s %q: not an integer", what, s)
	}
	return v, nil
}

// --- init ---

func cmdInit(args []string) error {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	home := homeFlag(fs)
	chainID := fs.String("chain-id", ogocoin.ChainID, "chain id")
	maxSupplyStr := fs.String("max-supply", "", "mandatory max supply in uogo (no default)")
	denom := fs.String("denom", ogocoin.Denom, "base denom (proposed: uogo)")
	// Issuance schedule flags (Constitution Article III). All schedule
	// VALUES are TBD in production; these flags are TEST-ONLY parameters
	// for the localnet. The schedule itself is mandatory: init fails
	// without the numeric schedule flags.
	schedStart := fs.Int64("schedule-start", 1, "first issuance height")
	schedPeriodLen := fs.Int64("schedule-period-length", 0, "blocks per issuance period (required)")
	schedPeriods := fs.Uint64("schedule-periods", 0, "number of issuance periods (required)")
	schedInitial := fs.String("schedule-initial-per-period", "", "issuance in period 0, uogo (required)")
	schedReduction := fs.String("schedule-reduction", "9/10", "per-period reduction factor numerator/denominator (TEST-ONLY default)")
	// Epoch timer flags (Milestone 2). A duration of 0 (the default)
	// leaves epoch mode disabled and keeps the legacy block-based
	// issuance path. Any positive duration enables epoch mode: the
	// monetary AfterEpochEnd hook becomes the only issuance path.
	epochDuration := fs.Int64("epoch-duration-seconds", 0, "epoch duration in seconds; 0 disables epoch mode (TEST-ONLY values for localnet)")
	epochIdentifier := fs.String("epoch-identifier", ogocoin.EpochIdentifierMonetary, "epoch identifier for the monetary schedule")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *maxSupplyStr == "" {
		return fmt.Errorf("--max-supply is mandatory and has no default; genesis cannot be created without it")
	}
	maxSupply, err := mustParseInt(*maxSupplyStr, "max-supply")
	if err != nil {
		return err
	}
	if !maxSupply.IsPositive() {
		return fmt.Errorf("--max-supply must be greater than zero")
	}
	if *schedPeriodLen <= 0 || *schedPeriods == 0 || *schedInitial == "" {
		return fmt.Errorf("issuance schedule is mandatory: pass --schedule-period-length, --schedule-periods and --schedule-initial-per-period (all TEST-ONLY values for localnet)")
	}
	initialPerPeriod, err := mustParseInt(*schedInitial, "schedule-initial-per-period")
	if err != nil {
		return err
	}
	var redNum, redDen int64
	if _, err := fmt.Sscanf(*schedReduction, "%d/%d", &redNum, &redDen); err != nil {
		return fmt.Errorf("bad --schedule-reduction %q, want NUM/DEN: %w", *schedReduction, err)
	}
	cfg := config.DefaultConfig()
	cfg.SetRoot(*home)
	for _, d := range []string{filepath.Join(*home, "config"), filepath.Join(*home, "data"), filepath.Join(*home, "keys")} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return err
		}
	}
	pv := privval.LoadOrGenFilePV(cfg.PrivValidatorKeyFile(), cfg.PrivValidatorStateFile())
	if _, err := p2p.LoadOrGenNodeKey(cfg.NodeKeyFile()); err != nil {
		return err
	}
	params := ogocoin.DefaultParams()
	params.MaxSupply = &maxSupply
	params.Denom = *denom
	params.IssuanceSchedule = &ogocoin.IssuanceSchedule{
		ScheduleType:         ogocoin.ScheduleTypeReducingPerPeriod,
		StartHeight:          *schedStart,
		PeriodLength:         *schedPeriodLen,
		Periods:              *schedPeriods,
		InitialPerPeriod:     initialPerPeriod,
		ReductionNumerator:   redNum,
		ReductionDenominator: redDen,
	}
	gs := ogocoin.GenesisState{Params: params, Balances: []ogocoin.Balance{}}
	if err := ogocoin.ValidateGenesis(gs); err != nil {
		return err
	}
	appState := map[string]any{"ogocoin": gs}
	if *epochDuration > 0 {
		epochCfg := ogocoin.EpochConfig{Identifier: *epochIdentifier, DurationSeconds: *epochDuration}
		if err := epochCfg.Validate(); err != nil {
			return err
		}
		appState["epochs"] = epochCfg
	}
	appStateBytes, err := json.Marshal(appState)
	if err != nil {
		return err
	}
	pubKey, err := pv.GetPubKey()
	if err != nil {
		return err
	}
	genDoc := &types.GenesisDoc{
		ChainID:         *chainID,
		GenesisTime:     time.Now().UTC(),
		InitialHeight:   1,
		ConsensusParams: types.DefaultConsensusParams(),
		Validators: []types.GenesisValidator{{
			Address: pv.GetAddress(),
			PubKey:  pubKey,
			Power:   10,
			Name:    "ogo-validator",
		}},
		AppState: appStateBytes,
	}
	if err := genDoc.SaveAs(cfg.GenesisFile()); err != nil {
		return err
	}
	fmt.Printf("initialized ogo-local home at %s\n", *home)
	fmt.Printf("chain-id: %s\n", *chainID)
	fmt.Printf("validator: %s\n", pv.GetAddress().String())
	fmt.Printf("max_supply: %s %s (TEST-ONLY value for localnet; production supply is TBD)\n", maxSupply.String(), *denom)
	fmt.Printf("issuance schedule: %s start=%d period_len=%d periods=%d initial=%s reduction=%d/%d total_scheduled=%s (TEST-ONLY values; production schedule TBD)\n",
		params.IssuanceSchedule.ScheduleType, params.IssuanceSchedule.StartHeight, params.IssuanceSchedule.PeriodLength,
		params.IssuanceSchedule.Periods, params.IssuanceSchedule.InitialPerPeriod.String(),
		params.IssuanceSchedule.ReductionNumerator, params.IssuanceSchedule.ReductionDenominator,
		params.IssuanceSchedule.TotalIssuance().String())
	if *epochDuration > 0 {
		fmt.Printf("epoch mode ENABLED: identifier=%s duration=%ds (TEST-ONLY; issuance runs only through the monetary AfterEpochEnd hook)\n", *epochIdentifier, *epochDuration)
	} else {
		fmt.Println("epoch mode disabled: issuance uses the legacy block-based path (pass --epoch-duration-seconds to enable epoch mode)")
	}
	fmt.Printf("genesis: %s\n", cfg.GenesisFile())
	return nil
}

// --- node identity and validator set ---

// cmdShowNodeID prints the P2P node ID for a home (used to build
// --persistent-peers for the multi-validator localnet).
func cmdShowNodeID(args []string) error {
	fs := flag.NewFlagSet("show-node-id", flag.ContinueOnError)
	home := homeFlag(fs)
	if err := fs.Parse(args); err != nil {
		return err
	}
	cfg := config.DefaultConfig()
	cfg.SetRoot(*home)
	nodeKey, err := p2p.LoadOrGenNodeKey(cfg.NodeKeyFile())
	if err != nil {
		return err
	}
	fmt.Println(nodeKey.ID())
	return nil
}

// cmdShowValidator prints the validator address and public key for a
// home, in the form the genesis validator set needs.
func cmdShowValidator(args []string) error {
	fs := flag.NewFlagSet("show-validator", flag.ContinueOnError)
	home := homeFlag(fs)
	if err := fs.Parse(args); err != nil {
		return err
	}
	cfg := config.DefaultConfig()
	cfg.SetRoot(*home)
	pv := privval.LoadOrGenFilePV(cfg.PrivValidatorKeyFile(), cfg.PrivValidatorStateFile())
	pubKey, err := pv.GetPubKey()
	if err != nil {
		return err
	}
	pubJSON, _ := json.Marshal(pubKey)
	fmt.Printf("address: %s\n", pv.GetAddress().String())
	fmt.Printf("pubkey: %s\n", string(pubJSON))
	return nil
}

// cmdSetGenesisValidators replaces the validator set in a home's
// genesis doc with the validators of the given homes (each home
// contributes its own priv_validator key, power 10). This is how the
// multi-validator localnet builds one shared genesis: init each home,
// run this on one home with all homes listed, then copy that
// genesis.json to every home.
func cmdSetGenesisValidators(args []string) error {
	fs := flag.NewFlagSet("set-genesis-validators", flag.ContinueOnError)
	home := homeFlag(fs)
	homes := fs.String("homes", "", "comma-separated validator home directories")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *homes == "" {
		return fmt.Errorf("usage: ogod set-genesis-validators --home <dir> --homes <dir1,dir2,...>")
	}
	genDoc, _, err := readGenesis(*home)
	if err != nil {
		return err
	}
	var validators []types.GenesisValidator
	for _, h := range strings.Split(*homes, ",") {
		h = strings.TrimSpace(h)
		if h == "" {
			continue
		}
		cfg := config.DefaultConfig()
		cfg.SetRoot(h)
		pv := privval.LoadOrGenFilePV(cfg.PrivValidatorKeyFile(), cfg.PrivValidatorStateFile())
		pubKey, err := pv.GetPubKey()
		if err != nil {
			return err
		}
		validators = append(validators, types.GenesisValidator{
			Address: pv.GetAddress(),
			PubKey:  pubKey,
			Power:   10,
			Name:    filepath.Base(h),
		})
	}
	if len(validators) == 0 {
		return fmt.Errorf("no validator homes given")
	}
	genDoc.Validators = validators
	if err := writeGenesis(*home, genDoc); err != nil {
		return err
	}
	fmt.Printf("genesis validators set to %d:\n", len(validators))
	for _, v := range validators {
		fmt.Printf("  %s power=%d name=%s\n", v.Address.String(), v.Power, v.Name)
	}
	return nil
}

// --- keys ---

type keyFile struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	PrivHex string `json:"privkey_hex"`
}

func cmdKeys(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: ogod keys <add|list> ...")
	}
	switch args[0] {
	case "add":
		fs := flag.NewFlagSet("keys add", flag.ContinueOnError)
		home := homeFlag(fs)
		if err := fs.Parse(flagsFirst(args[1:])); err != nil {
			return err
		}
		if fs.NArg() < 1 {
			return fmt.Errorf("usage: ogod keys add <name>")
		}
		name := fs.Arg(0)
		priv := secp256k1.GenPrivKey()
		addr := sdk.AccAddress(priv.PubKey().Address())
		kf := keyFile{Name: name, Address: addr.String(), PrivHex: hex.EncodeToString(priv.Bytes())}
		p := filepath.Join(*home, "keys", name+".json")
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			return err
		}
		b, _ := json.MarshalIndent(kf, "", "  ")
		if err := os.WriteFile(p, b, 0o600); err != nil {
			return err
		}
		fmt.Printf("key %q created: %s\n", name, kf.Address)
		fmt.Println("WARNING: key file is plaintext, LOCALNET ONLY. Never use for real value.")
		return nil
	case "list":
		fs := flag.NewFlagSet("keys list", flag.ContinueOnError)
		home := homeFlag(fs)
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		entries, err := os.ReadDir(filepath.Join(*home, "keys"))
		if err != nil {
			return err
		}
		for _, e := range entries {
			b, err := os.ReadFile(filepath.Join(*home, "keys", e.Name()))
			if err != nil {
				continue
			}
			var kf keyFile
			if json.Unmarshal(b, &kf) == nil {
				fmt.Printf("%s\t%s\n", kf.Name, kf.Address)
			}
		}
		return nil
	default:
		return fmt.Errorf("unknown keys subcommand %q", args[0])
	}
}

func loadKey(home, name string) (keyFile, error) {
	var kf keyFile
	b, err := os.ReadFile(filepath.Join(home, "keys", name+".json"))
	if err != nil {
		return kf, fmt.Errorf("key %q not found: %w", name, err)
	}
	if err := json.Unmarshal(b, &kf); err != nil {
		return kf, err
	}
	return kf, nil
}

// --- genesis accounts ---

// readGenesis reads a home's genesis doc and its full app_state as a
// raw map, so callers that edit one key (ogocoin balances, validator
// set) never drop the other keys (notably the "epochs" config).
func readGenesis(home string) (*types.GenesisDoc, map[string]json.RawMessage, error) {
	cfg := config.DefaultConfig()
	cfg.SetRoot(home)
	genDoc, err := types.GenesisDocFromFile(cfg.GenesisFile())
	if err != nil {
		return nil, nil, err
	}
	var appState map[string]json.RawMessage
	if err := json.Unmarshal(genDoc.AppState, &appState); err != nil {
		return nil, nil, err
	}
	return genDoc, appState, nil
}

// ogocoinState extracts the x/ogocoin genesis state from raw app_state.
func ogocoinState(appState map[string]json.RawMessage) (ogocoin.GenesisState, error) {
	var gs ogocoin.GenesisState
	raw, ok := appState["ogocoin"]
	if !ok {
		return gs, fmt.Errorf("genesis app_state has no ogocoin key")
	}
	if err := json.Unmarshal(raw, &gs); err != nil {
		return gs, err
	}
	return gs, nil
}

// setOgocoinState writes the x/ogocoin genesis state back into raw
// app_state, preserving every other key.
func setOgocoinState(appState map[string]json.RawMessage, gs ogocoin.GenesisState) error {
	b, err := json.Marshal(gs)
	if err != nil {
		return err
	}
	appState["ogocoin"] = b
	return nil
}

// writeGenesis writes the (possibly edited) genesis doc back to disk.
// The caller owns genDoc.AppState: use setOgocoinState plus
// syncAppState to edit the ogocoin key without dropping other keys.
func writeGenesis(home string, genDoc *types.GenesisDoc) error {
	cfg := config.DefaultConfig()
	cfg.SetRoot(home)
	return genDoc.SaveAs(cfg.GenesisFile())
}

// syncAppState marshals raw app_state back onto the genesis doc.
func syncAppState(genDoc *types.GenesisDoc, appState map[string]json.RawMessage) error {
	b, err := json.Marshal(appState)
	if err != nil {
		return err
	}
	genDoc.AppState = b
	return nil
}

func cmdAddGenesisAccount(args []string) error {
	fs := flag.NewFlagSet("add-genesis-account", flag.ContinueOnError)
	home := homeFlag(fs)
	purpose := fs.String("purpose", "public", "public-disclosure purpose label (founder/reserved purposes are rejected)")
	if err := fs.Parse(flagsFirst(args)); err != nil {
		return err
	}
	if fs.NArg() < 2 {
		return fmt.Errorf("usage: ogod add-genesis-account <address> <amount> [--purpose label]")
	}
	addr := fs.Arg(0)
	if _, err := sdk.AccAddressFromBech32(addr); err != nil {
		return fmt.Errorf("bad address %q: %w", addr, err)
	}
	amt, err := mustParseInt(fs.Arg(1), "amount")
	if err != nil {
		return err
	}
	genDoc, appState, err := readGenesis(*home)
	if err != nil {
		return err
	}
	gs, err := ogocoinState(appState)
	if err != nil {
		return err
	}
	gs.Balances = append(gs.Balances, ogocoin.Balance{Address: addr, Denom: gs.Params.Denom, Amount: amt, Purpose: *purpose})
	if err := ogocoin.ValidateGenesis(gs); err != nil {
		return fmt.Errorf("genesis validation failed: %w", err)
	}
	if err := setOgocoinState(appState, gs); err != nil {
		return err
	}
	if err := syncAppState(genDoc, appState); err != nil {
		return err
	}
	if err := writeGenesis(*home, genDoc); err != nil {
		return err
	}
	fmt.Printf("added genesis account %s: %s %s\n", addr, amt.String(), gs.Params.Denom)
	fmt.Printf("genesis total supply now: %s / max %s\n", ogocoin.TotalGenesisSupply(gs).String(), gs.Params.MaxSupply.String())
	return nil
}

func cmdValidateGenesis(args []string) error {
	fs := flag.NewFlagSet("validate-genesis", flag.ContinueOnError)
	home := homeFlag(fs)
	if err := fs.Parse(args); err != nil {
		return err
	}
	_, appState, err := readGenesis(*home)
	if err != nil {
		return err
	}
	gs, err := ogocoinState(appState)
	if err != nil {
		return err
	}
	if err := ogocoin.ValidateGenesis(gs); err != nil {
		return fmt.Errorf("genesis INVALID: %w", err)
	}
	fmt.Printf("genesis VALID: max_supply=%s total=%s denom=%s\n",
		gs.Params.MaxSupply.String(), ogocoin.TotalGenesisSupply(gs).String(), gs.Params.Denom)
	return nil
}

// --- start ---

func cmdStart(args []string) error {
	fs := flag.NewFlagSet("start", flag.ContinueOnError)
	home := homeFlag(fs)
	moniker := fs.String("moniker", "ogo-local-1", "node moniker")
	rpcLaddr := fs.String("rpc-laddr", "", "RPC listen address (default: CometBFT default tcp://127.0.0.1:26657)")
	p2pLaddr := fs.String("p2p-laddr", "", "P2P listen address (default: CometBFT default tcp://127.0.0.1:26656)")
	peers := fs.String("persistent-peers", "", "comma-separated persistent peers (id@host:port)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	cfg := config.DefaultConfig()
	cfg.SetRoot(*home)
	cfg.Moniker = *moniker
	if *rpcLaddr != "" {
		cfg.RPC.ListenAddress = *rpcLaddr
	}
	if *p2pLaddr != "" {
		cfg.P2P.ListenAddress = *p2pLaddr
	}
	if *peers != "" {
		cfg.P2P.PersistentPeers = *peers
	}
	// Localnet only: several validators share one IP (127.0.0.1) on
	// distinct ports. CometBFT rejects same-IP peers by default;
	// allow them here. Never enable this on a public network.
	cfg.P2P.AllowDuplicateIP = true
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))
	oapp := app.NewOgoApp(logger, *home)
	pv := privval.LoadOrGenFilePV(cfg.PrivValidatorKeyFile(), cfg.PrivValidatorStateFile())
	nodeKey, err := p2p.LoadOrGenNodeKey(cfg.NodeKeyFile())
	if err != nil {
		return err
	}
	// CometBFT 0.38 NewNode has no context argument:
	// NewNode(config, pv, nodeKey, clientCreator, genesisDocProvider,
	// dbProvider, metricsProvider, logger). DBProvider lives in config.
	n, err := node.NewNode(cfg, pv, nodeKey, proxy.NewLocalClientCreator(oapp),
		node.DefaultGenesisDocProviderFunc(cfg), config.DefaultDBProvider,
		node.DefaultMetricsProvider(cfg.Instrumentation), logger)
	if err != nil {
		return fmt.Errorf("failed to create node: %w", err)
	}
	if err := n.Start(); err != nil {
		return fmt.Errorf("failed to start node: %w", err)
	}
	fmt.Println("ogo-network node started")
	fmt.Printf("moniker: %s node-id: %s\n", cfg.Moniker, nodeKey.ID())
	fmt.Printf("RPC: %s  P2P: %s\n", cfg.RPC.ListenAddress, cfg.P2P.ListenAddress)
	fmt.Println("Press Ctrl-C to stop.")
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig
	fmt.Println("stopping node...")
	if err := n.Stop(); err != nil {
		return err
	}
	return nil
}

// --- tx ---

func rpcClient(nodeAddr string) (*rpchttp.HTTP, error) {
	return rpchttp.New(nodeAddr, "/websocket")
}

func cmdTx(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: ogod tx <send|burn> ...")
	}
	switch args[0] {
	case "send":
		fs := flag.NewFlagSet("tx send", flag.ContinueOnError)
		home := homeFlag(fs)
		fromKey := fs.String("from-key", "", "key name")
		to := fs.String("to", "", "recipient address")
		amount := fs.String("amount", "", "amount in uogo")
		fee := fs.String("fee", "0", "fee in uogo")
		nodeAddr := fs.String("node", defaultRPC, "node RPC address")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if *fromKey == "" || *to == "" || *amount == "" {
			return fmt.Errorf("tx send requires --from-key, --to and --amount")
		}
		kf, err := loadKey(*home, *fromKey)
		if err != nil {
			return err
		}
		raw, _ := json.Marshal(map[string]string{
			"type": "send", "from": kf.Address, "to": *to, "amount": *amount, "fee": *fee,
		})
		return broadcast(*nodeAddr, raw)
	case "burn":
		fs := flag.NewFlagSet("tx burn", flag.ContinueOnError)
		home := homeFlag(fs)
		fromKey := fs.String("from-key", "", "key name")
		amount := fs.String("amount", "", "amount in uogo")
		nodeAddr := fs.String("node", defaultRPC, "node RPC address")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if *fromKey == "" || *amount == "" {
			return fmt.Errorf("tx burn requires --from-key and --amount")
		}
		kf, err := loadKey(*home, *fromKey)
		if err != nil {
			return err
		}
		raw, _ := json.Marshal(map[string]string{
			"type": "burn", "from": kf.Address, "amount": *amount,
		})
		return broadcast(*nodeAddr, raw)
	default:
		return fmt.Errorf("unknown tx subcommand %q", args[0])
	}
}

func broadcast(nodeAddr string, raw []byte) error {
	c, err := rpcClient(nodeAddr)
	if err != nil {
		return err
	}
	res, err := c.BroadcastTxSync(context.Background(), raw)
	if err != nil {
		return err
	}
	fmt.Printf("tx hash: %X code: %d log: %s\n", res.Hash, res.Code, res.Log)
	if res.Code != 0 {
		return fmt.Errorf("tx rejected by CheckTx (code %d): %s", res.Code, res.Log)
	}
	return nil
}

// --- query ---

func cmdQuery(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: ogod query <balance|supply|params|schedule|epochs|global-supply> ...")
	}
	sub := args[0]
	fs := flag.NewFlagSet("query "+sub, flag.ContinueOnError)
	nodeAddr := fs.String("node", defaultRPC, "node RPC address")
	if err := fs.Parse(flagsFirst(args[1:])); err != nil {
		return err
	}
	c, err := rpcClient(*nodeAddr)
	if err != nil {
		return err
	}
	ctx := context.Background()
	switch sub {
	case "balance":
		if fs.NArg() < 1 {
			return fmt.Errorf("usage: ogod query balance <address>")
		}
		res, err := c.ABCIQuery(ctx, "balance", []byte(fs.Arg(0)))
		if err != nil {
			return err
		}
		return printQuery(res)
	case "supply":
		res, err := c.ABCIQuery(ctx, "supply", nil)
		if err != nil {
			return err
		}
		return printQuery(res)
	case "params":
		res, err := c.ABCIQuery(ctx, "params", nil)
		if err != nil {
			return err
		}
		return printQuery(res)
	case "schedule":
		res, err := c.ABCIQuery(ctx, "schedule", nil)
		if err != nil {
			return err
		}
		return printQuery(res)
	case "epochs":
		res, err := c.ABCIQuery(ctx, "epochs", nil)
		if err != nil {
			return err
		}
		return printQuery(res)
	case "global-supply":
		if fs.NArg() < 1 {
			return fmt.Errorf("usage: ogod query global-supply <channel>")
		}
		res, err := c.ABCIQuery(ctx, "global-supply", []byte(fs.Arg(0)))
		if err != nil {
			return err
		}
		return printQuery(res)
	default:
		return fmt.Errorf("unknown query %q", sub)
	}
}

// flagsFirst hoists flag arguments ahead of positional arguments before
// parsing. Go's flag package stops parsing at the first positional, so a
// command like `ogod keys add alice --home X` would silently ignore
// --home (and read the DEFAULT home) without this. Caught by the
// localnet proof run 2026-11-06: add-genesis-account silently read
// ~/.ogod instead of --home. Flag values in this CLI never start with
// "-", so hoisting is safe.
func flagsFirst(args []string) []string {
	var flags, rest []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		if len(a) > 0 && a[0] == '-' {
			flags = append(flags, a)
			// A flag without "=" consumes the next arg as its value;
			// hoist the pair together or --home would swallow the
			// positional (caught 2026-11-06: keys were named after the
			// home path). All flags in this CLI take a value.
			if !strings.Contains(a, "=") && i+1 < len(args) {
				i++
				flags = append(flags, args[i])
			}
		} else {
			rest = append(rest, a)
		}
	}
	return append(flags, rest...)
}

func printQuery(res *ctypes.ResultABCIQuery) error {
	r := res.Response
	if r.Code != 0 {
		return fmt.Errorf("query failed (code %d): %s", r.Code, r.Log)
	}
	var pretty map[string]string
	if json.Unmarshal(r.Value, &pretty) == nil {
		b, _ := json.MarshalIndent(pretty, "", "  ")
		fmt.Println(string(b))
		return nil
	}
	fmt.Println(string(r.Value))
	return nil
}
