// NavControlSystem/pkg/nats/tnats.go
package tnats

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
)

// Common errors
var (
	ErrNotConnected = errors.New("nats: not connected")
)

// Client - это наша обертка для NATS-соединения.
// Он управляет жизненным циклом подключения и предоставляет методы для создания Publisher/Subscriber.
type Client struct {
	conn   *nats.Conn
	js     nats.JetStreamContext // Для продвинутых возможностей (streams, consumers)
	logger *logrus.Logger
	mu     sync.RWMutex
	closed bool
}

// Publisher - это интерфейс для публикации сообщений.
type Publisher interface {
	Publish(subject string, data []byte) error
	Close() error
}

// Subscriber - это интерфейс для подписки на сообщения.
type Subscriber interface {
	Subscribe(subject string, handler nats.MsgHandler) (*nats.Subscription, error)
	Close() error
}

// NewClient создает новый NATS клиент.
// Он пытается установить соединение при создании.
func NewClient(url string, logger *logrus.Logger) (*Client, error) {
	if logger == nil {
		logger = logrus.New()
		logger.Warn("No logger provided to NATS client, using default")
	}

	opts := []nats.Option{
		nats.ReconnectWait(2 * time.Second),
		nats.MaxReconnects(5),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			logger.WithError(err).Warn("NATS connection disconnected")
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			logger.Info("NATS connection reconnected")
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			logger.Info("NATS connection closed")
		}),
	}

	nc, err := nats.Connect(url, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	// Включаем JetStream для будущих нужд (например, для гарантированной доставки)
	js, err := nc.JetStream()
	if err != nil {
		// Если JetStream не доступен, можно продолжить без него, но лучше предупредить
		logger.WithError(err).Warn("JetStream is not available, continuing without it")
	}

	logger.Info("Successfully connected to NATS")

	return &Client{
		conn:   nc,
		js:     js,
		logger: logger,
	}, nil
}

// Close закрывает соединение с NATS.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	if c.conn != nil {
		c.conn.Close()
	}
	c.closed = true
	c.logger.Info("NATS client connection closed")
	return nil
}

// IsConnected проверяет, активно ли соединение.
func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.conn != nil && c.conn.IsConnected() && !c.closed
}

// NewPublisher создает новый Publisher.
// В данном случае, сам Client реализует интерфейс Publisher для простоты.
func (c *Client) NewPublisher() (Publisher, error) {
	if !c.IsConnected() {
		return nil, ErrNotConnected
	}
	// Возвращаем сам клиент, т.к. он уже имеет метод Publish.
	// Это паттерн, когда один объект может выступать в разных ролях.
	return c, nil
}

// Publish публикует сообщение в указанный топик.
func (c *Client) Publish(subject string, data []byte) error {
	if !c.IsConnected() {
		return ErrNotConnected
	}

	err := c.conn.Publish(subject, data)
	if err != nil {
		c.logger.WithFields(logrus.Fields{
			"subject": subject,
			"error":   err,
		}).Error("Failed to publish message")
		return fmt.Errorf("nats publish failed: %w", err)
	}

	c.logger.WithField("subject", subject).Debug("Message published successfully")
	return nil
}

// NewSubscriber создает новый Subscriber.
// Аналогично Publisher, сам Client реализует этот интерфейс.
func (c *Client) NewSubscriber() (Subscriber, error) {
	if !c.IsConnected() {
		return nil, ErrNotConnected
	}
	return c, nil
}

// Subscribe подписывается на указанный топик и вызывает handler для каждого сообщения.
func (c *Client) Subscribe(subject string, handler nats.MsgHandler) (*nats.Subscription, error) {
	if !c.IsConnected() {
		return nil, ErrNotConnected
	}

	sub, err := c.conn.Subscribe(subject, handler)
	if err != nil {
		c.logger.WithFields(logrus.Fields{
			"subject": subject,
			"error":   err,
		}).Error("Failed to subscribe")
		return nil, fmt.Errorf("nats subscribe failed: %w", err)
	}

	c.logger.WithFields(logrus.Fields{
		"subject": subject,
		"queue":   sub.Queue,
	}).Info("Successfully subscribed to subject")
	return sub, nil
}

// QueueSubscribe подписывается с группой (queue). Только один подписчик в группе получит сообщение.
func (c *Client) QueueSubscribe(subject, queue string, handler nats.MsgHandler) (*nats.Subscription, error) {
	if !c.IsConnected() {
		return nil, ErrNotConnected
	}

	sub, err := c.conn.QueueSubscribe(subject, queue, handler)
	if err != nil {
		c.logger.WithFields(logrus.Fields{
			"subject": subject,
			"queue":   queue,
			"error":   err,
		}).Error("Failed to queue subscribe")
		return nil, fmt.Errorf("nats queue subscribe failed: %w", err)
	}

	c.logger.WithFields(logrus.Fields{
		"subject": subject,
		"queue":   queue,
	}).Info("Successfully queue subscribed to subject")
	return sub, nil
}
