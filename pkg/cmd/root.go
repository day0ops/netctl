package cmd

import (
	"fmt"
	"net"

	"github.com/spf13/cobra"

	"github.com/day0ops/netctl/pkg/config"
	"github.com/day0ops/netctl/pkg/log"
	"github.com/day0ops/netctl/pkg/network"
)

var versionInfo = config.AppVersion()

var rootCmdArgs struct {
	network.Network
	Verbose bool
}

var rootCmd = &cobra.Command{
	Use:     "netctl",
	Short:   "For managing libvirt domains and networks",
	Long:    "A tool for creating and managing networks using libvirt API. Useful if you need to run multiple clusters (for e.g. with minikube) on the same underlying network.",
	Version: versionInfo.Version,
	Args:    cobra.MaximumNArgs(1),
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		err := initLog()
		return err
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Error(err)
	}
}

// createCmd returns the create subcommand
func createCmd() *cobra.Command {
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create network",
		RunE:  createNet,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if isNotValidCIDR(rootCmdArgs.Subnet) {
				return fmt.Errorf("invalid CIDR value provided (for e.g. it should be of the form 10.89.0.1/24): %v", rootCmdArgs.Subnet)
			}
			return nil
		},
	}

	// add flags
	addCommonFlags(createCmd)
	createCmd.Flags().StringVarP(&rootCmdArgs.Bridge, "bridge", "b", "virbr0", "Name of the network bridge")
	createCmd.Flags().StringVarP(&rootCmdArgs.Subnet, "subnet-cidr", "s", "", "Subnet of the network (for e.g. 10.89.0.1/24")
	createCmd.Flags().StringVarP(&rootCmdArgs.ConnectionURI, "uri", "u", "qemu:///system", "libvirt connection URI")
	createCmd.MarkFlagRequired("subnet-cidr")

	return createCmd
}

func createNet(cmd *cobra.Command, args []string) error {
	n := rootCmdArgs.Network
	return n.EnsureNetwork()
}

// deleteCmd returns the delete subcommand
func deleteCmd() *cobra.Command {
	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete network",
		RunE:  deleteNet,
	}

	// add flags
	addCommonFlags(deleteCmd)

	return deleteCmd
}

func deleteNet(cmd *cobra.Command, args []string) error {
	n := rootCmdArgs.Network
	return n.DeleteNetwork()
}

func addCommonFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&rootCmdArgs.Name, "name", "n", "", "Name of the network")
	cmd.MarkFlagRequired("name")
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&rootCmdArgs.Verbose, "verbose", "v", rootCmdArgs.Verbose, "enable verbose log")

	rootCmd.AddCommand(createCmd())
	rootCmd.AddCommand(deleteCmd())
}

func initLog() error {
	if rootCmdArgs.Verbose {
		log.SetDebug(true)
	}
	return nil
}

func isNotValidCIDR(cidr string) bool {
	_, _, err := net.ParseCIDR(cidr)
	return err != nil
}
