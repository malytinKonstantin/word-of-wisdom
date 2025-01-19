package interfaces

import (
	"context"
	"net"
	"time"
)

// Server представляет интерфейс для сервера
type Server interface {
	Run(port, certPath, keyPath string) error
}

// ProofOfWorkHandler определяет методы для обработки Proof of Work
type ProofOfWorkHandler interface {
	HandleProofOfWork(ctx context.Context, conn net.Conn, challenge string, startTime time.Time) error
	VerifyProofOfWork(challenge, nonce string) bool
}

// DifficultyManager определяет методы для управления сложностью
type DifficultyManager interface {
	GetDifficulty() int
	SetDifficulty(difficulty int)
	AdjustDifficulty(solveTime time.Duration)
}

// QuoteStorage представляет интерфейс для хранилища цитат
type QuoteStorage interface {
	GetRandomQuote() string
}
