//go:build unit
// +build unit

package postgres_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ozontech/allure-go/pkg/allure"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/runner"
	"github.com/stretchr/testify/require"

	messengerErrors "quickflow/messenger_service/internal/errors"
	"quickflow/messenger_service/internal/repository/postgres"
	"quickflow/shared/models"
)

type StickerPackBuilder struct {
	id        uuid.UUID
	name      string
	creatorID uuid.UUID
	createdAt time.Time
	updatedAt time.Time
	stickers  []models.File
}

func NewStickerPackBuilder() *StickerPackBuilder {
	return &StickerPackBuilder{
		id:        uuid.New(),
		name:      "Sample Pack",
		creatorID: uuid.New(),
		createdAt: time.Now(),
		updatedAt: time.Now(),
		stickers:  []models.File{},
	}
}

func (b *StickerPackBuilder) WithID(id uuid.UUID) *StickerPackBuilder {
	b.id = id
	return b
}

func (b *StickerPackBuilder) WithName(name string) *StickerPackBuilder {
	b.name = name
	return b
}

func (b *StickerPackBuilder) WithCreator(id uuid.UUID) *StickerPackBuilder {
	b.creatorID = id
	return b
}

func (b *StickerPackBuilder) WithStickers(stickers ...models.File) *StickerPackBuilder {
	b.stickers = stickers
	return b
}

func (b *StickerPackBuilder) Build() models.StickerPack {
	res := models.StickerPack{
		Id:        b.id,
		Name:      b.name,
		CreatorId: b.creatorID,
		CreatedAt: b.createdAt,
		UpdatedAt: b.updatedAt,
	}

	for _, el := range b.stickers {
		res.Stickers = append(res.Stickers, &el)
	}
	return res
}

func TestStickerRepository_AddStickerPack(t *testing.T) {
	runner.Run(t, "Add Sticker Pack Tests", func(t provider.T) {
		t.Epic("Unit")
		t.Feature("Add Sticker Pack")
		t.Severity(allure.CRITICAL)
		t.Description("Test adding sticker packs to the repository")

		ctx := context.Background()

		tests := []struct {
			name    string
			pack    models.StickerPack
			mock    func(sqlmock.Sqlmock, models.StickerPack)
			wantErr bool
		}{
			{
				name: "success with stickers",
				pack: NewStickerPackBuilder().
					WithStickers(NewFileBuilder().WithURL("http://example.com/1.png").Build()).
					Build(),
				mock: func(mock sqlmock.Sqlmock, pack models.StickerPack) {
					mock.ExpectExec(`INSERT INTO sticker_pack`).
						WithArgs(pack.Id, pack.Name, pack.CreatedAt, pack.UpdatedAt, pack.CreatorId).
						WillReturnResult(sqlmock.NewResult(1, 1))
					mock.ExpectExec(`INSERT INTO sticker`).
						WithArgs(pack.Id, pack.Stickers[0].URL).
						WillReturnResult(sqlmock.NewResult(1, 1))
				},
			},
			{
				name: "error on pack insert",
				pack: NewStickerPackBuilder().Build(),
				mock: func(mock sqlmock.Sqlmock, pack models.StickerPack) {
					mock.ExpectExec(`INSERT INTO sticker_pack`).
						WillReturnError(errors.New("db error"))
				},
				wantErr: true,
			},
			{
				name: "error on sticker insert",
				pack: NewStickerPackBuilder().
					WithStickers(NewFileBuilder().Build()).
					Build(),
				mock: func(mock sqlmock.Sqlmock, pack models.StickerPack) {
					mock.ExpectExec(`INSERT INTO sticker_pack`).
						WillReturnResult(sqlmock.NewResult(1, 1))
					mock.ExpectExec(`INSERT INTO sticker`).
						WillReturnError(errors.New("db error"))
				},
				wantErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t provider.T) {
				t.Epic("Unit")
				t.Description(tt.name)

				db, mock, _ := sqlmock.New()
				defer db.Close()
				repo := postgres.NewPostgresStickerRepository(db)

				tt.mock(mock, tt.pack)
				err := repo.AddStickerPack(ctx, tt.pack)
				if tt.wantErr {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
				}
				require.NoError(t, mock.ExpectationsWereMet())
			})
		}
	})
}

