package internal

type Protocol interface {
	Start(port string) error
	Stop() error
	SendCommand(deviceID string, command []byte) error
	SetDataHandler(handler func(*model.NavRecord))
}

type ProtocolFactory interface {
	Create() Protocol
}
