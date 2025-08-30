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

	// ctx    context.Context    // Добавляем контекст для управления жизненным циклом
	// cancel context.CancelFunc // и функцию для его отмены

	connections map[string]*clientConnection // Ключ - адрес клиента
	clientData  ClientData                   // Зависимость для получения ID клиента

	// --- НОВЫЕ ПОЛЯ ДЛЯ УПРАВЛЕНИЯ КОНТЕКСТОМ ---
	internalCtx    context.Context
	internalCancel context.CancelFunc
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
func (cm *ConnectionManager) Start(parentCtx context.Context, port int, connectionHandler func(ctx context.Context, conn net.Conn, clientID string)) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.running {
		// Используем Warn, т.к. это нештатная ситуация
		logger.Warn("Connection manager is already running")
		return fmt.Errorf("connection manager is already running")
	}

	var err error
	// Создаем слушателя заново при каждом запуске, чтобы избежать ошибки "address already in use"
	cm.listener, err = net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		logger.Errorf("Failed to listen on port %d: %v", port, err)
		return fmt.Errorf("failed to listen on port %d: %w", port, err)
	}
	logger.Infof("Connection manager started listening on port %d", port)
	// --- ИЗМЕНЕНИЕ: Создаем внутренний контекст для управления жизненным циклом ---
	// parentCtx нужен, чтобы если родительский контекст (весь сервис) отменяется,
	// то и менеджер тоже остановился.
	cm.internalCtx, cm.internalCancel = context.WithCancel(parentCtx)

	cm.running = true
	cm.wg.Add(1)
	// Запускаем цикл приема соединений, передавая ему контекст менеджера
	// Передаем во внутренний контекст, а не внешний
	go func() {
		defer cm.wg.Done()
		cm.acceptLoop(cm.internalCtx, port, connectionHandler)
	}()

	return nil
}

// acceptLoop теперь использует внутренний контекст менеджера.
func (cm *ConnectionManager) acceptLoop(ctx context.Context, port int, connectionHandler func(ctx context.Context, conn net.Conn, clientID string)) {
	for {
		select {
		// --- ИЗМЕНЕНИЕ: Проверяем сигнал отмены из ВНУТРЕННЕГО контекста ---
		case <-ctx.Done():
			logger.Infof("Accept loop for port %d shutting down due to context cancellation.", port)
			return // Выходим из цикла, gracefully завершаем горутину

		default:
			conn, err := cm.listener.Accept()
			if err != nil {
				// Проверяем, не связана ли ошибка с нашим закрытием слушателя
				select {
				case <-ctx.Done():
					// Ошибка "use of closed network connection" ожидаема. Просто выходим.
					return
				default:
					logger.Errorf("Accept error on port %d: %v", port, err)
					time.Sleep(1 * time.Second)
					continue
				}
			}

			cm.wg.Add(1)
			go func(c net.Conn) {
				defer cm.wg.Done()
				// Передаем ВНУТРЕННИЙ контекст в обработчик соединения
				cm.handleNewConnection(ctx, c, port, connectionHandler)
			}(conn)
		}
	}
}

func (cm *ConnectionManager) handleNewConnection(parentCtx context.Context, conn net.Conn, port int, connectionHandler func(ctx context.Context, conn net.Conn, clientID string)) {
	defer conn.Close()

	clientAddr := conn.RemoteAddr().String()
	logger.Debugf("Handling new connection from %s", clientAddr)

	// 1. Авторизация через зависимость (обработчик протокола)
	clientID, err := cm.clientData.GetClientID(conn)
	if err != nil {
		logger.Errorf("Failed to authorize client %s: %v", clientAddr, err)
		return
	}
	//  protocolName:=cm.clientData.GetName()
	logger.Infof("Client %s authorized with ID: %s on port %d ", clientAddr, clientID, port)

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
		return nil
	}

	logger.Info("Stopping connection manager for protocol ")

	// --- ИЗМЕНЕНИЕ: Отправляем сигнал отмены через ВНУТРЕННЮЮ функцию ---
	if cm.internalCancel != nil {
		cm.internalCancel() // Это "разбудит" acceptLoop
	}

	// Отменяем контексты для всех активных подключений
	for addr, connInfo := range cm.connections {
		logger.Debugf("Cancelling context for client %s (ID: %s)", addr, connInfo.clientID)
		connInfo.cancelFunc()
		// connInfo.conn.Close() вызывается в defer внутри handleNewConnection
	}

	// Закрываем сам слушатель.
	if cm.listener != nil {
		if err := cm.listener.Close(); err != nil {
			logger.Errorf("Failed to close listener for protocol %s: %v", "cm.clientData.GetName()", err)
		}
	}

	cm.running = false
	cm.wg.Wait() // Ждем, пока acceptLoop и все handleConnection горутины завершатся

	logger.Infof("Connection manager for protocol %s stopped.", "cm.clientData.GetName()")
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

// GetName возвращает имя протокола.
