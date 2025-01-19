package interfaces

import (
	"context"
	"net"
	"time"
)

type Server interface {
	Run(port, certPath, keyPath string) error
}

type ProofOfWorkHandler interface {
	HandleProofOfWork(ctx context.Context, conn net.Conn, challenge string, startTime time.Time) error
	VerifyProofOfWork(challenge, nonce string) bool
}

type DifficultyManager interface {
	GetDifficulty() int
	SetDifficulty(difficulty int)
	AdjustDifficulty(solveTime time.Duration)
}

type QuoteStorage interface {
	GetRandomQuote() string
}
