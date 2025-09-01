// services/receiver/internal/portmanager/manager.go
package portmanager

import (
	"context"
	"fmt"
	"net"
	"os"
	"runtime"
	"sync"

	"github.com/nats-io/nats.go"
	"github.com/rackov/NavControlSystem/pkg/logger"
	"github.com/rackov/NavControlSystem/proto"
	"github.com/rackov/NavControlSystem/services/receiver/configs"

	"github.com/rackov/NavControlSystem/services/receiver/internal/handler/arnavi"
	"github.com/rackov/NavControlSystem/services/receiver/internal/protocol"
	"github.com/sirupsen/logrus"
)

// PortState хранит всю информацию о работающем порте
type PortState struct {
	Config     *configs.PortConfig
	Listener   net.Listener
	NatsConn   *nats.Conn
	CancelFunc context.CancelFunc // Для graceful shutdown
}

// Глобальная карта для регистрации обработчиков.
var handlers = make(map[string]protocol.Constructor)

// RegisterHandler регистрирует конструктор обработчика в глобальной карте.
// Эту функцию будут вызывать из других пакетов (например, из handlers.go).
func RegisterHandler(name string, constructor protocol.Constructor) {
	// ---- ДОБАВЛЯЕМ ЛОГИРОВАНИЕ ----
	if constructor == nil {
		fmt.Fprintf(os.Stderr, "PANIC PREVENTION: Attempted to register a nil constructor for protocol %s!\n", name)
		// Получаем информацию о том, кто вызвал эту функцию
		_, file, line, _ := runtime.Caller(1)
		fmt.Fprintf(os.Stderr, "PANIC PREVENTION: Called from %s:%d\n", file, line)
		// Можно даже вызвать панику с более понятным сообщением
		panic(fmt.Sprintf("attempted to register nil constructor for protocol %s", name))
	}
	// ---- КОНЕЦ ЛОГИРОВАНИЯ ----

	if _, exists := handlers[name]; exists {
		fmt.Fprintf(os.Stderr, "WARNING: Handler for protocol %s already registered, overwriting.\n", name)
	}
	handlers[name] = constructor
	fmt.Fprintf(os.Stderr, "INFO: Handler for protocol %s registered successfully.\n", name)
}

// PortManager управляет всеми портами
type PortManager struct {
	mu      sync.RWMutex
	ports   map[uint32]*PortState
	config  *configs.Config
	logger  *logger.Logger
	natsURL string
}

func NewPortManager(cfg *configs.Config, l *logger.Logger) *PortManager {
	// ---- ДОБАВЛЯЕМ ЛОГИРОВАНИЕ ----
	fmt.Fprintf(os.Stderr, "DEBUG: NewPortManager called. Logger is nil? %v\n", l == nil)
	// ---- КОНЕЦ ЛОГИРОВАНИЯ ----

	if l == nil {
		panic("logger passed to NewPortManager is nil")
	}

	pm := &PortManager{
		ports:   make(map[uint32]*PortState),
		config:  cfg,
		logger:  l,
		natsURL: cfg.NatsURL,
	}

	// ---- ДОБАВЛЯЕМ ЛОГИРОВАНИЕ ----
	fmt.Fprintf(os.Stderr, "DEBUG: About to register ARNAVI handler.\n")

	// ---- КОНЕЦ ЛОГИРОВАНИЯ ----

	RegisterHandler("ARNAVI", arnavi.NewHandler)

	// ... (регистрация других обработчиков) ...

	pm.logger.Info("PortManager created and handlers registered.")
	return pm
}

// getHandler находит конструктор обработчика по имени протокола.
func (pm *PortManager) getHandler(name string) (protocol.Constructor, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	constructor, exists := handlers[name]
	if !exists {
		return nil, fmt.Errorf("no handler registered for protocol: %s", name)
	}
	return constructor, nil
}

// --- МЕТОДЫ УПРАВЛЕНИЯ ПОРТАМИ ---

