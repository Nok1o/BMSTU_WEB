package client

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"net/http"
	"quickflow/cmd/tech_ui/models"
	"strings"
	"time"
)

type WSClient struct {
	conn     *websocket.Conn
	handlers map[string]func([]byte)
}

func (c *APIClient) ConnectWebSocket() (*WSClient, error) {
	url := "ws" + strings.TrimPrefix(c.BaseURL, "http") + "/api/v2/ws"

	headers := http.Header{}
	//headers.Add("Connection", "keep-alive")
	if c.SessionID != "" {
		headers.Add("Cookie", "session="+c.SessionID)
	}

	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}
	conn, _, err := dialer.Dial(url, headers)
	if err != nil {
		return nil, err
	}

	//// Set session cookie
	//if c.SessionID != "" {
	//	conn.WriteMessage(websocket.TextMessage, []byte("Cookie: session="+c.SessionID))
	//}

	wsClient := &WSClient{
		conn:     conn,
		handlers: make(map[string]func([]byte)),
	}

	return wsClient, nil
}

func (w *WSClient) SendMessage(messageType string, payload interface{}) error {
	message := models.WSMessage{
		Type: messageType,
	}

	payloadData, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	message.Payload = payloadData

	messageData, err := json.Marshal(message)
	if err != nil {
		return err
	}

	return w.conn.WriteMessage(websocket.TextMessage, messageData)
}

func (w *WSClient) SendTextMessage(text, chatID, receiverID string, media, audio, files []string) error {
	payload := models.WSMessagePayload{
		Text:  text,
		Media: media,
		Audio: audio,
		File:  files,
	}

	if chatID == "" && receiverID == "" {
		return fmt.Errorf("either chatID or receiverID must be provided")
	}
	if chatID == "" {
		payload.ReceiverID = receiverID
		payload.ChatID = uuid.Nil.String()
	}

	if receiverID == "" {
		payload.ChatID = chatID
		payload.ReceiverID = uuid.Nil.String()
	}

	return w.SendMessage("message", payload)
}

func (w *WSClient) MarkMessageRead(chatID, messageID string) error {
	payload := models.WSMessageRead{
		ChatID:    chatID,
		MessageID: messageID,
	}

	return w.SendMessage("message_read", payload)
}

func (w *WSClient) DeleteMessage(messageID string) error {
	payload := map[string]string{
		"message_id": messageID,
	}

	return w.SendMessage("message_delete", payload)
}

func (w *WSClient) DeleteChat(chatID string) error {
	payload := map[string]string{
		"chat_id": chatID,
	}

	return w.SendMessage("chat_delete", payload)
}

func (w *WSClient) OnMessage(handler func(messageType string, payload []byte)) {
	go func() {
		for {
			_, message, err := w.conn.ReadMessage()
			if err != nil {
				fmt.Printf("WebSocket read error: %v\n", err)
				return
			}

			var wsMessage models.WSMessage
			if err := json.Unmarshal(message, &wsMessage); err != nil {
				fmt.Printf("WebSocket parse error: %v\n", err)
				continue
			}

			handler(wsMessage.Type, wsMessage.Payload)
		}
	}()
}

func (w *WSClient) Close() error {
	return w.conn.Close()
}
