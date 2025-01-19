package config

import "github.com/spf13/viper"

type AppConfig struct {
	Server  ServerConfig  `mapstructure:"server"`
	PoW     PoWConfig     `mapstructure:"pow"`
	Logging LoggingConfig `mapstructure:"logging"`
}

type LoggingConfig struct {
	Level string `mapstructure:"level"`
}

func LoadConfig(configPath string) (*AppConfig, error) {
	viper.SetConfigFile(configPath)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var config AppConfig
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
