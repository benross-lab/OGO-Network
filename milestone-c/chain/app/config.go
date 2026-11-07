package app

import (
	"os"
	"path/filepath"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DefaultNodeHome is the default home directory for ogod.
var DefaultNodeHome string

func init() {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	DefaultNodeHome = filepath.Join(userHomeDir, ".ogod")

	// Bech32 prefixes for the OGO Network (local staging chain).
	// The account prefix "ogo" matches the auth module config in
	// app_config.go and the prototype's ogo1... addresses.
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("ogo", "ogopub")
	config.SetBech32PrefixForValidator("ogovaloper", "ogovaloperpub")
	config.SetBech32PrefixForConsensusNode("ogovalcons", "ogovalconspub")
	config.SetCoinType(118)
}
