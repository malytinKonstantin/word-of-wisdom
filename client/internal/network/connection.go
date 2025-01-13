package network

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"word-of-wisdom-client/internal/config"
)

// ConnectToServer устанавливает TLS-соединение с сервером по заданному адресу
func ConnectToServer(cfg *config.Config) (*tls.Conn, error) {
	log.Printf("🔒 Загрузка сертификата сервера...")
	// Загружаем сертификат сервера для проверки подлинности
	certPool := x509.NewCertPool()
	serverCert, err := ioutil.ReadFile("certs/server.crt")
	if err != nil {
		return nil, fmt.Errorf("Не удалось прочитать сертификат сервера: %v", err)
	}
	certPool.AppendCertsFromPEM(serverCert)
	log.Printf("✅ Сертификат сервера успешно загружен")

	// Настраиваем параметры TLS-соединения
	tlsConfig := &tls.Config{
		RootCAs:    certPool,         // Устанавливаем корневой сертификат сервера
		MinVersion: tls.VersionTLS12, // Используем минимум TLS версии 1.2
	}

	log.Printf("🔌 Попытка подключения к %s...", cfg.ServerAddr)
	// Устанавливаем TLS-соединение с сервером
	conn, err := tls.Dial("tcp", cfg.ServerAddr, tlsConfig)
	if err != nil {
		return nil, fmt.Errorf("Ошибка подключения к серверу: %v", err)
	}
	log.Printf("✅ TLS соединение установлено с %s", conn.RemoteAddr())
	return conn, nil
}

// ReceiveChallenge получает challenge и сложность для PoW от сервера
func ReceiveChallenge(conn net.Conn) (string, int, error) {
	reader := bufio.NewReader(conn)

	log.Print("📥 Ожидание challenge от сервера...")
	challenge, err := ReadLine(reader)
	if err != nil {
		return "", 0, fmt.Errorf("Ошибка получения challenge: %v", err)
	}
	log.Printf("✅ Получен challenge: '%s'", challenge)

	log.Print("📥 Ожидание сложности от сервера...")
	diffStr, err := ReadLine(reader)
	if err != nil {
		return "", 0, fmt.Errorf("Ошибка получения сложности: %v", err)
	}

	difficulty, err := strconv.Atoi(diffStr)
	if err != nil {
		return "", 0, fmt.Errorf("Некорректное значение сложности: %v", err)
	}
	log.Printf("✅ Получена сложность: %d", difficulty)

	return challenge, difficulty, nil
}

// SendNonceAndGetQuote отправляет найденный nonce серверу и получает цитату
func SendNonceAndGetQuote(conn net.Conn, nonce string) error {
	log.Printf("📤 Отправка nonce='%s' серверу...", nonce)
	startTime := time.Now()

	if _, err := fmt.Fprintln(conn, nonce); err != nil {
		return fmt.Errorf("Ошибка отправки nonce: %v", err)
	}

	reader := bufio.NewReader(conn)
	response, err := ReadLine(reader)
	if err != nil {
		return fmt.Errorf("Ошибка получения ответа: %v", err)
	}

	log.Printf("✨ Получена цитата (за %v):", time.Since(startTime))
	log.Printf("📜 %s", response)
	return nil
}

func ReadLine(reader *bufio.Reader) (string, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}
