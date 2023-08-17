package cmd

import (
	"fmt"
	"strconv"

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
		// verify home directory
		ok, err := storage.IsHomeWellFormed(homepath)
		if err != nil {
			fmt.Println(err)
			return
		} else if !ok {
			fmt.Printf("Home %v does not exist or is malformed. You can use 'init' to fix it.\n", homepath)
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

var pullBankCmd = &cobra.Command{
	Use:   "pull [flags] [fileNumber]",
	Short: "Download file from server bank",
	Long: `Downloads a file from a server's bank, verifies merkle proof, decrypts file.

Args:
  fileNumber: identifier of the file in the bank`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Printf("Missing positional arguments: fileNumber is required\n\n")
			cmd.Help()
			return
		}
		if len(args) > 1 {
			fmt.Printf("Unexpected positional arguments after %v\n\n", args[0])
			cmd.Help()
			return
		}
		fileNum, err := strconv.ParseInt(args[0], 10, 0)
		if err != nil {
			fmt.Printf("Positional arguments %v is not a valid int value\n\n", args[0])
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
		// verify home directory
		ok, err := storage.IsHomeWellFormed(homepath)
		if err != nil {
			fmt.Println(err)
			return
		} else if !ok {
			fmt.Printf("Home %v does not exist or is malformed. You can use 'init' to fix it.\n", homepath)
			return
		}

		if err := client.CallDownloadFiles(homepath, serverName, bankName, int(fileNum)); err != nil {
			fmt.Println(err)
			return
		}
	},
}

var listBankCmd = &cobra.Command{
	Use:   "list",
	Short: "List server banks, list bank contents",
	Long: `- List bank names for specified server (only provide server flag)
- List file names and identifiers for specified bank on specified server (provide server and bank-name flags)`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			fmt.Printf("Unexpected positional arguments\n\n")
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
		homepath, err := getHomePath(cmd)
		if err != nil {
			fmt.Println(err)
			return
		}
		// verify home directory
		ok, err := storage.IsHomeWellFormed(homepath)
		if err != nil {
			fmt.Println(err)
			return
		} else if !ok {
			fmt.Printf("Home %v does not exist or is malformed. You can use 'init' to fix it.\n", homepath)
			return
		}

		if bankName == "" { // listing banks from server
			banks, err := storage.Client_ListBanks(homepath, serverName)
			if err != nil {
				fmt.Printf("%v\n", err)
				return
			}
			fmt.Printf("Banks for server '%s'\n=====================================\n", serverName)
			for _, bank := range banks {
				fmt.Printf("\t%s\n", bank)
			}
		} else { // listing bank files
			files, err := storage.Client_ListBankFiles(homepath, serverName, bankName)
			if err != nil {
				fmt.Printf("%v\n", err)
				return
			}
			fmt.Printf("Files for bank '%s:%s'\n=====================================\n", serverName, bankName)
			for i, filename := range files {
				fmt.Printf("%5d  %s\n", i+1, filename)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(bankCmd)
	bankCmd.AddCommand(createBankCmd, pullBankCmd, listBankCmd)

	bankCmd.PersistentFlags().StringP("bank-name", "b", "", "unique local name for the filebank")
	bankCmd.PersistentFlags().StringP("server", "s", "", "unique local name for the server")
}
