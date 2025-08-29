package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/BurntSushi/toml"
	"github.com/google/uuid"
)

// ProtocolConfig теперь описывает один конкретный слушающий порт.
type ProtocolConfig struct {
	ID     string `toml:"id"`     // Уникальный идентификатор порта (UUID)
	Name   string `toml:"name"`   // Имя протокола (EGTS, ARNAVI, NDTP)
	Port   int    `toml:"port"`   // Номер порта
	Active bool   `toml:"active"` // Флаг, должен ли порт быть открыт
}

// Config описывает всю конфигурацию для сервиса RECEIVER.
type Config struct {
	mu sync.RWMutex // Добавляем мьютекс для безопасного доступа из разных горутин

	GrpcPort    int    `toml:"grpc_port"`
	MetricsPort int    `toml:"metrics_port"`
	NatsURL     string `toml:"nats_url"`
	LogLevel    string `toml:"log_level"`

	// Теперь это срез всех сконфигурированных портов.
	ProtocolConfigs []ProtocolConfig `toml:"protocols"`

	Logging struct {
		FilePath string `toml:"file_path"`
	} `toml:"logging"`

	// Добавляем путь к файлу, чтобы иметь возможность его перезаписать
	configPath string
}

// LoadConfig загружает и парсит TOML файл.
func LoadConfig(configPath *string) (*Config, error) {
	cfgFile := resolveConfigPath(configPath)

	data, err := os.ReadFile(cfgFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", cfgFile, err)
	}

	var cfg Config
	if _, err := toml.Decode(string(data), &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", cfgFile, err)
	}

	cfg.configPath = cfgFile // Сохраняем путь

	// Валидация и генерация ID для старых конфигов без ID
	for i := range cfg.ProtocolConfigs {
		if cfg.ProtocolConfigs[i].ID == "" {
			cfg.ProtocolConfigs[i].ID = uuid.New().String()
		}
	}

	if len(cfg.ProtocolConfigs) == 0 {
		return nil, fmt.Errorf("no protocol configurations found in %s", cfgFile)
	}
	if cfg.NatsURL == "" {
		return nil, fmt.Errorf("nats_url is not specified in %s", cfgFile)
	}

	if cfg.Logging.FilePath != "" {
		logDir := filepath.Dir(cfg.Logging.FilePath)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory %s: %w", logDir, err)
		}
	}

	fmt.Printf("Config loaded from: %s\n", cfgFile)
	return &cfg, nil
}

// resolveConfigPath определяет итоговый путь к файлу конфигурации.
func resolveConfigPath(configPath *string) string {
	if configPath != nil && *configPath != "" {
		return *configPath
	}
	if envPath := os.Getenv("RECEIVER_CONFIG_PATH"); envPath != "" {
		return envPath
	}
	return "./configs/receiver.toml" // Путь по умолчанию
}

// Save перезаписывает конфигурационный файл текущим состоянием.
func (c *Config) Save() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.configPath == "" {
		return fmt.Errorf("config path is not set, cannot save")
	}

	// Создаем временный файл для атомарной записи
	tmpFile := c.configPath + ".tmp"
	f, err := os.Create(tmpFile)
	if err != nil {
		return fmt.Errorf("failed to create temp config file: %w", err)
	}
	defer f.Close()

	// Используем toml.NewEncoder для красивого форматирования
	encoder := toml.NewEncoder(f)
	if err := encoder.Encode(c); err != nil {
		return fmt.Errorf("failed to encode config to TOML: %w", err)
	}

	// Атомарно заменяем старый файл новым
	if err := os.Rename(tmpFile, c.configPath); err != nil {
		return fmt.Errorf("failed to rename temp config file: %w", err)
	}

	fmt.Printf("Config successfully saved to: %s\n", c.configPath)
	return nil
}

// --- Методы для управления портами ---

// AddPort добавляет новую конфигурацию порта.
func (c *Config) AddPort(name string, port int) (*ProtocolConfig, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Проверяем, не занят ли порт
	for _, pCfg := range c.ProtocolConfigs {
		if pCfg.Port == port {
			return nil, fmt.Errorf("port %d is already in use", port)
		}
	}

	newPortCfg := ProtocolConfig{
		ID:     uuid.New().String(),
		Name:   strings.ToUpper(name),
		Port:   port,
		Active: true, // По умолчанию добавлен, но не активен
	}
	c.ProtocolConfigs = append(c.ProtocolConfigs, newPortCfg)

	// Убираем сохранение отсюда, оно будет после перезапуска обработчиков
	// // Сразу сохраняем изменения в файл
	// if err := c.Save(); err != nil {
	// 	// Откатываем изменение, если не удалось сохранить
	// 	c.ProtocolConfigs = c.ProtocolConfigs[:len(c.ProtocolConfigs)-1]
	// 	return nil, fmt.Errorf("failed to save config after adding port: %w", err)
	// }

	return &newPortCfg, nil
}

// DeletePort удаляет конфигурацию порта по его ID.
func (c *Config) DeletePort(id string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	found := false
	newConfigs := make([]ProtocolConfig, 0, len(c.ProtocolConfigs))
	for _, pCfg := range c.ProtocolConfigs {
		if pCfg.ID == id {
			found = true
		} else {
			newConfigs = append(newConfigs, pCfg)
		}
	}

	if !found {
		return fmt.Errorf("port with id %s not found", id)
	}

	c.ProtocolConfigs = newConfigs

	if err := c.Save(); err != nil {
		// Откатываем изменение
		c.ProtocolConfigs = append(c.ProtocolConfigs, ProtocolConfig{ID: id}) // Упрощенный откат
		return fmt.Errorf("failed to save config after deleting port: %w", err)
	}

	return nil
}

// SetPortState меняет состояние (активен/неактивен) порта по его ID.
func (c *Config) SetPortState(id string, active bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i, pCfg := range c.ProtocolConfigs {
		if pCfg.ID == id {
			if pCfg.Active == active {
				return nil // Состояние уже нужное, ничего не делаем
			}
			c.ProtocolConfigs[i].Active = active

			if err := c.Save(); err != nil {
				// Откатываем изменение
				c.ProtocolConfigs[i].Active = !active
				return fmt.Errorf("failed to save config after changing port state: %w", err)
			}
			return nil
		}
	}
	return fmt.Errorf("port with id %s not found", id)
}

// GetPortByID находит конфигурацию порта по ID.
func (c *Config) GetPortByID(id string) (*ProtocolConfig, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, pCfg := range c.ProtocolConfigs {
		if pCfg.ID == id {
			return &pCfg, nil
		}
	}
	return nil, fmt.Errorf("port with id %s not found", id)
}
