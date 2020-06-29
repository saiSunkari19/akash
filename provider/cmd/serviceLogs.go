package cmd

import (
	"context"
	"fmt"
	"net/url"

	ccontext "github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/ovrclk/akash/provider/gateway"
	mcli "github.com/ovrclk/akash/x/market/client/cli"
	mtypes "github.com/ovrclk/akash/x/market/types"
	pmodule "github.com/ovrclk/akash/x/provider"
)

func serviceLogsCmd(codec *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "service-logs",
		Short: "get service status",
		RunE: func(cmd *cobra.Command, args []string) error {
			return doServiceLogs(codec, cmd)
		},
	}

	mcli.AddBidIDFlags(cmd.Flags())
	mcli.MarkReqBidIDFlags(cmd)

	cmd.Flags().String("service", "", "")
	_ = cmd.MarkFlagRequired("service")

	cmd.Flags().BoolP("follow", "f", false, "Specify if the logs should be streamed. Defaults to false")
	cmd.Flags().Int64P("tail", "t", 0, "the number of lines from the end of the logs to show")
	cmd.Flags().String("format", "text", "Output format text|json. Defaults to text")
	return cmd
}

func doServiceLogs(codec *codec.Codec, cmd *cobra.Command) error {
	cctx := ccontext.NewCLIContext().WithCodec(codec)

	addr, err := mcli.ProviderFromFlagsWithoutCtx(cmd.Flags())
	if err != nil {
		return err
	}

	var svcName string
	if svcName, err = cmd.Flags().GetString("service"); err != nil {
		return err
	}

	var outputFormat string
	if outputFormat, err = cmd.Flags().GetString("format"); err != nil {
		return err
	}

	if outputFormat != "text" && outputFormat != "json" {
		return errors.Errorf("invalid output format %s", outputFormat)
	}

	pclient := pmodule.AppModuleBasic{}.GetQueryClient(cctx)
	provider, err := pclient.Provider(addr)
	if err != nil {
		return err
	}

	gclient := gateway.NewClient()

	bid, err := mcli.BidIDFromFlagsWithoutCtx(cmd.Flags())
	if err != nil {
		return err
	}

	lid := mtypes.MakeLeaseID(bid)

	follow, err := cmd.Flags().GetBool("follow")
	if err != nil {
		return err
	}

	var tailLines *int64

	if cmd.Flags().Changed("tail") {
		vl := new(int64)
		if *vl, err = cmd.Flags().GetInt64("tail"); err != nil {
			return err
		}

		tailLines = vl
	}

	uri, err := url.Parse(provider.HostURI)
	if err != nil {
		return err
	}

	switch uri.Scheme {
	case "ws", "http", "":
		uri.Scheme = "ws"
	case "wss", "https":
		uri.Scheme = "wss"
	default:
		return errors.Errorf("invalid uri scheme \"%s\"", uri.Scheme)
	}

	result, err := gclient.ServiceLogs(context.Background(), uri.String(), lid, svcName, follow, tailLines)
	if err != nil {
		return err
	}

	for res := range result.Stream {
		if outputFormat == "json" {
			if err = cctx.PrintOutput(res); err != nil {
				return err
			}
		} else {
			fmt.Printf("[%s] %s\n", res.Name, res.Message)
		}
	}

	return nil
}
