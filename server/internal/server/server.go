package server

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"word-of-wisdom-server/internal/config"
	"word-of-wisdom-server/internal/pow"
	"word-of-wisdom-server/internal/storage"
)

type Server struct {
	config            *config.Config
	difficultyManager *pow.DifficultyManager
	quoteStorage      *storage.QuoteStorage
	proofOfWork       *pow.ProofOfWork
	activeConnections int32
}

func New(config *config.Config) *Server {
	dm := pow.NewDifficultyManager(config)
	qs := storage.New()
	po := pow.New(config, dm, qs)
	return &Server{
		config:            config,
		difficultyManager: dm,
		quoteStorage:      qs,
		proofOfWork:       po,
	}
}

func (s *Server) Run(port, certPath, keyPath string) error {
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return fmt.Errorf("Ошибка загрузки сертификата: %v", err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}

	listener, err := tls.Listen("tcp", port, tlsConfig)
	if err != nil {
		return fmt.Errorf("ошибка запуска сервера: %v", err)
	}
	defer listener.Close()

	log.Printf("Сервер запущен на порту %s", port)

	quit := s.setupSignalHandler(listener)
	return s.acceptConnections(listener, quit)
}

func (s *Server) setupSignalHandler(listener net.Listener) chan os.Signal {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("Завершение работы сервера...")
		listener.Close()
		os.Exit(0)
	}()

	return quit
}
