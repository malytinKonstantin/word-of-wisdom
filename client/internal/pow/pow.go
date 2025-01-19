package pow

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
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

	log.Printf("🔄 Запуск решения PoW на %d CPU, требуемый префикс: '%s'", numCPU, prefix)
	startTime := time.Now()

	for i := 0; i < numCPU; i++ {
		wg.Add(1)
		go func(startNonce int64, workerID int) {
			defer wg.Done()
			nonce := startNonce
			localAttempts := uint64(0)

			challengeBytes := []byte(challenge)

			log.Printf("👷 Запущен worker %d с начальным nonce=%d", workerID, startNonce)
			workerStart := time.Now()

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
						log.Printf("🎯 Worker %d нашел решение! nonce='%s', hash='%s', попыток=%d, время=%v",
							workerID, nonceStr, hashStr, localAttempts, time.Since(workerStart))
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

	log.Printf("✅ PoW решен за %v, всего попыток: %d", time.Since(startTime), atomic.LoadUint64(&attemptsCounter))
	return nonce, nil
}
