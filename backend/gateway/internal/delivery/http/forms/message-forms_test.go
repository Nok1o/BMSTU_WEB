//go:build unit
// +build unit

package forms

import (
	"net/url"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	time2 "quickflow/config/time"
	"quickflow/shared/models"
)

func TestGetMessagesForm_GetParams(t *testing.T) {
	now := time.Now()
	nowStr := now.Format(time2.TimeStampLayout)

	tests := []struct {
		name        string
		values      url.Values
		expected    GetMessagesForm
		expectError bool
		errMessage  string
	}{
		{
			name: "Valid params with timestamp",
			values: url.Values{
				"messages_count": []string{"50"},
				"ts":             []string{nowStr},
			},
			expected: GetMessagesForm{
				MessagesCount: 50,
				Ts:            now,
			},
			expectError: false,
		},
		{
			name: "Missing messages_count",
			values: url.Values{
				"ts": []string{nowStr},
			},
			expectError: true,
			errMessage:  "messages_count parameter missing",
		},
		{
			name: "Invalid messages_count",
			values: url.Values{
				"messages_count": []string{"invalid"},
				"ts":             []string{nowStr},
			},
			expectError: true,
			errMessage:  "failed to parse messages_count",
		},
		{
			name: "Invalid timestamp",
			values: url.Values{
				"messages_count": []string{"20"},
				"ts":             []string{"invalid"},
			},
			expected: GetMessagesForm{
				MessagesCount: 20,
				Ts:            time.Now(), // Will be set to current time
			},
			expectError: false,
		},
		{
			name: "No timestamp provided",
			values: url.Values{
				"messages_count": []string{"10"},
			},
			expected: GetMessagesForm{
				MessagesCount: 10,
				Ts:            time.Now(), // Will be set to current time
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var form GetMessagesForm
			err := form.GetParams(tt.values)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMessage)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.MessagesCount, form.MessagesCount)

				// For timestamp, we can't compare directly due to potential slight differences
				if tt.values.Get("ts") == "invalid" || !tt.values.Has("ts") {
					assert.WithinDuration(t, time.Now(), form.Ts, time.Second)
				} else {
					assert.Equal(t, tt.expected.Ts.Format(time2.TimeStampLayout), form.Ts.Format(time2.TimeStampLayout))
				}
			}
		})
	}
}

