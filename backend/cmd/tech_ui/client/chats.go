package client

import (
	"fmt"
	"quickflow/cmd/tech_ui/models"
)

func (c *APIClient) GetChats(count int, ts string) ([]models.ChatOut, error) {
	path := fmt.Sprintf("/api/v2/chats?count=%d", count)
	if ts != "" {
		path += "&ts=" + ts
	}

	resp, err := c.doRequest("GET", path, nil, nil)
	if err != nil {
		return nil, err
	}

	var chats []models.ChatOut
	if err := c.parseResponse(resp, &chats); err != nil {
		return nil, err
	}

	return chats, nil
}

func (c *APIClient) GetChatMessages(chatID string, count int, ts string) ([]models.MessageOut, error) {
	path := fmt.Sprintf("/api/v2/chats/%s/messages?count=%d", chatID, count)
	if ts != "" {
		path += "&ts=" + ts
	}

	resp, err := c.doRequest("GET", path, nil, nil)
	if err != nil {
		return nil, err
	}

	messages := struct {
		Msgs []models.MessageOut `json:"messages"`
	}{}
	if err := c.parseResponse(resp, &messages); err != nil {
		return nil, err
	}

	return messages.Msgs, nil
}
