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

type DefaultNetworkClient struct {
	cfg       *config.Config
	tlsConfig *tls.Config
}

func NewDefaultNetworkClient(cfg *config.Config) *DefaultNetworkClient {
	// Загружаем сертификат сервера для TLS
	certPool := x509.NewCertPool()
	serverCert, err := ioutil.ReadFile("certs/server.crt")
	if err != nil {
		log.Fatalf("Не удалось прочитать сертификат сервера: %v", err)
	}
	certPool.AppendCertsFromPEM(serverCert)

	tlsConfig := &tls.Config{
		RootCAs:    certPool,
		MinVersion: tls.VersionTLS12,
	}

	return &DefaultNetworkClient{
		cfg:       cfg,
		tlsConfig: tlsConfig,
	}
}

func (nc *DefaultNetworkClient) Connect(serverAddr string) (net.Conn, error) {
	log.Printf("🔌 Попытка подключения к %s...", serverAddr)
	conn, err := tls.Dial("tcp", serverAddr, nc.tlsConfig)
	if err != nil {
		return nil, fmt.Errorf("Ошибка подключения к серверу: %v", err)
	}
	log.Printf("✅ TLS соединение установлено с %s", conn.RemoteAddr())
	return conn, nil
}

func (nc *DefaultNetworkClient) ReceiveChallenge(conn net.Conn) (string, int, error) {
	reader := bufio.NewReader(conn)

	log.Print("📥 Ожидание challenge от сервера...")
	challenge, err := readLine(reader)
	if err != nil {
		return "", 0, fmt.Errorf("Ошибка получения challenge: %v", err)
	}
	log.Printf("✅ Получен challenge: '%s'", challenge)

	log.Print("📥 Ожидание сложности от сервера...")
	diffStr, err := readLine(reader)
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

func (nc *DefaultNetworkClient) SendNonceAndGetQuote(conn net.Conn, nonce string) error {
	log.Printf("📤 Отправка nonce='%s' серверу...", nonce)
	startTime := time.Now()

	if _, err := fmt.Fprintln(conn, nonce); err != nil {
		return fmt.Errorf("Ошибка отправки nonce: %v", err)
	}

	reader := bufio.NewReader(conn)
	response, err := readLine(reader)
	if err != nil {
		return fmt.Errorf("Ошибка получения ответа: %v", err)
	}

	log.Printf("✨ Получена цитата (за %v):", time.Since(startTime))
	log.Printf("📜 %s", response)
	return nil
}

func readLine(reader *bufio.Reader) (string, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}
