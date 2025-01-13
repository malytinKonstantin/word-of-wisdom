package utils

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"sync"
)

// challengePool используется для повторного использования выделенной памяти
var challengePool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 20)
	},
}

// GenerateChallenge генерирует случайную строку для использования в качестве challenge
func GenerateChallenge() string {
	bytes := challengePool.Get().([]byte)
	defer challengePool.Put(bytes)

	if _, err := rand.Read(bytes); err != nil {
		log.Fatalf("Ошибка генерации challenge: %v", err)
	}
	return hex.EncodeToString(bytes)
}
