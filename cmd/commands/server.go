package cmd

import (
	"fmt"

	"github.com/oteffahi/merkle-filebank/client"
	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Manage servers",
	Long:  `Manage known server`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.Help()
		return nil
	},
}

var addServerCmd = &cobra.Command{
	Use:   "add [flags] <ServerName>",
	Short: "Add new server",
	Long: `Add new server running on known host

Args:
  ServerName: unique local name for the server`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Printf("Missing positional argument: ServerName is required\n\n")
			cmd.Help()
			return
		}
		if len(args) > 2 {
			fmt.Printf("Unexpected positional arguments after %v\n\n", args[0])
			cmd.Help()
			return
		}

		addr, err := cmd.Flags().GetString("address")
		if err != nil {
			fmt.Printf("%v\n\n", err)
			cmd.Help()
			return
		}
		if addr == "" {
			fmt.Printf("Missing flag: address flag is required\n\n")
			cmd.Help()
			return
		}

		port, err := cmd.Flags().GetInt16("port")
		if err != nil {
			fmt.Println(err)
			cmd.Help()
			return
		}
		serverName := args[0]

		if err := addServer(serverName, addr, port); err != nil {
			fmt.Println(err)
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
	serverCmd.AddCommand(addServerCmd)

	addServerCmd.Flags().StringP("address", "a", "", "hostname or IP address of server")
	addServerCmd.Flags().Int16P("port", "p", 5500, "TCP Port number on which the MerkleFileBank service is running")
}

func addServer(serverName string, host string, port int16) error {
	hostName := fmt.Sprintf("%s:%d", host, port)
	if err := client.CallAddNode(hostName, serverName); err != nil {
		return err
	}
	return nil
}
