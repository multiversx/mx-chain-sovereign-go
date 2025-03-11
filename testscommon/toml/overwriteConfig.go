package toml

import "github.com/multiversx/mx-chain-sovereign-go/config"

// OverrideConfig holds an array of configs to be overridden
type OverrideConfig struct {
	OverridableConfigTomlValues []config.OverridableConfig
}
