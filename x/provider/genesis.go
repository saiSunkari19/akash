package provider

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	
	"github.com/ovrclk/akash/x/provider/keeper"
	"github.com/ovrclk/akash/x/provider/types"
)

// GenesisState defines the basic genesis state used by provider module
type GenesisState struct {
	Providers []types.Provider `json:"providers"`
}

// ValidateGenesis does validation check of the Genesis and returns error incase of failure
func ValidateGenesis(data GenesisState) error {
	return nil
}

// InitGenesis initiate genesis state and return updated validator details
func InitGenesis(ctx sdk.Context, keeper keeper.Keeper, data GenesisState) []abci.ValidatorUpdate {
	for _, provider := range data.Providers {
		keeper.Create(ctx, provider)
	}
	
	return []abci.ValidatorUpdate{}
}

// ExportGenesis returns genesis state as raw bytes for the provider module
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) GenesisState {
	var providers []types.Provider
	
	k.WithProviders(ctx, func(provider types.Provider) bool {
		providers = append(providers, provider)
		return false
	})
	
	return GenesisState{Providers: providers}
}

// DefaultGenesisState returns default genesis state as raw bytes for the provider
// module.
func DefaultGenesisState() GenesisState {
	return GenesisState{}
}
