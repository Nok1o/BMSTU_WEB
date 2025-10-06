package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	messenger_errors "quickflow/messenger_service/internal/errors"
	postgres_models "quickflow/messenger_service/internal/repository/postgres-models"
	"quickflow/shared/models"
)

const (
	insertStickerPackQuery = `
		INSERT INTO sticker_pack (id, name, created_at, updated_at, creator_id)
		VALUES ($1, $2, $3, $4, $5)`

	insertStickerQuery = `
			INSERT INTO sticker (sticker_pack_id, sticker_url)
			VALUES ($1, $2)`

	getStickerPackQuery = `
		SELECT id, name, creator_id, created_at, updated_at
		FROM sticker_pack WHERE id = $1`

	getStickerPackByNameQuery = `
    SELECT id, name, creator_id, created_at, updated_at
    FROM sticker_pack WHERE name = $1
`

	getStickersQuery = `
		SELECT sticker_url
		FROM sticker WHERE sticker_pack_id = $1`

	getStickerPacksQuery = `
		SELECT sp.id, sp.name, sp.creator_id, sp.created_at, sp.updated_at
		FROM sticker_pack sp
		ORDER BY sp.created_at DESC
		LIMIT $1 OFFSET $2`
)

type StickerRepository struct {
	ConnPool *sql.DB
}

func NewPostgresStickerRepository(db *sql.DB) *StickerRepository {
	return &StickerRepository{ConnPool: db}
}

// Close закрывает пул соединений
func (s *StickerRepository) Close() {
	s.ConnPool.Close()
}

func (s *StickerRepository) AddStickerPack(ctx context.Context, stickerPack models.StickerPack) error {
	pgStickerPack := postgres_models.FromStickerPack(stickerPack)
	_, err := s.ConnPool.ExecContext(ctx, insertStickerPackQuery,
		pgStickerPack.ID, pgStickerPack.Name, pgStickerPack.CreatedAt, pgStickerPack.UpdatedAt, pgStickerPack.CreatorID,
	)
	if err != nil {
		return fmt.Errorf("could not insert sticker pack: %v", err)
	}

	for _, sticker := range pgStickerPack.Stickers {
		_, err = s.ConnPool.ExecContext(ctx, insertStickerQuery,
			stickerPack.Id, sticker.URL,
		)
		if err != nil {
			return fmt.Errorf("could not insert sticker: %v", err)
		}
	}

	return nil
}

func (s *StickerRepository) GetStickerPack(ctx context.Context, packId uuid.UUID) (models.StickerPack, error) {
	var stickerPack postgres_models.StickerPackPostgres

	row := s.ConnPool.QueryRowContext(ctx, getStickerPackQuery, packId)

	err := row.Scan(&stickerPack.ID, &stickerPack.Name, &stickerPack.CreatorID, &stickerPack.CreatedAt, &stickerPack.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return models.StickerPack{}, messenger_errors.ErrNotFound
	} else if err != nil {
		return models.StickerPack{}, fmt.Errorf("could not get sticker pack: %v", err)
	}

	rows, err := s.ConnPool.QueryContext(ctx, getStickersQuery, packId)
	if err != nil {
		return models.StickerPack{}, fmt.Errorf("could not get stickers: %v", err)
	}
	defer rows.Close()

	var stickers []postgres_models.PostgresFile
	for rows.Next() {
		var sticker postgres_models.PostgresFile
		err = rows.Scan(&sticker.URL)
		if err != nil {
			return models.StickerPack{}, fmt.Errorf("could not scan sticker: %v", err)
		}
		stickers = append(stickers, sticker)
	}

	stickerPack.Stickers = stickers
	return stickerPack.ToStickerPack(), nil
}

