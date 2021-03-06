package cli

import (
	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	dcli "github.com/ovrclk/akash/x/deployment/client/cli"
	"github.com/ovrclk/akash/x/market/query"
	"github.com/ovrclk/akash/x/market/types"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// AddOrderIDFlags add flags for order
func AddOrderIDFlags(flags *pflag.FlagSet) {
	dcli.AddGroupIDFlags(flags)
	flags.Uint32("oseq", 0, "Order Sequence")
}

// MarkReqOrderIDFlags marks flags required for order
func MarkReqOrderIDFlags(cmd *cobra.Command) {
	dcli.MarkReqGroupIDFlags(cmd)
	cmd.MarkFlagRequired("oseq")
}

// AddProviderFlag add provider flag to command flags set
func AddProviderFlag(flags *pflag.FlagSet) {
	flags.String("provider", "", "Provider")
}

// MarkReqProviderFlag marks provider flag as required
func MarkReqProviderFlag(cmd *cobra.Command) {
	_ = cmd.MarkFlagRequired("provider")
}

// OrderIDFromFlags returns OrderID with given flags and error if occurred
func OrderIDFromFlags(flags *pflag.FlagSet) (types.OrderID, error) {
	prev, err := dcli.GroupIDFromFlags(flags)
	if err != nil {
		return types.OrderID{}, err
	}
	val, err := flags.GetUint32("oseq")
	if err != nil {
		return types.OrderID{}, err
	}
	return types.MakeOrderID(prev, val), nil
}

// AddBidIDFlags add flags for bid
func AddBidIDFlags(flags *pflag.FlagSet) {
	AddOrderIDFlags(flags)
	AddProviderFlag(flags)
}

// AddQueryBidIDFlags add flags for bid in query commands
func AddQueryBidIDFlags(flags *pflag.FlagSet) {
	AddBidIDFlags(flags)
}

// MarkReqBidIDFlags marks flags required for bid
// Used in get bid query command
func MarkReqBidIDFlags(cmd *cobra.Command) {
	MarkReqOrderIDFlags(cmd)
	MarkReqProviderFlag(cmd)
}

// BidIDFromFlags returns BidID with given flags and error if occurred
func BidIDFromFlags(ctx context.CLIContext, flags *pflag.FlagSet) (types.BidID, error) {
	prev, err := OrderIDFromFlags(flags)
	if err != nil {
		return types.BidID{}, err
	}
	return types.MakeBidID(prev, ctx.GetFromAddress()), nil
}

// BidIDFromFlagsWithoutCtx returns BidID with given flags and error if occurred
// Here provider value is taken from flags
func BidIDFromFlagsWithoutCtx(flags *pflag.FlagSet) (types.BidID, error) {
	prev, err := OrderIDFromFlags(flags)
	if err != nil {
		return types.BidID{}, err
	}
	provider, err := flags.GetString("provider")
	if err != nil {
		return types.BidID{}, err
	}
	addr, err := sdk.AccAddressFromBech32(provider)
	if err != nil {
		return types.BidID{}, err
	}
	return types.MakeBidID(prev, addr), nil
}

// ProviderFromFlagsWithoutCtx returns Provider address with given flags and error if occurred
func ProviderFromFlagsWithoutCtx(flags *pflag.FlagSet) (sdk.AccAddress, error) {
	provider, err := flags.GetString("provider")
	if err != nil {
		return sdk.AccAddress{}, err
	}
	addr, err := sdk.AccAddressFromBech32(provider)
	if err != nil {
		return sdk.AccAddress{}, err
	}
	return addr, nil
}

// AddOrderFilterFlags add flags to filter for order list
func AddOrderFilterFlags(flags *pflag.FlagSet) {
	flags.String("owner", "", "order owner address to filter")
	flags.String("state", "", "order state to filter (open,matched,closed)")
}

// OrderFiltersFromFlags returns OrderFilters with given flags and error if occurred
func OrderFiltersFromFlags(flags *pflag.FlagSet) (query.OrderFilters, error) {
	gfilters, err := dcli.GroupFiltersFromFlags(flags)
	if err != nil {
		return query.OrderFilters{}, err
	}
	ofilters := query.OrderFilters{
		Owner:        gfilters.Owner,
		StateFlagVal: gfilters.StateFlagVal,
	}
	return ofilters, nil
}

// AddBidFilterFlags add flags to filter for bid list
func AddBidFilterFlags(flags *pflag.FlagSet) {
	flags.String("owner", "", "bid owner address to filter")
	flags.String("state", "", "bid state to filter (open,matched,lost,closed)")
}

// BidFiltersFromFlags returns BidFilters with given flags and error if occurred
func BidFiltersFromFlags(flags *pflag.FlagSet) (query.BidFilters, error) {
	ofilters, err := OrderFiltersFromFlags(flags)
	if err != nil {
		return query.BidFilters{}, err
	}
	bfilters := query.BidFilters{
		Owner:        ofilters.Owner,
		StateFlagVal: ofilters.StateFlagVal,
	}
	return bfilters, nil
}

// AddLeaseFilterFlags add flags to filter for lease list
func AddLeaseFilterFlags(flags *pflag.FlagSet) {
	flags.String("owner", "", "lease owner address to filter")
	flags.String("state", "", "lease state to filter (active,insufficient-funds,closed)")
}

// LeaseFiltersFromFlags returns LeaseFilters with given flags and error if occurred
func LeaseFiltersFromFlags(flags *pflag.FlagSet) (query.LeaseFilters, error) {
	ofilters, err := OrderFiltersFromFlags(flags)
	if err != nil {
		return query.LeaseFilters{}, err
	}
	lfilters := query.LeaseFilters{
		Owner:        ofilters.Owner,
		StateFlagVal: ofilters.StateFlagVal,
	}
	return lfilters, nil
}