func TestStickerRepository_GetStickerPack(t *testing.T) {
	runner.Run(t, "Get Sticker Pack Tests", func(t provider.T) {
		t.Epic("Unit")
		t.Feature("Get Sticker Pack")
		t.Severity(allure.CRITICAL)
		t.Description("Test retrieving sticker packs by ID")

		ctx := context.Background()
		pack := NewStickerPackBuilder().
			WithStickers(NewFileBuilder().Build()).
			Build()

		tests := []struct {
			name    string
			mock    func(sqlmock.Sqlmock)
			wantErr error
		}{
			{
				name: "success",
				mock: func(mock sqlmock.Sqlmock) {
					mock.ExpectQuery(`SELECT id, name, creator_id, created_at, updated_at`).
						WithArgs(pack.Id).
						WillReturnRows(sqlmock.NewRows([]string{"id", "name", "creator_id", "created_at", "updated_at"}).
							AddRow(pgtype.UUID{Bytes: pack.Id, Valid: true}, pack.Name,
								pgtype.UUID{Bytes: pack.CreatorId, Valid: true},
								pack.CreatedAt, pack.UpdatedAt))
					mock.ExpectQuery(`SELECT sticker_url FROM sticker`).
						WithArgs(pack.Id).
						WillReturnRows(sqlmock.NewRows([]string{"sticker_url"}).
							AddRow(pack.Stickers[0].URL))
				},
				wantErr: nil,
			},
			{
				name: "not found",
				mock: func(mock sqlmock.Sqlmock) {
					mock.ExpectQuery(`SELECT id, name, creator_id, created_at, updated_at`).
						WithArgs(pack.Id).
						WillReturnError(sql.ErrNoRows)
				},
				wantErr: messengerErrors.ErrNotFound,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t provider.T) {
				t.Epic("Unit")
				t.Description(tt.name)

				db, mock, _ := sqlmock.New()
				defer db.Close()
				repo := postgres.NewPostgresStickerRepository(db)

				tt.mock(mock)
				_, err := repo.GetStickerPack(ctx, pack.Id)
				if tt.wantErr != nil {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
				}
				require.NoError(t, mock.ExpectationsWereMet())
			})
		}
	})
}

func TestStickerRepository_GetStickerPacks(t *testing.T) {
	runner.Run(t, "Get Sticker Packs Tests", func(t provider.T) {
		t.Epic("Unit")
		t.Feature("Get Sticker Packs")
		t.Severity(allure.NORMAL)
		t.Description("Test retrieving multiple sticker packs with pagination")

		ctx := context.Background()
		pack1 := NewStickerPackBuilder().WithName("Pack 1").WithStickers(NewFileBuilder().WithURL("http://example.com/1.png").Build()).Build()
		pack2 := NewStickerPackBuilder().WithName("Pack 2").WithStickers(NewFileBuilder().WithURL("http://example.com/2.png").Build()).Build()

		tests := []struct {
			name    string
			mock    func(sqlmock.Sqlmock)
			wantErr bool
		}{
			{
				name: "success",
				mock: func(mock sqlmock.Sqlmock) {
					mock.ExpectQuery(`SELECT sp.id, sp.name, sp.creator_id, sp.created_at, sp.updated_at`).
						WithArgs(10, 0).
						WillReturnRows(sqlmock.NewRows([]string{"id", "name", "creator_id", "created_at", "updated_at"}).
							AddRow(pgtype.UUID{Bytes: pack1.Id, Valid: true}, pack1.Name,
								pgtype.UUID{Bytes: pack1.CreatorId, Valid: true}, pack1.CreatedAt, pack1.UpdatedAt).
							AddRow(pgtype.UUID{Bytes: pack2.Id, Valid: true}, pack2.Name,
								pgtype.UUID{Bytes: pack2.CreatorId, Valid: true}, pack2.CreatedAt, pack2.UpdatedAt))
					mock.ExpectQuery(`SELECT sticker_url FROM sticker`).
						WithArgs(pgtype.UUID{Bytes: pack1.Id, Valid: true}).
						WillReturnRows(sqlmock.NewRows([]string{"sticker_url"}).AddRow(pack1.Stickers[0].URL))
					mock.ExpectQuery(`SELECT sticker_url FROM sticker`).
						WithArgs(pgtype.UUID{Bytes: pack2.Id, Valid: true}).
						WillReturnRows(sqlmock.NewRows([]string{"sticker_url"}).AddRow(pack2.Stickers[0].URL))
				},
			},
			{
				name: "error on query packs",
				mock: func(mock sqlmock.Sqlmock) {
					mock.ExpectQuery(`SELECT sp.id, sp.name, sp.creator_id, sp.created_at, sp.updated_at`).
						WithArgs(10, 0).
						WillReturnError(errors.New("db error"))
				},
				wantErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t provider.T) {
				t.Epic("Unit")
				t.Description(tt.name)

				db, mock, _ := sqlmock.New()
				defer db.Close()
				repo := postgres.NewPostgresStickerRepository(db)

				tt.mock(mock)
				_, err := repo.GetStickerPacks(ctx, uuid.Nil, 10, 0)
				if tt.wantErr {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
				}
				require.NoError(t, mock.ExpectationsWereMet())
			})
		}
	})
}

