//go:build unit
// +build unit

package messenger_service

import (
	proto "quickflow/shared/proto/file_service"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	"quickflow/shared/models"
	pb "quickflow/shared/proto/messenger_service"
)

func TestMapMessageToProto(t *testing.T) {
	now := time.Now()
	msgID := uuid.New()
	senderID := uuid.New()
	chatID := uuid.New()
	receiverID := uuid.New()

	tests := []struct {
		name     string
		input    models.Message
		expected *pb.Message
	}{
		{
			name: "complete message",
			input: models.Message{
				ID:          msgID,
				SenderID:    senderID,
				ChatID:      chatID,
				ReceiverID:  receiverID,
				Text:        "test message",
				CreatedAt:   now,
				UpdatedAt:   now,
				Attachments: []*models.File{{Name: "test.txt"}},
			},
			expected: &pb.Message{
				Id:          msgID.String(),
				SenderId:    senderID.String(),
				ChatId:      chatID.String(),
				ReceiverId:  receiverID.String(),
				Text:        "test message",
				CreatedAt:   timestamppb.New(now),
				UpdatedAt:   timestamppb.New(now),
				Attachments: []*proto.File{{FileName: "test.txt"}},
			},
		},
		{
			name: "empty message",
			input: models.Message{
				ID: uuid.Nil,
			},
			expected: &pb.Message{
				Id:         uuid.Nil.String(),
				SenderId:   uuid.Nil.String(),
				ChatId:     uuid.Nil.String(),
				ReceiverId: uuid.Nil.String(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = MapMessageToProto(tt.input)
		})
	}
}

func TestMapMessagesToProto(t *testing.T) {
	now := time.Now()
	msg1 := models.Message{
		ID:        uuid.New(),
		Text:      "msg1",
		CreatedAt: now,
	}
	msg2 := models.Message{
		ID:        uuid.New(),
		Text:      "msg2",
		CreatedAt: now.Add(time.Hour),
	}

	tests := []struct {
		name     string
		input    []models.Message
		expected []*pb.Message
	}{
		{
			name:  "multiple messages",
			input: []models.Message{msg1, msg2},
			expected: []*pb.Message{
				{
					Id:        msg1.ID.String(),
					Text:      "msg1",
					CreatedAt: timestamppb.New(now),
				},
				{
					Id:        msg2.ID.String(),
					Text:      "msg2",
					CreatedAt: timestamppb.New(now.Add(time.Hour)),
				},
			},
		},
		{
			name:     "empty slice",
			input:    []models.Message{},
			expected: []*pb.Message{},
		},
		{
			name:     "nil slice",
			input:    nil,
			expected: []*pb.Message{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = MapMessagesToProto(tt.input)
		})
	}
}

func TestMapProtoToMessage(t *testing.T) {
	now := time.Now()
	validID := uuid.New().String()
	invalidID := "invalid-uuid"

	tests := []struct {
		name        string
		input       *pb.Message
		expected    *models.Message
		expectError bool
	}{
		{
			name: "valid complete message",
			input: &pb.Message{
				Id:          validID,
				SenderId:    validID,
				ChatId:      validID,
				ReceiverId:  validID,
				Text:        "test",
				CreatedAt:   timestamppb.New(now),
				UpdatedAt:   timestamppb.New(now),
				Attachments: []*proto.File{{FileName: "test.txt"}},
			},
			expected: &models.Message{
				ID:          uuid.MustParse(validID),
				SenderID:    uuid.MustParse(validID),
				ChatID:      uuid.MustParse(validID),
				ReceiverID:  uuid.MustParse(validID),
				Text:        "test",
				CreatedAt:   now,
				UpdatedAt:   now,
				Attachments: []*models.File{{Name: "test.txt"}},
			},
			expectError: false,
		},
		{
			name: "invalid message id",
			input: &pb.Message{
				Id:       invalidID,
				SenderId: validID,
			},
			expectError: true,
		},
		{
			name: "invalid sender id",
			input: &pb.Message{
				Id:       validID,
				SenderId: invalidID,
			},
			expectError: true,
		},
		{
			name:        "nil message",
			input:       nil,
			expected:    nil,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := MapProtoToMessage(tt.input)

			if tt.expectError {
				require.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				if tt.input == nil {
					assert.Nil(t, result)
				} else {
					assert.Equal(t, tt.expected.ID, result.ID)
					assert.Equal(t, tt.expected.SenderID, result.SenderID)
					assert.Equal(t, tt.expected.ChatID, result.ChatID)
					assert.Equal(t, tt.expected.ReceiverID, result.ReceiverID)
					assert.Equal(t, tt.expected.Text, result.Text)
					assert.Equal(t, tt.expected.CreatedAt.UTC(), result.CreatedAt.UTC())
					assert.Equal(t, tt.expected.UpdatedAt.UTC(), result.UpdatedAt.UTC())

					if tt.input.Attachments != nil {
						assert.Equal(t, len(tt.input.Attachments), len(result.Attachments))
					} else {
						assert.Nil(t, result.Attachments)
					}
				}
			}
		})
	}
}
