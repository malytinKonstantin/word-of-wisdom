package pow

import (
	"crypto/sha256"
	"encoding/hex"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

func SolveProofOfWork(challenge string, difficulty int) string {
	prefix := strings.Repeat("0", difficulty)
	numCPU := runtime.NumCPU()
	resultChan := make(chan string)
	var wg sync.WaitGroup

	for i := 0; i < numCPU; i++ {
		wg.Add(1)
		go func(startNonce int64) {
			defer wg.Done()
			nonce := startNonce
			for {
				select {
				default:
					data := challenge + strconv.FormatInt(nonce, 10)
					hash := sha256.Sum256([]byte(data))
					if strings.HasPrefix(hex.EncodeToString(hash[:]), prefix) {
						resultChan <- strconv.FormatInt(nonce, 10)
						return
					}
					nonce += int64(numCPU)
				}
			}
		}(int64(i))
	}

	nonce := <-resultChan
	return nonce
}
