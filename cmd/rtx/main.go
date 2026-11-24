package main

import (
	"os"

	"github.com/spf13/cobra"

	"rtx/cmd/rtx/commands"
	"rtx/src/utils"
)

var (
	verbose bool
)

var rootCmd = &cobra.Command{
	Use:   "rtx",
	Short: "RTX - Rotacionador de IP usando proxies publicas",
	Long:  utils.Banner(),
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		utils.RedColor.Print(err.Error())
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Ativar modo verboso")
	commands.Register(rootCmd)
}
