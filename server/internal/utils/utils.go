package utils

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"sync"
)

var challengePool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 20)
	},
}

func GenerateChallenge() string {
	bytes := challengePool.Get().([]byte)
	defer challengePool.Put(bytes)

	if _, err := rand.Read(bytes); err != nil {
		log.Fatalf("Ошибка генерации challenge: %v", err)
	}
	return hex.EncodeToString(bytes)
}
