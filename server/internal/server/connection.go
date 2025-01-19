package server

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"sync/atomic"
	"time"
	"word-of-wisdom-server/internal/logger"
	"word-of-wisdom-server/internal/utils"
)

func (s *Server) acceptConnections(listener net.Listener, quit chan os.Signal) error {
	logger.Log.Info().Int("max_connections", s.serverConfig.MaxConnections).Msg("Сервер начал прием подключений")

	for {
		// Проверяем лимит активных подключений
		if atomic.LoadInt32(&s.activeConnections) >= int32(s.serverConfig.MaxConnections) {
			logger.Log.Info().Int("max_connections", s.serverConfig.MaxConnections).Msg("Достигнуто максимальное количество подключений")
			time.Sleep(100 * time.Millisecond)
			continue
		}

		// Принимаем новое соединение
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-quit:
				return nil
			default:
				logger.Log.Error().Err(err).Msg("Ошибка принятия соединения")
				continue
			}
		}

		atomic.AddInt32(&s.activeConnections, 1)
		logger.Log.Info().Str("client", conn.RemoteAddr().String()).Msg("Новое подключение")

		go func(conn net.Conn) {
			defer func() {
				conn.Close()
				atomic.AddInt32(&s.activeConnections, -1)
				logger.Log.Info().Str("client", conn.RemoteAddr().String()).Msg("Соединение закрыто")
			}()

			// Создаем контекст для этого соединения
			ctx := context.Background()
			s.handleConnection(ctx, conn)
		}(conn)
	}
}

func (s *Server) handleConnection(ctx context.Context, conn net.Conn) {
	defer s.handlePanic(conn)
	defer conn.Close()

	clientAddr := conn.RemoteAddr().String()
	logger.Log.Info().Str("client", clientAddr).Msg("Начало обработки соединения")

	// Устанавливаем таймаут на соединение через контекст
	connCtx, cancel := context.WithTimeout(ctx, time.Duration(s.serverConfig.ReadTimeout)*time.Second)
	defer cancel()

	challenge := utils.GenerateChallenge()
	logger.Log.Info().Str("client", clientAddr).Str("challenge", challenge).Msg("Сгенерирован challenge для клиента")

	if err := s.sendChallenge(conn, challenge); err != nil {
		logger.Log.Error().Err(err).Str("client", clientAddr).Msg("Ошибка отправки challenge клиенту")
		return
	}

	if err := s.sendDifficulty(conn); err != nil {
		logger.Log.Error().Err(err).Str("client", clientAddr).Msg("Ошибка отправки сложности клиенту")
		return
	}

	startTime := time.Now()
	if err := s.handleProofOfWork(connCtx, conn, challenge, startTime); err != nil {
		logger.Log.Error().Err(err).Str("client", clientAddr).Msg("Ошибка обработки proof-of-work для клиента")
		return
	}

	solveTime := time.Since(startTime)
	logger.Log.Info().Str("client", clientAddr).Dur("solve_time", solveTime).Msg("Клиент успешно решил задачу")
}

func (s *Server) handlePanic(conn net.Conn) {
	if r := recover(); r != nil {
		logger.Log.Error().Interface("recover", r).Msg("Паника в goroutine")
	}
}

func (s *Server) readNonce(conn net.Conn) (string, error) {
	reader := bufio.NewReader(conn)
	nonce, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("ошибка чтения nonce: %v", err)
	}

	nonce = strings.TrimSpace(nonce)
	logger.Log.Info().Str("nonce", nonce).Msg("Получен nonce")

	if len(nonce) == 0 {
		logger.Log.Info().Msg("Получен пустой nonce")
		fmt.Fprintln(conn, "Ошибка: пустой nonce")
		return "", fmt.Errorf("пустой nonce")
	}

	return nonce, nil
}

func (s *Server) sendChallenge(conn net.Conn, challenge string) error {
	logger.Log.Info().Str("challenge", challenge).Msg("Отправка challenge клиенту")
	if _, err := fmt.Fprintln(conn, challenge); err != nil {
		logger.Log.Error().Err(err).Msg("Ошибка отправки challenge")
		return err
	}
	return nil
}

func (s *Server) sendDifficulty(conn net.Conn) error {
	difficulty := s.difficultyManager.GetDifficulty()
	logger.Log.Info().Int("difficulty", difficulty).Msg("Отправка сложности клиенту")
	if _, err := fmt.Fprintln(conn, difficulty); err != nil {
		logger.Log.Error().Err(err).Msg("Ошибка отправки сложности")
		return err
	}
	return nil
}

func (s *Server) sendQuote(conn net.Conn) error {
	quote := s.quoteStorage.GetRandomQuote()
	logger.Log.Info().Str("quote", quote).Msg("Отправка цитаты")
	if _, err := fmt.Fprintln(conn, quote); err != nil {
		logger.Log.Error().Err(err).Msg("Ошибка отправки цитаты")
		return err
	}
	return nil
}

func (s *Server) handleProofOfWork(ctx context.Context, conn net.Conn, challenge string, startTime time.Time) error {
	return s.proofOfWork.HandleProofOfWork(ctx, conn, challenge, startTime)
}
