package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/rackov/NavControlSystem/pkg/logger"
	"github.com/rackov/NavControlSystem/proto"
	"github.com/rackov/NavControlSystem/services/receiver/internal/handler/arnavi"
	"github.com/rackov/NavControlSystem/services/receiver/internal/protocol"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"google.golang.org/protobuf/types/known/wrapperspb"
)

// ReceiverServer реализует gRPC-сервер и управляет жизненным циклом
// сервиса RECEIVER.
type ReceiverServer struct {
	proto.UnimplementedReceiverControlServer

	cfg                  *Config
	nc                   *nats.Conn
	js                   nats.JetStreamContext
	handlers             map[string]protocol.ProtocolHandler // Карта обработчиков по имени
	handlersMu           sync.RWMutex
	grpcServer           *grpc.Server
	natsSubject          string    // Топик NATS для публикации данных
	natsStatusChangeChan chan bool // Канал для получения уведомлений о статусе NATS (true=connected, false=disconnected)
}

// NewReceiverServer создает новый экземпляр сервера.
func NewReceiverServer(cfg *Config) *ReceiverServer {
	return &ReceiverServer{
		cfg:         cfg,
		handlers:    make(map[string]protocol.ProtocolHandler),
		natsSubject: "nav.data.raw", // Стандартный топик для данных
		// Инициализируем канал при создании сервера
		natsStatusChangeChan: make(chan bool, 1), // Буферизированный канал на 1 сообщение
	}
}

// Start запускает все компоненты сервиса: NATS, обработчики протоколов и gRPC-сервер.
func (s *ReceiverServer) Start(ctx context.Context) error {
	logger.Info("Starting RECEIVER server...")

	// 1. Подключение к NATS
	if err := s.connectNats(); err != nil {
		logger.Errorf("failed to connect to NATS: %v", err)
		return fmt.Errorf("failed to connect to NATS: %w", err)
	}
	logger.Infof("Successfully connected to NATS at %s", s.cfg.NatsURL)
	ServiceMetrics.SetGauge("nats_connected", 1)

	// 2. Инициализация и запуск обработчиков протоколов
	if err := s.startProtocolHandlers(ctx); err != nil {
		logger.Errorf("Failed to start protocol handlers: %v", err)
		return fmt.Errorf("failed to start protocol handlers: %w", err)
	}

	// 3. Запуск gRPC сервера
	if err := s.startGrpcServer(); err != nil {
		logger.Errorf("Failed to start gRPC server: %v", err)
		return fmt.Errorf("failed to start gRPC server: %w", err)
	}
	logger.Infof("gRPC server started on port %d", s.cfg.GrpcPort)

	// 4. Запуск горутины для мониторинга состояния NATS
	go s.monitorNatsConnection(ctx)

	logger.Info("RECEIVER server started successfully")
	return nil
}

// Stop останавливает все компоненты сервиса gracefully.
func (s *ReceiverServer) Stop() {
	logger.Info("Shutting down RECEIVER server...")

	// Останавливаем gRPC сервер
	if s.grpcServer != nil {
		logger.Debug("Stopping gRPC server...")
		s.grpcServer.GracefulStop()
	}

	// Останавливаем обработчики протоколов
	logger.Debug("Stopping protocol handlers...")
	s.stopProtocolHandlers()

	// Закрываем соединение с NATS
	if s.nc != nil {
		logger.Debug("Closing NATS connection...")
		s.nc.Close()
		ServiceMetrics.SetGauge("nats_connected", 0)
	}
	logger.Info("RECEIVER server stopped.")
}

// --- NATS Management ---

// connectNats устанавливает соединение с NATS и назначает обработчики событий.
func (s *ReceiverServer) connectNats() error {
	var err error
	s.nc, err = nats.Connect(s.cfg.NatsURL,
		nats.ReconnectWait(2*time.Second),
		nats.MaxReconnects(5),
		// НАЗНАЧАЕМ НАШИ ФУНКЦИИ-ОБРАБОТЧИКИ
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			logger.Warnf("NATS connection disconnected: %v", err)
			// При дисконнекте мы не знаем, будет ли переподключение, поэтому не отправляем сигнал сюда.
			// Сигнал отправится только в ClosedHandler, если переподключение не удалось.
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			logger.Info("NATS connection reestablished")
			// Отправляем сигнал в канал, что NATS снова подключен
			select {
			case s.natsStatusChangeChan <- true:
				logger.Debugf("Sent 'connected' signal to NATS status channel")
			default:
				// Если канал уже занят (буфер полон), ничего не делаем,
				// чтобы не блокировать callback NATS.
				logger.Debugf("NATS status channel is full, 'connected' signal not sent")
			}
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			logger.Errorf("NATS connection closed permanently: %v", nc.LastError())
			// Отправляем сигнал, что NATS окончательно отключен
			select {
			case s.natsStatusChangeChan <- false:
				logger.Debugf("Sent 'disconnected' signal to NATS status channel")
			default:
				logger.Debugf("NATS status channel is full, 'disconnected' signal not sent")
			}
		}),
	)
	if err != nil {
		return err
	}

	s.js, err = s.nc.JetStream()
	if err != nil {
		logger.Warnf("JetStream not available, falling back to core NATS. Error: %v", err)
		s.js = nil
	}
	return nil
}

