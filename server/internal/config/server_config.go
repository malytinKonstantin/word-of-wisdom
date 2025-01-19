package config

// ServerConfig содержит настройки сервера
type ServerConfig struct {
	Port           string `mapstructure:"port"`            // Порт для запуска сервера
	ReadTimeout    int    `mapstructure:"read_timeout"`    // Таймаут чтения данных от клиента (в секундах)
	MaxConnections int    `mapstructure:"max_connections"` // Максимальное количество одновременных подключений
	CertPath       string `mapstructure:"cert_path"`       // Путь к файлу сертификата
	KeyPath        string `mapstructure:"key_path"`        // Путь к файлу ключа
}
