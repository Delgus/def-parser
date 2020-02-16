package sse

import (
	"encoding/json"
	"fmt"
	"net/http"

	sse "github.com/alexandrevicenzi/go-sse"
	"github.com/delgus/def-parser/internal/app"
)

// Notifier реализует интерфейс для оповещения пользователя
type Notifier struct {
	server *sse.Server
	route  string
}

// NewNotifier вернет *Notifier
func NewNotifier(route string) *Notifier {
	n := &Notifier{route: route, server: sse.NewServer(&sse.Options{
		// Increase default retry interval to 10s.
		RetryInterval: 10 * 1000,
		// Print debug info
		Logger: nil,
	})}
	http.Handle(n.route, n.server)
	return n
}

// Publish отправит сообщение клиенту
func (n *Notifier) Publish(stream string, site *app.Site) error {
	siteBytes, err := json.Marshal(site)
	if err != nil {
		return fmt.Errorf(`can not convert message to json: %v`, err)
	}
	n.server.SendMessage(n.route+stream, sse.SimpleMessage(string(siteBytes)))
	return nil
}

// CreateStream отправит сообщение ping чтобы установить соединение с клиентом
func (n *Notifier) CreateStream(stream string) {
	n.server.SendMessage(n.route+stream, sse.SimpleMessage("ping"))
}

// Shutdown Плавная остановка рассыльщика сообщений
func (n *Notifier) Shutdown() {
	n.server.Shutdown()
}
