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

func TestMapStickerPackToProto(t *testing.T) {
	now := time.Now()
	stickerPackID := uuid.New()
	creatorID := uuid.New()

	tests := []struct {
		name     string
		input    *models.StickerPack
		expected *pb.StickerPack
	}{
		{
			name: "full sticker pack with all fields",
			input: &models.StickerPack{
				Id:        stickerPackID,
				Name:      "Test Pack",
				CreatorId: creatorID,
				CreatedAt: now,
				UpdatedAt: now,
				Stickers: []*models.File{
					{Name: "sticker1.png"},
					{Name: "sticker2.png"},
				},
			},
			expected: &pb.StickerPack{
				Id:        stickerPackID.String(),
				Name:      "Test Pack",
				CreatorId: creatorID.String(),
				CreatedAt: timestamppb.New(now),
				UpdatedAt: timestamppb.New(now),
				Stickers: []*proto.File{
					{FileName: "sticker1.png"},
					{FileName: "sticker2.png"},
				},
			},
		},
		{
			name: "minimal sticker pack with required fields only",
			input: &models.StickerPack{
				Id:        stickerPackID,
				Name:      "Minimal Pack",
				CreatorId: creatorID,
				CreatedAt: now,
				UpdatedAt: now,
			},
			expected: &pb.StickerPack{
				Id:        stickerPackID.String(),
				Name:      "Minimal Pack",
				CreatorId: creatorID.String(),
				CreatedAt: timestamppb.New(now),
				UpdatedAt: timestamppb.New(now),
			},
		},
		{
			name:     "nil input",
			input:    nil,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MapStickerPackToProto(tt.input)
			if tt.expected == nil {
				assert.Nil(t, result)
				return
			}

			assert.Equal(t, tt.expected.Id, result.Id)
			assert.Equal(t, tt.expected.Name, result.Name)
			assert.Equal(t, tt.expected.CreatorId, result.CreatorId)
			assert.Equal(t, tt.expected.CreatedAt.AsTime(), result.CreatedAt.AsTime())
			assert.Equal(t, tt.expected.UpdatedAt.AsTime(), result.UpdatedAt.AsTime())

			if tt.input.Stickers != nil {
				assert.Len(t, result.Stickers, len(tt.input.Stickers))
				for i, sticker := range tt.input.Stickers {
					if result.Stickers[i] != nil {
						assert.Equal(t, sticker.Name, result.Stickers[i].FileName)
					}

				}
			} else {
				assert.Nil(t, result.Stickers)
			}
		})
	}
}

func TestMapStickerPacksToProto(t *testing.T) {
	now := time.Now()
	stickerPackID1 := uuid.New()
	stickerPackID2 := uuid.New()
	creatorID := uuid.New()

	tests := []struct {
		name     string
		input    []*models.StickerPack
		expected []*pb.StickerPack
	}{
		{
			name: "multiple sticker packs",
			input: []*models.StickerPack{
				{
					Id:        stickerPackID1,
					Name:      "Pack 1",
					CreatorId: creatorID,
					CreatedAt: now,
					UpdatedAt: now,
				},
				{
					Id:        stickerPackID2,
					Name:      "Pack 2",
					CreatorId: creatorID,
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
			expected: []*pb.StickerPack{
				{
					Id:        stickerPackID1.String(),
					Name:      "Pack 1",
					CreatorId: creatorID.String(),
					CreatedAt: timestamppb.New(now),
					UpdatedAt: timestamppb.New(now),
				},
				{
					Id:        stickerPackID2.String(),
					Name:      "Pack 2",
					CreatorId: creatorID.String(),
					CreatedAt: timestamppb.New(now),
					UpdatedAt: timestamppb.New(now),
				},
			},
		},
		{
			name:     "empty slice",
			input:    []*models.StickerPack{},
			expected: []*pb.StickerPack{},
		},
		{
			name:     "nil input",
			input:    nil,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MapStickerPacksToProto(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMapProtoToStickerPack(t *testing.T) {
	now := time.Now()
	stickerPackID := uuid.New()
	creatorID := uuid.New()

	tests := []struct {
		name     string
		input    *pb.StickerPack
		expected *models.StickerPack
		wantErr  bool
	}{
		{
			name: "full proto sticker pack",
			input: &pb.StickerPack{
				Id:        stickerPackID.String(),
				Name:      "Test Pack",
				CreatorId: creatorID.String(),
				CreatedAt: timestamppb.New(now),
				UpdatedAt: timestamppb.New(now),
				Stickers: []*proto.File{
					{FileName: "sticker1.png"},
					{FileName: "sticker2.png"},
				},
			},
			expected: &models.StickerPack{
				Id:        stickerPackID,
				Name:      "Test Pack",
				CreatorId: creatorID,
				CreatedAt: now,
				UpdatedAt: now,
				Stickers: []*models.File{
					{Name: "sticker1.png"},
					{Name: "sticker2.png"},
				},
			},
			wantErr: false,
		},
		{
			name: "minimal proto sticker pack",
			input: &pb.StickerPack{
				Id:        stickerPackID.String(),
				Name:      "Minimal Pack",
				CreatorId: creatorID.String(),
				CreatedAt: timestamppb.New(now),
				UpdatedAt: timestamppb.New(now),
			},
			expected: &models.StickerPack{
				Id:        stickerPackID,
				Name:      "Minimal Pack",
				CreatorId: creatorID,
				CreatedAt: now,
				UpdatedAt: now,
			},
			wantErr: false,
		},
		{
			name: "invalid pack UUID",
			input: &pb.StickerPack{
				Id:        "invalid-uuid",
				Name:      "Invalid Pack",
				CreatorId: creatorID.String(),
				CreatedAt: timestamppb.New(now),
				UpdatedAt: timestamppb.New(now),
			},
			expected: nil,
			wantErr:  true,
		},
		{
			name: "invalid creator UUID",
			input: &pb.StickerPack{
				Id:        stickerPackID.String(),
				Name:      "Invalid Creator",
				CreatorId: "invalid-uuid",
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
			result, err := MapProtoToStickerPack(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				if tt.expected == nil {
					assert.Nil(t, result)
					return
				}

				assert.Equal(t, tt.expected.Id, result.Id)
				assert.Equal(t, tt.expected.Name, result.Name)
				assert.Equal(t, tt.expected.CreatorId, result.CreatorId)

				if tt.input.Stickers != nil {
					assert.Len(t, result.Stickers, len(tt.input.Stickers))
					for i, sticker := range tt.input.Stickers {
						assert.Equal(t, sticker.FileName, result.Stickers[i].Name)
					}
				} else {
					assert.Nil(t, result.Stickers)
				}
			}
		})
	}
}
