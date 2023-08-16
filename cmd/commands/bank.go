package cmd

import (
	"fmt"

	"github.com/oteffahi/merkle-filebank/client"
	"github.com/oteffahi/merkle-filebank/storage"
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

var createBankCmd = &cobra.Command{
	Use:   "create [flags] [paths...]",
	Short: "Create new bank on server",
	Long: `Encrypts files, generates merkle tree, uploads files to server, saves merkle root and cryptographic parameters.

Args:
  paths: Space-seperated paths to files or directories. Files will be added recursively from directories.
         Does not support regular expressions.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Printf("Missing positional arguments: at least one filepath is required\n\n")
			cmd.Help()
			return
		}

		serverName, err := cmd.Flags().GetString("server")
		if err != nil {
			fmt.Printf("%v\n\n", err)
			cmd.Help()
			return
		}
		if serverName == "" {
			fmt.Printf("Missing flag: server flag is required\n\n")
			cmd.Help()
			return
		}

		bankName, err := cmd.Flags().GetString("bank-name")
		if err != nil {
			fmt.Printf("%v\n\n", err)
			cmd.Help()
			return
		}
		if bankName == "" {
			fmt.Printf("Missing flag: bank-name flag is required\n\n")
			cmd.Help()
			return
		}

		homepath, err := getHomePath(cmd)
		if err != nil {
			fmt.Println(err)
			return
		}

		var paths []string
		for _, arg := range args {
			content, err := storage.GetAllFilesPaths(arg)
			if err != nil {
				fmt.Printf("Error while processing positional argument %v:\n%v\n\n", arg, err)
				cmd.Help()
				return
			}
			paths = append(paths, content...)
		}

		if err := client.CallUploadFiles(homepath, serverName, bankName, paths); err != nil {
			fmt.Println(err)
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(bankCmd)
	bankCmd.AddCommand(createBankCmd)

	bankCmd.PersistentFlags().StringP("bank-name", "b", "", "unique local name for the filebank")
	bankCmd.PersistentFlags().StringP("server", "s", "", "unique local name for the server")
}
