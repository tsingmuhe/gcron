package gcron

import (
	"fmt"
	"net"
	"time"

	"github.com/spf13/viper"
)

const DefaultBindPort int = 7946

type Config struct {
	Server          bool `mapstructure:"server"`
	BootstrapExpect int  `mapstructure:"bootstrap-expect"`

	//serf
	Profile             string            `mapstructure:"profile"`
	ReconnectTimeoutRaw string            `mapstructure:"reconnect_timeout"`
	ReconnectTimeout    time.Duration     `mapstructure:"-"`
	NodeName            string            `mapstructure:"node_name"`
	Tags                map[string]string `mapstructure:"tags"`
	BindAddr            string            `mapstructure:"bind"`
}

func (c *Config) GetTags() map[string]string {
	return c.Tags
}

func (c *Config) AddrParts(address string) (string, int, error) {
	checkAddr := address

START:
	_, _, err := net.SplitHostPort(checkAddr)
	if ae, ok := err.(*net.AddrError); ok && ae.Err == "missing port in address" {
		checkAddr = fmt.Sprintf("%s:%d", checkAddr, DefaultBindPort)
		goto START
	}

	if err != nil {
		return "", 0, err
	}

	// Get the address
	addr, err := net.ResolveTCPAddr("tcp", checkAddr)
	if err != nil {
		return "", 0, err
	}

	return addr.IP.String(), addr.Port, nil
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

	if len(b.Profile) > 0 {
		a.Profile = b.Profile
	}

	if len(b.NodeName) > 0 {
		a.NodeName = b.NodeName
	}

	return &result
}
