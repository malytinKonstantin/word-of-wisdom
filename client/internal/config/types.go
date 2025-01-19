package config

import "time"

type ServerConfig struct {
	Address string
}

type NetworkConfig struct {
	Timeout time.Duration
}

type LoggingConfig struct {
	Level string
}

type Config struct {
	Server  ServerConfig
	Network NetworkConfig
	Logging LoggingConfig
}