func TestStickerRepository_DeleteStickerPack(t *testing.T) {
	runner.Run(t, "Delete Sticker Pack Tests", func(t provider.T) {
		t.Epic("Unit")
		t.Feature("Delete Sticker Pack")
		t.Severity(allure.CRITICAL)
		t.Description("Test deleting sticker packs with ownership validation")

		ctx := context.Background()
		userID := uuid.New()
		pack := NewStickerPackBuilder().WithCreator(userID).Build()

		tests := []struct {
			name    string
			userID  uuid.UUID
			mock    func(sqlmock.Sqlmock)
			wantErr bool
		}{
			{
				name:   "success",
				userID: userID,
				mock: func(mock sqlmock.Sqlmock) {
					mock.ExpectQuery(`SELECT creator_id FROM sticker_pack`).
						WithArgs(pgtype.UUID{Bytes: pack.Id, Valid: true}).
						WillReturnRows(sqlmock.NewRows([]string{"creator_id"}).AddRow(userID))
					mock.ExpectExec(`DELETE FROM sticker`).
						WithArgs(pgtype.UUID{Bytes: pack.Id, Valid: true}).
						WillReturnResult(sqlmock.NewResult(1, 1))
					mock.ExpectExec(`DELETE FROM sticker_pack`).
						WithArgs(pgtype.UUID{Bytes: pack.Id, Valid: true}).
						WillReturnResult(sqlmock.NewResult(1, 1))
				},
			},
			{
				name:   "not owner",
				userID: uuid.New(),
				mock: func(mock sqlmock.Sqlmock) {
					mock.ExpectQuery(`SELECT creator_id FROM sticker_pack`).
						WithArgs(pgtype.UUID{Bytes: pack.Id, Valid: true}).
						WillReturnRows(sqlmock.NewRows([]string{"creator_id"}).AddRow(userID))
				},
				wantErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t provider.T) {
				t.Epic("Unit")
				t.Description(tt.name)

				db, mock, _ := sqlmock.New()
				defer db.Close()
				repo := postgres.NewPostgresStickerRepository(db)

				tt.mock(mock)
				err := repo.DeleteStickerPack(ctx, tt.userID, pack.Id)
				if tt.wantErr {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
				}
				require.NoError(t, mock.ExpectationsWereMet())
			})
		}
	})
}

