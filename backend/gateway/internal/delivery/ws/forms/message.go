package forms

import (
    "encoding/json"

    "github.com/google/uuid"
)

type MessageRequest struct {
    Type    string          `json:"type"`
    Payload json.RawMessage `json:"payload"`
}

type MarkReadPayload struct {
    ChatId    uuid.UUID `json:"chat_id"`
    MessageId uuid.UUID `json:"message_id"`
}

type NotifyMessageRead struct {
    ChatId    uuid.UUID `json:"chat_id"`
    MessageId uuid.UUID `json:"message_id"`
    Timestamp string    `json:"ts"`
    SenderId  uuid.UUID `json:"sender_id"`
}

//easyjson:json
type DeleteMessagePayload struct {
    MessageId uuid.UUID `json:"message_id"`
}

type NotifyDeleteMessage struct {
    ChatId    uuid.UUID `json:"chat_id"`
    MessageId uuid.UUID `json:"message_id"`
}

type DeleteChatPayload struct {
    ChatId uuid.UUID `json:"chat_id"`
}
