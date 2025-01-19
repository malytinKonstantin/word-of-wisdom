package pow

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"word-of-wisdom-client/internal/logger"
)

type DefaultPoWSolver struct{}

func NewDefaultPoWSolver() *DefaultPoWSolver {
	return &DefaultPoWSolver{}
}

func (ps *DefaultPoWSolver) SolveProofOfWork(ctx context.Context, challenge string, difficulty int) (string, error) {
	prefix := strings.Repeat("0", difficulty)
	numCPU := runtime.NumCPU()
	resultChan := make(chan string, 1)
	done := make(chan struct{})
	var wg sync.WaitGroup
	var attemptsCounter uint64

	logger.Log.Info().
		Int("num_cpu", numCPU).
		Str("prefix", prefix).
		Msg("Запуск решения Proof of Work")

	for i := 0; i < numCPU; i++ {
		wg.Add(1)
		go func(startNonce int64, workerID int) {
			defer wg.Done()
			nonce := startNonce
			localAttempts := uint64(0)

			challengeBytes := []byte(challenge)

			for {
				select {
				case <-ctx.Done():
					return
				case <-done:
					return
				default:
					data := append(challengeBytes, strconv.FormatInt(nonce, 10)...)
					hash := sha256.Sum256(data)
					hashStr := hex.EncodeToString(hash[:])
					atomic.AddUint64(&attemptsCounter, 1)
					localAttempts++

					if strings.HasPrefix(hashStr, prefix) {
						nonceStr := strconv.FormatInt(nonce, 10)
						logger.Log.Debug().
							Int("worker_id", workerID).
							Str("nonce", nonceStr).
							Str("hash", hashStr).
							Uint64("local_attempts", localAttempts).
							Msg("Найдено решение")
						select {
						case resultChan <- nonceStr:
						case <-done:
						}
						return
					}
					nonce += int64(numCPU)
				}
			}
		}(int64(i), i)
	}

	var nonce string
	select {
	case <-ctx.Done():
		close(done)
		wg.Wait()
		return "", ctx.Err()
	case nonce = <-resultChan:
		close(done)
		wg.Wait()
	}

	logger.Log.Info().
		Uint64("total_attempts", atomic.LoadUint64(&attemptsCounter)).
		Msg("Proof of Work решен")
	return nonce, nil
}
