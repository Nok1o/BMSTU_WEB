package forms

import (
	"time"

	"github.com/google/uuid"

	time_config "quickflow/config/time"
	"quickflow/shared/models"
)

//easyjson:json
type StickerPackForm struct {
	Name     string   `json:"name"`
	Stickers []string `json:"stickers"`
}

func (s *StickerPackForm) ToStickerPackModel(creatorId uuid.UUID) *models.StickerPack {
	var stickers []*models.File
	for _, sticker := range s.Stickers {
		stickers = append(stickers, &models.File{
			URL:         sticker,
			DisplayType: models.DisplayTypeSticker,
		})
	}

	return &models.StickerPack{
		Id:        uuid.New(),
		Name:      s.Name,
		CreatorId: creatorId,
		Stickers:  stickers,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

//easyjson:json
type StickerPackOut struct {
	Id        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatorId uuid.UUID `json:"creator_id"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
	Stickers  []string  `json:"stickers"`
}

func ToStickerPackOut(stickerPack *models.StickerPack) StickerPackOut {
	var stickers []string
	for _, sticker := range stickerPack.Stickers {
		stickers = append(stickers, sticker.URL)
	}

	return StickerPackOut{
		Id:        stickerPack.Id,
		Name:      stickerPack.Name,
		CreatorId: stickerPack.CreatorId,
		CreatedAt: stickerPack.CreatedAt.Format(time_config.TimeStampLayout),
		UpdatedAt: stickerPack.UpdatedAt.Format(time_config.TimeStampLayout),
		Stickers:  stickers,
	}
}

func ToStickerPacksOut(stickerPacks []*models.StickerPack) []StickerPackOut {
	if stickerPacks == nil {
		return nil
	}
	res := make([]StickerPackOut, len(stickerPacks))
	for i, stickerPack := range stickerPacks {
		res[i] = ToStickerPackOut(stickerPack)
	}
	return res
}
