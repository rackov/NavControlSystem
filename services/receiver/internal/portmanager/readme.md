Интеграция метрик в бизнес-логику (Пример для portmanager)

package portmanager

import (
	// ... импорты ...
	"time"
	"github.com/rackov/NavControlSystem/pkg/monitoring" // <-- Импортируем пакет мониторинга
)

// ... (код структуры PortManager без изменений) ...

func (pm *PortManager) CreatePort(req *proto.CreatePortRequest) (*proto.PortResponse, error) {
	start := time.Now() // Засекаем время начала операции
	
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.ports[req.PortNumber]; exists {
		// В случае ошибки, тоже фиксируем метрику
		monitoring.ObserveOperation("create_port", "failure", start)
		return nil, fmt.Errorf("port %d already exists", req.PortNumber)
	}
	
	// ... (логика создания порта) ...

	if err := pm.config.SaveConfig(); err != nil {
		monitoring.ObserveOperation("create_port", "failure", start)
		// ... откат и возврат ошибки ...
	}

	pm.logger.WithField("port", req.PortNumber).Info("Port configuration created")
	
	// Фиксируем успешное завершение операции
	monitoring.ObserveOperation("create_port", "success", start)
	
	return &proto.PortResponse{ /* ... */ }, nil
}

func (pm *PortManager) OpenPort(portNum uint32) (*proto.PortResponse, error) {
	start := time.Now()
	
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// ... (логика открытия порта) ...

	if err := pm.config.SaveConfig(); err != nil {
		monitoring.ObserveOperation("open_port", "failure", start)
		// ... откат и возврат ошибки ...
	}

	pm.logger.WithField("port", portNum).Info("Port opened successfully")
	monitoring.ObserveOperation("open_port", "success", start)
	
	return &proto.PortResponse{ /* ... */ }, nil
}
