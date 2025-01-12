package client

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"net"
	"strings"
	"testing"
	"time"

	"word-of-wisdom-client/internal/config"
	"word-of-wisdom-client/internal/network"
	"word-of-wisdom-client/internal/pow"
)

// Тестирование конфигурации
func TestConfig(t *testing.T) {
	cfg := config.NewDefault()

	if cfg.ServerAddr != "server:3333" {
		t.Errorf("Неверный ServerAddr, ожидалось 'server:3333', получено '%s'", cfg.ServerAddr)
	}

	expectedTimeout := 30 * time.Second
	if cfg.Timeout != expectedTimeout {
		t.Errorf("Неверный Timeout, ожидалось %v, получено %v", expectedTimeout, cfg.Timeout)
	}
}

// Тестирование Proof of Work
func TestSolveProofOfWork(t *testing.T) {
	testCases := []struct {
		name       string
		challenge  string
		difficulty int
	}{
		{
			name:       "Сложность 1",
			challenge:  "test1",
			difficulty: 1,
		},
		{
			name:       "Сложность 2",
			challenge:  "test2",
			difficulty: 2,
		},
		{
			name:       "Сложность 3",
			challenge:  "test3",
			difficulty: 3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			nonce := pow.SolveProofOfWork(tc.challenge, tc.difficulty)

			// Проверяем результат
			data := tc.challenge + nonce
			hash := sha256.Sum256([]byte(data))
			hashStr := hex.EncodeToString(hash[:])
			prefix := strings.Repeat("0", tc.difficulty)

			if !strings.HasPrefix(hashStr, prefix) {
				t.Errorf("Неверное решение PoW: хеш %s не начинается с %s", hashStr, prefix)
			}
		})
	}
}

// Тестирование функции чтения строк
func TestReadLine(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
		hasError bool
	}{
		{
			name:     "Обычная строка",
			input:    "test\n",
			expected: "test",
			hasError: false,
		},
		{
			name:     "Пустая строка",
			input:    "\n",
			expected: "",
			hasError: false,
		},
		{
			name:     "Строка с пробелами",
			input:    "  test  \n",
			expected: "test",
			hasError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reader := bufio.NewReader(strings.NewReader(tc.input))
			result, err := network.ReadLine(reader)

			if tc.hasError && err == nil {
				t.Error("Ожидалась ошибка, но её не было")
			}

			if !tc.hasError && err != nil {
				t.Errorf("Неожиданная ошибка: %v", err)
			}

			if result != tc.expected {
				t.Errorf("Ожидалось: %s, получено: %s", tc.expected, result)
			}
		})
	}
}

// Мок для тестирования сетевых функций
type mockConn struct {
	net.Conn
	readData  string
	writeData string
}

func (m *mockConn) Read(b []byte) (n int, err error) {
	copy(b, m.readData)
	return len(m.readData), nil
}

func (m *mockConn) Write(b []byte) (n int, err error) {
	m.writeData = string(b)
	return len(b), nil
}

func (m *mockConn) Close() error {
	return nil
}

// Тестирование отправки nonce и получения цитаты
func TestSendNonceAndGetQuote(t *testing.T) {
	mockConn := &mockConn{
		readData: "Тестовая цитата\n",
	}

	err := network.SendNonceAndGetQuote(mockConn, "12345")
	if err != nil {
		t.Errorf("Неожиданная ошибка при отправке nonce: %v", err)
	}

	expectedWrite := "12345\n"
	if mockConn.writeData != expectedWrite {
		t.Errorf("Неверные данные отправлены: ожидалось '%s', получено '%s'",
			expectedWrite, mockConn.writeData)
	}
}

// Тестирование получения challenge и сложности
func TestReceiveChallenge(t *testing.T) {
	mockConn := &mockConn{
		readData: "testchallenge\n4\n",
	}

	challenge, difficulty, err := network.ReceiveChallenge(mockConn)
	if err != nil {
		t.Errorf("Неожиданная ошибка при получении challenge: %v", err)
	}

	if challenge != "testchallenge" {
		t.Errorf("Неверный challenge: ожидался 'testchallenge', получен '%s'", challenge)
	}

	if difficulty != 4 {
		t.Errorf("Неверная сложность: ожидалось 4, получено %d", difficulty)
	}
}
