package pow

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"strings"
	"time"

	"word-of-wisdom-server/internal/config"
	"word-of-wisdom-server/internal/interfaces"
	"word-of-wisdom-server/internal/logger"
)

// ProofOfWork содержит логику работы механизма Proof of Work
type ProofOfWork struct {
	difficultyManager interfaces.DifficultyManager
	config            *config.Config
	quoteStorage      interfaces.QuoteStorage
}

func New(config *config.Config, dm interfaces.DifficultyManager, qs interfaces.QuoteStorage) *ProofOfWork {
	return &ProofOfWork{
		config:            config,
		difficultyManager: dm,
		quoteStorage:      qs,
	}
}

var _ interfaces.ProofOfWorkHandler = (*ProofOfWork)(nil)

func (p *ProofOfWork) HandleProofOfWork(ctx context.Context, conn net.Conn, challenge string, startTime time.Time) error {
	clientAddr := conn.RemoteAddr().String()

	// Считываем nonce от клиента с использованием контекста
	nonce, err := p.readNonce(ctx, conn)
	if err != nil {
		logger.Log.Error().Err(err).Str("client", clientAddr).Msg("Ошибка чтения nonce")
		return err
	}
	logger.Log.Info().Str("client", clientAddr).Str("nonce", nonce).Msg("Получен nonce")

	// Проверяем корректность решения
	if !p.VerifyProofOfWork(challenge, nonce) {
		logger.Log.Error().Str("client", clientAddr).Str("challenge", challenge).Str("nonce", nonce).Msg("Неверное решение")
		fmt.Fprintln(conn, "Ошибка: неверное решение")
		return fmt.Errorf("неверное решение от %v", clientAddr)
	}

	// Вычисляем время решения и корректируем сложность
	solveTime := time.Since(startTime)
	logger.Log.Info().Str("client", clientAddr).Dur("solve_time", solveTime).Msg("Успешная верификация PoW")

	p.difficultyManager.AdjustDifficulty(solveTime)
	newDifficulty := p.difficultyManager.GetDifficulty()
	logger.Log.Info().Int("new_difficulty", newDifficulty).Msg("Сложность скорректирована")

	// Отправляем цитату клиенту
	quote := p.quoteStorage.GetRandomQuote()
	logger.Log.Info().Str("client", clientAddr).Str("quote", quote).Msg("Отправка цитаты клиенту")
	if _, err := fmt.Fprintln(conn, quote); err != nil {
		logger.Log.Error().Err(err).Str("client", clientAddr).Msg("Ошибка отправки цитаты")
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
		logger.Log.Debug().Str("nonce", nonce).Msg("Nonce успешно прочитан")
		return nonce, nil
	}
}
