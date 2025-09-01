package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/nats-io/nats.go"
	"github.com/rackov/NavControlSystem/pkg/logger"
	"github.com/rackov/NavControlSystem/pkg/monitoring"
	"github.com/rackov/NavControlSystem/proto"
	"github.com/rackov/NavControlSystem/services/receiver/configs"
	"github.com/rackov/NavControlSystem/services/receiver/internal/portmanager"
	"github.com/rackov/NavControlSystem/services/receiver/internal/rgprs"
	"google.golang.org/grpc"
)

func main() {
	// 1. Загрузка конфигурации
	var configPath string
	flag.StringVar(&configPath, "config", "/home/vladimir/go/project/NavControlSystem/configs/receiver.toml", "Path to the TOML configuration file (e.g., ./configs/receiver.toml)")
	flag.Parse()
	fmt.Println("Загрузка конфигурации")
	cfg, err := configs.LoadConfig(configPath) // <-- Путь к конфигу относительно main.go
	if err != nil {
		fmt.Printf("Failed to load config: %v \n", err)
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}
	fmt.Println("Инициализация логгера")
	// 2. Инициализация логгера
	log, err := logger.NewLogger(cfg.LogConfig)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	defer log.Close()
	log.Info("Logger initialized")
	fmt.Println(" Инициализация менеджера портов")

	// 3. Инициализация менеджера портов
	portManager := portmanager.NewPortManager(cfg, log)

	// 4. Подключение к NATS и установка обработчиков событий
	var nc *nats.Conn

	// Создаем опции подключения с обработчиками
	opts := []nats.Option{
		// Обработчик отключения. Сигнатура: func(nc *nats.Conn)
		nats.DisconnectHandler(func(nc *nats.Conn) {
			// Здесь мы не получаем ошибку напрямую, но можем проверить статус соединения
			if nc.LastError() != nil {
				// Теперь мы можем получить ошибку и залогировать ее
				log.WithError(nc.LastError()).Warn("NATS disconnected")
			} else {
				log.Warn("NATS disconnected without specific error")
			}
			portManager.OnNatsDisconnect()
		}),
		// Обработчик переподключения. Сигнатура: func(nc *nats.Conn)
		nats.ReconnectHandler(func(nc *nats.Conn) {
			log.Info("NATS reconnected")
			portManager.OnNatsReconnect()
		}),
		// Обработчик закрытия соединения. Сигнатура: func(nc *nats.Conn)
		nats.ClosedHandler(func(nc *nats.Conn) {
			if nc.LastError() != nil {
				log.WithError(nc.LastError()).Warn("NATS connection closed")
			} else {
				log.Warn("NATS connection closed")
			}
		}),
	}

	// Подключаемся с использованием опций
	nc, err = nats.Connect(cfg.NatsURL, opts...)
	if err != nil {
		log.WithError(err).Fatal("Failed to connect to NATS")
	}
	defer nc.Close()
	log.Info("Connected to NATS")

	// 5. Запуск gRPC сервера
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
	if err != nil {
		log.WithError(err).Fatal("Failed to listen for gRPC")
	}
	grpcServer := grpc.NewServer()
	proto.RegisterReceiverServiceServer(grpcServer, rgprs.NewService(portManager))
	go func() {
		log.WithField("port", cfg.GRPCPort).Info("Starting gRPC server")
		if err := grpcServer.Serve(lis); err != nil {
			log.WithError(err).Fatal("Failed to serve gRPC")
		}
	}()
	defer grpcServer.GracefulStop()

	// 6. Запуск сервера метрик Prometheus
	go func() {
		log.WithField("port", cfg.MonitoringPort).Info("Starting monitoring server")
		if err := monitoring.Start(cfg.MonitoringPort); err != nil {
			log.WithError(err).Fatal("Failed to start monitoring server")
		}
	}()

	// 7. Восстановление состояния
	log.Info("Restoring active ports from config...")
	for _, pCfg := range cfg.Ports {
		if pCfg.IsActive {
			if _, err := portManager.OpenPort(pCfg.PortNumber); err != nil {
				log.WithError(err).WithField("port", pCfg.PortNumber).Error("Failed to restore port on startup")
			}
		}
	}

	// 8. Ожидание сигналов для graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("Shutting down receiver service...")

	log.Info("Closing all managed ports...")
	for _, pCfg := range cfg.Ports {
		if pCfg.IsActive {
			if _, err := portManager.ClosePort(pCfg.PortNumber); err != nil {
				log.WithError(err).WithField("port", pCfg.PortNumber).Error("Error during shutdown port closing")
			}
		}
	}
	log.Info("Receiver service stopped")
}
