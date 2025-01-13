package config

import "time"

// Config содержит настройки клиента для подключения к серверу
type Config struct {
	ServerAddr string        // Адрес сервера (хост и порт)
	Timeout    time.Duration // Таймаут для сетевых операций
}

func NewDefault() *Config {
	return &Config{
		ServerAddr: "server:3333",
		Timeout:    30 * time.Second,
	}
}
