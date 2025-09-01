package configs

import (
	"fmt"
	"os"
	"sync"

	"github.com/BurntSushi/toml"
	"github.com/rackov/NavControlSystem/pkg/logger"
)

// PortConfig описывает конфигурацию одного порта
type PortConfig struct {
	PortNumber uint32 `toml:"port_number"`
	Protocol   string `toml:"protocol"` // "ARNAVI", "EGTS", "NDTP"
	NatsTopic  string `toml:"nats_topic"`
	IsActive   bool   `toml:"is_active"` // Состояние: открыт или закрыт
}

// Config описывает полную конфигурацию сервиса receiver
type Config struct {
	GRPCPort       uint32        `toml:"grpc_port"`
	MonitoringPort uint32        `toml:"monitoring_port"`
	NatsURL        string        `toml:"nats_url"`
	LogConfig      logger.Config `toml:"log"`
	Ports          []PortConfig  `toml:"ports"` // Список всех сконфигурированных портов
	configPath     string        // Путь к файлу для сохранения
	mu             sync.Mutex
}

// LoadConfig загружает конфигурацию из TOML-файла
func LoadConfig(path string) (*Config, error) {
	cfg := &Config{configPath: path}
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	cfg.mu.Lock()
	defer cfg.mu.Unlock()
	if _, err := toml.Decode(string(file), cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}
	return cfg, nil
}

// SaveConfig сохраняет текущую конфигурацию в TOML-файл
func (c *Config) SaveConfig() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 1. Определяем пути к основному и временному файлам
	configPath := c.configPath
	tempPath := configPath + ".tmp"

	// 2. Создаем временный файл
	file, err := os.Create(tempPath)
	if err != nil {
		return fmt.Errorf("failed to create temp config file: %w", err)
	}

	// 3. Используем defer для гарантированного закрытия файла
	// Если произойдет ошибка до переименования, временный файл будет удален.

	defer func() {
		file.Close()
		// Если временный файл все еще существует, значит, что-то пошло не так, удаляем его.
		if _, err := os.Stat(tempPath); err == nil {
			os.Remove(tempPath)
		}
	}()

	// 4. Кодируем конфигурацию и записываем во временный файл
	encoder := toml.NewEncoder(file)
	if err := encoder.Encode(c); err != nil {
		return fmt.Errorf("failed to encode config to temp file: %w", err)
	}

	// 5. Принудительно сбрасываем буферы на диск. Это критически важный шаг!
	if err := file.Sync(); err != nil {
		return fmt.Errorf("failed to sync temp config file to disk: %w", err)
	}

	// 6. Закрываем файл перед переименованием
	if err := file.Close(); err != nil {
		return fmt.Errorf("failed to close temp config file: %w", err)
	}

	// 7. Атомарно переименовываем временный файл в постоянный
	// Это момент, когда старый файл заменяется новым.
	if err := os.Rename(tempPath, configPath); err != nil {
		return fmt.Errorf("failed to rename temp config file to permanent: %w", err)
	}

	// Если все прошло успешно, файл сохранен.
	return nil
}
