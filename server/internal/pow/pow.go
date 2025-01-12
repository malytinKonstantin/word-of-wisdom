package pow

import (
	"bufio"
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

type ProofOfWork struct {
	difficultyManager *DifficultyManager
	config            *config.Config
	quoteStorage      *storage.QuoteStorage
}

func New(config *config.Config, dm *DifficultyManager, qs *storage.QuoteStorage) *ProofOfWork {
	return &ProofOfWork{
		config:            config,
		difficultyManager: dm,
		quoteStorage:      qs,
	}
}

func (p *ProofOfWork) HandleProofOfWork(conn net.Conn, challenge string, startTime time.Time) error {
	nonce, err := readNonce(conn)
	if err != nil {
		return err
	}

	if !p.VerifyProofOfWork(challenge, nonce) {
		log.Printf("Неверное решение от клиента")
		fmt.Fprintln(conn, "Ошибка: неверное решение")
		return fmt.Errorf("неверное решение")
	}

	quote := p.quoteStorage.GetRandomQuote()
	if _, err := fmt.Fprintln(conn, quote); err != nil {
		return fmt.Errorf("ошибка отправки цитаты: %v", err)
	}

	return nil
}

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

func readNonce(conn net.Conn) (string, error) {
	reader := bufio.NewReader(conn)
	nonce, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("ошибка чтения nonce: %v", err)
	}
	return strings.TrimSpace(nonce), nil
}
