package pow

import (
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

func SolveProofOfWork(challenge string, difficulty int) string {
	prefix := strings.Repeat("0", difficulty)
	numCPU := runtime.NumCPU()
	resultChan := make(chan string)
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

			log.Printf("👷 Запущен worker %d с начальным nonce=%d", workerID, startNonce)
			workerStart := time.Now()

			for {
				select {
				case <-done:
					return
				default:
					data := challenge + strconv.FormatInt(nonce, 10)
					hash := sha256.Sum256([]byte(data))
					hashStr := hex.EncodeToString(hash[:])
					atomic.AddUint64(&attemptsCounter, 1)
					localAttempts++

					if strings.HasPrefix(hashStr, prefix) {
						nonceStr := strconv.FormatInt(nonce, 10)
						log.Printf("🎯 Worker %d нашел решение! nonce='%s', hash='%s', попыток=%d, время=%v",
							workerID, nonceStr, hashStr, localAttempts, time.Since(workerStart))
						resultChan <- nonceStr
						return
					}
					nonce += int64(numCPU)
				}
			}
		}(int64(i), i)
	}

	nonce := <-resultChan
	close(done)
	wg.Wait()
	log.Printf("✅ PoW решен за %v, всего попыток: %d", time.Since(startTime), attemptsCounter)
	return nonce
}
