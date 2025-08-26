package internal

import (
	"github.com/rackov/NavControlSystem/pkg/models"
)

type Protocol interface {
	Start(port string) error
	Stop() error
	SendCommand(deviceID string, command []byte) error
	SetDataHandler(handler func(*models.NavRecord))
}

type ProtocolFactory interface {
	Create() Protocol
}
