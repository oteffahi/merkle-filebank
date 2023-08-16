package cmd

import (
	"fmt"

	"github.com/oteffahi/merkle-filebank/client"
	"github.com/oteffahi/merkle-filebank/server"
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

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Run server",
	Long:  `Start server instance on local machine`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			fmt.Printf("Unexpected positional arguments\n\n")
			cmd.Help()
			return
		}
		addr, err := cmd.Flags().GetString("address")
		if err != nil {
			fmt.Printf("%v\n\n", err)
			cmd.Help()
			return
		}

		port, err := cmd.Flags().GetInt16("port")
		if err != nil {
			fmt.Println(err)
			cmd.Help()
			return
		}

		homepath, err := getHomePath(cmd)
		if err != nil {
			fmt.Println(err)
			cmd.Help()
			return
		}

		if err := startServer(addr, port, homepath); err != nil {
			fmt.Println(err)
			return
		}
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
		if len(args) > 1 {
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

		homepath, err := getHomePath(cmd)
		if err != nil {
			fmt.Println(err)
			cmd.Help()
			return
		}
		if err := addServer(homepath, serverName, addr, port); err != nil {
			fmt.Println(err)
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
	rootCmd.AddCommand(startCmd)
	serverCmd.AddCommand(addServerCmd)

	addServerCmd.Flags().StringP("address", "a", "", "hostname or IP address of server")
	addServerCmd.Flags().Int16P("port", "p", 5500, "TCP Port number on which the MerkleFileBank service is running")

	startCmd.Flags().StringP("address", "a", "0.0.0.0", "hostname or IP address of server")
	startCmd.Flags().Int16P("port", "p", 5500, "TCP Port number on which the MerkleFileBank service will run")
}

func addServer(homepath, serverName string, host string, port int16) error {
	hostName := fmt.Sprintf("%s:%d", host, port)
	if err := client.CallAddNode(hostName, homepath, serverName); err != nil {
		return err
	}
	return nil
}

func startServer(host string, port int16, homepath string) error {
	endpoint := fmt.Sprintf("%s:%d", host, port)
	if err := server.SetBankHome(homepath); err != nil {
		return err
	}
	server.RunServer(endpoint)
	return nil
}
