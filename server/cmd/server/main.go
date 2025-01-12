package main

import (
	"flag"
	"log"
	"runtime"
	"word-of-wisdom-server/internal/config"
	"word-of-wisdom-server/internal/server"
)

func main() {
	config := config.NewDefaultConfig()

	port := flag.String("port", config.Port, "Порт для запуска сервера")
	certPath := flag.String("cert", "certs/server.crt", "Путь к файлу сертификата")
	keyPath := flag.String("key", "certs/server.key", "Путь к файлу ключа")
	numCPU := flag.Int("cpu", runtime.NumCPU(), "Количество используемых CPU")
	maxConn := flag.Int("max-conn", config.MaxConnections, "Максимальное количество одновременных подключений")
	flag.Parse()

	config.MaxConnections = *maxConn
	runtime.GOMAXPROCS(*numCPU)

	server := server.New(config)

	if err := server.Run(*port, *certPath, *keyPath); err != nil {
		log.Fatal(err)
	}
}
