package server

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"strings"
	"time"
	"word-of-wisdom-server/internal/log"
	"word-of-wisdom-server/internal/utils"
)

func (s *Server) acceptConnections(ctx context.Context, listener net.Listener) error {
	log.Info().Int("max_connections", s.serverConfig.MaxConnections).Msg("Сервер начал прием подключений")

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("Получен сигнал завершения, прекращаем прием подключений")
			return nil
		default:
		}

		conn, err := listener.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				continue
			}
			log.Error().Err(err).Msg("Ошибка принятия соединения")
			continue
		}

		select {
		case s.workerPool <- struct{}{}:
			go func(conn net.Conn) {
				defer func() { <-s.workerPool }()
				s.handleConnection(conn)
			}(conn)
		default:
			log.Warn().Str("client", conn.RemoteAddr().String()).Msg("Превышено максимальное количество воркеров")
			conn.Close()
		}
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer s.handlePanic(conn)

	clientAddr := conn.RemoteAddr().String()
	log.Info().Str("client", clientAddr).Msg("Начало обработки соединения")

	conn.SetReadDeadline(time.Now().Add(time.Duration(s.serverConfig.ReadTimeout) * time.Second))

	challenge := utils.GenerateChallenge()

	if err := s.sendChallenge(conn, challenge); err != nil {
		log.Error().Err(err).Str("client", clientAddr).Msg("Ошибка отправки challenge")
		return
	}

	if err := s.sendDifficulty(conn); err != nil {
		log.Error().Err(err).Str("client", clientAddr).Msg("Ошибка отправки сложности")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(s.serverConfig.ReadTimeout)*time.Second)
	defer cancel()

	startTime := time.Now()

	if err := s.handleProofOfWork(ctx, conn, challenge, startTime); err != nil {
		log.Error().Err(err).Str("client", clientAddr).Msg("Ошибка обработки Proof of Work")
		return
	}

	log.Info().
		Str("client", clientAddr).
		Dur("solve_time", time.Since(startTime)).
		Msg("Клиент успешно решил задачу")
}

func (s *Server) handlePanic(conn net.Conn) {
	if r := recover(); r != nil {
		log.Error().Interface("recover", r).Msg("Паника в goroutine")
	}
}

func (s *Server) readNonce(conn net.Conn) (string, error) {
	reader := bufio.NewReader(conn)
	nonce, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("ошибка чтения nonce: %v", err)
	}

	nonce = strings.TrimSpace(nonce)
	log.Info().Str("nonce", nonce).Msg("Получен nonce")

	if len(nonce) == 0 {
		log.Info().Msg("Получен пустой nonce")
		fmt.Fprintln(conn, "Ошибка: пустой nonce")
		return "", fmt.Errorf("пустой nonce")
	}

	return nonce, nil
}

func (s *Server) sendChallenge(conn net.Conn, challenge string) error {
	log.Info().Str("challenge", challenge).Msg("Отправка challenge клиенту")
	if _, err := fmt.Fprintln(conn, challenge); err != nil {
		log.Error().Err(err).Msg("Ошибка отправки challenge")
		return err
	}
	return nil
}

func (s *Server) sendDifficulty(conn net.Conn) error {
	difficulty := s.difficultyManager.GetDifficulty()
	log.Info().Int("difficulty", difficulty).Msg("Отправка сложности клиенту")
	if _, err := fmt.Fprintln(conn, difficulty); err != nil {
		log.Error().Err(err).Msg("Ошибка отправки сложности")
		return err
	}
	return nil
}

func (s *Server) sendQuote(conn net.Conn) error {
	quote := s.quoteStorage.GetRandomQuote()
	log.Info().Str("quote", quote).Msg("Отправка цитаты")
	if _, err := fmt.Fprintln(conn, quote); err != nil {
		log.Error().Err(err).Msg("Ошибка отправки цитаты")
		return err
	}
	return nil
}

func (s *Server) handleProofOfWork(ctx context.Context, conn net.Conn, challenge string, startTime time.Time) error {
	return s.proofOfWork.HandleProofOfWork(ctx, conn, challenge, startTime)
}
