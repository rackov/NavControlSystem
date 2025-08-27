package connectionmanager

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/rackov/NavControlSystem/pkg/logger"
	"github.com/rackov/NavControlSystem/services/receiver/internal/protocol"
)

// ClientData — это интерфейс, который должен реализовывать обработчик протокола,
// чтобы получить данные для авторизации клиента.
type ClientData interface {
	GetClientID(conn net.Conn) (string, error)
}

// ConnectionManager управляет всеми активными TCP-подключениями для одного протокола.
type ConnectionManager struct {
	listener net.Listener
	wg       sync.WaitGroup
	mu       sync.Mutex
	running  bool

	connections map[string]*clientConnection // Ключ - адрес клиента
	clientData  ClientData                   // Зависимость для получения ID клиента
}

// clientConnection хранит данные о конкретном подключении.
type clientConnection struct {
	conn        net.Conn
	clientID    string
	connectedAt time.Time
	cancelFunc  context.CancelFunc
}

// NewConnectionManager создает новый экземпляр менеджера.
func NewConnectionManager(cd ClientData) *ConnectionManager {
	return &ConnectionManager{
		connections: make(map[string]*clientConnection),
		clientData:  cd,
	}
}

// Start запускает прослушивание порта и принимает подключения.
func (cm *ConnectionManager) Start(ctx context.Context, port int, connectionHandler func(ctx context.Context, conn net.Conn, clientID string)) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.running {
		// Используем Warn, т.к. это нештатная ситуация
		logger.Warn("Connection manager is already running")
		return fmt.Errorf("connection manager is already running")
	}

	var err error
	cm.listener, err = net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		logger.Errorf("Failed to listen on port %d: %v", port, err)
		return fmt.Errorf("failed to listen on port %d: %w", port, err)
	}
	logger.Infof("Connection manager started listening on port %d", port)

	cm.running = true
	cm.wg.Add(1)

	go func() {
		defer cm.wg.Done()
		cm.acceptLoop(ctx, connectionHandler)
	}()

	return nil
}

// acceptLoop — метод ConnectionManager, который непрерывно принимает входящие соединения
// @param ctx: Контекст, используемый для корректного завершения работы
// @param connectionHandler: Функция, обрабатывающая новые соединения с указанием контекста, соединения и идентификатора клиента
func (cm *ConnectionManager) acceptLoop(ctx context.Context, connectionHandler func(ctx context.Context, conn net.Conn, clientID string)) {
	// Infinite loop to continuously accept new connections
	for {
		// Accept a new connection from the listener
		conn, err := cm.listener.Accept()
		if err != nil {
			// Check if the context is done (shutdown signal)
			select {
			case <-ctx.Done():
				// If context is done, log shutdown info and exit
				logger.Info("Connection manager shutting down...")
				return
			default:
				// Ошибки принятия соединения логируем, но не останавливаем сервис
				logger.Errorf("Accept error: %v", err)
				time.Sleep(1 * time.Second)
				continue
			}
		}

		cm.wg.Add(1)
		go func(c net.Conn) {
			defer cm.wg.Done()
			cm.handleNewConnection(ctx, c, connectionHandler)
		}(conn)
	}
}

func (cm *ConnectionManager) handleNewConnection(parentCtx context.Context, conn net.Conn, connectionHandler func(ctx context.Context, conn net.Conn, clientID string)) {
	defer conn.Close()

	clientAddr := conn.RemoteAddr().String()
	logger.Debugf("Handling new connection from %s", clientAddr)

	// 1. Авторизация через зависимость (обработчик протокола)
	clientID, err := cm.clientData.GetClientID(conn)
	if err != nil {
		logger.Errorf("Failed to authorize client %s: %v", clientAddr, err)
		return
	}
	logger.Infof("Client %s authorized with ID: %s", clientAddr, clientID)

	// 2. Регистрация подключения
	connCtx, cancel := context.WithCancel(parentCtx)

	cm.mu.Lock()
	cm.connections[clientAddr] = &clientConnection{
		conn:        conn,
		clientID:    clientID,
		connectedAt: time.Now(),
		cancelFunc:  cancel,
	}
	cm.mu.Unlock()

	// 3. Удаление при выходе
	defer func() {
		cm.mu.Lock()
		delete(cm.connections, clientAddr)
		cm.mu.Unlock()
		logger.Infof("Client %s (ID: %s) disconnected", clientAddr, clientID)
	}()

	// 4. Передача управления обработчику протокола
	logger.Debugf("Passing control for client %s (ID: %s) to protocol handler", clientAddr, clientID)
	connectionHandler(connCtx, conn, clientID)
}

// Stop останавливает менеджер и все активные подключения.
func (cm *ConnectionManager) Stop() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if !cm.running {
		logger.Debug("Connection manager is not running, nothing to stop")
		return nil
	}

	logger.Info("Stopping connection manager...")

	// Отменяем контексты для всех активных подключений
	for addr, connInfo := range cm.connections {
		logger.Debugf("Cancelling context for client %s (ID: %s)", addr, connInfo.clientID)
		connInfo.cancelFunc()
		if err := connInfo.conn.Close(); err != nil {
			// Ошибку при закрытии логируем, но не прерываем процесс
			logger.Errorf("Error closing connection for client %s: %v", addr, err)
		}
	}

	if cm.listener != nil {
		if err := cm.listener.Close(); err != nil {
			logger.Errorf("Failed to close listener: %v", err)
			return fmt.Errorf("failed to close listener: %w", err)
		}
	}

	cm.running = false
	cm.wg.Wait()
	logger.Info("Connection manager stopped.")
	return nil
}

// --- Методы для получения информации и управления ---

func (cm *ConnectionManager) GetActiveConnectionsCount() int {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	count := len(cm.connections)
	logger.Debugf("Current active connections count: %d", count)
	return count
}

// GetConnectedClients возвращает срез с информацией о всех подключенных клиентах
func (cm *ConnectionManager) GetConnectedClients() []protocol.ClientInfo {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	clients := make([]protocol.ClientInfo, 0, len(cm.connections))
	for addr, connInfo := range cm.connections {
		clients = append(clients, protocol.ClientInfo{
			ID:    connInfo.clientID,
			Addr:  addr,
			Since: connInfo.connectedAt,
		})
	}
	logger.Debugf("Retrieved list of %d connected clients", len(clients))
	return clients
}

// DisconnectClient находит подключение по адресу и закрывает его
func (cm *ConnectionManager) DisconnectClient(clientAddr string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	connInfo, found := cm.connections[clientAddr]
	if !found {
		logger.Warnf("Попытка отключить несуществующего клиента с адресом %s", clientAddr)
		return fmt.Errorf("client with address %s not found", clientAddr)
	}

	logger.Infof("Отключение клиента %s (ID: %s) по запросу сервера", clientAddr, connInfo.clientID)
	connInfo.cancelFunc()
	if err := connInfo.conn.Close(); err != nil {
		logger.Errorf("Ошибка закрытия соединения для клиента %s во время отключения: %v", clientAddr, err)
		// Не возвращаем ошибку, т.к. цель (отключение клиента) достигнута
	}
	return nil
}

func (cm *ConnectionManager) IsRunning() bool {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	return cm.running
}
