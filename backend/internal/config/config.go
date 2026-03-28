package config

// Config holds application configuration
type Config struct {
	Server ServerConfig `yaml:"server"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port int    `yaml:"port"`
	Mode string `yaml:"mode"`
}

// Default returns the default configuration
func Default() *Config {
	return &Config{
		Server: ServerConfig{
			Port: 8040,
			Mode: "release",
		},
	}
}