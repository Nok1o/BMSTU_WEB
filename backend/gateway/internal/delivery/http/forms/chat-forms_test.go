//go:build unit
// +build unit

package forms

import (
	"errors"
	"net/url"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	time2 "quickflow/config/time"
	"quickflow/shared/models"
)

func TestGetChatsForm_GetParams(t *testing.T) {
	now := time.Now()
	nowStr := now.Format(time2.TimeStampLayout)

	tests := []struct {
		name        string
		values      url.Values
		expected    GetChatsForm
		expectedErr error
	}{
		{
			name: "valid params with timestamp",
			values: url.Values{
				"chats_count": []string{"10"},
				"ts":          []string{nowStr},
			},
			expected: GetChatsForm{
				ChatsCount: 10,
				Ts:         now,
			},
		},
		{
			name: "valid params without timestamp",
			values: url.Values{
				"chats_count": []string{"5"},
			},
			expected: GetChatsForm{
				ChatsCount: 5,
				Ts:         time.Now(), // will be set to current time
			},
		},
		{
			name: "missing chats_count",
			values: url.Values{
				"ts": []string{nowStr},
			},
			expectedErr: errors.New("chats_count parameter missing"),
		},
		{
			name: "invalid chats_count format",
			values: url.Values{
				"chats_count": []string{"invalid"},
				"ts":          []string{nowStr},
			},
			expectedErr: errors.New("failed to parse chats_count"),
		},
		{
			name: "invalid timestamp format",
			values: url.Values{
				"chats_count": []string{"10"},
				"ts":          []string{"invalid"},
			},
			expected: GetChatsForm{
				ChatsCount: 10,
				Ts:         time.Now(), // will be set to current time
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var form GetChatsForm
			err := form.GetParams(tt.values)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.ChatsCount, form.ChatsCount)

				// For timestamp, we can't compare directly due to possible minor time differences
				if tt.values.Has("ts") && tt.values.Get("ts") != "invalid" {
					assert.Equal(t, tt.expected.Ts.Format(time2.TimeStampLayout), form.Ts.Format(time2.TimeStampLayout))
				} else {
					// Just check it's set to current time (with some leeway)
					assert.WithinDuration(t, time.Now(), form.Ts, time.Second)
				}
			}
		})
	}
}

