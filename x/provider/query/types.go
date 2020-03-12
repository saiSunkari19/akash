package query

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ovrclk/akash/x/provider/types"
)

type (
	// Provider type
	Provider types.Provider
	// Providers - Slice of Provider Struct
	Providers []Provider
)

func (p Provider) String() string {
	return "TODO see deployment/query/types.go"
}

func (p Providers) String() string {
	return "TODO see deployment/query/types.go"
}

// Address returns the provider owner account address
func (p *Provider) Address() sdk.AccAddress {
	return p.Owner
}
