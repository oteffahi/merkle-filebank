package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "filebankd",
	Short: "A file storage system based on merkle trees",
	Long: `MerkleFileBank is a CLI tool for secure file storage on servers.
Files are encrypted before upload to server, and merkle trees are used to guarantee file intergrity after download from server`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	userHome, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Errorf("Error: cannot get user home directory path\n"))
	}
	rootCmd.PersistentFlags().String("home", userHome+"/.filebankd", "root directory for MerkleFileBank storage")
}
