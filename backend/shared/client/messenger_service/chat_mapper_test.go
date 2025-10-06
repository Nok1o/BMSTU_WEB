//go:build unit
// +build unit

package messenger_service

import (
	proto "quickflow/shared/proto/file_service"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"

	"quickflow/shared/models"
	pb "quickflow/shared/proto/messenger_service"
)

func TestMapChatToProto(t *testing.T) {
	now := time.Now()
	chatID := uuid.New()
	messageID := uuid.New()

	tests := []struct {
		name     string
		input    models.Chat
		expected *pb.Chat
	}{
		{
			name: "full chat with all fields",
			input: models.Chat{
				ID:        chatID,
				Name:      "Test Chat",
				Type:      models.ChatTypeGroup,
				AvatarURL: "http://example.com/avatar.jpg",
				CreatedAt: now,
				UpdatedAt: now,
				LastMessage: models.Message{
					ID: messageID,
				},
				LastReadByOther: &now,
				LastReadByMe:    &now,
			},
			expected: &pb.Chat{
				Id:               chatID.String(),
				Name:             "Test Chat",
				Type:             pb.ChatType_CHAT_TYPE_GROUP,
				AvatarUrl:        "http://example.com/avatar.jpg",
				CreatedAt:        timestamppb.New(now),
				UpdatedAt:        timestamppb.New(now),
				LastReadByOthers: timestamppb.New(now),
				LastReadByMe:     timestamppb.New(now),
				LastMessage: &pb.Message{
					Id: messageID.String(),
				},
			},
		},
		{
			name: "minimal chat with required fields only",
			input: models.Chat{
				ID:        chatID,
				Name:      "Minimal Chat",
				Type:      models.ChatTypePrivate,
				CreatedAt: now,
				UpdatedAt: now,
			},
			expected: &pb.Chat{
				Id:        chatID.String(),
				Name:      "Minimal Chat",
				Type:      pb.ChatType_CHAT_TYPE_PRIVATE,
				CreatedAt: timestamppb.New(now),
				UpdatedAt: timestamppb.New(now),
			},
		},
		{
			name: "chat with nil timestamps",
			input: models.Chat{
				ID:              chatID,
				Name:            "No Timestamps Chat",
				Type:            models.ChatTypePrivate,
				CreatedAt:       now,
				UpdatedAt:       now,
				LastReadByOther: nil,
				LastReadByMe:    nil,
			},
			expected: &pb.Chat{
				Id:        chatID.String(),
				Name:      "No Timestamps Chat",
				Type:      pb.ChatType_CHAT_TYPE_PRIVATE,
				CreatedAt: timestamppb.New(now),
				UpdatedAt: timestamppb.New(now),
			},
		},
		{
			name: "chat with empty last message",
			input: models.Chat{
				ID:          chatID,
				Name:        "Empty Last Message",
				Type:        models.ChatTypePrivate,
				CreatedAt:   now,
				UpdatedAt:   now,
				LastMessage: models.Message{ID: uuid.Nil},
			},
			expected: &pb.Chat{
				Id:        chatID.String(),
				Name:      "Empty Last Message",
				Type:      pb.ChatType_CHAT_TYPE_PRIVATE,
				CreatedAt: timestamppb.New(now),
				UpdatedAt: timestamppb.New(now),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = MapChatToProto(tt.input)
		})
	}
}

