package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rackov/NavControlSystem/pkg/monitoring" // Импортируем наш пакет
)

// Предположим, у нас есть функция для обработки одного подключения
func handleConnection(conn net.Conn, metrics *monitoring.ServiceMetrics) {
	defer conn.Close()

	// --- Начало замера времени операции ---
	startTime := time.Now()

	// ... ваша логика обработки данных из conn ...
	// Например, читаем данные, парсим, отправляем в NATS.
	fmt.Printf("Handling connection from %s\n", conn.RemoteAddr())

	// Имитируем успешную обработку
	operationName := "packet_received"
	metrics.IncOperationCounter(operationName) // Увеличиваем счетчик обработанных пакетов

	// Имитируем ошибку для примера
	// if some_error_occurs {
	// 	metrics.IncErrorCounter("nats_publish_failed")
	// }

	// --- Конец замера времени операции ---
	duration := time.Since(startTime)
	metrics.ObserveOperationDuration(operationName, duration) // Записываем длительность

	// Обновляем gauge (например, количество активных соединений)
	// Это можно делать в отдельной горутине, которая периодически обновляет значение
	metrics.SetGauge("active_connections", float64(10))
}

func main() {
	// 1. Инициализируем метрики для этого сервиса
	// "receiver" - это имя сервиса, которое будет добавлено как лейбл ко всем метрикам
	serviceMetrics := monitoring.NewServiceMetrics("receiver")

	// 2. Запускаем HTTP-сервер для экспорта метрик в отдельной горутине
	// Предположим, порт для мониторинга указан в конфигурации
	metricsPort := 9091
	go func() {
		if err := monitoring.StartMetricsServer(metricsPort); err != nil {
			// В реальном приложении лучше использовать логгер
			fmt.Printf("Failed to start metrics server: %v\n", err)
			os.Exit(1)
		}
	}()

	// 3. Основная логика вашего сервиса
	// ... код для открытия TCP-порта и приема данных ...
	listener, err := net.Listen("tcp", ":9999")
	if err != nil {
		panic(err)
	}
	defer listener.Close()
	fmt.Println("RECEIVER service is listening on :9999")

	// Горутина для graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				fmt.Printf("Accept error: %v\n", err)
				continue
			}
			// Передаем метрики в обработчик, чтобы он мог их обновлять
			go handleConnection(conn, serviceMetrics)
		}
	}()

	<-done // Ждем сигнала для остановки
	fmt.Println("RECEIVER service is shutting down...")
}
