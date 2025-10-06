package ws

import (
    "context"
    "encoding/json"
    "fmt"

    `quickflow/gateway/internal/delivery/ws/interfaces`
    "quickflow/shared/models"
)

// WebSocketRouter - Внутренний роутер для обработки команд, передаваемых через WebSocket
type WebSocketRouter struct {
    handlers map[string]interfaces.CommandHandler
}

func (r *WebSocketRouter) RegisterHandler(command string, handler interfaces.CommandHandler) {
    if r.handlers == nil {
        r.handlers = make(map[string]interfaces.CommandHandler)
    }
    r.handlers[command] = handler
}

func (r *WebSocketRouter) Route(ctx context.Context, command string, user models.User, payload json.RawMessage) error {
    handler, found := r.handlers[command]
    if !found {
        return fmt.Errorf("no handler for command: %s", command)
    }
    return handler(ctx, user, payload)
}

func NewWebSocketRouter() *WebSocketRouter {
    return &WebSocketRouter{}
}
