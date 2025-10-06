package ws

import (
    "encoding/json"
    "errors"
    "fmt"
    "log"
    "net/http"

    "github.com/gorilla/websocket"
    "github.com/microcosm-cc/bluemonday"

    "quickflow/gateway/internal/delivery/http/forms"
    `quickflow/gateway/internal/delivery/http/interfaces`
    forms2 "quickflow/gateway/internal/delivery/ws/forms"
    interfaces2 `quickflow/gateway/internal/delivery/ws/interfaces`
    errors2 "quickflow/gateway/internal/errors"
    http2 "quickflow/gateway/utils/http"
    "quickflow/shared/logger"
    "quickflow/shared/models"
)

// MessageListenerWS Обработчик сообщений
type MessageListenerWS struct {
    profileUseCase   interfaces.ProfileUseCase
    WebSocketManager interfaces2.IWebSocketConnectionManager
    WebSocketRouter  interfaces2.IWebSocketRouter
    policy           *bluemonday.Policy
}

func NewMessageListenerWS(profileUseCase interfaces.ProfileUseCase,
    webSocketManager interfaces2.IWebSocketConnectionManager,
    webSocketRouter interfaces2.IWebSocketRouter, policy *bluemonday.Policy) *MessageListenerWS {
    return &MessageListenerWS{
        profileUseCase:   profileUseCase,
        WebSocketManager: webSocketManager,
        policy:           policy,
        WebSocketRouter:  webSocketRouter,
    }
}

// HandleMessages godoc
// @Summary SendMessage incoming messages
// @Description SendMessage incoming messages
// @Tags WebSocket
// @Accept json
// @Produce json
// @Param message body forms.MessageForm true "Message"
// @Success 200 {object} forms.MessageOut "Message"
// @Failure 400 {object} forms.ErrorForm "Invalid data"
// @Failure 500 {object} forms.ErrorForm "Server error"
// @Router /api/ws [get]
func (m *MessageListenerWS) HandleMessages(w http.ResponseWriter, r *http.Request) {
    ctx := http2.SetRequestId(r.Context())

    user, ok := ctx.Value("user").(models.User)
    if !ok {
        logger.Error(ctx, "Failed to get user from context while handling messages")
        http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to get user from context", http.StatusInternalServerError))
        return
    }

    conn, found := m.WebSocketManager.IsConnected(user.Id)
    if !found {
        logger.Error(ctx, "WebSocket connection not found for user: %s", user.Id)
        http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "WebSocket connection not found", http.StatusInternalServerError))
        return
    }

    defer func() {
        if err := m.profileUseCase.UpdateLastSeen(ctx, user.Id); err != nil {
            err = errors2.FromGRPCError(err)
            logger.Error(ctx, "Failed to update last seen: %s", err)
            http2.WriteJSONError(w, err)
        }
    }()

    for {
        var messageRequest forms2.MessageRequest

        _, msg, err := conn.ReadMessage()
        if err != nil {
            var closeErr *websocket.CloseError
            if errors.As(err, &closeErr) {
                logger.Info(ctx, "WebSocket closed by user %v: %v", user.Id, err)
            } else {
                logger.Error(ctx, "Error reading WS message for user %v: %v", user.Id, err)
            }
            return
        }

        if err := json.Unmarshal(msg, &messageRequest); err != nil {
            logger.Error(ctx, "Failed to unmarshal WS message: %v", err)
            writeErrorToWS(conn, fmt.Sprintf("Invalid message format: %v", err))
            continue
        }

        if err := m.WebSocketRouter.Route(ctx, messageRequest.Type, user, messageRequest.Payload); err != nil {
            logger.Error(ctx, "Failed to route WS message: %v", err)
            writeErrorToWS(conn, fmt.Sprintf("Failed to process message: %v", err))
            continue
        }
    }
}

func writeErrorToWS(conn *websocket.Conn, errMsg string) {
    if err := conn.WriteJSON(forms.ErrorForm{ErrorCode: errMsg}); err != nil {
        log.Printf("Failed to send WS error message: %v", err)
    }
}
