package utils

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"word-of-wisdom-server/internal/logger"
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
		logger.Log.Fatal().Err(err).Msg("Ошибка генерации challenge")
	}
	return hex.EncodeToString(bytes)
}
