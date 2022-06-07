package gcron

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Server          bool `mapstructure:"server"`
	BootstrapExpect int  `mapstructure:"bootstrap-expect"`
}

func (c *Config) normalizeAddrs() error {
	return nil
}

func DefaultConfig() *Config {
	return &Config{
		Server:          false,
		BootstrapExpect: 0,
	}
}

func ReadConfigPaths(path string) (*Config, error) {
	viper.SetConfigFile(path)
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("Error reading '%s'", path)
	}

	result := new(Config)
	if err := viper.Unmarshal(result); err != nil {
		return nil, fmt.Errorf("Error reading '%s'", path)
	}

	return result, nil
}

func MergeConfig(a, b *Config) *Config {
	var result = *a

	if b.Server {
		a.Server = b.Server
	}

	if b.BootstrapExpect >= 0 {
		a.BootstrapExpect = b.BootstrapExpect
	}

	return &result
}
