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
	"word-of-wisdom-client/internal/log"
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
		log.Init("info")
		log.Log.Fatal().Err(err).Msg("Ошибка загрузки конфигурации")
	}

	log.Init(cfg.Logging.Level)

	log.Log.Info().
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

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Network.Timeout)
	defer cancel()

	if err := run(ctx, cfg, netClient, powSolver, startTime); err != nil {
		log.Log.Error().Err(err).Msg("Ошибка во время выполнения клиента")
		os.Exit(1)
	}

	log.Log.Info().
		Dur("total_time", time.Since(startTime)).
		Msg("Клиент завершил работу")
}

func run(ctx context.Context, cfg *config.Config, netClient interfaces.NetworkClient, powSolver interfaces.PoWSolver, startTime time.Time) error {
	conn, err := netClient.Connect(cfg.Server.Address)
	if err != nil {
		log.Log.Fatal().Err(err).Msg("Ошибка подключения к серверу")
	}
	defer conn.Close()

	log.Log.Info().
		Str("server_address", cfg.Server.Address).
		Dur("connection_time", time.Since(startTime)).
		Msg("Успешное подключение к серверу")

	challenge, difficulty, err := netClient.ReceiveChallenge(conn)
	if err != nil {
		log.Log.Fatal().Err(err).Msg("Ошибка получения challenge")
	}
	log.Log.Info().
		Str("challenge", challenge).
		Int("difficulty", difficulty).
		Msg("Получен challenge и сложность")

	powStartTime := time.Now()
	log.Log.Info().Msg("Начало решения Proof of Work")

	nonce, err := powSolver.SolveProofOfWork(ctx, challenge, difficulty)
	if err != nil {
		log.Log.Fatal().Err(err).Msg("Ошибка при решении Proof of Work")
	}

	log.Log.Info().
		Dur("pow_time", time.Since(powStartTime)).
		Str("nonce", nonce).
		Msg("Proof of Work решен")

	if err := netClient.SendNonceAndGetQuote(conn, nonce); err != nil {
		return fmt.Errorf("ошибка при обмене данными с сервером: %v", err)
	}

	return nil
}