// monitorNatsConnection слушает канал с уведомлениями и управляет обработчиками.
func (s *ReceiverServer) monitorNatsConnection(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			logger.Info("NATS monitoring stopped by context cancellation.")
			return

		// СЛУШАЕМ НАШ КАНАЛ, А НЕ CALLBACK
		case isConnected := <-s.natsStatusChangeChan:
			logger.Infof("Received NATS status change event. Connected: %v", isConnected)
			if isConnected {
				logger.Info("NATS reconnected, restarting protocol handlers.")
				ServiceMetrics.SetGauge("nats_connected", 1)
				if err := s.startProtocolHandlers(ctx); err != nil {
					logger.Errorf("Failed to restart protocol handlers after NATS reconnect: %v", err)
				}
			} else {
				logger.Warn("NATS connection lost, stopping protocol handlers.")
				ServiceMetrics.SetGauge("nats_connected", 0)
				s.stopProtocolHandlers()
			}
		}
	}
}

// Реализация интерфейса DataPublisher для передачи в обработчики.
func (s *ReceiverServer) Publish(data *protocol.NavRecord) error {
	if s.nc == nil || !s.nc.IsConnected() {
		return fmt.Errorf("NATS is not connected")
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		ServiceMetrics.IncErrorCounter("nats_marshal_failed")
		return fmt.Errorf("failed to marshal navigation data: %w", err)
	}

	if s.js != nil {
		_, err = s.js.Publish(s.natsSubject, jsonData)
	} else {
		err = s.nc.Publish(s.natsSubject, jsonData)
	}

	if err != nil {
		ServiceMetrics.IncErrorCounter("nats_publish_failed")
		return fmt.Errorf("failed to publish to NATS: %w", err)
	}

	ServiceMetrics.IncOperationCounter("nats_published")
	return nil
}

func (s *ReceiverServer) IsConnected() bool {
	return s.nc != nil && s.nc.IsConnected()
}

// --- Protocol Handlers Management ---

// startProtocolHandlers инициализирует и запускает обработчики для каждого
// протокола, указанного в конфигурации.
func (s *ReceiverServer) startProtocolHandlers(ctx context.Context) error {
	s.handlersMu.Lock()
	defer s.handlersMu.Unlock()

	// Останавливаем старые обработчики, если они были
	s.stopProtocolHandlersInternal()

	for _, protoCfg := range s.cfg.ProtocolConfigs {
		var handler protocol.ProtocolHandler
		switch protoCfg.Name {
		case "ARNAVI":
			handler = arnavi.NewArnaviHandler()
		// case "EGTS":
		// 	handler = arnavi.NewArnaviHandler()
		// case "NDTP":
		// 	handler = ndtp.NewNdtpHandler()
		default:
			logger.Warnf("Unsupported protocol: %s, skipping", protoCfg.Name)
			continue
		}

		if err := handler.Start(ctx, s, protoCfg.Port); err != nil {
			logger.Errorf("Failed to start %s handler on port %d: %v", protoCfg.Name, protoCfg.Port, err)
			// Если не удалось запустить один из обработчиков, откатываем все
			s.stopProtocolHandlersInternal()
			return fmt.Errorf("failed to start %s handler", protoCfg.Name)
		}

		s.handlers[protoCfg.Name] = handler
		logger.Infof("%s handler started successfully on port %d", protoCfg.Name, protoCfg.Port)
	}
	return nil
}

// stopProtocolHandlers останавливает все активные обработчики.
func (s *ReceiverServer) stopProtocolHandlers() {
	s.handlersMu.Lock()
	defer s.handlersMu.Unlock()
	s.stopProtocolHandlersInternal()
}

// stopProtocolHandlersInternal - внутренняя версия без мьютекса.
func (s *ReceiverServer) stopProtocolHandlersInternal() {
	if len(s.handlers) == 0 {
		return
	}

	logger.Info("Stopping all protocol handlers...")
	var wg sync.WaitGroup
	for name, handler := range s.handlers {
		wg.Add(1)
		go func(name string, h protocol.ProtocolHandler) {
			defer wg.Done()
			if err := h.Stop(); err != nil {
				logger.Errorf("Failed to stop %s handler: %v", name, err)
			}
		}(name, handler)
	}
	wg.Wait()
	s.handlers = make(map[string]protocol.ProtocolHandler)
	logger.Info("All protocol handlers stopped.")
}

// --- gRPC Server Management ---

// startGrpcServer инициализирует и запускает gRPC-сервер в отдельной горутине.
func (s *ReceiverServer) startGrpcServer() error {
	s.grpcServer = grpc.NewServer()
	proto.RegisterReceiverControlServer(s.grpcServer, s)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.cfg.GrpcPort))
	if err != nil {
		return err
	}

	go func() {
		if err := s.grpcServer.Serve(lis); err != nil && err != grpc.ErrServerStopped {
			logger.Errorf("gRPC server error: %v", err)
		}
	}()

	return nil
}

