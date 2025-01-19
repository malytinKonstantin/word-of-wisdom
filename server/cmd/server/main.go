package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

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
	rootCmd := &cobra.Command{
		Use:   "server",
		Short: "Word of Wisdom Server",
		Run: func(cmd *cobra.Command, args []string) {
			runServer(configPath)
		},
	}
	rootCmd.Flags().StringVarP(&configPath, "config", "c", "config.yaml", "Путь к файлу конфигурации")
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runServer(configPath string) {
	appConfig, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Printf("Ошибка загрузки конфигурации: %v\n", err)
		os.Exit(1)
	}

	logger.Init(appConfig.Logging.Level)
	runtime.GOMAXPROCS(runtime.NumCPU())

	c := container.New()
	c.Register("config", appConfig)

	dm := pow.NewDifficultyManager(appConfig.PoW)
	qs := storage.New()
	po := pow.New(c)
	srv := server.New(c)

	c.Register("difficultyManager", dm)
	c.Register("quoteStorage", qs)
	c.Register("proofOfWorkHandler", po)
	c.Register("server", srv)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-quit
		logger.Log.Info().Msg("Получен сигнал завершения работы")
		cancel()
	}()

	if err := srv.Run(ctx); err != nil {
		logger.Log.Fatal().Err(err).Msg("Ошибка запуска сервера")
	}
}
