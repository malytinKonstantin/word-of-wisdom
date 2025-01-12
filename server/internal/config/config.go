package config

import "time"

type Config struct {
	Port           string
	BaseDifficulty int
	MaxDifficulty  int
	MinDifficulty  int
	ReadTimeout    time.Duration
	MinSolveTime   time.Duration
	MaxSolveTime   time.Duration
	MaxConnections int
}

func NewDefaultConfig() *Config {
	return &Config{
		Port:           ":3333",
		BaseDifficulty: 4,
		MaxDifficulty:  6,
		MinDifficulty:  3,
		ReadTimeout:    30 * time.Second,
		MinSolveTime:   5 * time.Second,
		MaxSolveTime:   10 * time.Second,
		MaxConnections: 1000,
	}
}
