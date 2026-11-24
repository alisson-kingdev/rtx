package commands

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"rtx/src/services"
	"rtx/src/utils"
)

func newSetProxyCmd() *cobra.Command {
	var opts rotationOptions

	cmd := &cobra.Command{
		Use:   "set-proxy [PROXY_URL | default]",
		Short: "Configura proxy no ambiente do sistema",
		Long: "Define variáveis de ambiente de proxy nos arquivos de configuração do shell e, " +
			"quando aplicável, nas configurações de rede do sistema.\n\n" +
			"Use 'default' para coletar automaticamente um proxy funcional das APIs públicas " +
			"gratuitas (gimmeproxy, hproxy, proxyscraper).",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			if isVerbose(cmd) {
				fmt.Fprintln(out, utils.Info("Modo verboso ativado"))
			}
			fmt.Fprintln(out, utils.Banner())

			proxyURL := args[0]
			if proxyURL == "default" {
				config := opts.config()
				ctx, cancel := context.WithTimeout(cmd.Context(), config.Timeout)
				defer cancel()

				selected, err := selectPublicProxy(ctx, out, config.MaxProxies)
				if err != nil {
					return err
				}
				proxyURL = selected
			} else {
				normalized, err := utils.NormalizeProxyURL(proxyURL)
				if err != nil {
					return fmt.Errorf("proxy inválido: %w", err)
				}
				proxyURL = normalized
				fmt.Fprintln(out, utils.Section("Configurando proxy informado"))
				fmt.Fprint(out, utils.PrintKV("Proxy URL", utils.RedactURL(proxyURL)))
			}

			result, err := services.SetSystemProxy(cmd.Context(), proxyURL)
			if err != nil {
				return fmt.Errorf("configurar proxy: %w", err)
			}

			fmt.Fprintln(out, utils.Success("Proxy configurado com sucesso no sistema!"))
			fmt.Fprint(out, utils.PrintKV("Shell", result.Shell))
			fmt.Fprint(out, utils.PrintKV("Ação", result.Action))
			fmt.Fprint(out, utils.PrintKV("Proxy", result.ProxyURL))
			fmt.Fprintln(out, utils.Info("Arquivos atualizados:"))
			for _, file := range result.FilesUpdated {
				fmt.Fprintf(out, "   %s%s\n", utils.Colorize(utils.CyanColor, "  - "), file)
			}
			fmt.Fprintln(out)
			fmt.Fprintln(out, utils.Warning("Reinicie o terminal ou execute: source ~/.bashrc (ou ~/.zshrc / config.fish)"))
			fmt.Fprintln(out, utils.Info("Em seguida, use: rtx rotate"))
			fmt.Fprintln(out, utils.Colorize(utils.MagentaColor, "[*] Script finalizado."))
			return nil
		},
	}

	addRotationFlags(cmd, &opts)
	return cmd
}

func selectPublicProxy(ctx context.Context, out io.Writer, maxProxies int) (string, error) {
	fmt.Fprintln(out, utils.Section("Coletando proxies públicas automaticamente"))
	fmt.Fprintln(out, utils.Info("Consultando APIs gratuitas: gimmeproxy, hproxy, proxyscraper..."))

	proxies, err := services.CollectProxies(ctx)
	if err != nil {
		return "", fmt.Errorf("coletar proxies públicas: %w", err)
	}
	if maxProxies > 0 && len(proxies) > maxProxies {
		proxies = proxies[:maxProxies]
	}
	fmt.Fprintln(out, utils.Success(fmt.Sprintf("%d proxies coletados. Testando cada um...", len(proxies))))

	for _, proxy := range proxies {
		if err := ctx.Err(); err != nil {
			return "", err
		}

		displayURL := utils.RedactURL(proxy.URL)
		fmt.Fprintf(out, "   Testando %s (fonte: %s) ... ", displayURL, proxy.Source)

		ip, err := services.GetPublicIPViaProxy(ctx, proxy.URL)
		if err != nil || ip == "" {
			fmt.Fprintln(out, utils.RedColor.Sprint("FALHOU"))
			continue
		}

		fmt.Fprintln(out, utils.GreenColor.Sprint(fmt.Sprintf("OK -> IP: %s", ip)))
		fmt.Fprintln(out)
		fmt.Fprintln(out, utils.Section("Proxy selecionado automaticamente"))
		fmt.Fprint(out, utils.PrintKV("Proxy URL", displayURL))
		fmt.Fprint(out, utils.PrintKV("Fonte", proxy.Source))
		fmt.Fprint(out, utils.PrintKV("IP via proxy", ip))
		return proxy.URL, nil
	}

	return "", errors.New("nenhum proxy funcional encontrado; tente novamente ou informe um proxy manualmente")
}

func newUnsetProxyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "unset-proxy",
		Short: "Remove configuração de proxy do ambiente",
		Long:  "Remove blocos gerados pelo RTX dos arquivos de configuração do shell e reseta as configurações de rede do sistema.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			out := cmd.OutOrStdout()
			if isVerbose(cmd) {
				fmt.Fprintln(out, utils.Info("Modo verboso ativado"))
			}
			fmt.Fprintln(out, utils.Banner())
			fmt.Fprintln(out, utils.Section("Removendo proxy do sistema"))

			result, err := services.UnsetSystemProxy(cmd.Context())
			if err != nil {
				return fmt.Errorf("remover proxy: %w", err)
			}

			fmt.Fprintln(out, utils.Success("Proxy removido do sistema!"))
			fmt.Fprint(out, utils.PrintKV("Shell", result.Shell))
			fmt.Fprint(out, utils.PrintKV("Ação", result.Action))
			if len(result.FilesUpdated) > 0 {
				fmt.Fprintln(out, utils.Info("Arquivos atualizados:"))
				for _, file := range result.FilesUpdated {
					fmt.Fprintf(out, "   %s%s\n", utils.Colorize(utils.CyanColor, "  - "), file)
				}
			} else {
				fmt.Fprintln(out, utils.Info("Nenhum bloco RTX encontrado para remover."))
			}
			fmt.Fprintln(out, utils.Warning("Reinicie o terminal para aplicar."))
			fmt.Fprintln(out, utils.Colorize(utils.MagentaColor, "[*] Script finalizado."))
			return nil
		},
	}
}
