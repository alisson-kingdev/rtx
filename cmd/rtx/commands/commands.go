package commands

import "github.com/spf13/cobra"

func Register(rootCmd *cobra.Command) {
	rotateCmd.Flags().BoolP("verbose", "v", false, "Ativar modo verboso")
	setProxyCmd.Flags().BoolP("verbose", "v", false, "Ativar modo verboso")
	unsetProxyCmd.Flags().BoolP("verbose", "v", false, "Ativar modo verboso")
	rootCmd.AddCommand(rotateCmd)
	rootCmd.AddCommand(setProxyCmd)
	rootCmd.AddCommand(unsetProxyCmd)
}
