package commands

import (
	"fmt"
	"path/filepath"

	"github.com/rubixchain/rubix-nexus/config"
	"github.com/spf13/cobra"
)

func configCommands() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Configuration related sub-commands",
		Long:  "Configuration related sub-commands",
		Args:  cobra.NoArgs,
	}

	cmd.AddCommand(
		cmdInit(),
		cmdValidate(),
	)

	return cmd
}

func cmdInit() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize default configuration",
		Long:  "Initialize default configuration in $HOME/.rubix-nexus/config.toml",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := config.GenerateConfig(flagHomeDir); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error: failed to initialize config: %v\n", err)
				return nil
			}

			configPath := filepath.Join(flagHomeDir, ".rubix-nexus", "config.toml")
			cmd.Printf("Configuration initialized at %v", configPath)
			return nil
		},
	}
	cmd.SilenceUsage = true
	return cmd
}

func cmdValidate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate configuration file",
		Long:  "Validate the configuration file at $HOME/.rubix-nexus/config.toml",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := config.ValidateConfig(flagHomeDir); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error: invalid configuration: %v\n", err)
				return nil
			}
			cmd.Println("Configuration is valid")
			return nil
		},
	}
	cmd.SilenceUsage = true
	return cmd
}
