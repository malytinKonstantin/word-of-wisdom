package config

import "time"

// Config содержит настройки сервера
type Config struct {
	Port           string        // Порт для запуска сервера
	BaseDifficulty int           // Начальная сложность PoW
	MaxDifficulty  int           // Максимальная сложность PoW
	MinDifficulty  int           // Минимальная сложность PoW
	ReadTimeout    time.Duration // Таймаут чтения данных от клиента
	MinSolveTime   time.Duration // Минимальное время решения для регулировки сложности
	MaxSolveTime   time.Duration // Максимальное время решения для регулировки сложности
	MaxConnections int           // Максимальное количество одновременных подключений
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
