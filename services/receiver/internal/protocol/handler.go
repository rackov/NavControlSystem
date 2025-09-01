package protocol

import (
	"context"
	"net"

	"github.com/nats-io/nats.go"
	"github.com/rackov/NavControlSystem/pkg/logger"
)

// Handler определяет интерфейс для обработчика протокола
type Handler interface {
	Process()
	SendCommand(cmd []byte) error
	Close() error
}

// Constructor - это тип для функции-конструктора обработчика.
// Пакеты, которые реализуют Handler, будут предоставлять такую функцию.
type Constructor func(conn net.Conn, nc *nats.Conn, topic string, log *logger.Logger, ctx context.Context) Handler
