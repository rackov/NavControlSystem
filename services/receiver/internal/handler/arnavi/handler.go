package arnavi

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/rackov/NavControlSystem/pkg/logger"
	"github.com/rackov/NavControlSystem/pkg/models"
	"github.com/rackov/NavControlSystem/services/receiver/internal/protocol"
)

// Handler реализует интерфейс protocol.Handler для протокола ARNAVI
type Handler struct {
	conn     net.Conn
	nc       *nats.Conn
	topic    string
	deviceID string // Эмулируем ID устройства, для простоты возьмем IP-адрес
	logger   *logger.Logger
	stopChan chan struct{} // Канал для graceful shutdown

	ctx    context.Context    // <-- СОХРАНЯЕМ КОНТЕКСТ
	cancel context.CancelFunc // <-- И ЕГО ФУНКЦИЮ ОТМЕНЫ
}

// NewHandler создает новый экземпляр обработчика ARNAVI
func NewHandler(conn net.Conn, nc *nats.Conn, topic string, log *logger.Logger, ctx context.Context) protocol.Handler { // <-- ПРИНИМАЕМ КОНТЕКСТ
	deviceID := conn.RemoteAddr().String()
	// Создаем дочерний контекст, который можно отменить индивидуально для этого обработчика
	handlerCtx, handlerCancel := context.WithCancel(ctx)
	h := &Handler{
		conn:     conn,
		nc:       nc,
		topic:    topic,
		deviceID: deviceID,
		logger:   log,
		stopChan: make(chan struct{}),
		ctx:      handlerCtx,
		cancel:   handlerCancel,
	}
	h.logger.WithField("device_id", h.deviceID).Info("Created new ARNAVI handler")
	return h
}

// Process запускает основной цикл обработки данных от клиента
func (h *Handler) Process() {
	defer h.conn.Close()
	h.logger.WithField("device_id", h.deviceID).Info("Starting ARNAVI handler process")

	// Подписываемся на команды для этого устройства
	commandTopic := fmt.Sprintf("commands.%s", h.deviceID)
	sub, err := h.nc.Subscribe(commandTopic, h.handleCommand)
	if err != nil {
		h.logger.WithError(err).WithField("device_id", h.deviceID).Error("Failed to subscribe to command topic")
		return
	}
	defer sub.Unsubscribe()

	scanner := bufio.NewScanner(h.conn)
	for {
		select {
		// ПРОВЕРЯЕМ КОНТЕКСТ. ЕСЛИ ОН ОТМЕНЕН - ВЫХОДИМ ИЗ ЦИКЛА
		case <-h.ctx.Done():
			h.logger.WithField("device_id", h.deviceID).Info("Handler stopped by context cancellation")
			return
		case <-h.stopChan:
			h.logger.WithField("device_id", h.deviceID).Info("Handler stopped by stop signal")
			return
		default:
			if !scanner.Scan() {
				// Сканер вернул false, значит соединение закрыто или ошибка
				if err := scanner.Err(); err != nil {
					h.logger.WithError(err).WithField("device_id", h.deviceID).Error("Error reading from connection")
				} else {
					h.logger.WithField("device_id", h.deviceID).Info("Client closed connection")
				}
				return
			}

			// Эмулируем получение данных. Предположим, что каждая строка - это пакет.
			rawData := scanner.Text()
			h.logger.WithField("device_id", h.deviceID).WithField("raw_data", rawData).Info("Received data")

			// Эмулируем парсинг и создание NavRecords
			navRecord := models.NavRecord{
				Latitude:            5575, // Москва
				Longitude:           3761,
				Speed:               50,
				Course:              90,
				NavigationTimestamp: uint32(time.Now().Unix()),
			}
			navData := models.NavRecords{
				PacketType: 1, // Условный тип пакета
				PacketID:   123,
				RecNav:     []models.NavRecord{navRecord},
			}

			// Публикуем в NATS
			data, err := json.Marshal(navData)
			if err != nil {
				h.logger.WithError(err).WithField("device_id", h.deviceID).Error("Failed to marshal NavRecords")
				continue
			}

			// ПРОВЕРЯЕМ КОНТЕКСТ ПЕРЕД ПУБЛИКАЦИЕЙ
			select {
			case <-h.ctx.Done():
				h.logger.WithField("device_id", h.deviceID).Info("Aborting publish due to context cancellation")
				return
			default:
				if err := h.nc.Publish(h.topic, data); err != nil {
					h.logger.WithError(err).WithField("device_id", h.deviceID).Error("Failed to publish to NATS")
					continue
				}
				h.logger.WithField("device_id", h.deviceID).WithField("topic", h.topic).Info("Successfully published data to NATS")
			}
			h.logger.WithField("device_id", h.deviceID).WithField("topic", h.topic).Info("Successfully published data to NATS")
		}
	}
}

// handleCommand обрабатывает входящие команды из NATS
func (h *Handler) handleCommand(msg *nats.Msg) {
	h.logger.WithField("device_id", h.deviceID).WithField("command_topic", msg.Subject).Info("Received command from NATS")

	// Эмулируем отправку команды на устройство
	// Просто отправляем полученные данные обратно как текст
	commandText := string(msg.Data)
	_, err := h.conn.Write([]byte(fmt.Sprintf("COMMAND: %s\n", commandText)))
	if err != nil {
		h.logger.WithError(err).WithField("device_id", h.deviceID).Error("Failed to send command to device")
	} else {
		h.logger.WithField("device_id", h.deviceID).WithField("command", commandText).Info("Command sent to device")
	}
}

// SendCommand отправляет команду на устройство (реализация интерфейса)
// В нашем упрощенном примере мы публикуем команду в NATS, а handleCommand уже отправляет ее на устройство.
// Но для полноты реализации интерфейса, добавим логику.
func (h *Handler) SendCommand(cmd []byte) error {
	commandTopic := fmt.Sprintf("commands.%s", h.deviceID)
	h.logger.WithField("device_id", h.deviceID).WithField("command_topic", commandTopic).Info("Sending command via NATS")
	return h.nc.Publish(commandTopic, cmd)
}

// Close корректно завершает работу обработчика (реализация интерфейса)
func (h *Handler) Close() error {
	close(h.stopChan) // Сигнализируем циклу Process о необходимости остановиться
	h.logger.WithField("device_id", h.deviceID).Info("ARNAVI handler closing...")
	return h.conn.Close()
}
