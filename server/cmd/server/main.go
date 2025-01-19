package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"

	"word-of-wisdom-server/internal/config"
	"word-of-wisdom-server/internal/container"
	"word-of-wisdom-server/internal/logger"
	"word-of-wisdom-server/internal/pow"
	"word-of-wisdom-server/internal/server"
	"word-of-wisdom-server/internal/storage"
)

func main() {
	var configPath string

	// Создаем корневую команду Cobra
	rootCmd := &cobra.Command{
		Use:   "server",
		Short: "Word of Wisdom Server",
		Run: func(cmd *cobra.Command, args []string) {
			runServer(configPath)
		},
	}

	// Добавляем флаг для указания пути к конфигурационному файлу
	rootCmd.Flags().StringVarP(&configPath, "config", "c", "config.yaml", "Путь к файлу конфигурации")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runServer(configPath string) {
	// Загружаем конфигурацию
	appConfig, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Printf("Ошибка загрузки конфигурации: %v\n", err)
		os.Exit(1)
	}

	// Инициализируем логгер с уровнем из конфигурации
	logger.Init(appConfig.Logging.Level)

	// Устанавливаем количество используемых CPU
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Создаем DI-контейнер
	c := container.New()

	// Регистрируем конфигурацию в контейнере
	c.Register("config", appConfig)

	// Инициализация зависимостей
	dm := pow.NewDifficultyManager(appConfig.PoW)
	qs := storage.New()
	po := pow.New(c)
	srv := server.New(c)

	// Регистрируем зависимости в контейнере
	c.Register("difficultyManager", dm)
	c.Register("quoteStorage", qs)
	c.Register("proofOfWorkHandler", po)
	c.Register("server", srv)

	// Запуск сервера
	if err := srv.Run(appConfig.Server); err != nil {
		logger.Log.Fatal().Err(err).Msg("Ошибка запуска сервера")
	}
}
