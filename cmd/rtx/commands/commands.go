package commands

import (
	"time"

	"github.com/spf13/cobra"

	"rtx/src/services"
)

// rotationOptions holds CLI-bound limits shared by commands that test proxies.
type rotationOptions struct {
	timeout     time.Duration
	maxProxies  int
	concurrency int
}

func addRotationFlags(cmd *cobra.Command, opts *rotationOptions) {
	defaults := services.DefaultRotationConfig()
	cmd.Flags().DurationVar(&opts.timeout, "timeout", defaults.Timeout, "Tempo limite total da operação")
	cmd.Flags().IntVar(&opts.maxProxies, "max-proxies", defaults.MaxProxies, "Número máximo de proxies testados")
	cmd.Flags().IntVar(&opts.concurrency, "concurrency", defaults.Concurrency, "Quantidade de testes de proxy em paralelo")
}

func (o *rotationOptions) config() services.RotationConfig {
	return services.RotationConfig{
		Timeout:     o.timeout,
		MaxProxies:  o.maxProxies,
		Concurrency: o.concurrency,
	}
}

func isVerbose(cmd *cobra.Command) bool {
	value, err := cmd.Flags().GetBool("verbose")
	return err == nil && value
}

// Register adds all RTX subcommands to the root command.
func Register(rootCmd *cobra.Command) {
	rootCmd.AddCommand(newRotateCmd())
	rootCmd.AddCommand(newSetProxyCmd())
	rootCmd.AddCommand(newUnsetProxyCmd())
}
