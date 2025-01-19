package config

import (
	"time"

	"github.com/spf13/viper"
)

func LoadConfig(configPath string) (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configPath)

	viper.SetDefault("server.address", "localhost:3333")
	viper.SetDefault("network.timeout", "30s")
	viper.SetDefault("logging.level", "info")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	timeoutStr := viper.GetString("network.timeout")
	timeout, err := time.ParseDuration(timeoutStr)
	if err != nil {
		return nil, err
	}
	cfg.Network.Timeout = timeout

	return &cfg, nil
}
