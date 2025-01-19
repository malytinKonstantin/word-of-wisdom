package utils

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"word-of-wisdom-server/internal/logger"
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
		logger.Log.Fatal().Err(err).Msg("Ошибка генерации challenge")
	}
	return hex.EncodeToString(bytes)
}
