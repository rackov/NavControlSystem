package tnats

import (
	"encoding/json"
	"log"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/rackov/NavControlSystem/pkg/models"
)

// Publisher - это интерфейс для публикации сообщений.
type Publisher struct {
	nc             *nats.Conn
	js             nats.JetStreamContext
	statusChan     chan bool // true для подключен, false для отключен
	natsURL        string
	reconnectDelay int
}

func NewPublisher(url string, delay int) *Publisher {
	return &Publisher{
		natsURL:        url,
		statusChan:     make(chan bool, 1),
		reconnectDelay: delay,
	}
}

// Connect пытается подключиться к NATS и начинает отслеживать состояние соединения.
func (p *Publisher) Connect() {
	var err error
	opts := []nats.Option{
		nats.ReconnectWait(time.Duration(p.reconnectDelay) * time.Second),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			if err != nil {
				log.Printf("NATS отключен: %v", err)
			}
			p.sendStatus(false)
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			log.Printf("NATS переподключен к %s", nc.ConnectedUrl())
			p.sendStatus(true)
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			log.Printf("Соединение NATS закрыто: %v", nc.LastError())
			p.sendStatus(false)
		}),
	}

	p.nc, err = nats.Connect(p.natsURL, opts...)
	if err != nil {
		log.Printf("Не удалось подключиться к NATS на %s, будет повторная попытка: %v", p.natsURL, err)
		p.sendStatus(false)
		// Запускаем цикл повторных попыток в горутине
		go p.retryConnect()
		return
	}

	p.js, err = p.nc.JetStream()
	if err != nil {
		log.Printf("Не удалось получить контекст JetStream, будет использован стандартный NATS: %v", err)
	}

	log.Printf("Успешное подключение к NATS на %s", p.natsURL)
	p.sendStatus(true)
}

// retryConnect управляет логикой повтора и переподключения.
func (p *Publisher) retryConnect() {
	for {
		time.Sleep(time.Duration(p.reconnectDelay) * time.Second)
		log.Println("Попытка переподключения к NATS...")
		var err error
		p.nc, err = nats.Connect(p.natsURL, nats.ReconnectWait(time.Duration(p.reconnectDelay)*time.Second))
		if err == nil {
			p.js, _ = p.nc.JetStream() // Пытаемся получить контекст JS
			log.Printf("NATS успешно переподключен к %s", p.natsURL)
			p.sendStatus(true)
			return
		}
		log.Printf("Переподключение NATS не удалось: %v", err)
	}
}

func (p *Publisher) sendStatus(connected bool) {
	// Неблокирующая отправка
	select {
	case p.statusChan <- connected:
	default:
	}
}

func (p *Publisher) StatusChannel() <-chan bool {
	return p.statusChan
}

func (p *Publisher) Publish(subject string, data *models.NavRecord) error {
	if p.nc == nil || !p.nc.IsConnected() {
		return nats.ErrConnectionClosed
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// Используем JetStream Publish, если доступно, иначе стандартный Publish
	if p.js != nil {
		_, err = p.js.Publish(subject, jsonData)
	} else {
		err = p.nc.Publish(subject, jsonData)
	}

	return err
}

func (p *Publisher) Close() {
	if p.nc != nil {
		p.nc.Close()
	}
}
