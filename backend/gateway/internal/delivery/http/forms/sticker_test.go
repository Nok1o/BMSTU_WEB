//go:build unit
// +build unit

package forms

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	time_config "quickflow/config/time"
	"quickflow/shared/models"
)

func TestToStickerPackModel(t *testing.T) {
	// Генерация тестовых данных
	creatorId := uuid.New()
	stickerPackForm := &StickerPackForm{
		Name: "Test Sticker Pack",
		Stickers: []string{
			"sticker1_url",
			"sticker2_url",
		},
	}

	// Преобразование в модель
	stickerPack := stickerPackForm.ToStickerPackModel(creatorId)

	// Проверка, что все поля корректно установлены
	assert.NotNil(t, stickerPack)
	assert.Equal(t, stickerPack.Name, stickerPackForm.Name)
	assert.Equal(t, len(stickerPack.Stickers), len(stickerPackForm.Stickers))
	assert.Equal(t, stickerPack.CreatorId, creatorId)
	assert.WithinDuration(t, time.Now(), stickerPack.CreatedAt, time.Second)
	assert.WithinDuration(t, time.Now(), stickerPack.UpdatedAt, time.Second)

	// Проверка на то, что URL стикеров установлены корректно
	assert.Equal(t, stickerPack.Stickers[0].URL, stickerPackForm.Stickers[0])
	assert.Equal(t, stickerPack.Stickers[1].URL, stickerPackForm.Stickers[1])
}

func TestToStickerPackOut(t *testing.T) {
	// Генерация тестовых данных
	stickerPack := &models.StickerPack{
		Id:        uuid.New(),
		Name:      "Test Sticker Pack",
		CreatorId: uuid.New(),
		Stickers: []*models.File{
			{URL: "sticker1_url", DisplayType: models.DisplayTypeSticker},
			{URL: "sticker2_url", DisplayType: models.DisplayTypeSticker},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Преобразование в выводную модель
	stickerPackOut := ToStickerPackOut(stickerPack)

	// Проверка, что все поля корректно установлены
	assert.Equal(t, stickerPackOut.Id, stickerPack.Id)
	assert.Equal(t, stickerPackOut.Name, stickerPack.Name)
	assert.Equal(t, stickerPackOut.CreatorId, stickerPack.CreatorId)
	assert.Equal(t, len(stickerPackOut.Stickers), len(stickerPack.Stickers))
	assert.Equal(t, stickerPackOut.Stickers[0], stickerPack.Stickers[0].URL)
	assert.Equal(t, stickerPackOut.Stickers[1], stickerPack.Stickers[1].URL)

	// Проверка формата времени
	createdAt, err := time.Parse(time_config.TimeStampLayout, stickerPackOut.CreatedAt)
	assert.NoError(t, err)
	assert.WithinDuration(t, stickerPack.CreatedAt, createdAt, time.Second)

	updatedAt, err := time.Parse(time_config.TimeStampLayout, stickerPackOut.UpdatedAt)
	assert.NoError(t, err)
	assert.WithinDuration(t, stickerPack.UpdatedAt, updatedAt, time.Second)
}

func TestToStickerPacksOut(t *testing.T) {
	// Генерация тестовых данных
	stickerPacks := []*models.StickerPack{
		{
			Id:        uuid.New(),
			Name:      "Test Sticker Pack 1",
			CreatorId: uuid.New(),
			Stickers: []*models.File{
				{URL: "sticker1_url", DisplayType: models.DisplayTypeSticker},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			Id:        uuid.New(),
			Name:      "Test Sticker Pack 2",
			CreatorId: uuid.New(),
			Stickers: []*models.File{
				{URL: "sticker2_url", DisplayType: models.DisplayTypeSticker},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	// Преобразование в выводные модели
	stickerPacksOut := ToStickerPacksOut(stickerPacks)

	// Проверка, что преобразовано нужное количество элементов
	assert.Equal(t, len(stickerPacksOut), len(stickerPacks))

	// Проверка содержимого каждого элемента
	for i, packOut := range stickerPacksOut {
		assert.Equal(t, packOut.Id, stickerPacks[i].Id)
		assert.Equal(t, packOut.Name, stickerPacks[i].Name)
		assert.Equal(t, packOut.CreatorId, stickerPacks[i].CreatorId)
		assert.Equal(t, len(packOut.Stickers), len(stickerPacks[i].Stickers))
		assert.Equal(t, packOut.Stickers[0], stickerPacks[i].Stickers[0].URL)

		// Проверка времени
		createdAt, err := time.Parse(time_config.TimeStampLayout, packOut.CreatedAt)
		assert.NoError(t, err)
		assert.WithinDuration(t, stickerPacks[i].CreatedAt, createdAt, time.Second)

		updatedAt, err := time.Parse(time_config.TimeStampLayout, packOut.UpdatedAt)
		assert.NoError(t, err)
		assert.WithinDuration(t, stickerPacks[i].UpdatedAt, updatedAt, time.Second)
	}
}
