package server

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"runtime/debug"
	"strings"
	"sync/atomic"
	"time"
	"word-of-wisdom-server/internal/utils"
)

func (s *Server) acceptConnections(listener net.Listener, quit chan os.Signal) error {
	log.Printf("Сервер начал прием подключений. Максимальное количество одновременных подключений: %d", s.config.MaxConnections)

	for {
		// Проверяем лимит активных подключений
		if atomic.LoadInt32(&s.activeConnections) >= int32(s.config.MaxConnections) {
			log.Printf("Достигнуто максимальное количество подключений (%d). Ожидание освобождения...", s.config.MaxConnections)
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
				log.Printf("Ошибка принятия соединения от %v: %v", conn.RemoteAddr(), err)
				continue
			}
		}

		currentConnections := atomic.AddInt32(&s.activeConnections, 1)
		log.Printf("Новое подключение от %v. Текущее количество активных подключений: %d", conn.RemoteAddr(), currentConnections)

		// Обрабатываем соединение в отдельной горутине
		go func() {
			s.handleConnection(conn)
			remainingConnections := atomic.AddInt32(&s.activeConnections, -1)
			log.Printf("Соединение с %v закрыто. Осталось активных подключений: %d", conn.RemoteAddr(), remainingConnections)
		}()
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer s.handlePanic(conn)
	defer conn.Close()

	clientAddr := conn.RemoteAddr()
	log.Printf("Начало обработки соединения от %v", clientAddr)

	if err := conn.SetDeadline(time.Now().Add(s.config.ReadTimeout)); err != nil {
		log.Printf("Ошибка установки таймаута для %v: %v", clientAddr, err)
		return
	}

	challenge := utils.GenerateChallenge()
	log.Printf("Сгенерирован challenge для %v: %s", clientAddr, challenge)

	if err := s.sendChallenge(conn, challenge); err != nil {
		log.Printf("Ошибка отправки challenge клиенту %v: %v", clientAddr, err)
		return
	}

	difficulty := s.difficultyManager.GetDifficulty()
	log.Printf("Текущая сложность для %v: %d", clientAddr, difficulty)

	if err := s.sendDifficulty(conn); err != nil {
		log.Printf("Ошибка отправки сложности клиенту %v: %v", clientAddr, err)
		return
	}

	startTime := time.Now()
	if err := s.handleProofOfWork(conn, challenge, startTime); err != nil {
		log.Printf("Ошибка обработки proof-of-work для %v: %v", clientAddr, err)
		return
	}

	solveTime := time.Since(startTime)
	log.Printf("Клиент %v успешно решил задачу за %v", clientAddr, solveTime)
}

func (s *Server) handlePanic(conn net.Conn) {
	if r := recover(); r != nil {
		log.Printf("Восстановление после паники: %v\nСтек вызовов:\n%s", r, debug.Stack())
	}
}

func (s *Server) readNonce(conn net.Conn) (string, error) {
	reader := bufio.NewReader(conn)
	nonce, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("ошибка чтения nonce: %v", err)
	}

	nonce = strings.TrimSpace(nonce)
	log.Printf("Получен nonce: %s", nonce)

	if len(nonce) == 0 {
		log.Printf("Получен пустой nonce")
		fmt.Fprintln(conn, "Ошибка: пустой nonce")
		return "", fmt.Errorf("пустой nonce")
	}

	return nonce, nil
}

func (s *Server) sendChallenge(conn net.Conn, challenge string) error {
	log.Printf("Отправка challenge клиенту: %s", challenge)
	if _, err := fmt.Fprintln(conn, challenge); err != nil {
		log.Printf("Ошибка отправки challenge: %v", err)
		return err
	}
	return nil
}

func (s *Server) sendDifficulty(conn net.Conn) error {
	difficulty := s.difficultyManager.GetDifficulty()
	log.Printf("Отправка сложности клиенту: %d", difficulty)
	if _, err := fmt.Fprintln(conn, difficulty); err != nil {
		log.Printf("Ошибка отправки сложности: %v", err)
		return err
	}
	return nil
}

func (s *Server) sendQuote(conn net.Conn) error {
	quote := s.quoteStorage.GetRandomQuote()
	log.Printf("Отправка цитаты: %s", quote)
	if _, err := fmt.Fprintln(conn, quote); err != nil {
		log.Printf("Ошибка отправки цитаты: %v", err)
		return err
	}
	return nil
}

func (s *Server) handleProofOfWork(conn net.Conn, challenge string, startTime time.Time) error {
	return s.proofOfWork.HandleProofOfWork(conn, challenge, startTime)
}
