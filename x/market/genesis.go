package market

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	
	"github.com/ovrclk/akash/x/market/keeper"
	"github.com/ovrclk/akash/x/market/types"
)

// GenesisState defines the basic genesis state used by market module
type GenesisState struct {
	Orders []types.Order `json:"orders"`
	Bids   []types.Bid   `json:"bids"`
	Leases []types.Lease `json:"leases"`
}

// ValidateGenesis does validation check of the Genesis
func ValidateGenesis(data GenesisState) error {
	return nil
}

// DefaultGenesisState returns default genesis state as raw bytes for the market
// module.
func DefaultGenesisState() GenesisState {
	return GenesisState{}
}

// InitGenesis initiate genesis state and return updated validator details
func InitGenesis(ctx sdk.Context, keeper keeper.Keeper, data GenesisState) []abci.ValidatorUpdate {
	
	for _, order := range data.Orders {
		keeper.CreateOrder(ctx, order.GroupID(), order.Spec)
	}
	
	for _, bid := range data.Bids {
		keeper.CreateBid(ctx, bid.OrderID(), bid.Provider, bid.Price)
	}
	
	for _, leases := range data.Leases {
		bid, _ := keeper.GetBid(ctx, leases.BidID())
		keeper.CreateLease(ctx, bid)
	}
	return []abci.ValidatorUpdate{}
}

// ExportGenesis returns genesis state as raw bytes for the market module
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) GenesisState {
	var orders []types.Order
	var bids []types.Bid
	var leases []types.Lease
	
	k.WithOrders(ctx, func(order types.Order) bool {
		orders = append(orders, order)
		return false
	})
	
	k.WithBids(ctx, func(bid types.Bid) bool {
		bids = append(bids, bid)
		return false
	})
	
	k.WithLeases(ctx, func(lease types.Lease) bool {
		leases = append(leases, lease)
		return false
	})
	
	return GenesisState{orders, bids, leases}
}
