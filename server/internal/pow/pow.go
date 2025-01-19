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
	"word-of-wisdom-server/internal/container"
	"word-of-wisdom-server/internal/interfaces"
	"word-of-wisdom-server/internal/log"
)

// ProofOfWork содержит логику работы механизма Proof of Work
type ProofOfWork struct {
	difficultyManager interfaces.DifficultyManager
	config            *config.Config
	quoteStorage      interfaces.QuoteStorage
}

func New(c *container.Container) *ProofOfWork {
	config := c.Resolve("config").(*config.Config)
	dm := c.Resolve("difficultyManager").(interfaces.DifficultyManager)
	qs := c.Resolve("quoteStorage").(interfaces.QuoteStorage)

	return &ProofOfWork{
		config:            config,
		difficultyManager: dm,
		quoteStorage:      qs,
	}
}

var _ interfaces.ProofOfWorkHandler = (*ProofOfWork)(nil)

func (p *ProofOfWork) HandleProofOfWork(ctx context.Context, conn net.Conn, challenge string, startTime time.Time) error {
	clientAddr := conn.RemoteAddr().String()

	nonce, err := p.readNonce(ctx, conn)
	if err != nil {
		log.Error().Err(err).Str("client", clientAddr).Msg("Ошибка чтения nonce")
		return err
	}
	log.Info().Str("client", clientAddr).Str("nonce", nonce).Msg("Получен nonce")

	if !p.VerifyProofOfWork(challenge, nonce) {
		log.Error().Str("client", clientAddr).Str("challenge", challenge).Str("nonce", nonce).Msg("Неверное решение")
		fmt.Fprintln(conn, "Ошибка: неверное решение")
		return fmt.Errorf("неверное решение от %v", clientAddr)
	}

	solveTime := time.Since(startTime)
	log.Info().Str("client", clientAddr).Dur("solve_time", solveTime).Msg("Успешная верификация PoW")

	p.difficultyManager.AdjustDifficulty(solveTime)
	newDifficulty := p.difficultyManager.GetDifficulty()
	log.Info().Int("new_difficulty", newDifficulty).Msg("Сложность скорректирована")

	quote := p.quoteStorage.GetRandomQuote()
	log.Info().Str("client", clientAddr).Str("quote", quote).Msg("Отправка цитаты клиенту")
	if _, err := fmt.Fprintln(conn, quote); err != nil {
		log.Error().Err(err).Str("client", clientAddr).Msg("Ошибка отправки цитаты")
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
		log.Debug().Str("nonce", nonce).Msg("Nonce успешно прочитан")
		return nonce, nil
	}
}