func TestToChatOut(t *testing.T) {
	chatID := uuid.New()
	messageID := uuid.New()
	senderID := uuid.New()
	createdAt := time.Now().Add(-time.Hour)
	updatedAt := time.Now().Add(-30 * time.Minute)
	lastReadByOther := time.Now().Add(-15 * time.Minute)
	lastReadByMe := time.Now().Add(-10 * time.Minute)

	tests := []struct {
		name                    string
		chat                    models.Chat
		lastMessageSenderInfo   models.PublicUserInfo
		privateChatOnlineStatus *PrivateChatInfo
		expected                ChatOut
	}{
		{
			name: "private chat with last message and online status",
			chat: models.Chat{
				ID:        chatID,
				Name:      "Private Chat",
				Type:      models.ChatTypePrivate,
				CreatedAt: createdAt,
				UpdatedAt: updatedAt,
				AvatarURL: "http://example.com/avatar.jpg",
				LastMessage: models.Message{
					ID:        messageID,
					SenderID:  senderID,
					Text:      "Hello!",
					CreatedAt: updatedAt,
				},
				LastReadByOther: &lastReadByOther,
				LastReadByMe:    &lastReadByMe,
			},
			lastMessageSenderInfo: models.PublicUserInfo{
				Id:       senderID,
				Username: "sender_user",
			},
			privateChatOnlineStatus: &PrivateChatInfo{
				Username: "other_user",
				Activity: Activity{
					IsOnline: true,
					LastSeen: "",
				},
			},
			expected: ChatOut{
				ID:                chatID.String(),
				Name:              "Private Chat",
				CreatedAt:         createdAt.Format(time2.TimeStampLayout),
				UpdatedAt:         updatedAt.Format(time2.TimeStampLayout),
				AvatarURL:         "http://example.com/avatar.jpg",
				Type:              "private",
				LastReadByOther:   lastReadByOther.Format(time2.TimeStampLayout),
				LastReadByMe:      lastReadByMe.Format(time2.TimeStampLayout),
				IsOnline:          func(b bool) *bool { return &b }(true),
				Username:          "other_user",
				NumUnreadMessages: 0,
				LastMessage: &MessageOut{
					ID:        messageID,
					Text:      "Hello!",
					CreatedAt: updatedAt.Format(time2.TimeStampLayout),
					Sender: PublicUserInfoOut{
						ID:       senderID.String(),
						Username: "sender_user",
					},
				},
			},
		},
		{
			name: "group chat without last message",
			chat: models.Chat{
				ID:        chatID,
				Name:      "Group Chat",
				Type:      models.ChatTypeGroup,
				CreatedAt: createdAt,
				UpdatedAt: updatedAt,
				AvatarURL: "http://example.com/group.jpg",
			},
			expected: ChatOut{
				ID:                chatID.String(),
				Name:              "Group Chat",
				CreatedAt:         createdAt.Format(time2.TimeStampLayout),
				UpdatedAt:         updatedAt.Format(time2.TimeStampLayout),
				AvatarURL:         "http://example.com/group.jpg",
				Type:              "group",
				NumUnreadMessages: 0,
			},
		},
		{
			name: "private chat offline with last seen",
			chat: models.Chat{
				ID:        chatID,
				Type:      models.ChatTypePrivate,
				CreatedAt: createdAt,
				UpdatedAt: updatedAt,
			},
			privateChatOnlineStatus: &PrivateChatInfo{
				Username: "offline_user",
				Activity: Activity{
					IsOnline: false,
					LastSeen: "2 minutes ago",
				},
			},
			expected: ChatOut{
				ID:                chatID.String(),
				CreatedAt:         createdAt.Format(time2.TimeStampLayout),
				UpdatedAt:         updatedAt.Format(time2.TimeStampLayout),
				Type:              "private",
				IsOnline:          func(b bool) *bool { return &b }(false),
				LastSeen:          "2 minutes ago",
				Username:          "offline_user",
				NumUnreadMessages: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToChatOut(tt.chat, tt.lastMessageSenderInfo, tt.privateChatOnlineStatus)
			assert.Equal(t, tt.expected.ID, result.ID)
		})
	}
}

func TestToChatsOut(t *testing.T) {
	chatID1 := uuid.New()
	chatID2 := uuid.New()
	senderID := uuid.New()
	now := time.Now()

	privateChat := models.Chat{
		ID:        chatID1,
		Name:      "Private Chat",
		Type:      models.ChatTypePrivate,
		CreatedAt: now.Add(-time.Hour),
		UpdatedAt: now.Add(-30 * time.Minute),
		LastMessage: models.Message{
			ID:        uuid.New(),
			SenderID:  senderID,
			Text:      "Private message",
			CreatedAt: now.Add(-25 * time.Minute),
		},
	}

	groupChat := models.Chat{
		ID:        chatID2,
		Name:      "Group Chat",
		Type:      models.ChatTypeGroup,
		CreatedAt: now.Add(-2 * time.Hour),
		UpdatedAt: now.Add(-15 * time.Minute),
		LastMessage: models.Message{
			ID:        uuid.New(),
			SenderID:  senderID,
			Text:      "Group message",
			CreatedAt: now.Add(-20 * time.Minute),
		},
	}

	lastMessageSenderInfo := map[uuid.UUID]models.PublicUserInfo{
		senderID: {
			Id:       senderID,
			Username: "sender_user",
		},
	}

	privateChatsOnlineStatus := map[uuid.UUID]PrivateChatInfo{
		chatID1: {
			Username: "private_user",
			Activity: Activity{
				IsOnline: true,
			},
		},
	}

	t.Run("multiple chats conversion", func(t *testing.T) {
		chats := []models.Chat{privateChat, groupChat}
		result := ToChatsOut(chats, lastMessageSenderInfo, privateChatsOnlineStatus)

		assert.Len(t, result, 2)

		// Check private chat
		assert.Equal(t, chatID1.String(), result[0].ID)
		assert.Equal(t, "private", result[0].Type)
		assert.NotNil(t, result[0].IsOnline)
		assert.True(t, *result[0].IsOnline)
		assert.Equal(t, "private_user", result[0].Username)
		assert.NotNil(t, result[0].LastMessage)
		assert.Equal(t, "Private message", result[0].LastMessage.Text)

		// Check group chat
		assert.Equal(t, chatID2.String(), result[1].ID)
		assert.Equal(t, "group", result[1].Type)
		assert.Nil(t, result[1].IsOnline)
		assert.Equal(t, "", result[1].Username)
		assert.NotNil(t, result[1].LastMessage)
		assert.Equal(t, "Group message", result[1].LastMessage.Text)
	})

	t.Run("empty chats slice", func(t *testing.T) {
		result := ToChatsOut([]models.Chat{}, lastMessageSenderInfo, privateChatsOnlineStatus)
		assert.Empty(t, result)
	})

	t.Run("nil lastMessageSenderInfo", func(t *testing.T) {
		chats := []models.Chat{privateChat, groupChat}
		result := ToChatsOut(chats, nil, privateChatsOnlineStatus)

		assert.Len(t, result, 2)
		assert.NotNil(t, result[0].LastMessage)
		assert.Equal(t, "Private message", result[0].LastMessage.Text)
		// Sender info should be empty
		assert.Equal(t, "", result[0].LastMessage.Sender.Username)
	})

	t.Run("nil privateChatsOnlineStatus", func(t *testing.T) {
		chats := []models.Chat{privateChat, groupChat}
		result := ToChatsOut(chats, lastMessageSenderInfo, nil)

		assert.Len(t, result, 2)
		// Private chat online status should be nil
		assert.Nil(t, result[0].IsOnline)
		assert.Equal(t, "", result[0].Username)
	})
}

