// NavControlSystem/pkg/logger/logger.go
package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger - это наша основная структура, которая инкапсулирует logrus и lumberjack.
// Мы экспортируем ее, чтобы пользователь мог работать с экземпляром.
type Logger struct {
	*logrus.Logger                    // Встраиваем стандартный логгер logrus
	fileHook       *lumberjack.Logger // Сохраняем ссылку на lumberjack для доступа к Sync()
}

var (
	defaultLogger *Logger
	once          sync.Once
)

// Config содержит все необходимые настройки для инициализации логгера.
type Config struct {
	LogLevel    string `toml:"log_level"`
	LogFilePath string `toml:"log_file_path"`
	MaxSize     int    `toml:"max_size"`
	MaxBackups  int    `toml:"max_backups"`
	MaxAge      int    `toml:"max_age"`
	Compress    bool   `toml:"compress"`
}

// NewLogger создает и возвращает новый настроенный экземпляр Logger.
// Это основная функция для создания логгера.
func NewLogger(cfg Config) (*Logger, error) {
	var l *Logger
	var err error

	once.Do(func() {
		// 1. Создаем lumberjack для ротации логов
		fileHook := &lumberjack.Logger{
			Filename:   cfg.LogFilePath,
			MaxSize:    cfg.MaxSize,    // megabytes
			MaxBackups: cfg.MaxBackups, // files
			MaxAge:     cfg.MaxAge,     // days
			Compress:   cfg.Compress,   // compress
		}

		// 2. Создаем стандартный логгер logrus
		baseLogger := logrus.New()

		// 3. Устанавливаем уровень логирования
		level, err := logrus.ParseLevel(cfg.LogLevel)
		if err != nil {
			level = logrus.InfoLevel
			fmt.Fprintf(os.Stderr, "Failed to parse log level '%s', defaulting to 'info'. Error: %v\n", cfg.LogLevel, err)
		}
		baseLogger.SetLevel(level)

		// 4. Устанавливаем форматтер
		baseLogger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
			CallerPrettyfier: func(frame *runtime.Frame) (function string, file string) {
				return "", fmt.Sprintf("%s:%d", filepath.Base(frame.File), frame.Line)
			},
		})
		baseLogger.SetReportCaller(true)

		// 5. Устанавливаем вывод. lumberjack будет основным.
		// Можно добавить вывод в консоль для разработки, раскомментировав строку ниже.
		// baseLogger.SetOutput(io.MultiWriter(fileHook, os.Stdout))
		baseLogger.SetOutput(fileHook)

		// 6. Создаем наш экземпляр Logger
		l = &Logger{
			Logger:   baseLogger,
			fileHook: fileHook,
		}

		defaultLogger = l // Сохраняем как "логгер по умолчанию"
	})

	if l == nil {
		return nil, fmt.Errorf("failed to initialize logger after Do block")
	}
	return l, err
}

// Sync сбрасывает буферы на диск. Реализует интерфейс syncer.
// Теперь это метод нашего Logger, что делает API последовательным.
func (l *Logger) Close() error {
	if l.fileHook != nil {
		return l.fileHook.Close()
	}
	return nil
}

// SetLevel динамически изменяет уровень логирования.
// Также сделано методом для работы с экземпляром.
func (l *Logger) SetLevel(levelStr string) error {
	level, err := logrus.ParseLevel(levelStr)
	if err != nil {
		return fmt.Errorf("invalid log level '%s': %w", levelStr, err)
	}
	l.Logger.SetLevel(level)
	l.Infof("Log level changed to %s", level.String())
	return nil
}

// Default возвращает "логгер по умолчанию".
// Полезно, если не хочется передавать логгер через всю иерархию вызовов.
func Default() *Logger {
	if defaultLogger == nil {
		// Можно вернуть логгер с настройками по умолчанию, если он еще не был создан
		// или паниковать, это зависит от политики проекта.
		panic("Logger not initialized. Call NewLogger first.")
	}
	return defaultLogger
}
