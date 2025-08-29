package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/rackov/NavControlSystem/pkg/logger"
	"github.com/rackov/NavControlSystem/pkg/monitoring"
)

func main() {
	// 1. Определяем флаг для пути к конфигурационному файлу
	var configPath string
	flag.StringVar(&configPath, "config", "", "Path to the TOML configuration file (e.g., ./configs/receiver.toml)")
	flag.Parse()

	// 2. Загружаем конфигурацию
	cfg, err := LoadConfig(&configPath) // Передаем указатель на флаг
	if err != nil {
		// На этом этапе логгер еще не настроен, используем стандартный вывод
		panic(fmt.Sprintf("Failed to load configuration: %v", err))
	}

	// 3. Инициализация логгера с параметрами из конфигурации
	// Преобразуем строку уровня в logrus.Level
	logLevel, err := logger.ParseLevel(cfg.LogLevel)
	if err != nil {
		panic(fmt.Sprintf("Invalid log level in config: %v", err))
	}

	// Инициализируем глобальный логгер
	// Если путь к файлу в конфиге пуст, логи будут только в консоли.
	logger.Init(logLevel, cfg.Logging.FilePath)

	logger.Info("Application starting...")
	logger.Info("Config loaded successfully")
	logger.Debugf("Config details: %+v", cfg)

	// 4. Инициализация метрик Prometheus
	InitMetrics("receiver")
	go func() {
		if err := monitoring.StartMetricsServer(cfg.MetricsPort); err != nil {
			logger.Errorf("Failed to start metrics server: %v", err)
		}
	}()

	// 5. Создание и запуск основного сервера
	receiverServer := NewReceiverServer(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Запускаем воркер для изменений конфигурации
	receiverServer.startConfigWorker(ctx)

	if err := receiverServer.Start(ctx); err != nil {
		logger.Errorf("Failed to start receiver server: %v", err)
		panic(err)
	}

	// 6. Настройка graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan // Ждем сигнала
	logger.Info("Shutdown signal received, stopping server...")

	receiverServer.Stop()
	logger.Info("Application stopped.")
}