// CreatePort создает конфигурацию порта в памяти и сохраняет ее в файл
func (pm *PortManager) CreatePort(req *proto.CreatePortRequest) (*proto.PortResponse, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.ports[req.PortNumber]; exists {
		return nil, fmt.Errorf("port %d already exists", req.PortNumber)
	}

	// Проверяем, зарегистрирован ли обработчик для такого протокола
	if _, err := pm.getHandler(req.Protocol); err != nil {
		return nil, fmt.Errorf("failed to create port: %w", err)
	}

	newPortCfg := &configs.PortConfig{
		PortNumber: req.PortNumber,
		Protocol:   req.Protocol,
		NatsTopic:  req.NatsTopic,
		IsActive:   false,
	}

	pm.config.Ports = append(pm.config.Ports, *newPortCfg)
	if err := pm.config.SaveConfig(); err != nil {
		pm.logger.WithError(err).Error("Failed to save config after creating port")
		pm.config.Ports = pm.config.Ports[:len(pm.config.Ports)-1]
		return nil, fmt.Errorf("failed to save configuration: %w", err)
	}

	pm.logger.WithField("port", req.PortNumber).Info("Port configuration created")
	return &proto.PortResponse{
		PortNumber: req.PortNumber,
		Protocol:   req.Protocol,
		NatsTopic:  req.NatsTopic,
		Status:     "CREATED",
		Message:    "Port configuration created successfully",
	}, nil
}

