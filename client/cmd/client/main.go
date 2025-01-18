package main

import (
	"context"
	"log"
	"os"
	"time"

	"word-of-wisdom-client/internal/config"
	"word-of-wisdom-client/internal/network"
	"word-of-wisdom-client/internal/pow"
)

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	log.SetOutput(os.Stdout)
}

func main() {
	cfg := config.NewDefault()
	log.Printf("Запуск клиента с конфигурацией: адрес сервера=%s, таймаут=%v",
		cfg.ServerAddr, cfg.Timeout)

	startTime := time.Now()
	conn, err := network.ConnectToServer(cfg)
	if err != nil {
		log.Fatalf("❌ Ошибка подключения к серверу: %v", err)
	}
	defer conn.Close()
	log.Printf("✅ Успешное подключение к серверу %s (заняло %v)",
		cfg.ServerAddr, time.Since(startTime))

	challenge, difficulty, err := network.ReceiveChallenge(conn)
	if err != nil {
		log.Fatalf("❌ Ошибка получения challenge: %v", err)
	}
	log.Printf("📥 Получен challenge='%s' и сложность=%d", challenge, difficulty)

	powStartTime := time.Now()
	log.Printf("⚙️ Начало решения Proof of Work...")

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	nonce, err := pow.SolveProofOfWork(ctx, challenge, difficulty)
	if err != nil {
		log.Fatalf("❌ Ошибка при решении Proof of Work: %v", err)
	}

	log.Printf("✅ Proof of Work решен за %v, найденный nonce='%s'",
		time.Since(powStartTime), nonce)

	if err := network.SendNonceAndGetQuote(conn, nonce); err != nil {
		log.Fatalf("❌ Ошибка при обмене данными с сервером: %v", err)
	}
	log.Printf("✨ Общее время работы клиента: %v", time.Since(startTime))
}
