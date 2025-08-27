package logger

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Уровни логирования, соответствующие logrus
const (
	TRACE = logrus.TraceLevel
	DEBUG = logrus.DebugLevel
	INFO  = logrus.InfoLevel
	WARN  = logrus.WarnLevel
	ERROR = logrus.ErrorLevel
)

// Logger - это наша обертка над logrus.Logger
type Logger struct {
	*logrus.Logger
}

var (
	defaultLogger = New(INFO) // Создаем экземпляр с уровнем по умолчанию
)

// New создает и настраивает новый экземпляр логгера.
// filePath - путь к файлу для записи логов. Если пустой, логи пишутся только в stdout.
func New(level logrus.Level, filePath ...string) *Logger {
	l := logrus.New()

	// Устанавливаем текстовый форматтер (как в стандартном log)
	// Формат: time="2023-10-27T10:30:00+03:00" level=info msg="Some message"
	l.SetFormatter(&logrus.TextFormatter{
		DisableColors:   true,                  // Отключаем цвета для вывода в файл
		FullTimestamp:   true,                  // Включаем полную временную метку
		TimestampFormat: "2006-01-02 15:04:05", // Формат времени
	})

	l.SetOutput(os.Stdout) // По умолчанию выводим в консоль
	l.SetLevel(level)      // Устанавливаем уровень

	// Если передан путь к файлу, настраиваем запись в файл с ротацией
	if len(filePath) > 0 && filePath[0] != "" {
		// Создаем MultiWriter для записи и в консоль, и в файл
		multiWriter := io.MultiWriter(os.Stdout, &lumberjack.Logger{
			Filename:   filePath[0],
			MaxSize:    100,  // MB
			MaxBackups: 3,    // Хранить 3 старых файла логов
			MaxAge:     28,   // Дней
			Compress:   true, // Сжимать старые логи
		})
		l.SetOutput(multiWriter)
	}

	return &Logger{l}
}

// --- Методы для установки уровня ---

// SetLevel динамически меняет уровень логирования
func (l *Logger) SetLevel(level logrus.Level) {
	l.Logger.SetLevel(level)
}

// SetGlobalLevel устанавливает уровень для глобального логгера
func SetGlobalLevel(level logrus.Level) {
	defaultLogger.SetLevel(level)
}

// --- Глобальные функции-хелперы для удобства ---

// Они вызывают методы у глобального экземпляра defaultLogger
func Trace(args ...interface{}) {
	defaultLogger.Trace(args...)
}
func Tracef(format string, args ...interface{}) {
	defaultLogger.Tracef(format, args...)
}
func Info(args ...interface{}) {
	defaultLogger.Info(args...)
}

func Debug(args ...interface{}) {
	defaultLogger.Debug(args...)
}

func Warn(args ...interface{}) {
	defaultLogger.Warn(args...)
}

func Error(args ...interface{}) {
	defaultLogger.Error(args...)
}

func Infof(format string, args ...interface{}) {
	defaultLogger.Infof(format, args...)
}

func Debugf(format string, args ...interface{}) {
	defaultLogger.Debugf(format, args...)
}

func Warnf(format string, args ...interface{}) {
	defaultLogger.Warnf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	defaultLogger.Errorf(format, args...)
}

// GetLogger возвращает экземпляр глобального логгера.
// Это может быть полезно, если какой-то компонент хочет получить
// прямой доступ к логгеру, а не использовать глобальные функции.
func GetLogger() *Logger {
	return defaultLogger
}