// OpenPort открывает порт и начинает слушать подключения
func (pm *PortManager) OpenPort(portNum uint32) (*proto.PortResponse, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	nc, err := nats.Connect(pm.natsURL)
	if err != nil {
		pm.logger.WithError(err).Error("Failed to connect to NATS")
		return nil, fmt.Errorf("NATS connection failed: %w", err)
	}

	var portCfg *configs.PortConfig
	for i, p := range pm.config.Ports {
		if p.PortNumber == portNum {
			portCfg = &pm.config.Ports[i]
			break
		}
	}
	if portCfg == nil {
		return nil, fmt.Errorf("port %d not found", portNum)
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", portNum))
	if err != nil {
		pm.logger.WithError(err).WithField("port", portNum).Error("Failed to listen on port")
		return nil, fmt.Errorf("failed to listen on port %d: %w", portNum, err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	pm.ports[portNum] = &PortState{
		Config:     portCfg,
		Listener:   listener,
		NatsConn:   nc,
		CancelFunc: cancel,
	}

	portCfg.IsActive = true
	if err := pm.config.SaveConfig(); err != nil {
		pm.logger.WithError(err).Error("Failed to save config after opening port")
		cancel()
		listener.Close()
		nc.Close()
		delete(pm.ports, portNum)
		portCfg.IsActive = false
		return nil, fmt.Errorf("failed to save configuration: %w", err)
	}

	go pm.acceptConnections(ctx, portNum, listener, nc, portCfg)

	pm.logger.WithField("port", portNum).Info("Port opened successfully")
	return &proto.PortResponse{
		PortNumber: portNum,
		Status:     "OPEN",
		Message:    "Port opened successfully",
	}, nil
}

// ClosePort закрывает порт и отключает всех клиентов
func (pm *PortManager) ClosePort(portNum uint32) (*proto.PortResponse, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	portState, exists := pm.ports[portNum]
	if !exists {
		return nil, fmt.Errorf("port %d is not open", portNum)
	}

	portState.CancelFunc()
	portState.Listener.Close()
	portState.NatsConn.Close()

	for i, p := range pm.config.Ports {
		if p.PortNumber == portNum {
			pm.config.Ports[i].IsActive = false
			break
		}
	}
	if err := pm.config.SaveConfig(); err != nil {
		pm.logger.WithError(err).Error("Failed to save config after closing port")
	}

	delete(pm.ports, portNum)

	pm.logger.WithField("port", portNum).Info("Port closed successfully")
	return &proto.PortResponse{
		PortNumber: portNum,
		Status:     "CLOSED",
		Message:    "Port closed successfully",
	}, nil
}

// DeletePort удаляет конфигурацию порта
func (pm *PortManager) DeletePort(portNum uint32) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.ports[portNum]; exists {
		if _, err := pm.ClosePort(portNum); err != nil {
			return err
		}
	}

	found := false
	newPorts := pm.config.Ports[:0]
	for _, p := range pm.config.Ports {
		if p.PortNumber != portNum {
			newPorts = append(newPorts, p)
		} else {
			found = true
		}
	}
	if !found {
		return fmt.Errorf("port %d not found in configuration", portNum)
	}
	pm.config.Ports = newPorts

	if err := pm.config.SaveConfig(); err != nil {
		pm.logger.WithError(err).Error("Failed to save config after deleting port")
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	pm.logger.WithField("port", portNum).Info("Port configuration deleted")
	return nil
}

// ListPorts возвращает статус всех портов
func (pm *PortManager) ListPorts() *proto.ListPortsResponse {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	response := &proto.ListPortsResponse{}
	for _, pCfg := range pm.config.Ports {
		status := "CLOSED"
		if _, isOpen := pm.ports[pCfg.PortNumber]; isOpen {
			status = "OPEN"
		}
		response.Ports = append(response.Ports, &proto.PortResponse{
			PortNumber: pCfg.PortNumber,
			Protocol:   pCfg.Protocol,
			NatsTopic:  pCfg.NatsTopic,
			Status:     status,
		})
	}
	return response
}

// --- ВСПОМОГАТЕЛЬНЫЕ МЕТОДЫ ---

// acceptConnections принимает подключения в отдельной горутине
func (pm *PortManager) acceptConnections(ctx context.Context, portNum uint32, listener net.Listener, nc *nats.Conn, cfg *configs.PortConfig) {
	pm.logger.WithField("port", portNum).Info("Started accepting connections")
	for {
		select {
		case <-ctx.Done(): // <-- ПРОВЕРЯЕМ КОНТЕКСТ
			pm.logger.WithField("port", portNum).Info("Stopped accepting connections due to context cancellation")
			return // Выходим из горутины, слушающей порт
		default:
			conn, err := listener.Accept()
			if err != nil {
				if opErr, ok := err.(*net.OpError); ok && opErr.Op == "accept" {
					return // Слушатель закрыт, выходим
				}
				pm.logger.WithError(err).WithField("port", portNum).Error("Error accepting connection")
				continue
			}

			constructor, err := pm.getHandler(cfg.Protocol)
			if err != nil {
				pm.logger.WithError(err).WithField("port", portNum).Error("Failed to get handler for protocol")
				conn.Close()
				continue
			}

			// ПЕРЕДАЕМ КОНТЕКСТ В ОБРАБОТЧИК
			handler := constructor(conn, nc, cfg.NatsTopic, pm.logger, ctx)
			pm.logger.WithFields(logrus.Fields{
				"port":     portNum,
				"protocol": cfg.Protocol,
				"client":   conn.RemoteAddr(),
			}).Info("New client connected, starting handler")
			go handler.Process()
		}
	}
}

// OnNatsDisconnect вызывается при потере соединения с NATS
// OnNatsDisconnect вызывается при потере соединения с NATS
func (pm *PortManager) OnNatsDisconnect() {
	pm.logger.Warn("NATS connection lost, closing all ports and active handlers")
	pm.mu.Lock()
	defer pm.mu.Unlock()

	for portNum, state := range pm.ports {
		// 1. Сигнализируем всем обработчикам на этом порту, что пора завершать работу.
		if state.CancelFunc != nil {
			state.CancelFunc()
		}

		// 2. Закрываем слушающий порт, чтобы новые подключения не принимались.
		state.Listener.Close()

		// 3. Закрываем наше NATS-соединение для этого порта.
		state.NatsConn.Close()

		// 4. Обновляем конфиг.
		for i, p := range pm.config.Ports {
			if p.PortNumber == portNum {
				pm.config.Ports[i].IsActive = false
				break
			}
		}
		delete(pm.ports, portNum)
	}
	// pm.config.SaveConfig()
}

// OnNatsReconnect вызывается при восстановлении соединения с NATS
func (pm *PortManager) OnNatsReconnect() {
	pm.logger.Info("NATS connection restored, reopening active ports")
	for _, pCfg := range pm.config.Ports {
		if pCfg.IsActive {
			if _, err := pm.OpenPort(pCfg.PortNumber); err != nil {
				pm.logger.WithError(err).WithField("port", pCfg.PortNumber).Error("Failed to reopen port on NATS reconnect")
			}
		}
	}
}
