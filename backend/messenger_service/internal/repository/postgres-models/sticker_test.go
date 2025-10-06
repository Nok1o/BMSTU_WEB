//go:build unit
// +build unit

package postgres_models_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"

	"quickflow/messenger_service/internal/repository/postgres-models"
	"quickflow/shared/models"
)

func TestToStickerPack(t *testing.T) {
	// Тестовые данные
	stickerPackPostgres := postgres_models.StickerPackPostgres{
		ID:        pgtype.UUID{Bytes: uuid.New(), Valid: true},
		Name:      pgtype.Text{String: "Sample Sticker Pack", Valid: true},
		CreatorID: pgtype.UUID{Bytes: uuid.New(), Valid: true},
		CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
		UpdatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
		Stickers: []postgres_models.PostgresFile{
			{URL: pgtype.Text{String: "http://example.com/sticker1.png", Valid: true}},
		},
	}

	// Вызов метода ToStickerPack
	result := stickerPackPostgres.ToStickerPack()

	// Проверка результатов
	assert.Equal(t, stickerPackPostgres.Name.String, result.Name)
	assert.Len(t, result.Stickers, 1)
	assert.Equal(t, "http://example.com/sticker1.png", result.Stickers[0].URL)
}

func TestFromStickerPack(t *testing.T) {
	// Тестовые данные
	stickerPack := models.StickerPack{
		Id:        uuid.New(),
		Name:      "Sample Sticker Pack",
		CreatorId: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Stickers: []*models.File{
			{
				URL:         "http://example.com/sticker1.png",
				DisplayType: models.DisplayTypeSticker,
				Name:        "sticker1",
			},
		},
	}

	// Вызов метода FromStickerPack
	result := postgres_models.FromStickerPack(stickerPack)

	// Проверка результатов
	assert.Equal(t, stickerPack.Name, result.Name.String)
	assert.Len(t, result.Stickers, 1)
	assert.Equal(t, "http://example.com/sticker1.png", result.Stickers[0].URL.String)
	assert.Equal(t, "sticker1", result.Stickers[0].Name.String)
	assert.Equal(t, string(models.DisplayTypeSticker), result.Stickers[0].DisplayType.String)
}