func (s *StickerRepository) GetStickerPacks(ctx context.Context, _ uuid.UUID, count, offset int) ([]models.StickerPack, error) {
	var stickerPacks []models.StickerPack

	rows, err := s.ConnPool.QueryContext(ctx, getStickerPacksQuery, count, offset)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, messenger_errors.ErrNotFound
	} else if err != nil {
		return nil, fmt.Errorf("could not get sticker packs: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var stickerPack postgres_models.StickerPackPostgres
		err = rows.Scan(&stickerPack.ID, &stickerPack.Name, &stickerPack.CreatorID, &stickerPack.CreatedAt, &stickerPack.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("could not scan sticker pack: %v", err)
		}

		stickerRows, err := s.ConnPool.QueryContext(ctx, getStickersQuery, stickerPack.ID)
		if err != nil {
			return nil, fmt.Errorf("could not get stickers for pack: %v", err)
		}

		var stickers []postgres_models.PostgresFile
		for stickerRows.Next() {
			var sticker postgres_models.PostgresFile
			err = stickerRows.Scan(&sticker.URL)
			if err != nil {
				return nil, fmt.Errorf("could not scan sticker: %v", err)
			}
			stickers = append(stickers, sticker)
		}
		stickerPack.Stickers = stickers
		stickerRows.Close()

		stickerPacks = append(stickerPacks, stickerPack.ToStickerPack())
	}

	return stickerPacks, nil
}

func (s *StickerRepository) DeleteStickerPack(ctx context.Context, userId, packId uuid.UUID) error {
	var creatorId uuid.UUID
	postgresPackId := pgtype.UUID{Bytes: packId, Valid: true}
	err := s.ConnPool.QueryRowContext(ctx, `
		SELECT creator_id FROM sticker_pack WHERE id = $1`, postgresPackId).Scan(&creatorId)
	if err != nil {
		return fmt.Errorf("could not check creator: %v", err)
	}

	if creatorId != userId {
		return fmt.Errorf("user does not own this sticker pack")
	}

	_, err = s.ConnPool.ExecContext(ctx, `
		DELETE FROM sticker WHERE sticker_pack_id = $1`, postgresPackId)
	if err != nil {
		return fmt.Errorf("could not delete stickers: %v", err)
	}

	_, err = s.ConnPool.ExecContext(ctx, `
		DELETE FROM sticker_pack WHERE id = $1`, postgresPackId)
	if err != nil {
		return fmt.Errorf("could not delete sticker pack: %v", err)
	}

	return nil
}

func (s *StickerRepository) GetStickerPackByName(ctx context.Context, name string) (models.StickerPack, error) {
	var stickerPack postgres_models.StickerPackPostgres

	row := s.ConnPool.QueryRowContext(ctx, getStickerPackByNameQuery, name)

	err := row.Scan(&stickerPack.ID, &stickerPack.Name, &stickerPack.CreatorID, &stickerPack.CreatedAt, &stickerPack.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return models.StickerPack{}, messenger_errors.ErrNotFound
	} else if err != nil {
		return models.StickerPack{}, fmt.Errorf("could not get sticker pack: %v", err)
	}

	rows, err := s.ConnPool.QueryContext(ctx, getStickersQuery, stickerPack.ID)
	if err != nil {
		return models.StickerPack{}, fmt.Errorf("could not get stickers: %v", err)
	}
	defer rows.Close()

	var stickers []postgres_models.PostgresFile
	for rows.Next() {
		var sticker postgres_models.PostgresFile
		err = rows.Scan(&sticker.URL)
		if err != nil {
			return models.StickerPack{}, fmt.Errorf("could not scan sticker: %v", err)
		}
		stickers = append(stickers, sticker)
	}

	stickerPack.Stickers = stickers
	return stickerPack.ToStickerPack(), nil
}

func (s *StickerRepository) BelongsTo(ctx context.Context, userId uuid.UUID, packId uuid.UUID) (bool, error) {
	var creatorId uuid.UUID
	err := s.ConnPool.QueryRowContext(ctx, `
		SELECT creator_id FROM sticker_pack WHERE id = $1`, pgtype.UUID{Bytes: packId, Valid: true}).
		Scan(&creatorId)
	if errors.Is(err, sql.ErrNoRows) {
		return false, messenger_errors.ErrNotFound
	} else if err != nil {
		return false, fmt.Errorf("could not check creator: %v", err)
	}

	return creatorId == userId, nil
}