func TestToFileOut(t *testing.T) {
	tests := []struct {
		name     string
		file     models.File
		expected FileOut
	}{
		{
			name: "File with name",
			file: models.File{
				URL:  "http://example.com/file.pdf",
				Name: "document.pdf",
			},
			expected: FileOut{
				URL:  "http://example.com/file.pdf",
				Name: "document.pdf",
			},
		},
		{
			name: "File without name",
			file: models.File{
				URL: "http://example.com/image.jpg",
			},
			expected: FileOut{
				URL: "http://example.com/image.jpg",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToFileOut(tt.file)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestToMessageOut(t *testing.T) {
	messageID := uuid.New()
	chatID := uuid.New()
	senderID := uuid.New()
	now := time.Now()

	userInfo := models.PublicUserInfo{
		Id:        senderID,
		Firstname: "John",
		Lastname:  "Doe",
		AvatarURL: "avatar.jpg",
	}

	tests := []struct {
		name     string
		message  models.Message
		userInfo models.PublicUserInfo
		expected MessageOut
	}{
		{
			name: "Message with all attachments",
			message: models.Message{
				ID:        messageID,
				Text:      "Hello world",
				CreatedAt: now,
				UpdatedAt: now,
				Attachments: []*models.File{
					{URL: "media1.jpg", DisplayType: models.DisplayTypeMedia},
					{URL: "audio1.mp3", DisplayType: models.DisplayTypeAudio},
					{URL: "file1.pdf", DisplayType: models.DisplayTypeFile},
					{URL: "sticker1.png", DisplayType: models.DisplayTypeSticker},
				},
				SenderID: senderID,
				ChatID:   chatID,
			},
			userInfo: userInfo,
			expected: MessageOut{
				ID:          messageID,
				Text:        "Hello world",
				CreatedAt:   now.Format(time2.TimeStampLayout),
				UpdatedAt:   now.Format(time2.TimeStampLayout),
				MediaURLs:   []FileOut{{URL: "media1.jpg"}},
				AudioURLs:   []FileOut{{URL: "audio1.mp3"}},
				FileURLs:    []FileOut{{URL: "file1.pdf"}},
				StickerUrls: []FileOut{{URL: "sticker1.png"}},
				Sender:      PublicUserInfoToOut(userInfo, ""),
				ChatId:      chatID,
			},
		},
		{
			name: "Message with text only",
			message: models.Message{
				ID:        messageID,
				Text:      "Simple message",
				CreatedAt: now,
				UpdatedAt: now,
				SenderID:  senderID,
				ChatID:    chatID,
			},
			userInfo: userInfo,
			expected: MessageOut{
				ID:        messageID,
				Text:      "Simple message",
				CreatedAt: now.Format(time2.TimeStampLayout),
				UpdatedAt: now.Format(time2.TimeStampLayout),
				Sender:    PublicUserInfoToOut(userInfo, ""),
				ChatId:    chatID,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToMessageOut(tt.message, tt.userInfo)

			assert.Equal(t, tt.expected.ID, result.ID)
			assert.Equal(t, tt.expected.Text, result.Text)
			assert.Equal(t, tt.expected.CreatedAt, result.CreatedAt)
			assert.Equal(t, tt.expected.UpdatedAt, result.UpdatedAt)
			assert.Equal(t, tt.expected.ChatId, result.ChatId)

			assert.Len(t, result.MediaURLs, len(tt.expected.MediaURLs))
			assert.Len(t, result.AudioURLs, len(tt.expected.AudioURLs))
			assert.Len(t, result.FileURLs, len(tt.expected.FileURLs))
			assert.Len(t, result.StickerUrls, len(tt.expected.StickerUrls))

			// Check sender info
			assert.Equal(t, tt.userInfo.Id.String(), result.Sender.ID)
			assert.Equal(t, tt.userInfo.Firstname, result.Sender.FirstName)
			assert.Equal(t, tt.userInfo.Lastname, result.Sender.LastName)
			assert.Equal(t, tt.userInfo.AvatarURL, result.Sender.AvatarURL)
		})
	}
}

func TestToMessagesOut(t *testing.T) {
	message1ID := uuid.New()
	message2ID := uuid.New()
	chatID := uuid.New()
	sender1ID := uuid.New()
	sender2ID := uuid.New()
	now := time.Now()

	userInfoMap := map[uuid.UUID]models.PublicUserInfo{
		sender1ID: {
			Id:        sender1ID,
			Firstname: "John",
			Lastname:  "Doe",
		},
		sender2ID: {
			Id:        sender2ID,
			Firstname: "Jane",
			Lastname:  "Smith",
		},
	}

	messages := []*models.Message{
		{
			ID:        message1ID,
			Text:      "First message",
			CreatedAt: now,
			UpdatedAt: now,
			SenderID:  sender1ID,
			ChatID:    chatID,
		},
		{
			ID:        message2ID,
			Text:      "Second message",
			CreatedAt: now.Add(time.Minute),
			UpdatedAt: now.Add(time.Minute),
			SenderID:  sender2ID,
			ChatID:    chatID,
		},
	}

	result := ToMessagesOut(messages, userInfoMap)

	assert.Len(t, result, 2)
	assert.Equal(t, "First message", result[0].Text)
	assert.Equal(t, "John", result[0].Sender.FirstName)
	assert.Equal(t, "Second message", result[1].Text)
	assert.Equal(t, "Jane", result[1].Sender.FirstName)
}

func TestMessageForm_ToMessageModel(t *testing.T) {
	chatID := uuid.New()
	receiverID := uuid.New()
	senderID := uuid.New()

	tests := []struct {
		name     string
		form     MessageForm
		expected models.Message
	}{
		{
			name: "Full message form",
			form: MessageForm{
				Text:       "Hello",
				ChatId:     chatID,
				Media:      []string{"media1.jpg", "media2.jpg"},
				Audio:      []string{"audio1.mp3"},
				File:       []string{"file1.pdf"},
				Stickers:   []string{"sticker1.png"},
				ReceiverId: receiverID,
				SenderId:   senderID,
			},
			expected: models.Message{
				Text:       "Hello",
				ChatID:     chatID,
				ReceiverID: receiverID,
				SenderID:   senderID,
			},
		},
		{
			name: "Text only message",
			form: MessageForm{
				Text:     "Simple",
				ChatId:   chatID,
				SenderId: senderID,
			},
			expected: models.Message{
				Text:     "Simple",
				ChatID:   chatID,
				SenderID: senderID,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.form.ToMessageModel()

			assert.NotEqual(t, uuid.Nil, result.ID)
			assert.Equal(t, tt.expected.Text, result.Text)
			assert.Equal(t, tt.expected.ChatID, result.ChatID)
			assert.Equal(t, tt.expected.ReceiverID, result.ReceiverID)
			assert.Equal(t, tt.expected.SenderID, result.SenderID)
			assert.WithinDuration(t, time.Now(), result.CreatedAt, time.Second)
			assert.WithinDuration(t, time.Now(), result.UpdatedAt, time.Second)

			// Check attachments
			expectedAttachments := len(tt.form.Media) + len(tt.form.Audio) + len(tt.form.File) + len(tt.form.Stickers)
			assert.Len(t, result.Attachments, expectedAttachments)
		})
	}
}
