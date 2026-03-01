/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

// dockerCmd represents the docker command
var dockerCmd = &cobra.Command{
	Use:   "docker",
	Short: "genere le Dockerfile",
	Long: `docker permet d'extraire des informations dans l'enclave.lock
	 pour generer le Dockerfile pret a l'emploi`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
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
