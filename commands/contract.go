package commands

import (
	"fmt"

	"github.com/rubixchain/rubix-nexus/contract"
	"github.com/spf13/cobra"
)

func contractCommands() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "contract",
		Short: "Contract related sub-commands",
		Long:  "Contract related sub-commands",
		Args:  cobra.NoArgs,
	}

	cmd.AddCommand(
		cmdBootstrap(),
		cmdDeploy(),
		cmdExecute(),
	)

	return cmd
}

func cmdBootstrap() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bootstrap [contract-name]",
		Short: "Bootstrap a new Rust smart contract project",
		Long:  "Bootstrap a new Rust smart contract project with the necessary structure and dependencies",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			contractName := args[0]
			if err := contract.Bootstrap(contractName); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error: failed to bootstrap contract: %v\n", err)
				return nil
			}
			cmd.Printf("Contract project '%s' bootstrapped successfully\n", contractName)
			return nil
		},
	}
	cmd.SilenceUsage = true
	return cmd
}

func cmdDeploy() *cobra.Command {
	var (
		contractDir      string
		deployerDid      string
		deployAmt        float64
	)

	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy a smart contract",
		Long:  "Deploy a smart contract from a Rust project directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			if contractDir == "" {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error: --contract-dir is required\n")
				return nil
			}
			if deployerDid == "" {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error: --deployer-did is required\n")
				return nil
			}

			// Print stage messages
			onStage := func(stage contract.DeploymentStage) {
				switch stage {
				case contract.StageBuild:
					cmd.Println("Building contract...")
				case contract.StageGenerate:
					cmd.Println("Generating smart contract...")
				case contract.StageDeploy:
					cmd.Println("Deploying smart contract...")
				}
			}

			result, err := contract.Deploy(contractDir, flagHomeDir, deployerDid, deployAmt, onStage)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error: deployment failed: %v\n", err)
				return nil
			}

			if result.Success {
				cmd.Printf("Contract deployed successfully with hash: %s\n", result.ContractHash)
			} else {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error: deployment failed: %s\n", result.Message)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&contractDir, "contract-dir", "", "Directory containing the Rust contract project")
	cmd.Flags().StringVar(&deployerDid, "deployer-did", "", "Contract Deployer DID")
	cmd.Flags().Float64Var(&deployAmt, "deploy-amt", 0.001, "RBT amount to deploy the contract")
	cmd.SilenceUsage = true
	return cmd
}

func cmdExecute() *cobra.Command {
	var (
		contractHash    string
		executorDid     string
		contractDir     string
		contractMsgFile string
	)

	cmd := &cobra.Command{
		Use:   "execute",
		Short: "Execute a deployed smart contract",
		Long:  "Execute a deployed smart contract with a JSON message",
		RunE: func(cmd *cobra.Command, args []string) error {
			if contractHash == "" {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error: --contract-hash is required\n")
				return nil
			}
			if contractMsgFile == "" {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error: --contract-msg-file is required\n")
				return nil
			}
			if contractDir == "" {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error: --contract-dir is required\n")
			}

			cmd.Println("Executing smart contract...")
			result, err := contract.Execute(contractHash, executorDid, flagHomeDir, contractDir, contractMsgFile)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error: execution failed: %v\n", err)
				return nil
			}

			if result.Success {
				cmd.Printf("Contract Result: %v\n", result.ContractResult)
			} else {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error: execution failed: %s\n", result.Message)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&contractHash, "contract-hash", "", "Hash of the deployed contract")
	cmd.Flags().StringVar(&executorDid, "executor-did", "", "Executor DID")
	cmd.Flags().StringVar(&contractDir, "contract-dir", "", "Directory containing the Rust contract project")
	cmd.Flags().StringVar(&contractMsgFile, "contract-msg-file", "", "File containing the JSON message for contract execution")
	
	cmd.SilenceUsage = true
	return cmd
}
