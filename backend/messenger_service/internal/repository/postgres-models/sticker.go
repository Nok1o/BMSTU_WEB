package postgres_models

import (
	"github.com/jackc/pgx/v5/pgtype"

	"quickflow/shared/models"
)

type StickerPackPostgres struct {
	ID        pgtype.UUID
	Name      pgtype.Text
	CreatorID pgtype.UUID
	CreatedAt pgtype.Timestamptz
	UpdatedAt pgtype.Timestamptz
	Stickers  []PostgresFile
}

func (sp *StickerPackPostgres) ToStickerPack() models.StickerPack {
	var stickerFiles []*models.File
	for _, sticker := range sp.Stickers {
		stickerFiles = append(stickerFiles, sticker.ToFile())
	}

	return models.StickerPack{
		Id:        sp.ID.Bytes,
		Name:      sp.Name.String,
		Stickers:  stickerFiles,
		CreatorId: sp.CreatorID.Bytes,
		CreatedAt: sp.CreatedAt.Time,
		UpdatedAt: sp.UpdatedAt.Time,
	}
}

func FromStickerPack(stickerPack models.StickerPack) StickerPackPostgres {
	var stickerFiles []PostgresFile
	for _, sticker := range stickerPack.Stickers {
		pgSticker := PostgresFile{
			URL:         pgtype.Text{String: sticker.URL, Valid: true},
			DisplayType: pgtype.Text{String: string(sticker.DisplayType), Valid: true},
			Name:        pgtype.Text{String: sticker.Name, Valid: true},
		}
		stickerFiles = append(stickerFiles, pgSticker)
	}

	return StickerPackPostgres{
		ID:        pgtype.UUID{Bytes: stickerPack.Id, Valid: true},
		Name:      pgtype.Text{String: stickerPack.Name, Valid: true},
		CreatorID: pgtype.UUID{Bytes: stickerPack.CreatorId, Valid: true},
		CreatedAt: pgtype.Timestamptz{Time: stickerPack.CreatedAt, Valid: true},
		UpdatedAt: pgtype.Timestamptz{Time: stickerPack.UpdatedAt, Valid: true},
		Stickers:  stickerFiles,
	}
}
