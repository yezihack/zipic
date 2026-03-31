package config

import (
	"github.com/spf13/viper"
)

// Config holds application configuration
type Config struct {
	Server ServerConfig `yaml:"server" mapstructure:"server"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port    int    `yaml:"port" mapstructure:"port"`
	Address string `yaml:"address" mapstructure:"address"`
	Mode    string `yaml:"mode" mapstructure:"mode"`
}

// Default returns the default configuration
func Default() *Config {
	return &Config{
		Server: ServerConfig{
			Port:    8040,
			Address: "127.0.0.1",
			Mode:    "release",
		},
	}
}

// Load reads configuration from viper and returns Config
func Load() *Config {
	cfg := Default()

	// Read from viper if available
	if viper.IsSet("server.port") {
		cfg.Server.Port = viper.GetInt("server.port")
	}
	if viper.IsSet("server.address") {
		cfg.Server.Address = viper.GetString("server.address")
	}
	if viper.IsSet("server.mode") {
		cfg.Server.Mode = viper.GetString("server.mode")
	}

	return cfg
}

// LoadFromFile loads config from a specific file path
func LoadFromFile(path string) (*Config, error) {
	viper.SetConfigFile(path)
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}
	return Load(), nil
}