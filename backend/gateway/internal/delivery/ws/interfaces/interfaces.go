package interfaces

import (
    `context`
    `encoding/json`

    `github.com/google/uuid`
    `github.com/gorilla/websocket`

    `quickflow/gateway/internal/delivery/http/forms`
    `quickflow/shared/models`
)

type IWebSocketManager interface {
    SendMessageToUser(ctx context.Context, userId uuid.UUID, message forms.MessageOut) error
    SendMessageToChat(ctx context.Context, message models.Message, publicSenderInfo models.PublicUserInfo, chatParticipants []models.User) error
    IsConnected(userId uuid.UUID) (*websocket.Conn, bool)
    HandlePing(conn *websocket.Conn)
    AddConnection(userId uuid.UUID, conn *websocket.Conn)
    RemoveAndCloseConnection(userId uuid.UUID)
}

// IWebSocketConnectionManager интерфейс для управления соединениями
type IWebSocketConnectionManager interface {
    AddConnection(userId uuid.UUID, conn *websocket.Conn)
    RemoveAndCloseConnection(userId uuid.UUID)
    IsConnected(userId uuid.UUID) (*websocket.Conn, bool)
}

type CommandHandler func(ctx context.Context, user models.User, payload json.RawMessage) error
type IWebSocketRouter interface {
    RegisterHandler(command string, handler CommandHandler)
    Route(ctx context.Context, command string, user models.User, payload json.RawMessage) error
}
