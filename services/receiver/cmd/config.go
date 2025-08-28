package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// ProtocolConfig описывает настройки для одного протокола.
type ProtocolConfig struct {
	Name string `toml:"name"` // EGTS, Arnavi, NDTP
	Port int    `toml:"port"`
}

// Config описывает всю конфигурацию для сервиса RECEIVER.
// Теги `toml:"..."` указывают, как поля структуры будут отображаться на ключи в TOML файле.
type Config struct {
	// Основные настройки сервиса
	GrpcPort    int    `toml:"grpc_port"`
	MetricsPort int    `toml:"metrics_port"`
	NatsURL     string `toml:"nats_url"`
	LogLevel    string `toml:"log_level"` // debug, info, warn, error

	// Настройки протоколов, представленные как срез структур
	ProtocolConfigs []ProtocolConfig `toml:"protocols"`

	// Настройки логирования
	Logging struct {
		FilePath string `toml:"file_path"` // Путь к файлу для записи логов
	} `toml:"logging"`
}

// LoadConfig загружает конфигурацию из TOML файла.
// Путь к файлу можно передать аргументом или взять из переменной окружения.
func LoadConfig(configPath *string) (*Config, error) {
	// 1. Определяем путь к конфигурационному файлу
	var cfgFile string
	if configPath != nil && *configPath != "" {
		// Путь передан через флаг командной строки
		cfgFile = *configPath
	} else {
		// Пытаемся взять путь из переменной окружения
		cfgFile = os.Getenv("RECEIVER_CONFIG_PATH")
		if cfgFile == "" {
			// Если переменная окружения не задана, используем путь по умолчанию
			cfgFile = "./configs/receiver.toml"
		}
	}

	// 2. Проверяем существование файла
	if _, err := os.Stat(cfgFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("configuration file not found at %s", cfgFile)
	}

	// 3. Читаем и парсим TOML файл
	data, err := os.ReadFile(cfgFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", cfgFile, err)
	}

	var cfg Config
	if _, err := toml.Decode(string(data), &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", cfgFile, err)
	}

	// 4. Базовая валидация загруженных данных
	if len(cfg.ProtocolConfigs) == 0 {
		return nil, fmt.Errorf("no protocol configurations found in %s", cfgFile)
	}
	if cfg.NatsURL == "" {
		return nil, fmt.Errorf("nats_url is not specified in %s", cfgFile)
	}

	// 5. (Опционально) Создаем директорию для логов, если она указана
	if cfg.Logging.FilePath != "" {
		logDir := filepath.Dir(cfg.Logging.FilePath)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory %s: %w", logDir, err)
		}
	}

	fmt.Printf("Config loaded from: %s\n", cfgFile)
	return &cfg, nil
}
