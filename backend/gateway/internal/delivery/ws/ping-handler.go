package ws

import (
	"context"
	"time"

	"github.com/gorilla/websocket"

	"quickflow/shared/logger"
)

type PingHandler interface {
	Handle(ctx context.Context, conn *websocket.Conn)
}

// PingHandlerWS - Обработчик Ping сообщений
type PingHandlerWS struct{}

func NewPingHandlerWS() *PingHandlerWS {
	return &PingHandlerWS{}
}

func (wm *PingHandlerWS) Handle(ctx context.Context, conn *websocket.Conn) {
	conn.SetPongHandler(func(appData string) error {
		logger.Info(ctx, "Received pong: %v", appData)
		return nil
	})

	go func() {
		for {
			time.Sleep(30 * time.Second) // отправка ping каждые 30 секунд
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				logger.Info(ctx, "Failed to send ping: %v", err)
				return
			}
		}
	}()
	return
}
