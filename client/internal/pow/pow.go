package pow

import (
	"crypto/sha256"
	"encoding/hex"
	"log"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

func SolveProofOfWork(challenge string, difficulty int) string {
	prefix := strings.Repeat("0", difficulty)
	numCPU := runtime.NumCPU()
	resultChan := make(chan string)
	var wg sync.WaitGroup

	log.Printf("🔄 Запуск решения PoW на %d CPU, требуемый префикс: '%s'", numCPU, prefix)
	startTime := time.Now()
	attemptsCounter := uint64(0)

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
				default:
					data := challenge + strconv.FormatInt(nonce, 10)
					hash := sha256.Sum256([]byte(data))
					hashStr := hex.EncodeToString(hash[:])
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
	log.Printf("✅ PoW решен за %v, всего попыток: %d", time.Since(startTime), attemptsCounter)
	return nonce
}
