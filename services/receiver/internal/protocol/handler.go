package protocol

import (
	"context"
	"encoding/json"
	"time"
)

type NavRecord struct {
	Client              uint32       `json:"tid"`
	PacketID            uint32       `json:"pk_id"`
	NavigationTimestamp uint32       `json:"nav_time"`
	ReceivedTimestamp   uint32       `json:"rec_time"`
	Latitude            uint32       `json:"lat"`
	Longitude           uint32       `json:"lng"`
	Speed               uint16       `json:"speed"`
	FlagPos             byte         `json:"flag"`
	Pdop                uint16       `json:"pdop"`
	Hdop                uint16       `json:"hdop"`
	Vdop                uint16       `json:"vdop"`
	Nsat                uint8        `json:"nsat"`
	Ns                  uint16       `json:"ns"`
	DigInput            byte         `json:"din"`
	Odometer            uint32       `json:"odm"`
	Course              uint8        `json:"course"`
	Imei                string       `json:"imei"`
	Imsi                string       `json:"imsi"`
	AnSenAbs            []Sensor     `json:"in_abs_in"`  // одного аналогового входа
	DigSenAbs           []DiSensor   `json:"in_abs_dig"` // одного дискретного входа
	AnSensors           []DopAnIn    `json:"in_an"`      // дополнительных аналоговых входов
	DigSenonrs          []DopDigIn   `json:"in_dig"`     // дополнительных дискретного входа
	DigSenOuts          []int        `json:"out_dig"`    // дополнительных дискретного выхода
	LiquidSensors       LiquidSensor `json:"sn_liq"`     // данных о показаниях ДУТ
}

func (eep *NavRecord) ToBytes() ([]byte, error) {
	return json.Marshal(eep)
}

type DopAnIn struct {
	Asfe byte      `json:"asfe"`
	Ansi [8]uint32 `json:"ansi"`
}

type DopDigIn struct {
	Dioe byte    `json:"dioe"`
	Adio [8]byte `json:"adio"`
}

type DiSensor struct {
	StateNumber byte `json:"st_num"`
	Number      byte `json:"num"`
}

type LiquidSensor struct {
	FlagLiqNum uint8     `json:"fl_ln"`
	Value      [8]uint32 `json:"val_l"`
}

type Sensor struct {
	SensorNumber uint8  `json:"sn_n"`
	Value        uint32 `json:"val"`
}

type NavRecords struct {
	PacketType int16       `json:"packet_type"`
	PacketID   uint32      `json:"packet_id"`
	RecNav     []NavRecord `json:"record"`
}

// DataPublisher - интерфейс для публикации данных (будет реализован основным сервисом)
type DataPublisher interface {
	Publish(data *NavRecord) error
	IsConnected() bool
}

// ClientInfo содержит информацию о подключенном клиенте
type ClientInfo struct {
	ID    string    // ID устройства (например, из EGTS)
	Addr  string    // Сетевой адрес клиента (IP:Port)
	Since time.Time // Время подключения
}

// ProtocolHandler - интерфейс, который должен реализовать каждый обработчик протокола
type ProtocolHandler interface {
	// GetName возвращает имя протокола (EGTS, Arnavi и т.д.)
	GetName() string
	// Start запускает прослушивание TCP-порта
	Start(ctx context.Context, publisher DataPublisher, port int) error
	// Stop останавливает прослушивание и закрывает порт
	Stop() error
	// IsRunning проверяет, запущен ли обработчик
	IsRunning() bool

	// --- Новые методы ---

	// GetActiveConnectionsCount возвращает количество активных подключений
	GetActiveConnectionsCount() int

	// GetConnectedClients возвращает список информации о подключенных клиентах
	GetConnectedClients() []ClientInfo

	// DisconnectClient отключает конкретного клиента по его сетевому адресу
	DisconnectClient(clientAddr string) error
}
