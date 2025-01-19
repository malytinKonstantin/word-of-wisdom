package main

import (
	"flag"
	"runtime"

	"word-of-wisdom-server/internal/config"
	"word-of-wisdom-server/internal/container"
	"word-of-wisdom-server/internal/logger"
	"word-of-wisdom-server/internal/pow"
	"word-of-wisdom-server/internal/server"
	"word-of-wisdom-server/internal/storage"
)

func main() {
	logger.Init()

	// Создаём DI-контейнер
	c := container.New()

	config := config.NewDefaultConfig()

	// Обработка флагов командной строки для переопределения настроек
	port := flag.String("port", config.Port, "Порт для запуска сервера")
	certPath := flag.String("cert", "certs/server.crt", "Путь к файлу сертификата")
	keyPath := flag.String("key", "certs/server.key", "Путь к файлу ключа")
	numCPU := flag.Int("cpu", runtime.NumCPU(), "Количество используемых CPU")
	maxConn := flag.Int("max-conn", config.MaxConnections, "Максимальное количество одновременных подключений")
	flag.Parse()

	// Обновляем конфигурацию на основе аргументов командной строки
	config.MaxConnections = *maxConn
	runtime.GOMAXPROCS(*numCPU)

	// Регистрируем конфигурацию в контейнере
	c.Register("config", config)

	// Инициализация зависимостей
	dm := pow.NewDifficultyManager(config)
	qs := storage.New()
	po := pow.New(c)
	srv := server.New(c)

	// Регистрируем зависимости в контейнере
	c.Register("difficultyManager", dm)
	c.Register("quoteStorage", qs)
	c.Register("proofOfWorkHandler", po)
	c.Register("server", srv)

	if err := srv.Run(*port, *certPath, *keyPath); err != nil {
		logger.Log.Fatal().Err(err).Msg("Ошибка запуска сервера")
	}
}
