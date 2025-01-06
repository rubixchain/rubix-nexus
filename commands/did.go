package commands

import (
	"fmt"

	"github.com/rubixchain/rubix-nexus/did"
	"github.com/spf13/cobra"
)

func didCommands() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "did",
		Short: "DID related sub-commands",
		Long:  "DID related sub-commands",
		Args:  cobra.NoArgs,
	}

	cmd.AddCommand(
		cmdCreate(),
	)

	return cmd
}

func cmdCreate() *cobra.Command {
	var flagLocalnet bool

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new DID",
		Long:  "Create a new DID",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			did, err := did.CreateDID(flagHomeDir, flagLocalnet)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error: failed to create DID: %v\n", err)
				return nil
			}
			cmd.Printf("DID created successfully: %s\n", did)
			return nil
		},
	}

	cmd.Flags().BoolVar(&flagLocalnet, "localnet", false, "It indicates whether the deployer node is running on a localnet setup")

	cmd.SilenceUsage = true
	return cmd
}
