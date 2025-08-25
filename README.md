# NavControlSystem
Управление навигацией
NavControlSystem/
├── .github/                  # CI/CD пайплайны (например, GitHub Actions)
│   └── workflows/
│       └── main.yml
├── proto/                    # Единый .proto файл для всех gRPC сервисов
│   └── service.proto         # Определения всех gRPC сервисов и сообщений
├── pkg/                      # Общие Go-пакеты (библиотеки)
│   ├── config/               # Управление конфигурацией (через TOML)
│   ├── logger/               # Централизованная реализация логгера
│   ├── nats/                 # Обертка для NATS клиента и издателей/подписчиков
│   ├── postgres/             # Помощники для подключения к БД и выполнения запросов
│   └── telemetry/            # Настройка метрик Prometheus
├── services/                 # Директория для всех микросервисов
│   ├── rest-api/             # Сервис 1: Веб-приложение и REST API шлюз
│   ├── receiver/             # Сервис 3: Сервис приёма данных
│   │   ├── cmd/
│   │   │   └── main.go
│   │   ├── internal/
│   │   │   ├── handler/
│   │   │   │   ├── arnavi/
│   │   │   │   ├── egts/
│   │   │   │   └── ndtp/
│   │   │   ├── server/
│   │   │   └── service/
│   │   └── configs/
│   │       └── config.toml   # Конфигурация в формате TOML
│   ├── writer/               # Сервис 4: Сервис записи данных
│   │   ├── cmd/
│   │   │   └── main.go
│   │   ├── internal/
│   │   │   ├── nats/
│   │   │   └── postgres/
│   │   └── configs/
│   │       └── config.toml
│   └── retranslator/         # Сервис 5: Сервис ретрансляции данных
│       ├── cmd/
│       │   └── main.go
│       ├── internal/
│       │   ├── postgres/
│       │   └── sender/
│       └── configs/
│           └── config.toml
├── configs/                  # Глобальные конфигурации (например, docker-compose)
│   └── docker-compose.yml
├── .gitignore
├── go.mod
└── README.md
