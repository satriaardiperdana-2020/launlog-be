package config

import (
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	DBURL      string `yaml:"database_url"`
	ServerPort string `yaml:"server_port"`
	JWTSecret  string `yaml:"jwt_secret"`
}

func Load() (*Config, error) {
	data, err := os.ReadFile("config-development.yml")
	if err != nil {
		return nil, err
	}
	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
