package config

import (
	"fmt"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env         string `yaml:"env" env-default:"local"`
	StorageLink string `yaml:"storage_link" env-required:"true"`
	HTTPServer  `yaml:"http_server"`
}

type HTTPServer struct {
	Host        string        `yaml:"host" env-default:"0.0.0.0"`
	Port        string        `yaml:"port" env-default:"8080"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

// Load reads the configuration from the specified file path or the environment variable.
func Load(configPaths ...string) (*Config, error) {
	configPath, err := getConfigPath(configPaths)
	if err != nil {
		return nil, err
	}

	config := &Config{}
	if err := cleanenv.ReadConfig(configPath, config); err != nil {
		return nil, fmt.Errorf("cannot read config: %w", err)
	}

	return config, nil
}

func getConfigPath(configPaths []string) (string, error) {
	configPath := os.Getenv("CONFIG_PATH")
	if len(configPaths) == 0 {
		if configPath == "" {
			return "", fmt.Errorf("config file not found: CONFIG_PATH env variable is not set")
		}
	} else if len(configPaths) == 1 {
		if configPaths[0] == "" {
			if configPath == "" {
				return "", fmt.Errorf("config file not found: CONFIG_PATH env variable is not set")
			}
		} else {
			configPath = configPaths[0]
		}
	} else {
		return "", fmt.Errorf("config file not found")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return "", fmt.Errorf("config file not found: %s", configPath)
	}

	return configPath, nil
}
