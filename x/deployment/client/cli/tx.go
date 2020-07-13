package cli

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	
	"github.com/ovrclk/akash/sdl"
	"github.com/ovrclk/akash/x/deployment/query"
	"github.com/ovrclk/akash/x/deployment/types"
	
	"github.com/spf13/cobra"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd(key string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Deployment transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	cmd.AddCommand(flags.PostCommands(
		cmdCreate(key, cdc),
		cmdUpdate(key, cdc),
		cmdClose(key, cdc),
	)...)
	return cmd
}

func cmdCreate(key string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create [sdl-file] [count]",
		Short: fmt.Sprintf("Create %s", key),
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.NewCLIContext().WithCodec(cdc)
			bldr := auth.NewTxBuilderFromCLI(os.Stdin).WithTxEncoder(utils.GetTxEncoder(cdc))
			
			var msgs []sdk.Msg
			
			sdl, err := sdl.ReadFile(args[0])
			if err != nil {
				return err
			}
			groups, err := sdl.DeploymentGroups()
			if err != nil {
				return err
			}
			
			msgCount, err := strconv.Atoi(args[1])
			if err != nil {
				return err
			}
			
			seq, err := QueryLastDeploymentID(ctx)
			if err != nil {
				return err
			}
			
			fmt.Println("MSG Count ", seq, msgCount)

			for i := seq; i <= seq+uint64(msgCount); i++ {
				id, err := DeploymentIDFromFlags(cmd.Flags(), ctx.GetFromAddress().String())
				if err != nil {
					return err
				}
				
				// Using Account Sequence in place of Dseq
				accGetter := auth.NewAccountRetriever(ctx)
				acc, err := accGetter.GetAccount(ctx.GetFromAddress())
				if err != nil {
					return err
				}
				
				
				id.DSeq = acc.GetSequence() + uint64(i)
				msg := types.MsgCreateDeployment{
					ID: id,
					// Version:  []byte{0x1, 0x2},
					Groups: make([]types.GroupSpec, 0, len(groups)),
				}
				
				for _, group := range groups {
					msg.Groups = append(msg.Groups, *group)
				}
				
				if err := msg.ValidateBasic(); err != nil {
					return err
				}
				
				msgs = append(msgs, msg)
			}
			
			fmt.Println("MSGS", len(msgs))
			return utils.GenerateOrBroadcastMsgs(ctx, bldr, msgs)
		},
	}
	AddDeploymentIDFlags(cmd.Flags())
	
	return cmd
}

func QueryLastDeploymentID(ctx context.CLIContext) (uint64, error) {
	path := fmt.Sprintf("custom/deployment/deployments/%s/", ctx.GetFromAddress().String())
	
	res, _, err := ctx.QueryWithData(path, nil)
	if err != nil {
		return 0, err
	}
	if len(res) == 0 {
		return 1, nil
	}
	
	var deployments query.Deployments
	if err := ctx.Codec.UnmarshalJSON(res, &deployments); err != nil {
		return 0, nil
	}
	
	if len(deployments) == 0 {
		return 0, nil
	}
	
	fmt.Println("ToTal Deployments", len(deployments))
	
	sort.Slice(deployments, func(i, j int) bool {
		return deployments[i].DSeq < deployments[j].DSeq
	})
	
	fmt.Println("TOTAL DEPLOYMENTS", len(deployments))
	return deployments[len(deployments)-1].DSeq, nil
}

func cmdClose(key string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "close",
		Short: fmt.Sprintf("Close %s", key),
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.NewCLIContext().WithCodec(cdc)
			bldr := auth.NewTxBuilderFromCLI(os.Stdin).WithTxEncoder(utils.GetTxEncoder(cdc))
			
			id, err := DeploymentIDFromFlags(cmd.Flags(), ctx.GetFromAddress().String())
			if err != nil {
				return err
			}
			
			msg := types.MsgCloseDeployment{ID: id}
			
			return utils.GenerateOrBroadcastMsgs(ctx, bldr, []sdk.Msg{msg})
		},
	}
	AddDeploymentIDFlags(cmd.Flags())
	return cmd
}

func cmdUpdate(key string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update [sdl-file]",
		Short: fmt.Sprintf("update %s", key),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.NewCLIContext().WithCodec(cdc)
			bldr := auth.NewTxBuilderFromCLI(os.Stdin).WithTxEncoder(utils.GetTxEncoder(cdc))
			
			id, err := DeploymentIDFromFlags(cmd.Flags(), ctx.GetFromAddress().String())
			if err != nil {
				return err
			}
			
			msg := types.MsgUpdateDeployment{
				ID: id,
			}
			
			return utils.GenerateOrBroadcastMsgs(ctx, bldr, []sdk.Msg{msg})
		},
	}
	AddDeploymentIDFlags(cmd.Flags())
	return cmd
}
