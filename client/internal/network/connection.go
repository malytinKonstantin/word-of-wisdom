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

	"word-of-wisdom-client/internal/config"
)

func ConnectToServer(cfg *config.Config) (*tls.Conn, error) {
	certPool := x509.NewCertPool()
	serverCert, err := ioutil.ReadFile("certs/server.crt")
	if err != nil {
		return nil, fmt.Errorf("Не удалось прочитать сертификат сервера: %v", err)
	}
	certPool.AppendCertsFromPEM(serverCert)

	tlsConfig := &tls.Config{
		RootCAs:    certPool,
		MinVersion: tls.VersionTLS12,
	}

	conn, err := tls.Dial("tcp", cfg.ServerAddr, tlsConfig)
	if err != nil {
		return nil, fmt.Errorf("Ошибка подключения к серверу: %v", err)
	}
	return conn, nil
}

func ReceiveChallenge(conn net.Conn) (string, int, error) {
	reader := bufio.NewReader(conn)
	challenge, err := ReadLine(reader)
	if err != nil {
		return "", 0, fmt.Errorf("Ошибка получения challenge: %v", err)
	}

	diffStr, err := ReadLine(reader)
	if err != nil {
		return "", 0, fmt.Errorf("Ошибка получения сложности: %v", err)
	}
	difficulty, err := strconv.Atoi(diffStr)
	if err != nil {
		return "", 0, fmt.Errorf("Некорректное значение сложности: %v", err)
	}

	return challenge, difficulty, nil
}

func SendNonceAndGetQuote(conn net.Conn, nonce string) error {
	if _, err := fmt.Fprintln(conn, nonce); err != nil {
		return fmt.Errorf("Ошибка отправки nonce: %v", err)
	}

	reader := bufio.NewReader(conn)
	response, err := ReadLine(reader)
	if err != nil {
		return fmt.Errorf("Ошибка получения ответа: %v", err)
	}

	fmt.Println("Цитата от сервера:", response)
	return nil
}

func ReadLine(reader *bufio.Reader) (string, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}
