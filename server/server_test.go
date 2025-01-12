package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"word-of-wisdom-server/internal/config"
	"word-of-wisdom-server/internal/pow"
	"word-of-wisdom-server/internal/storage"
)

func TestVerifyProofOfWork(t *testing.T) {
	cfg := config.NewDefaultConfig()
	dm := pow.NewDifficultyManager(cfg)
	qs := storage.New()
	powHandler := pow.New(cfg, dm, qs)

	challenge := "testchallenge"
	difficulty := 2

	// Устанавливаем фиксированную сложность для теста
	dm.SetDifficulty(difficulty)

	// Генерируем корректный nonce
	nonce, err := findValidNonce(challenge, difficulty)
	if err != nil {
		t.Fatalf("Не удалось найти подходящий nonce для теста: %v", err)
	}

	// Проверяем функцию VerifyProofOfWork
	if !powHandler.VerifyProofOfWork(challenge, nonce) {
		t.Errorf("VerifyProofOfWork вернула false для корректных данных")
	}

	// Проверяем на некорректный nonce
	wrongNonce := "invalid_nonce"
	if powHandler.VerifyProofOfWork(challenge, wrongNonce) {
		t.Errorf("VerifyProofOfWork вернула true для некорректных данных")
	}
}

func findValidNonce(challenge string, difficulty int) (string, error) {
	prefix := strings.Repeat("0", difficulty)
	for i := 0; i < 1000000; i++ {
		nonce := strconv.Itoa(i)
		data := challenge + nonce
		hash := sha256.Sum256([]byte(data))
		hashStr := hex.EncodeToString(hash[:])
		if strings.HasPrefix(hashStr, prefix) {
			return nonce, nil
		}
	}
	return "", fmt.Errorf("Не найден подходящий nonce после 1 000 000 попыток")
}

func TestAdjustDifficulty(t *testing.T) {
	cfg := config.NewDefaultConfig()
	dm := pow.NewDifficultyManager(cfg)

	// Тестируем увеличение сложности
	dm.SetDifficulty(3)
	dm.AdjustDifficulty(3 * time.Second) // Время меньше MinSolveTime
	if dm.GetDifficulty() != 4 {
		t.Errorf("Ожидалась сложность 4, но получена %d", dm.GetDifficulty())
	}

	// Тестируем уменьшение сложности
	dm.SetDifficulty(4)
	dm.AdjustDifficulty(11 * time.Second) // Время больше MaxSolveTime
	if dm.GetDifficulty() != 3 {
		t.Errorf("Ожидалась сложность 3, но получена %d", dm.GetDifficulty())
	}

	// Проверяем пределы сложности (максимум)
	dm.SetDifficulty(cfg.MaxDifficulty)
	dm.AdjustDifficulty(3 * time.Second)
	if dm.GetDifficulty() != cfg.MaxDifficulty {
		t.Errorf("Сложность не должна превышать %d", cfg.MaxDifficulty)
	}

	// Проверяем пределы сложности (минимум)
	dm.SetDifficulty(cfg.MinDifficulty)
	dm.AdjustDifficulty(11 * time.Second)
	if dm.GetDifficulty() != cfg.MinDifficulty {
		t.Errorf("Сложность не должна быть меньше %d", cfg.MinDifficulty)
	}
}

func TestGetRandomQuote(t *testing.T) {
	qs := storage.New()
	quote := qs.GetRandomQuote()

	if quote == "" {
		t.Errorf("Получена пустая цитата")
	}

	// Проверяем, что цитата входит в список возможных цитат
	allQuotes := qs.GetAllQuotes()
	found := false
	for _, q := range allQuotes {
		if q == quote {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Цитата не найдена в списке цитат")
	}
}