func TestChatOut_JSON(t *testing.T) {
	chatID := uuid.New()
	now := time.Now()

	chat := ChatOut{
		ID:        chatID.String(),
		Name:      "Test Chat",
		CreatedAt: now.Format(time2.TimeStampLayout),
		Type:      "private",
	}

	t.Run("marshal and unmarshal", func(t *testing.T) {
		// Marshal
		jsonData, err := chat.MarshalJSON()
		assert.NoError(t, err)

		// Unmarshal
		var newChat ChatOut
		err = newChat.UnmarshalJSON(jsonData)
		assert.NoError(t, err)

		assert.Equal(t, chat, newChat)
	})
}

func TestChatsOut_JSON(t *testing.T) {
	chat1 := ChatOut{
		ID:   uuid.New().String(),
		Name: "Chat 1",
		Type: "private",
	}
	chat2 := ChatOut{
		ID:   uuid.New().String(),
		Name: "Chat 2",
		Type: "group",
	}

	chats := ChatsOut{chat1, chat2}

	t.Run("marshal and unmarshal", func(t *testing.T) {
		// Marshal
		jsonData, err := chats.MarshalJSON()
		assert.NoError(t, err)

		// Unmarshal
		var newChats ChatsOut
		err = newChats.UnmarshalJSON(jsonData)
		assert.NoError(t, err)

		assert.Equal(t, chats, newChats)
	})
}

func TestGetNumUnreadChatsForm_JSON(t *testing.T) {
	form := GetNumUnreadChatsForm{
		ChatsCount: 10,
	}

	t.Run("marshal and unmarshal", func(t *testing.T) {
		// Marshal
		jsonData, err := form.MarshalJSON()
		assert.NoError(t, err)

		// Unmarshal
		var newForm GetNumUnreadChatsForm
		err = newForm.UnmarshalJSON(jsonData)
		assert.NoError(t, err)

		assert.Equal(t, form, newForm)
	})
}

func TestPrivateChatInfo_JSON(t *testing.T) {
	info := PrivateChatInfo{
		Username: "test_user",
		Activity: Activity{
			IsOnline: true,
			LastSeen: "now",
		},
	}

	t.Run("marshal and unmarshal", func(t *testing.T) {
		// Marshal
		jsonData, err := info.MarshalJSON()
		assert.NoError(t, err)

		// Unmarshal
		var newInfo PrivateChatInfo
		err = newInfo.UnmarshalJSON(jsonData)
		assert.NoError(t, err)

		assert.Equal(t, info, newInfo)
	})
}
