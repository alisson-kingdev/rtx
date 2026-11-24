package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"rtx/src/services"
	"rtx/src/utils"
)

var setProxyCmd = &cobra.Command{
	Use:   "set-proxy [PROXY_URL | default]",
	Short: "Configura proxy no ambiente do sistema",
	Long:  "Define variaveis de ambiente de proxy (http_proxy, https_proxy, etc.) nos arquivos de configuracao do shell (.bashrc, .zshrc, config.fish, etc.) e, quando aplicavel, configuracoes de rede do sistema (Windows WinHTTP, macOS networksetup).\n\nUse 'default' para que o RTX colete automaticamente um proxy funcional das APIs publicas (gimmeproxy, hproxy, proxyscraper) e o configure para voce.\n\nSe ja existir um bloco RTX anterior, ele e removido antes de gravar o novo, evitando duplicatas/loops.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		proxyURL := args[0]

		fmt.Println(utils.Banner())

		if proxyURL == "default" {
			fmt.Println(utils.Section("Coletando proxies publicas automaticamente"))
			fmt.Println(utils.Info("Consultando APIs: gimmeproxy, hproxy, proxyscraper..."))

			proxies := services.CollectProxies()
			if len(proxies) == 0 {
				fmt.Println(utils.Error("Nenhum proxy coletado das APIs publicas."))
				fmt.Println(utils.Warning("Tente novamente ou informe um proxy manualmente: rtx set-proxy http://host:porta"))
				os.Exit(1)
			}

			fmt.Println(utils.Success(fmt.Sprintf("%d proxies coletados. Testando cada um...", len(proxies))))

			var selectedProxy string
			var selectedIP string
			var selectedSource string

			for _, p := range proxies {
				fmt.Printf("   Testando %s (fonte: %s) ... ", p.URL, p.Source)
				ip, err := services.GetPublicIPViaProxy(p.URL)
				if err != nil || ip == "" {
					fmt.Println(utils.RedColor.Sprint("FALHOU"))
					continue
				}
				fmt.Println(utils.GreenColor.Sprint(fmt.Sprintf("OK -> IP: %s", ip)))
				selectedProxy = p.URL
				selectedIP = ip
				selectedSource = p.Source
				break
			}

			if selectedProxy == "" {
				fmt.Println(utils.Error("Nenhum proxy funcional encontrado no momento."))
				fmt.Println(utils.Warning("Proxies publicas podem estar offline. Tente novamente mais tarde ou informe um proxy manualmente."))
				os.Exit(1)
			}

			proxyURL = selectedProxy

			fmt.Println()
			fmt.Println(utils.Section("Proxy selecionado automaticamente"))
			utils.PrintKV("Proxy URL", selectedProxy)
			utils.PrintKV("Fonte", selectedSource)
			utils.PrintKV("IP via proxy", selectedIP)
		} else {
			fmt.Println(utils.Section("Configurando proxy informado"))
			utils.PrintKV("Proxy URL", proxyURL)
		}

		result := services.SetSystemProxy(proxyURL)

		if result.Error != nil {
			fmt.Println(utils.Error(fmt.Sprintf("Erro: %v", result.Error)))
			os.Exit(1)
		}

		if result.Success {
			fmt.Println(utils.Success("Proxy configurado com sucesso no sistema!"))
			utils.PrintKV("Shell", result.Shell)
			utils.PrintKV("Acao", result.Action)
			fmt.Println(utils.Info("Arquivos atualizados:"))
			for _, f := range result.FilesUpdated {
				fmt.Printf("   %s%s\n", utils.Colorize(utils.CyanColor, "  - "), f)
			}
			fmt.Println()
			fmt.Println(utils.Warning("Reinicie seu terminal ou execute: source ~/.bashrc (ou ~/.zshrc / config.fish) para aplicar."))
			fmt.Println(utils.Info("Em seguida, use: rtx rotate  (para rotacionar o IP via outro proxy)"))
		}

		fmt.Println(utils.Colorize(utils.MagentaColor, "[*] Script finalizado."))
	},
}

var unsetProxyCmd = &cobra.Command{
	Use:   "unset-proxy",
	Short: "Remove configuracao de proxy do ambiente",
	Long:  "Remove blocos de proxy gerados pelo RTX dos arquivos de configuracao do shell e reseta configuracoes de rede (Windows WinHTTP, macOS networksetup).",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(utils.Banner())

		fmt.Println(utils.Section("Removendo proxy do sistema"))

		result := services.UnsetSystemProxy()

		if result.Error != nil {
			fmt.Println(utils.Error(fmt.Sprintf("Erro: %v", result.Error)))
			os.Exit(1)
		}

		if result.Success {
			fmt.Println(utils.Success("Proxy removido do sistema!"))
			utils.PrintKV("Shell", result.Shell)
			utils.PrintKV("Acao", result.Action)
			if len(result.FilesUpdated) > 0 {
				fmt.Println(utils.Info("Arquivos atualizados:"))
				for _, f := range result.FilesUpdated {
					fmt.Printf("   %s%s\n", utils.Colorize(utils.CyanColor, "  - "), f)
				}
			} else {
				fmt.Println(utils.Info("Nenhum bloco RTX encontrado para remover."))
			}
			fmt.Println(utils.Warning("Reinicie seu terminal para aplicar."))
		}

		fmt.Println(utils.Colorize(utils.MagentaColor, "[*] Script finalizado."))
	},
}
