package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"rtx/src/controllers"
	"rtx/src/utils"
)

var rotateCmd = &cobra.Command{
	Use:   "rotate",
	Short: "Executar rotacao de IP",
	Long:  "Rotaciona o IP usando proxies publicas coletadas de APIs",
	Run: func(cmd *cobra.Command, args []string) {
		isVerbose, _ := cmd.Flags().GetBool("verbose")
		if isVerbose {
			fmt.Println(utils.Info("Modo verboso ativado"))
		}

		fmt.Println(utils.Banner())

		fmt.Println(utils.Section("Obtendo IP antigo (IP real)"))
		result := controllers.RotateIP()

		if result.Error != nil {
			fmt.Println(utils.Error(fmt.Sprintf("Erro ao executar rotacao: %v", result.Error)))
			os.Exit(1)
		}

		fmt.Println(utils.Success(fmt.Sprintf("IP Antigo: %s", result.OldIP)))

		fmt.Println(utils.Section("Rotacionando IP via proxies publicas"))

		fmt.Println(utils.Colorize(utils.CyanColor, "========================================"))
		fmt.Println(utils.Colorize(utils.MagentaColor, "        RESULTADO DA ROTACAO"))
		fmt.Println(utils.Colorize(utils.CyanColor, "========================================"))
		utils.PrintKV("Proxy URL", result.Proxy)
		utils.PrintKV("Proxy Source", result.Source)
		utils.PrintKV("IP Antigo", result.OldIP)
		utils.PrintKV("IP Novo", result.NewIP)
		fmt.Println(utils.Colorize(utils.CyanColor, "========================================"))

		if result.Changed {
			fmt.Println(utils.Success("IP rotacionado com sucesso!"))
		} else {
			fmt.Println(utils.Warning("O IP nao mudou apos a rotacao."))
		}

		fmt.Println(utils.Colorize(utils.MagentaColor, "[*] Script finalizado."))
	},
}
