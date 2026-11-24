package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/spf13/cobra"

	"rtx/cmd/rtx/commands"
	"rtx/src/utils"
)

// version can be overridden at build time with -ldflags "-X main.version=...".
var version = utils.GetVersion()

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "rtx",
		Short:         "RTX - Rotacionador de IP usando proxies públicas",
		Long:          utils.Banner(),
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
	root.PersistentFlags().BoolP("verbose", "v", false, "Ativar modo verboso")
	commands.Register(root)
	return root
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := newRootCmd().ExecuteContext(ctx); err != nil {
		fmt.Fprintln(os.Stderr, utils.Error(err.Error()))
		os.Exit(1)
	}
}
