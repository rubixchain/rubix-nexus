# Rubix Nexus

A CLI tool to bootstrap, build, deploy and execute Rubix WASM based smart contracts

## Installation

1. Clone the repo:

```
git clone https://github.com/rubixchain/rubix-nexus
cd rubix-nexus
```

2. Install Rubix Nexus

```
make install
```

## Getting started

1. Initialise the configuration

```
rubix-nexus config init
```

A `config.toml` file will be created in the following default directory:

- Linux and MacOS : `$HOME/.rubix-nexus`
- Windows : `%USERPROFILE%\.rubix-nexus`

Use `--home` flag to generate the configuaration in a different location. For every CLI operation, the `--home` flag has to provided the config is not present in the default directory.

Following is sample config:

```toml
[network]
deployer_node_url = 'http://localhost:20011'

```

The `deployer_node_url` refers to the Rubix node where the contracts will be deployed.

To validate the configuration, run the following:

```
rubix-nexus config validate
```

2. Boostrap a simple Rubix Smart Contract project

Run the following to bootstrap 

```
rubix-nexus contract bootstrap <contract-name>
```

A Cargo project will be generated under the `<contract-name>`. A template `src/lib.rs` is created for a better understanding of the structure of a Rubix Smart Contract.

3. Create a DID

DIDs can be created (and eventually register) using the following command:

```
rubix-nexus create did
```

If you are connected to a localnet and want to use test RBT tokens, run the following:

```
rubix-nexus create did --localnet
```

4. Deploy the contract

Once you have built the contract, run the following to deploy your contract on the network:

```
rubix-nexus contract deploy --contract-dir <project-directory> --deployer-did <DID deploying the contract>
```

Upon successful completion of the command, a Smart Contract hash will be genrated starting with `Qm`. This will be utilised in the `execute` CLI (explained below).

5. Execute the contract

To execute a deployed contract, run the following:

```
rubix-nexus contract execute --contract-dir <project-directory> --contract-hash <contract-hash> --contract-msg-file <path/to/smart-contract-msg-json> --executor-did <DID executing the contract>
```

The command requires the contract message to provided in a JSON file.
