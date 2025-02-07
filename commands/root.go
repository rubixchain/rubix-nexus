package commands

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"
)

var flagHomeDir string

var rootCmd = &cobra.Command{
	Use:                        "rubix-nexus",
	Short:                      "Rubix Nexus - Smart Contract Deployer and Executor",
	Long:                       "Rubix Nexus - Smart Contract Deployer and Executor",
	SuggestionsMinimumDistance: 2,
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.AddCommand(
		contractCommands(),
		configCommands(),
		didCommands(),
	)

	var errHomeDir error
	flagHomeDir, errHomeDir = getDefaultHomeDir()
	if errHomeDir != nil {
		fmt.Fprintf(os.Stderr, "unable to initialise Rubix Nexus CLI, err: %v", errHomeDir)
		os.Exit(1)
	}

	rootCmd.PersistentFlags().StringVar(&flagHomeDir, "home", flagHomeDir, "Set the home directory for configuration")

	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})
}

func getDefaultHomeDir() (string, error) {
	var defaultHomeDir = ""

	switch runtime.GOOS {
	case "windows":
		defaultHomeDir = os.Getenv("USERPROFILE")
	case "linux", "darwin":
		defaultHomeDir = os.Getenv("HOME")
	}

	if defaultHomeDir == "" {
		return "", fmt.Errorf("unable to fetch Home directory")
	}

	return defaultHomeDir, nil
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