// --- gRPC Service Implementation ---

// SetLogLevel изменяет глобальный уровень логирования.
func (s *ReceiverServer) SetLogLevel(ctx context.Context, req *proto.SetLogLevelRequest) (*proto.SetLogLevelResponse, error) {
	level, err := logger.ParseLevel(req.Level) // Предполагаем, что в logger есть ParseLevel
	if err != nil {
		logger.Errorf("Failed to parse log level '%s': %v", req.Level, err)
		return nil, status.Errorf(codes.InvalidArgument, "invalid log level: %s", req.Level)
	}

	logger.SetGlobalLevel(level)
	logger.Infof("Log level set to %s", req.Level)
	return &proto.SetLogLevelResponse{Success: true}, nil
}

// GetStatus возвращает текущий статус сервиса: состояние NATS и открытые порты.
func (s *ReceiverServer) GetStatus(ctx context.Context, req *proto.GetStatusRequest) (*proto.GetStatusResponse, error) {
	s.handlersMu.RLock()
	defer s.handlersMu.RUnlock()

	response := &proto.GetStatusResponse{
		NatsConnected: s.IsConnected(),
	}

	for name, handler := range s.handlers {
		var port int32
		// Ищем порт в конфигурации, так как сам обработчик его не хранит
		for _, cfg := range s.cfg.ProtocolConfigs {
			if cfg.Name == name {
				port = int32(cfg.Port)
				break
			}
		}

		response.Ports = append(response.Ports, &proto.PortStatus{
			Name:   name,
			Port:   port,
			IsOpen: handler.IsRunning(),
		})
	}

	logger.Debug("GRPC call: GetStatus")
	return response, nil
}

// GetActiveConnectionsCount возвращает количество активных подключений для указанного протокола.
func (s *ReceiverServer) GetActiveConnectionsCount(ctx context.Context, req *proto.GetClientsRequest) (*wrapperspb.Int32Value, error) {
	s.handlersMu.RLock()
	defer s.handlersMu.RUnlock()

	handler, exists := s.handlers[req.ProtocolName]
	if !exists {
		logger.Warnf("GRPC call GetActiveConnectionsCount for unknown protocol: %s", req.ProtocolName)
		return nil, status.Errorf(codes.NotFound, "protocol handler '%s' not found", req.ProtocolName)
	}

	count := handler.GetActiveConnectionsCount()
	logger.Infof("GRPC call: GetActiveConnectionsCount for %s = %d", req.ProtocolName, count)
	return wrapperspb.Int32(int32(count)), nil
}

// GetConnectedClients возвращает список информации о подключенных клиентах для протокола.
func (s *ReceiverServer) GetConnectedClients(ctx context.Context, req *proto.GetClientsRequest) (*proto.GetClientsResponse, error) {
	s.handlersMu.RLock()
	defer s.handlersMu.RUnlock()

	handler, exists := s.handlers[req.ProtocolName]
	if !exists {
		logger.Warnf("GRPC call GetConnectedClients for unknown protocol: %s", req.ProtocolName)
		return nil, status.Errorf(codes.NotFound, "protocol handler '%s' not found", req.ProtocolName)
	}

	connectedClients := handler.GetConnectedClients()

	grpcClients := make([]*proto.ClientInfo, 0, len(connectedClients))
	for _, client := range connectedClients {
		grpcClients = append(grpcClients, &proto.ClientInfo{
			Id:             client.ID,
			Address:        client.Addr,
			ConnectedSince: client.Since.Unix(),
		})
	}

	logger.Infof("GRPC call: GetConnectedClients for %s, found %d clients", req.ProtocolName, len(grpcClients))
	return &proto.GetClientsResponse{Clients: grpcClients}, nil
}

// DisconnectClient принудительно отключает клиента по его адресу.
func (s *ReceiverServer) DisconnectClient(ctx context.Context, req *proto.DisconnectClientRequest) (*proto.DisconnectClientResponse, error) {
	s.handlersMu.RLock()
	defer s.handlersMu.RUnlock()

	handler, exists := s.handlers[req.ProtocolName]
	if !exists {
		logger.Warnf("GRPC call DisconnectClient for unknown protocol: %s", req.ProtocolName)
		return nil, status.Errorf(codes.NotFound, "protocol handler '%s' not found", req.ProtocolName)
	}

	logger.Infof("GRPC call: DisconnectClient for %s, address %s", req.ProtocolName, req.ClientAddress)
	err := handler.DisconnectClient(req.ClientAddress)
	if err != nil {
		logger.Errorf("Failed to disconnect client %s for protocol %s: %v", req.ClientAddress, req.ProtocolName, err)
		return &proto.DisconnectClientResponse{Success: false}, status.Errorf(codes.Internal, "could not disconnect client: %v", err)
	}

	logger.Infof("Successfully disconnected client %s for protocol %s", req.ClientAddress, req.ProtocolName)
	return &proto.DisconnectClientResponse{Success: true}, nil
}
