package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"word-of-wisdom-client/internal/config"
	"word-of-wisdom-client/internal/container"
	"word-of-wisdom-client/internal/interfaces"
	"word-of-wisdom-client/internal/logger"
	"word-of-wisdom-client/internal/network"
	"word-of-wisdom-client/internal/pow"
)

var cfgPath string

func main() {
	rootCmd := &cobra.Command{
		Use:   "client",
		Short: "Word of Wisdom Client",
		Run:   runClient,
	}

	rootCmd.Flags().StringVar(&cfgPath, "config", "./configs", "Путь к директории с файлом конфигурации")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runClient(cmd *cobra.Command, args []string) {
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("Ошибка загрузки конфигурации")
	}

	logger.Init(cfg.Logging.Level)

	logger.Log.Info().
		Str("server_address", cfg.Server.Address).
		Dur("timeout", cfg.Network.Timeout).
		Msg("Запуск клиента с конфигурацией")

	c := container.New()

	// Регистрируем зависимости
	c.Register("config", cfg)
	c.Register("networkClient", network.NewDefaultNetworkClient(cfg))
	c.Register("powSolver", pow.NewDefaultPoWSolver())

	startTime := time.Now()

	netClient := c.Resolve("networkClient").(interfaces.NetworkClient)
	powSolver := c.Resolve("powSolver").(interfaces.PoWSolver)

	conn, err := netClient.Connect(cfg.Server.Address)
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("Ошибка подключения к серверу")
	}
	defer conn.Close()

	logger.Log.Info().
		Str("server_address", cfg.Server.Address).
		Dur("connection_time", time.Since(startTime)).
		Msg("Успешное подключение к серверу")

	challenge, difficulty, err := netClient.ReceiveChallenge(conn)
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("Ошибка получения challenge")
	}
	logger.Log.Info().
		Str("challenge", challenge).
		Int("difficulty", difficulty).
		Msg("Получен challenge и сложность")

	powStartTime := time.Now()
	logger.Log.Info().Msg("Начало решения Proof of Work")

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Network.Timeout)
	defer cancel()

	nonce, err := powSolver.SolveProofOfWork(ctx, challenge, difficulty)
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("Ошибка при решении Proof of Work")
	}

	logger.Log.Info().
		Dur("pow_time", time.Since(powStartTime)).
		Str("nonce", nonce).
		Msg("Proof of Work решен")

	if err := netClient.SendNonceAndGetQuote(conn, nonce); err != nil {
		logger.Log.Fatal().Err(err).Msg("Ошибка при обмене данными с сервером")
	}
	logger.Log.Info().
		Dur("total_time", time.Since(startTime)).
		Msg("Клиент завершил работу")
}
