package models

import (
	"time"

	"github.com/google/uuid"
)

type StickerPack struct {
	Id        uuid.UUID
	Name      string
	Stickers  []*File
	CreatorId uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
}
