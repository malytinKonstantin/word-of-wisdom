package interfaces

import (
	"context"
	"net"
)

type NetworkClient interface {
	Connect(serverAddr string) (net.Conn, error)
	ReceiveChallenge(conn net.Conn) (string, int, error)
	SendNonceAndGetQuote(conn net.Conn, nonce string) error
}

type PoWSolver interface {
	SolveProofOfWork(ctx context.Context, challenge string, difficulty int) (string, error)
}