func TestMapChatsToProto(t *testing.T) {
	now := time.Now()
	chatID1 := uuid.New()
	chatID2 := uuid.New()

	tests := []struct {
		name     string
		input    []models.Chat
		expected []*pb.Chat
	}{
		{
			name: "multiple chats",
			input: []models.Chat{
				{
					ID:        chatID1,
					Name:      "Chat 1",
					Type:      models.ChatTypePrivate,
					CreatedAt: now,
					UpdatedAt: now,
				},
				{
					ID:        chatID2,
					Name:      "Chat 2",
					Type:      models.ChatTypeGroup,
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
			expected: []*pb.Chat{
				{
					Id:        chatID1.String(),
					Name:      "Chat 1",
					Type:      pb.ChatType_CHAT_TYPE_PRIVATE,
					CreatedAt: timestamppb.New(now),
					UpdatedAt: timestamppb.New(now),
				},
				{
					Id:        chatID2.String(),
					Name:      "Chat 2",
					Type:      pb.ChatType_CHAT_TYPE_GROUP,
					CreatedAt: timestamppb.New(now),
					UpdatedAt: timestamppb.New(now),
				},
			},
		},
		{
			name:     "empty slice",
			input:    []models.Chat{},
			expected: []*pb.Chat{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MapChatsToProto(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMapProtoToChat(t *testing.T) {
	now := time.Now()
	chatID := uuid.New()
	messageID := uuid.New()

	tests := []struct {
		name     string
		input    *pb.Chat
		expected *models.Chat
		wantErr  bool
	}{
		{
			name: "full proto chat",
			input: &pb.Chat{
				Id:               chatID.String(),
				Name:             "Test Chat",
				Type:             pb.ChatType_CHAT_TYPE_GROUP,
				AvatarUrl:        "http://example.com/avatar.jpg",
				CreatedAt:        timestamppb.New(now),
				UpdatedAt:        timestamppb.New(now),
				LastReadByOthers: timestamppb.New(now),
				LastReadByMe:     timestamppb.New(now),
				LastMessage: &pb.Message{
					Id: messageID.String(),
				},
			},
			expected: &models.Chat{
				ID:              chatID,
				Name:            "Test Chat",
				Type:            models.ChatTypeGroup,
				AvatarURL:       "http://example.com/avatar.jpg",
				CreatedAt:       now,
				UpdatedAt:       now,
				LastReadByOther: &now,
				LastReadByMe:    &now,
				LastMessage: models.Message{
					ID: messageID,
				},
			},
			wantErr: false,
		},
		{
			name: "minimal proto chat",
			input: &pb.Chat{
				Id:        chatID.String(),
				Name:      "Minimal Chat",
				Type:      pb.ChatType_CHAT_TYPE_PRIVATE,
				CreatedAt: timestamppb.New(now),
				UpdatedAt: timestamppb.New(now),
			},
			expected: &models.Chat{
				ID:        chatID,
				Name:      "Minimal Chat",
				Type:      models.ChatTypePrivate,
				CreatedAt: now,
				UpdatedAt: now,
			},
			wantErr: false,
		},
		{
			name: "nil timestamps",
			input: &pb.Chat{
				Id:               chatID.String(),
				Name:             "No Timestamps",
				Type:             pb.ChatType_CHAT_TYPE_PRIVATE,
				CreatedAt:        timestamppb.New(now),
				UpdatedAt:        timestamppb.New(now),
				LastReadByOthers: nil,
				LastReadByMe:     nil,
			},
			expected: &models.Chat{
				ID:              chatID,
				Name:            "No Timestamps",
				Type:            models.ChatTypePrivate,
				CreatedAt:       now,
				UpdatedAt:       now,
				LastReadByOther: nil,
				LastReadByMe:    nil,
			},
			wantErr: false,
		},
		{
			name: "invalid UUID",
			input: &pb.Chat{
				Id:        "invalid-uuid",
				Name:      "Invalid UUID",
				Type:      pb.ChatType_CHAT_TYPE_PRIVATE,
				CreatedAt: timestamppb.New(now),
				UpdatedAt: timestamppb.New(now),
			},
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "nil input",
			input:    nil,
			expected: nil,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MapProtoToChat(tt.input)
			if tt.wantErr {
				assert.Nil(t, result)
			}
		})
	}
}

func TestMapProtoToChats(t *testing.T) {
	now := time.Now()
	chatID1 := uuid.New()
	chatID2 := uuid.New()

	tests := []struct {
		name     string
		input    []*pb.Chat
		expected []models.Chat
		wantErr  bool
	}{
		{
			name: "multiple proto chats",
			input: []*pb.Chat{
				{
					Id:        chatID1.String(),
					Name:      "Chat 1",
					Type:      pb.ChatType_CHAT_TYPE_PRIVATE,
					CreatedAt: timestamppb.New(now),
					UpdatedAt: timestamppb.New(now),
				},
				{
					Id:        chatID2.String(),
					Name:      "Chat 2",
					Type:      pb.ChatType_CHAT_TYPE_GROUP,
					CreatedAt: timestamppb.New(now),
					UpdatedAt: timestamppb.New(now),
				},
			},
			expected: []models.Chat{
				{
					ID:        chatID1,
					Name:      "Chat 1",
					Type:      models.ChatTypePrivate,
					CreatedAt: now,
					UpdatedAt: now,
				},
				{
					ID:        chatID2,
					Name:      "Chat 2",
					Type:      models.ChatTypeGroup,
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
			wantErr: false,
		},
		{
			name:     "empty slice",
			input:    []*pb.Chat{},
			expected: []models.Chat{},
			wantErr:  false,
		},
		{
			name: "slice with invalid chat",
			input: []*pb.Chat{
				{
					Id:        "invalid-uuid",
					Name:      "Invalid Chat",
					Type:      pb.ChatType_CHAT_TYPE_PRIVATE,
					CreatedAt: timestamppb.New(now),
					UpdatedAt: timestamppb.New(now),
				},
			},
			expected: []models.Chat{
				{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MapProtoToChats(tt.input)
			if !tt.wantErr {
			} else {
				// For error case, we expect zero values in the result
				for _, chat := range result {
					assert.Equal(t, uuid.Nil, chat.ID)
				}
			}
		})
	}
}

func TestMapProtoCreationInfoToModel(t *testing.T) {
	tests := []struct {
		name     string
		input    *pb.ChatCreationInfo
		expected models.ChatCreationInfo
	}{
		{
			name: "full creation info",
			input: &pb.ChatCreationInfo{
				Name:   "New Chat",
				Type:   pb.ChatType_CHAT_TYPE_GROUP,
				Avatar: &proto.File{FileName: "avatar.jpg"},
			},
			expected: models.ChatCreationInfo{
				Name:   "New Chat",
				Type:   models.ChatTypeGroup,
				Avatar: &models.File{Name: "avatar.jpg"},
			},
		},
		{
			name: "minimal creation info",
			input: &pb.ChatCreationInfo{
				Name: "Minimal Chat",
				Type: pb.ChatType_CHAT_TYPE_PRIVATE,
			},
			expected: models.ChatCreationInfo{
				Name: "Minimal Chat",
				Type: models.ChatTypePrivate,
			},
		},
		{
			name:     "nil input",
			input:    nil,
			expected: models.ChatCreationInfo{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = MapProtoCreationInfoToModel(tt.input)
		})
	}
}
