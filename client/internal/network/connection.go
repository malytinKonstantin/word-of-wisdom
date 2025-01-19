package network

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"
	"strconv"
	"strings"
	"time"

	"word-of-wisdom-client/internal/config"
	"word-of-wisdom-client/internal/logger"
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
		logger.Log.Fatal().Err(err).Msg("Не удалось прочитать сертификат сервера")
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
	logger.Log.Info().Str("server_address", serverAddr).Msg("Попытка подключения к серверу")
	conn, err := tls.Dial("tcp", serverAddr, nc.tlsConfig)
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к серверу: %v", err)
	}
	logger.Log.Info().Str("remote_address", conn.RemoteAddr().String()).Msg("TLS соединение установлено")
	return conn, nil
}

func (nc *DefaultNetworkClient) ReceiveChallenge(conn net.Conn) (string, int, error) {
	reader := bufio.NewReader(conn)

	logger.Log.Info().Msg("Ожидание challenge от сервера")
	challenge, err := readLine(reader)
	if err != nil {
		return "", 0, fmt.Errorf("ошибка получения challenge: %v", err)
	}
	logger.Log.Info().Str("challenge", challenge).Msg("Получен challenge")

	logger.Log.Info().Msg("Ожидание сложности от сервера")
	diffStr, err := readLine(reader)
	if err != nil {
		return "", 0, fmt.Errorf("ошибка получения сложности: %v", err)
	}

	difficulty, err := strconv.Atoi(diffStr)
	if err != nil {
		return "", 0, fmt.Errorf("некорректное значение сложности: %v", err)
	}
	logger.Log.Info().Int("difficulty", difficulty).Msg("Получена сложность")

	return challenge, difficulty, nil
}

func (nc *DefaultNetworkClient) SendNonceAndGetQuote(conn net.Conn, nonce string) error {
	logger.Log.Info().Str("nonce", nonce).Msg("Отправка nonce серверу")
	startTime := time.Now()

	if _, err := fmt.Fprintln(conn, nonce); err != nil {
		return fmt.Errorf("ошибка отправки nonce: %v", err)
	}

	reader := bufio.NewReader(conn)
	response, err := readLine(reader)
	if err != nil {
		return fmt.Errorf("ошибка получения ответа: %v", err)
	}

	logger.Log.Info().
		Dur("response_time", time.Since(startTime)).
		Str("quote", response).
		Msg("Получена цитата от сервера")
	return nil
}

func readLine(reader *bufio.Reader) (string, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}
