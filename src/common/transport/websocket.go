package transport

import (
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"io"
	"net/url"
)

type websocketClient struct {
	conn       *websocket.Conn
	connClosed chan struct{}
	handler    MessageHandler
}

type WebsocketClient interface {
	SendMessage(message []byte)
	Close()
}

type MessageHandler func([]byte)

func NewWebsocketClient(addr string, handler MessageHandler) WebsocketClient {
	u := url.URL{Scheme: "ws", Host: addr, Path: "/ws"}
	log.WithField("addr", u.String()).Info("connecting to websocket")

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.WithError(err).Fatal("websocket dial error")
	}

	c := &websocketClient{conn, make(chan struct{}), handler}
	go c.readMessages()

	return c
}

func (c *websocketClient) SendMessage(message []byte) {
	err := c.conn.WriteMessage(websocket.TextMessage, message)
	if err != nil {
		log.WithError(err).Fatal("write message")
	}
}

func (c *websocketClient) Close() {
	if c.conn != nil {
		_ = c.conn.Close()
		<-c.connClosed
	}
}

func (c *websocketClient) readMessages() {
	defer close(c.connClosed)
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil || err == io.EOF {
			log.WithError(err).Error("Error reading")
			return
		}

		log.WithField("m", message).Info("message from websocket")
		if c.handler != nil {
			c.handler(message)
		}
	}
}
