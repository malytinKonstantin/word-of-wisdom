package server

import (
	"context"
	"crypto/tls"
	"fmt"

	"word-of-wisdom-server/internal/config"
	"word-of-wisdom-server/internal/container"
	"word-of-wisdom-server/internal/interfaces"
	"word-of-wisdom-server/internal/log"
)

// Server представляет серверное приложение, обрабатывающее клиентские запросы
type Server struct {
	workerPool        chan struct{}
	serverConfig      config.ServerConfig
	difficultyManager interfaces.DifficultyManager
	quoteStorage      interfaces.QuoteStorage
	proofOfWork       interfaces.ProofOfWorkHandler
	activeConnections int32
}

func New(c *container.Container) *Server {
	appConfig := c.Resolve("config").(*config.AppConfig)
	dm := c.Resolve("difficultyManager").(interfaces.DifficultyManager)
	qs := c.Resolve("quoteStorage").(interfaces.QuoteStorage)
	po := c.Resolve("proofOfWorkHandler").(interfaces.ProofOfWorkHandler)

	return &Server{
		serverConfig:      appConfig.Server,
		difficultyManager: dm,
		quoteStorage:      qs,
		proofOfWork:       po,
		workerPool:        make(chan struct{}, appConfig.Server.MaxWorkers),
	}
}

func (s *Server) Run(ctx context.Context) error {
	cert, err := tls.LoadX509KeyPair(s.serverConfig.CertPath, s.serverConfig.KeyPath)
	if err != nil {
		return fmt.Errorf("Ошибка загрузки сертификата: %v", err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}

	listener, err := tls.Listen("tcp", s.serverConfig.Port, tlsConfig)
	if err != nil {
		return fmt.Errorf("Ошибка запуска сервера: %v", err)
	}
	defer listener.Close()

	log.Info().Msgf("Сервер запущен на порту %s", s.serverConfig.Port)

	errCh := make(chan error, 1)
	go func() {
		errCh <- s.acceptConnections(ctx, listener)
	}()

	select {
	case <-ctx.Done():
		log.Info().Msg("Контекст отменён, завершаем работу сервера")
		return nil
	case err := <-errCh:
		return err
	}
}
