package monitoring

import (
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// ServiceMetrics - это структура, которая хранит все метрики для одного сервиса.
// Это удобный способ передавать метрики в разные части приложения.
type ServiceMetrics struct {
	// OperationCounters должен хранить тип-вектор.
	OperationCounters map[string]*prometheus.CounterVec

	// OperationDurations должен хранить тип-вектор.
	OperationDurations map[string]*prometheus.HistogramVec

	// ErrorCounters должен хранить тип-вектор.
	ErrorCounters map[string]*prometheus.CounterVec

	// ServiceGauges корректен как простой Gauge.
	ServiceGauges map[string]prometheus.Gauge
}

// NewServiceMetrics - это конструктор для ServiceMetrics.
// Он создает и регистрирует все необходимые метрики.
func NewServiceMetrics(serviceName string) *ServiceMetrics {
	commonLabels := prometheus.Labels{"service": serviceName}

	// Эта инициализация теперь соответствует типам полей структуры.
	m := &ServiceMetrics{
		OperationCounters:  make(map[string]*prometheus.CounterVec),
		OperationDurations: make(map[string]*prometheus.HistogramVec),
		ErrorCounters:      make(map[string]*prometheus.CounterVec),
		ServiceGauges:      make(map[string]prometheus.Gauge),
	}

	// --- Предопределенные метрики ---

	// Сохраняем сам CounterVec, а не конкретный Counter.
	m.OperationCounters["operations_total"] = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "app_operations_total",
			Help:        "The total number of processed operations.",
			ConstLabels: commonLabels,
		},
		[]string{"operation_name"},
	) // <-- УДАЛИТЬ .WithLabelValues("default")

	// Сохраняем сам HistogramVec.
	m.OperationDurations["operation_duration_seconds"] = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:        "app_operation_duration_seconds",
			Help:        "A histogram of the operation duration in seconds.",
			ConstLabels: commonLabels,
			Buckets:     prometheus.DefBuckets,
		},
		[]string{"operation_name"},
	)

	// Сохраняем сам CounterVec.
	m.ErrorCounters["errors_total"] = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "app_errors_total",
			Help:        "The total number of errors.",
			ConstLabels: commonLabels,
		},
		[]string{"error_type"},
	)

	// Инициализация Gauge остается корректной.
	m.ServiceGauges["active_connections"] = promauto.NewGauge(prometheus.GaugeOpts{
		Name:        "app_active_connections",
		Help:        "The current number of active connections.",
		ConstLabels: commonLabels,
	})

	return m
}

// StartMetricsServer - это хэлпер, который запускает HTTP-сервер для экспорта метрик.
// Он блокирует выполнение, поэтому его следует запускать в отдельной горутине.
func StartMetricsServer(port int) error {
	// Регистрируем стандартный обработчик Prometheus, который будет отдавать метрики
	// по пути /metrics.
	http.Handle("/metrics", promhttp.Handler())

	addr := fmt.Sprintf(":%d", port)
	server := &http.Server{
		Addr:         addr,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	fmt.Printf("Starting Prometheus metrics server on %s\n", addr)
	// ListenAndServe блокирует выполнение, поэтому его нужно вызывать в горутине.
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start metrics server: %w", err)
	}

	return nil
}

// --- Вспомогательные методы для удобства ---

// IncOperationCounter увеличивает счетчик операций для конкретного имени операции.
func (m *ServiceMetrics) IncOperationCounter(operationName string) {
	// 'counterVec' теперь имеет корректный тип *prometheus.CounterVec
	if counterVec, ok := m.OperationCounters["operations_total"]; ok {
		counterVec.WithLabelValues(operationName).Inc()
	}
}

// ObserveOperationDuration записывает длительность операции в гистограмму.
func (m *ServiceMetrics) ObserveOperationDuration(operationName string, duration time.Duration) {
	// 'histogramVec' теперь имеет корректный тип *prometheus.HistogramVec
	if histogramVec, ok := m.OperationDurations["operation_duration_seconds"]; ok {
		histogramVec.WithLabelValues(operationName).Observe(duration.Seconds())
	}
}

// IncErrorCounter увеличивает счетчик ошибок для конкретного типа ошибки.
func (m *ServiceMetrics) IncErrorCounter(errorType string) {
	// 'counterVec' теперь имеет корректный тип *prometheus.CounterVec
	if counterVec, ok := m.ErrorCounters["errors_total"]; ok {
		counterVec.WithLabelValues(errorType).Inc()
	}
}

// SetGauge устанавливает значение для gauge-метрики.
func (m *ServiceMetrics) SetGauge(gaugeName string, value float64) {
	if gauge, ok := m.ServiceGauges[gaugeName]; ok {
		gauge.Set(value)
	}
}
