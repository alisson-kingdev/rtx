package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"rtx/src/controllers"
	"rtx/src/utils"
)

func newRotateCmd() *cobra.Command {
	var opts rotationOptions

	cmd := &cobra.Command{
		Use:   "rotate",
		Short: "Executar rotação de IP",
		Long:  "Rotaciona o IP usando proxies públicas coletadas de APIs gratuitas, com fallback para APIs diretas.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			out := cmd.OutOrStdout()
			if isVerbose(cmd) {
				fmt.Fprintln(out, utils.Info("Modo verboso ativado"))
			}

			fmt.Fprintln(out, utils.Banner())
			fmt.Fprintln(out, utils.Section("Obtendo IP antigo (IP real)"))

			result, err := controllers.RotateIP(cmd.Context(), opts.config())
			if err != nil {
				return fmt.Errorf("executar rotação: %w", err)
			}

			fmt.Fprintln(out, utils.Success(fmt.Sprintf("IP Antigo: %s", result.OldIP)))
			fmt.Fprintln(out, utils.Section("Rotacionando IP via proxies públicas"))
			fmt.Fprintln(out, utils.Colorize(utils.CyanColor, "========================================"))
			fmt.Fprintln(out, utils.Colorize(utils.MagentaColor, "        RESULTADO DA ROTAÇÃO"))
			proxyDisplay := result.Proxy
			if strings.Contains(proxyDisplay, "://") {
				proxyDisplay = utils.RedactURL(proxyDisplay)
			}
			fmt.Fprintln(out, utils.Colorize(utils.CyanColor, "========================================"))
			fmt.Fprint(out, utils.PrintKV("Proxy URL", proxyDisplay))
			fmt.Fprint(out, utils.PrintKV("Proxy Source", result.Source))
			fmt.Fprint(out, utils.PrintKV("IP Antigo", result.OldIP))
			fmt.Fprint(out, utils.PrintKV("IP Novo", result.NewIP))
			fmt.Fprintln(out, utils.Colorize(utils.CyanColor, "========================================"))

			if result.Changed {
				fmt.Fprintln(out, utils.Success("IP rotacionado com sucesso!"))
			} else {
				fmt.Fprintln(out, utils.Warning("O IP não mudou após a rotação."))
			}
			fmt.Fprintln(out, utils.Colorize(utils.MagentaColor, "[*] Script finalizado."))
			return nil
		},
	}

	addRotationFlags(cmd, &opts)
	return cmd
}
