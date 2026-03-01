/*
Copyright © 2026 Abdoulaye BALDE <[EMAIL_ADDRESS]>
*/
package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

// trackCmd represents the track command
var trackCmd = &cobra.Command{
	Use:   "track",
	Short: "track les depenances systeme d'une commande",
	Long:  `Avec cette commande, tu peux tracker les depenances systeme d'une commande`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Erreur : Veuillez spécifier une commande a tracker (ex : enclos track ls -l)")
			os.Exit(1)
		}

		straceArgs := []string{"-f", "-e", "trace=execve"}
		straceArgs = append(straceArgs, args...)

		sysCmd := exec.Command("strace", straceArgs...)
		sysCmd.Stdout = os.Stdout
		var stderrBuffer bytes.Buffer
		sysCmd.Stderr = &stderrBuffer
		fmt.Println("tracking en cours pour la commande : ", args)
		fmt.Println("------------------------------------------------------------")

		err := sysCmd.Run()
		fmt.Println("analyse des dependances ... ")
		traceText := stderrBuffer.String()
		lines := strings.Split(traceText, "\n")

		dependancesUniques := make(map[string]bool)

		for _, line := range lines {
			if strings.Contains(line, "execve(") {
				parts := strings.Split(line, `"`)

				if len(parts) >= 3 {
					binaryPath := parts[1]
					dependancesUniques[binaryPath] = true
				}
			}
		}

		file, errfile := os.Create("enclave.lock")
		if errfile != nil {
			fmt.Println("erreur lors de la creation du fichier enclave.lock", errfile)
		} else {
			defer file.Close()
			file.WriteString("#fichier generer automatiquement par enclos \n")
			file.WriteString("depenencies:\n")
			for chemin := range dependancesUniques {
				file.WriteString(fmt.Sprintf("  -%s\n", chemin))
			}
			fmt.Println("fichier enclave.lock generer avec succes !")
		}

		if err != nil {
			fmt.Println("Erreur lors de l'execution de la commande : ", err)
			os.Exit(1)
		}
		fmt.Println("\n tracking termine avec succes")

	},
}

func init() {
	rootCmd.AddCommand(trackCmd)
	trackCmd.DisableFlagParsing = true

	// Here you will define your flags and configurati	on settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// trackCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// trackCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
