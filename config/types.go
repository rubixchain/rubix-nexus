package config

type Config struct {
	Network NetworkConfig `toml:"network"`
}

type NetworkConfig struct {
	DeployerNodeURL string `toml:"deployer_node_url"`
}

