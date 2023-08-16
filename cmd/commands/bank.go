package cmd

import (
	"github.com/spf13/cobra"
)

var bankCmd = &cobra.Command{
	Use:   "bank",
	Short: "Manage banks",
	Long: `- Create new bank on server
- Download a file from a bank on server`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(bankCmd)
}
