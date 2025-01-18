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

func SolveProofOfWork(ctx context.Context, challenge string, difficulty int) (string, error) {
	prefix := strings.Repeat("0", difficulty)
	numCPU := runtime.NumCPU()
	resultChan := make(chan string, 1)
	done := make(chan struct{})
	var wg sync.WaitGroup
	var attemptsCounter uint64

	bufferPool := sync.Pool{
		New: func() interface{} {
			return make([]byte, 0, 64)
		},
	}

	log.Printf("🔄 Запуск решения PoW на %d CPU, требуемый префикс: '%s'", numCPU, prefix)
	startTime := time.Now()

	for i := 0; i < numCPU; i++ {
		wg.Add(1)
		go func(startNonce int64, workerID int) {
			defer wg.Done()
			nonce := startNonce
			localAttempts := uint64(0)

			buf := bufferPool.Get().([]byte)
			defer bufferPool.Put(buf)

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
					buf = buf[:0]
					buf = append(buf, challengeBytes...)
					buf = strconv.AppendInt(buf, nonce, 10)

					hash := sha256.Sum256(buf)
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

	// Ожидаем результат или отмену контекста
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

func hasLeadingZeros(hash []byte, difficulty int) bool {
	bytes := difficulty / 8
	bits := difficulty % 8

	for i := 0; i < bytes; i++ {
		if hash[i] != 0 {
			return false
		}
	}
	if bits > 0 {
		mask := byte(0xFF << (8 - bits))
		if hash[bytes]&mask != 0 {
			return false
		}
	}
	return true
}
