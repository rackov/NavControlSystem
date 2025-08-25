package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/sirupsen/logrus"
)

const (
	logFileName = "app.log"
)

var (
	logInstance *logrus.Logger
	once        sync.Once
)

// GetLogger возвращает экземпляр логгера (синглтон).
func GetLogger() *logrus.Logger {
	once.Do(func() {
		logInstance = logrus.New()

		// --- 1. Настройка вывода в КОНСОЛЬ ---
		// Устанавливаем вывод по умолчанию в stdout (консоль)
		logInstance.SetOutput(os.Stdout)
		// Устанавливаем текстовый форматер для консоли
		// НОВЫЙ ФОРМАТ ВРЕМЕНИ
		customTimeFormat := "2006-01-02 15:04:05.000"

		logInstance.SetFormatter(&logrus.TextFormatter{
			// Раскрашиваем логи для удобства
			ForceColors: true,
			// Отображать полное время, а не только время от запуска
			FullTimestamp:   true,
			TimestampFormat: customTimeFormat,
		})
		// Устанавливаем уровень логирования
		logInstance.SetLevel(logrus.InfoLevel)

		// --- 2. Настройка вывода в ФАЙЛ через ХУК ---
		logFilePath := filepath.Join("logs", logFileName)
		file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			// Если не удалось открыть файл, логируем ошибку в stderr и выходим
			fmt.Fprintf(os.Stderr, "Не удалось открыть лог-файл: %v\n", err)
			os.Exit(1)
		}

		jsonFormatter := &logrus.JSONFormatter{
			// Устанавливаем тот же формат времени для файла
			TimestampFormat: customTimeFormat,
		}
		// Создаем и добавляем хук
		logInstance.AddHook(&fileHook{
			writer:    file,
			formatter: jsonFormatter,     // Формат для файла - JSON
			level:     logrus.TraceLevel, // Записывать в файл все логи уровня Debug и выше
		})
	})
	return logInstance
}

// SetLevel - вспомогательная функция для изменения уровня логирования.
func SetLevel(level string) error {
	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		return err
	}
	GetLogger().SetLevel(lvl)
	GetLogger().Infof("Уровень логирования изменен на %s", level)
	return nil
}

// GetLogFilePath возвращает путь к лог-файлу.
func GetLogFilePath() string {
	GetLogger()
	return filepath.Join("logs", logFileName)
}

// --- НОВЫЙ КОД: Наш собственный хук ---
// fileHook реализует интерфейс logrus.Hook
type fileHook struct {
	writer    io.Writer
	formatter logrus.Formatter
	level     logrus.Level
}

// Fire вызывается, когда логируется событие.
// Мы решаем, что делать с записью (entry).
func (hook *fileHook) Fire(entry *logrus.Entry) error {
	// Форматируем запись согласно нашему форматеру (JSON)
	msg, err := hook.formatter.Format(entry)
	if err != nil {
		return err
	}

	// Записываем отформатированное сообщение в наш writer (файл)
	_, err = hook.writer.Write(msg)
	return err
}

// Levels возвращает уровни логирования, на которые этот хук должен реагировать.
func (hook *fileHook) Levels() []logrus.Level {
	// Мы хотим, чтобы хук срабатывал на всех уровнях, начиная с hook.level
	return logrus.AllLevels[:hook.level+1]
}
