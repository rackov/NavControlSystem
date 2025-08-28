package main // Важный момент: этот файл относится к пакету main

import "github.com/rackov/NavControlSystem/pkg/monitoring"

// ServiceMetrics - это ГЛОБАЛЬНАЯ переменная.
// Она будет доступна из любого файла внутри пакета main (cmd/receiver).
var ServiceMetrics *monitoring.ServiceMetrics

// InitMetrics инициализирует и регистрирует все метрики для сервиса RECEIVER.
func InitMetrics(serviceName string) {
	// Мы используем конструктор NewServiceMetrics из нашего пакета pkg/monitoring
	// и присваиваем результат нашей глобальной переменной.
	ServiceMetrics = monitoring.NewServiceMetrics(serviceName)

	// Здесь можно добавить и другие, специфичные для RECEIVER, метрики,
	// если они не покрываются стандартным набором из pkg/monitoring.
	// Например, мы можем заранее создать gauge для NATS.
	ServiceMetrics.SetGauge("nats_connected", 0) // Инициализируем значением по умолчанию (0 - отключен)
}
