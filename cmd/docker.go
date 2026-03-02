/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// dockerCmd represents the docker command
var dockerCmd = &cobra.Command{
	Use:   "docker",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		contenu, err := os.ReadFile("enclave.lock")
		if err != nil {
			fmt.Println("impossible de lire enclave.lock . as tu lancé enclos track")
			os.Exit(1)
		}

		texte := string(contenu)

		traducteur := map[string]string{
			"node":          "nodejs",
			"node22":        "nodejs",
			"npm":           "npm",
			"awk":           "gawk",
			"react-scripts": "",
		}

		paquetsAInstaller := make(map[string]bool)
		lignes := strings.Split(texte, "\n")

		for _, ligne := range lignes {

			if !strings.Contains(ligne, " - ") {
				continue
			}

			chemin := strings.TrimSpace(strings.ReplaceAll(ligne, "- ", ""))

			morceaux := strings.Split(chemin, "/")
			binaire := morceaux[len(morceaux)-1]

			if binaire == "sh" || binaire == "ls" || binaire == "ps" || binaire == "sed" || binaire == "lsof" {
				continue
			}
			nomPaquet, estConnu := traducteur[binaire]
			if estConnu && nomPaquet == "" {

			}

			if !estConnu {
				nomPaquet = binaire
			}

			paquetsAInstaller[nomPaquet] = true

		}

		dockerfile, _ := os.Create("Dockerfile")
		defer dockerfile.Close()

		dockerfile.WriteString("FROM ubuntu:22.04 \n\n")
		ligneInstall := "RUN apt-get update && apt-get install -y"

		for paquet := range paquetsAInstaller {
			ligneInstall += " " + paquet
		}

		dockerfile.WriteString(ligneInstall + "\n\n")
		dockerfile.WriteString("WORKDIR /app\n")
		dockerfile.WriteString("COPY . .\n")
		fmt.Println("fichier Dockerfile cree avec success")

	},
}

func init() {
	exportCmd.AddCommand(dockerCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// dockerCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// dockerCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
