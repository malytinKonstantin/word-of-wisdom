package pow

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"word-of-wisdom-server/internal/config"
	"word-of-wisdom-server/internal/storage"
)

// ProofOfWork содержит логику работы механизма Proof of Work
type ProofOfWork struct {
	difficultyManager *DifficultyManager    // Менеджер сложности
	config            *config.Config        // Конфигурация сервера
	quoteStorage      *storage.QuoteStorage // Хранилище цитат
}

func New(config *config.Config, dm *DifficultyManager, qs *storage.QuoteStorage) *ProofOfWork {
	return &ProofOfWork{
		config:            config,
		difficultyManager: dm,
		quoteStorage:      qs,
	}
}

func (p *ProofOfWork) HandleProofOfWork(ctx context.Context, conn net.Conn, challenge string, startTime time.Time) error {
	clientAddr := conn.RemoteAddr()

	// Считываем nonce от клиента с использованием контекста
	nonce, err := p.readNonce(ctx, conn)
	if err != nil {
		log.Printf("Ошибка чтения nonce от %v: %v", clientAddr, err)
		return err
	}
	log.Printf("Получен nonce от %v: %s", clientAddr, nonce)

	// Проверяем корректность решения
	if !p.VerifyProofOfWork(challenge, nonce) {
		log.Printf("Неверное решение от %v. Challenge: %s, Nonce: %s", clientAddr, challenge, nonce)
		fmt.Fprintln(conn, "Ошибка: неверное решение")
		return fmt.Errorf("неверное решение от %v", clientAddr)
	}

	// Вычисляем время решения и корректируем сложность
	solveTime := time.Since(startTime)
	log.Printf("Успешная верификация PoW от %v. Время решения: %v", clientAddr, solveTime)

	p.difficultyManager.AdjustDifficulty(solveTime)
	newDifficulty := p.difficultyManager.GetDifficulty()
	log.Printf("Сложность скорректирована. Новое значение: %d", newDifficulty)

	// Отправляем цитату клиенту
	quote := p.quoteStorage.GetRandomQuote()
	log.Printf("Отправка цитаты клиенту %v: %s", clientAddr, quote)
	if _, err := fmt.Fprintln(conn, quote); err != nil {
		log.Printf("Ошибка отправки цитаты клиенту %v: %v", clientAddr, err)
		return fmt.Errorf("ошибка отправки цитаты: %v", err)
	}

	return nil
}

// VerifyProofOfWork проверяет корректность решения PoW клиента
func (p *ProofOfWork) VerifyProofOfWork(challenge, nonce string) bool {
	if challenge == "" || nonce == "" {
		return false
	}

	data := challenge + nonce
	hash := sha256.Sum256([]byte(data))
	hashStr := hex.EncodeToString(hash[:])
	prefix := strings.Repeat("0", p.difficultyManager.GetDifficulty())
	return strings.HasPrefix(hashStr, prefix)
}

// readNonce считывает nonce от клиента
func (p *ProofOfWork) readNonce(ctx context.Context, conn net.Conn) (string, error) {
	reader := bufio.NewReader(conn)

	nonceChan := make(chan string, 1)
	errChan := make(chan error, 1)

	go func() {
		nonce, err := reader.ReadString('\n')
		if err != nil {
			errChan <- fmt.Errorf("ошибка чтения nonce: %v", err)
			return
		}
		nonceChan <- strings.TrimSpace(nonce)
	}()

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case err := <-errChan:
		return "", err
	case nonce := <-nonceChan:
		return nonce, nil
	}
}
