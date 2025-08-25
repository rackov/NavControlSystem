// NavControlSystem/services/receiver/cmd/main.go
package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/rackov/NavControlSystem/pkg/logger" // Предполагаем, что у нас есть pkg/logger
	"github.com/rackov/NavControlSystem/pkg/tnats"
	"github.com/sirupsen/logrus"
)

func main() {
	// 1. Инициализация логгера
	appLogger := logger.GetLogger()

	// 2. Инициализация NATS клиента
	natsURL := "nats://nats:4222" // Из конфига
	natsClient, err := tnats.NewClient(natsURL, appLogger)
	if err != nil {
		appLogger.WithError(err).Fatal("Failed to initialize NATS client")
	}
	defer natsClient.Close() // Гарантированно закрываем соединение при выходе

	// 3. Создание Publisher (например, для отправки данных)
	publisher, err := natsClient.NewPublisher()
	if err != nil {
		appLogger.WithError(err).Fatal("Failed to create NATS publisher")
	}

	// 4. Создание Subscriber (например, для получения команд)
	subscriber, err := natsClient.NewSubscriber()
	if err != nil {
		appLogger.WithError(err).Fatal("Failed to create NATS subscriber")
	}

	// --- Пример использования ---

	// Пример публикации данных
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop() // Этот вызов остановит тикер и закроет ticker.C

		// Используем идиоматичный цикл for range.
		// Цикл будет автоматически прерван, когда ticker.C будет закрыт (вызовом ticker.Stop()).
		for range ticker.C {
			data := []byte("sample navigation data")
			err := publisher.Publish("nav.data.raw", data)
			if err != nil {
				appLogger.WithError(err).Error("Example publish failed")
			}
		}
	}()

	// Пример подписки на команды
	_, err = subscriber.Subscribe("nav.commands.device123", func(msg *nats.Msg) {
		appLogger.WithFields(logrus.Fields{
			"subject": msg.Subject,
			"data":    string(msg.Data),
		}).Info("Received command")
		// Здесь логика обработки команды и отправки ее на устройство
	})
	if err != nil {
		appLogger.WithError(err).Fatal("Failed to subscribe to commands")
	}

	appLogger.Info("Receiver service started successfully")

	// 5. Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	appLogger.Info("Shutting down receiver service...")

	// Здесь можно добавить логику для остановки горутин и завершения активных задач
	// natsClient.Close() вызовется благодаря defer
}
