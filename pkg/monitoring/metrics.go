package monitoring

import (
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// --- Общие метрики, которые будут собираться для всех сервисов ---

// Метрики для стандартного коллектора Go (горутины, GC, память и т.д.)
// Они регистрируются автоматически.
var (
	// Пример кастомной метрики, которую могут использовать сервисы
	// Это просто шаблон, сервисы могут создавать свои метрики аналогичным образом.
	OperationCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "receiver_operations_total",
		Help: "Total number of operations processed by the receiver.",
	}, []string{"operation_type", "status"}) // operation_type: create_port, status: success, failure

	OperationDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "receiver_operation_duration_seconds",
		Help:    "Duration of receiver operations.",
		Buckets: prometheus.DefBuckets,
	}, []string{"operation_type"})
)

// Start запускает HTTP-сервер для экспорта метрик Prometheus.
// Он блокирует выполнение, поэтому его следует запускать в отдельной горутине.
func Start(port uint32) error {
	if port == 0 {
		return fmt.Errorf("monitoring port cannot be 0")
	}

	// Регистрируем стандартные коллекторы Go и процесса.
	// promauto.Register() делает это автоматически для метрик, созданных через promauto.
	// Если бы мы создавали метрики вручную, нам пришлось бы регистрировать их в реестре.
	// prometheus.MustRegister(prometheus.NewGoCollector())
	// prometheus.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))

	// Создаем обработчик для пути /metrics
	handler := promhttp.Handler()

	// Регистрируем его в стандартном mux (роутере) Go
	http.Handle("/metrics", handler)

	// Создаем и запускаем HTTP-сервер
	addr := fmt.Sprintf(":%d", port)
	server := &http.Server{
		Addr:         addr,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	fmt.Printf("Starting monitoring server on %s\n", addr)

	// Запускаем сервер. Эта функция блокируется, пока сервер не будет остановлен.
	// Если сервер не может стартовать (например, порт занят), он вернет ошибку.
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start monitoring server: %w", err)
	}

	return nil
}

// ObserveOperation - вспомогательная функция для удобного обновления метрик.
// Сервисы могут использовать ее, чтобы не дублировать код.
func ObserveOperation(operation string, status string, start time.Time) {
	OperationDuration.WithLabelValues(operation).Observe(time.Since(start).Seconds())
	OperationCount.WithLabelValues(operation, status).Inc()
}
