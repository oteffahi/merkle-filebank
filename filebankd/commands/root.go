package cmd

import (
	"fmt"
	"os"

	"github.com/oteffahi/merkle-filebank/storage"
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

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize home directory for filebankd",
	Long:  `Initialize home directory for filebankd. Use --home flag to overwrite default path`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			fmt.Printf("Unexpected positional arguments\n\n")
			cmd.Help()
			return
		}

		homepath, err := getHomePath(cmd)
		if err != nil {
			fmt.Println(err)
			cmd.Help()
			return
		}

		if IsHomeWellFormed, err := storage.IsHomeWellFormed(homepath); err != nil {
			fmt.Println(err)
			return
		} else if IsHomeWellFormed {
			fmt.Printf("%v is already a well-formed filebankd home directory\n", homepath)
			return
		}

		if err := storage.InitHome(homepath); err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("Initialized filebankd home directory at %v\n", homepath)
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

	rootCmd.AddCommand(initCmd)
	rootCmd.PersistentFlags().String("home", userHome+"/.filebankd", "root directory for MerkleFileBank storage")
}

func getHomePath(cmd *cobra.Command) (string, error) {
	homepath, err := cmd.Flags().GetString("home")
	if err != nil {
		return "", err
	}
	return homepath, nil
}
