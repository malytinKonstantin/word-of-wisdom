package server

import (
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"word-of-wisdom-server/internal/logger"

	"word-of-wisdom-server/internal/config"
	"word-of-wisdom-server/internal/interfaces"
)

// Server представляет серверное приложение, обрабатывающее клиентские запросы
type Server struct {
	config            *config.Config
	difficultyManager interfaces.DifficultyManager
	quoteStorage      interfaces.QuoteStorage
	proofOfWork       interfaces.ProofOfWorkHandler
	activeConnections int32
}

func New(config *config.Config, dm interfaces.DifficultyManager, qs interfaces.QuoteStorage, po interfaces.ProofOfWorkHandler) *Server {
	return &Server{
		config:            config,
		difficultyManager: dm,
		quoteStorage:      qs,
		proofOfWork:       po,
	}
}

func (s *Server) Run(port, certPath, keyPath string) error {
	// Загружаем сертификат и ключ для TLS
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return fmt.Errorf("Ошибка загрузки сертификата: %v", err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}

	// Запускаем TLS-листенер
	listener, err := tls.Listen("tcp", port, tlsConfig)
	if err != nil {
		return fmt.Errorf("Ошибка запуска сервера: %v", err)
	}
	defer listener.Close()

	logger.Log.Info().Msgf("Сервер запущен на порту %s", port)

	// Обработка системных сигналов для корректного завершения работы
	quit := s.setupSignalHandler(listener)
	return s.acceptConnections(listener, quit)
}

func (s *Server) setupSignalHandler(listener net.Listener) chan os.Signal {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		logger.Log.Info().Msg("Завершение работы сервера...")
		listener.Close()
		os.Exit(0)
	}()

	return quit
}