func TestStickerRepository_GetStickerPackByName(t *testing.T) {
	runner.Run(t, "Get Sticker Pack By Name Tests", func(t provider.T) {
		t.Epic("Unit")
		t.Feature("Get Sticker Pack By Name")
		t.Severity(allure.NORMAL)
		t.Description("Test retrieving sticker packs by name")

		ctx := context.Background()
		pack := NewStickerPackBuilder().WithName("Cool Pack").
			WithStickers(NewFileBuilder().WithURL("http://example.com/cool.png").Build()).
			Build()

		tests := []struct {
			name    string
			mock    func(sqlmock.Sqlmock)
			wantErr bool
		}{
			{
				name: "success",
				mock: func(mock sqlmock.Sqlmock) {
					mock.ExpectQuery(`SELECT id, name, creator_id, created_at, updated_at FROM sticker_pack`).
						WithArgs(pack.Name).
						WillReturnRows(sqlmock.NewRows([]string{"id", "name", "creator_id", "created_at", "updated_at"}).
							AddRow(pgtype.UUID{Bytes: pack.Id, Valid: true}, pack.Name,
								pgtype.UUID{Bytes: pack.CreatorId, Valid: true}, pack.CreatedAt, pack.UpdatedAt))
					mock.ExpectQuery(`SELECT sticker_url FROM sticker`).
						WithArgs(pgtype.UUID{Bytes: pack.Id, Valid: true}).
						WillReturnRows(sqlmock.NewRows([]string{"sticker_url"}).AddRow(pack.Stickers[0].URL))
				},
			},
			{
				name: "not found",
				mock: func(mock sqlmock.Sqlmock) {
					mock.ExpectQuery(`SELECT id, name, creator_id, created_at, updated_at FROM sticker_pack`).
						WithArgs(pack.Name).
						WillReturnError(sql.ErrNoRows)
				},
				wantErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t provider.T) {
				t.Epic("Unit")
				t.Description(tt.name)

				db, mock, _ := sqlmock.New()
				defer db.Close()
				repo := postgres.NewPostgresStickerRepository(db)

				tt.mock(mock)
				_, err := repo.GetStickerPackByName(ctx, pack.Name)
				if tt.wantErr {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
				}
				require.NoError(t, mock.ExpectationsWereMet())
			})
		}
	})
}

func TestStickerRepository_BelongsTo(t *testing.T) {
	runner.Run(t, "Check Sticker Pack Ownership Tests", func(t provider.T) {
		t.Epic("Unit")
		t.Feature("Check Sticker Pack Ownership")
		t.Severity(allure.NORMAL)
		t.Description("Test checking if a user owns a sticker pack")

		ctx := context.Background()
		userID := uuid.New()
		pack := NewStickerPackBuilder().WithCreator(userID).Build()

		tests := []struct {
			name    string
			userID  uuid.UUID
			mock    func(sqlmock.Sqlmock)
			want    bool
			wantErr bool
		}{
			{
				name:   "belongs",
				userID: userID,
				mock: func(mock sqlmock.Sqlmock) {
					mock.ExpectQuery(`SELECT creator_id FROM sticker_pack`).
						WithArgs(pgtype.UUID{Bytes: pack.Id, Valid: true}).
						WillReturnRows(sqlmock.NewRows([]string{"creator_id"}).AddRow(userID))
				},
				want:    true,
				wantErr: false,
			},
			{
				name:   "does not belong",
				userID: uuid.New(),
				mock: func(mock sqlmock.Sqlmock) {
					mock.ExpectQuery(`SELECT creator_id FROM sticker_pack`).
						WithArgs(pgtype.UUID{Bytes: pack.Id, Valid: true}).
						WillReturnRows(sqlmock.NewRows([]string{"creator_id"}).AddRow(userID))
				},
				want:    false,
				wantErr: false,
			},
			{
				name:   "not found",
				userID: userID,
				mock: func(mock sqlmock.Sqlmock) {
					mock.ExpectQuery(`SELECT creator_id FROM sticker_pack`).
						WithArgs(pgtype.UUID{Bytes: pack.Id, Valid: true}).
						WillReturnError(sql.ErrNoRows)
				},
				wantErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t provider.T) {
				t.Epic("Unit")
				t.Description(tt.name)

				db, mock, _ := sqlmock.New()
				defer db.Close()
				repo := postgres.NewPostgresStickerRepository(db)

				tt.mock(mock)
				got, err := repo.BelongsTo(ctx, tt.userID, pack.Id)
				if tt.wantErr {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
					require.Equal(t, tt.want, got)
				}
				require.NoError(t, mock.ExpectationsWereMet())
			})
		}
	})
}
