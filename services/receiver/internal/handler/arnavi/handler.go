package arnavi

import (
	"context"
	"net"
	"time"

	"github.com/rackov/NavControlSystem/pkg/logger"
	"github.com/rackov/NavControlSystem/services/receiver/internal/connectionmanager"
	"github.com/rackov/NavControlSystem/services/receiver/internal/protocol"
)

type ArnaviHandler struct {
	connManager *connectionmanager.ConnectionManager
	publisher   protocol.DataPublisher // Храним publisher для доступа в handleConnection
}

func NewArnaviHandler() *ArnaviHandler {
	// Передаем сам обработчик (который реализует ClientData) в конструктор менеджера
	h := &ArnaviHandler{}
	// Теперь, когда 'h' создан, мы можем передать его
	h.connManager = connectionmanager.NewConnectionManager(h)
	return h
}

// Start запускает обработчик, делегируя управление соединениями ConnectionManager
func (h *ArnaviHandler) Start(ctx context.Context, publisher protocol.DataPublisher, port int) error {
	h.publisher = publisher
	// Передаем функцию, которая будет вызываться для каждого авторизованного клиента
	return h.connManager.Start(ctx, port, h.handleConnection)
}

// GetName возвращает имя протокола
func (h *ArnaviHandler) GetName() string {
	return "ARNAVI"
}

// Stop останавливает ConnectionManager
func (h *ArnaviHandler) Stop() error {
	logger.Info("Stopping Arnavi handler...")

	return h.connManager.Stop()
}

// IsRunning проверяет состояние ConnectionManager
func (h *ArnaviHandler) IsRunning() bool {
	return h.connManager.IsRunning()
}

// --- Методы, которые просто делегируют вызовы ConnectionManager ---

func (h *ArnaviHandler) GetActiveConnectionsCount() int {
	return h.connManager.GetActiveConnectionsCount()
}

func (h *ArnaviHandler) GetConnectedClients() []protocol.ClientInfo {
	return h.connManager.GetConnectedClients()
}

func (h *ArnaviHandler) DisconnectClient(clientAddr string) error {
	return h.connManager.DisconnectClient(clientAddr)
}

// GetClientID реализует интерфейс ClientData для авторизации
func (h *ArnaviHandler) GetClientID(conn net.Conn) (string, error) {
	// // Устанавливаем таймаут на чтение, чтобы не зависнуть, если клиент ничего не присылает
	// conn.SetReadDeadline(time.Now().Add(5 * time.Second))

	// buf := make([]byte, 4)
	// n, err := conn.Read(buf)
	// if err != nil || n != 4 {
	// 	return "", fmt.Errorf("failed to read client ID: %w", err)
	// }

	// // Сбрасываем таймаут
	// conn.SetReadDeadline(time.Time{})

	// id := uint32(buf[0])<<24 | uint32(buf[1])<<16 | uint32(buf[2])<<8 | uint32(buf[3])
	return "1", nil //strconv.FormatUint(uint64(id), 10), nil
}

// handleConnection содержит логику, специфичную для Arnavi, после авторизации
func (h *ArnaviHandler) handleConnection(ctx context.Context, conn net.Conn, clientID string) {
	logger.Infof("Starting Arnavi data processing for client ID: %s", clientID)

	for {
		select {
		case <-ctx.Done():
			logger.Infof("Arnavi processing for client ID %s cancelled", clientID)
			if conn != nil {
				conn.Close()
			}
			return
		default:
			// ... логика парсинга пакета EGTS ...
			navData := protocol.NavRecord{Client: 1}
			// navData, err := ParseEgtsPacket(nil)
			// var err error
			// if err != nil {
			// 	logger.Errorf("Failed to parse Arnavi packet for client ID %s: %v", clientID, err)
			// 	continue
			// }

			if h.publisher.IsConnected() {
				if err := h.publisher.Publish(&navData); err != nil {
					logger.Errorf("Failed to publish Arnavi data for client ID %s: %v", clientID, err)
					// conn.Close()
				} else {
					logger.Debugf("Arnavi data for client ID %s published", clientID)
				}
			} else {
				if conn != nil {
					conn.Close()
				}

				logger.Warnf("NATS is not connected, Arnavi data for client ID %s not published", clientID)
			}
			time.Sleep(5 * time.Second)
		}
	}
}
