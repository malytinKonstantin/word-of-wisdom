package config

import "time"

type Config struct {
	ServerAddr string
	Timeout    time.Duration
}

func NewDefault() *Config {
	return &Config{
		ServerAddr: "server:3333",
		Timeout:    30 * time.Second,
	}
}
