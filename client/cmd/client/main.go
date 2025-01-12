package main

import (
	"log"

	"word-of-wisdom-client/internal/config"
	"word-of-wisdom-client/internal/network"
	"word-of-wisdom-client/internal/pow"
)

func main() {
	cfg := config.NewDefault()

	conn, err := network.ConnectToServer(cfg)
	if err != nil {
		log.Fatalf("Ошибка подключения к серверу: %v", err)
	}
	defer conn.Close()

	challenge, difficulty, err := network.ReceiveChallenge(conn)
	if err != nil {
		log.Fatalf("Ошибка получения challenge: %v", err)
	}

	nonce := pow.SolveProofOfWork(challenge, difficulty)

	if err := network.SendNonceAndGetQuote(conn, nonce); err != nil {
		log.Fatalf("Ошибка при обмене данными с сервером: %v", err)
	}
}
