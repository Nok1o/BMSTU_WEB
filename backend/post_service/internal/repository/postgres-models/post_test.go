//go:build unit
// +build unit

package postgres_models_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"

	postgres_models "quickflow/post_service/internal/repository/postgres-models"
	"quickflow/shared/models"
)

func TestConvertPostToPostgres(t *testing.T) {
	// Подготовка данных
	post := models.Post{
		Id:           uuid.New(),
		CreatorId:    uuid.New(),
		CreatorType:  models.PostUser,
		Desc:         "This is a test post",
		Files:        []*models.File{{URL: "file_url_1", DisplayType: models.DisplayTypeSticker, Name: "image_1"}},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now().Add(time.Hour),
		LikeCount:    10,
		RepostCount:  5,
		CommentCount: 3,
		IsRepost:     false,
		IsLiked:      true,
	}

	// Преобразование в Postgres модель
	postgresPost := postgres_models.ConvertPostToPostgres(post)

	// Проверки
	assert.True(t, postgresPost.Id.Valid)
	assert.True(t, postgresPost.CreatorId.Valid)
	assert.Equal(t, string(post.CreatorType), postgresPost.CreatorType.String)
	assert.Equal(t, post.Desc, postgresPost.Desc.String)
	assert.Equal(t, len(post.Files), len(postgresPost.Files))
	assert.Equal(t, post.Files[0].URL, postgresPost.Files[0].URL.String)
	assert.Equal(t, post.Files[0].DisplayType, models.DisplayType(postgresPost.Files[0].DisplayType.String))
	assert.Equal(t, post.Files[0].Name, postgresPost.Files[0].Name.String)
	assert.True(t, postgresPost.CreatedAt.Valid)
	assert.Equal(t, post.CreatedAt, postgresPost.CreatedAt.Time)
	assert.True(t, postgresPost.UpdatedAt.Valid)
	assert.Equal(t, post.UpdatedAt, postgresPost.UpdatedAt.Time)
	assert.True(t, postgresPost.LikeCount.Valid)
	assert.Equal(t, int64(post.LikeCount), postgresPost.LikeCount.Int64)
	assert.True(t, postgresPost.RepostCount.Valid)
	assert.Equal(t, int64(post.RepostCount), postgresPost.RepostCount.Int64)
	assert.True(t, postgresPost.CommentCount.Valid)
	assert.Equal(t, int64(post.CommentCount), postgresPost.CommentCount.Int64)
	assert.True(t, postgresPost.IsRepost.Valid)
	assert.Equal(t, post.IsRepost, postgresPost.IsRepost.Bool)
	assert.True(t, postgresPost.IsLiked.Valid)
	assert.Equal(t, post.IsLiked, postgresPost.IsLiked.Bool)
}

func TestToPost(t *testing.T) {
	// Подготовка данных
	postgresPost := postgres_models.PostPostgres{
		Id:           pgtype.UUID{Bytes: uuid.New(), Valid: true},
		CreatorId:    pgtype.UUID{Bytes: uuid.New(), Valid: true},
		CreatorType:  pgtype.Text{String: string(models.PostUser), Valid: true},
		Desc:         pgtype.Text{String: "This is a test post", Valid: true},
		Files:        []postgres_models.PostgresFile{{URL: pgtype.Text{String: "file_url_1", Valid: true}, DisplayType: pgtype.Text{String: "image", Valid: true}, Name: pgtype.Text{String: "image_1", Valid: true}}},
		CreatedAt:    pgtype.Timestamptz{Time: time.Now(), Valid: true},
		UpdatedAt:    pgtype.Timestamptz{Time: time.Now().Add(time.Hour), Valid: true},
		LikeCount:    pgtype.Int8{Int64: 10, Valid: true},
		RepostCount:  pgtype.Int8{Int64: 5, Valid: true},
		CommentCount: pgtype.Int8{Int64: 3, Valid: true},
		IsRepost:     pgtype.Bool{Bool: false, Valid: true},
		IsLiked:      pgtype.Bool{Bool: true, Valid: true},
	}

	// Преобразование в модель Post
	post := postgresPost.ToPost()

	// Проверки
	assert.Equal(t, models.PostCreatorType(post.CreatorType), post.CreatorType)
	assert.Equal(t, postgresPost.Desc.String, post.Desc)
	assert.Equal(t, len(postgresPost.Files), len(post.Files))
	assert.Equal(t, postgresPost.Files[0].URL.String, post.Files[0].URL)
	assert.Equal(t, models.DisplayType(postgresPost.Files[0].DisplayType.String), post.Files[0].DisplayType)
	assert.Equal(t, postgresPost.Files[0].Name.String, post.Files[0].Name)
	assert.Equal(t, postgresPost.CreatedAt.Time, post.CreatedAt)
	assert.Equal(t, postgresPost.UpdatedAt.Time, post.UpdatedAt)
	assert.Equal(t, int(postgresPost.LikeCount.Int64), post.LikeCount)
	assert.Equal(t, int(postgresPost.RepostCount.Int64), post.RepostCount)
	assert.Equal(t, int(postgresPost.CommentCount.Int64), post.CommentCount)
	assert.Equal(t, postgresPost.IsRepost.Bool, post.IsRepost)
	assert.Equal(t, postgresPost.IsLiked.Bool, post.IsLiked)
}

func TestConvertPostToPostgres_EmptyDesc(t *testing.T) {
	// Подготовка данных с пустым описанием
	post := models.Post{
		Id:           uuid.New(),
		CreatorId:    uuid.New(),
		CreatorType:  models.PostUser,
		Desc:         "",
		Files:        nil,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		LikeCount:    10,
		RepostCount:  5,
		CommentCount: 3,
		IsRepost:     false,
		IsLiked:      true,
	}

	// Преобразование в Postgres модель
	postgresPost := postgres_models.ConvertPostToPostgres(post)

	// Проверка
	assert.False(t, postgresPost.Desc.Valid)
}

func TestToPost_EmptyDesc(t *testing.T) {
	// Подготовка данных с пустым описанием
	postgresPost := postgres_models.PostPostgres{
		Id:           pgtype.UUID{Bytes: uuid.New(), Valid: true},
		CreatorId:    pgtype.UUID{Bytes: uuid.New(), Valid: true},
		CreatorType:  pgtype.Text{String: string(models.PostUser), Valid: true},
		Desc:         pgtype.Text{String: "", Valid: false},
		Files:        nil,
		CreatedAt:    pgtype.Timestamptz{Time: time.Now(), Valid: true},
		UpdatedAt:    pgtype.Timestamptz{Time: time.Now().Add(time.Hour), Valid: true},
		LikeCount:    pgtype.Int8{Int64: 10, Valid: true},
		RepostCount:  pgtype.Int8{Int64: 5, Valid: true},
		CommentCount: pgtype.Int8{Int64: 3, Valid: true},
		IsRepost:     pgtype.Bool{Bool: false, Valid: true},
		IsLiked:      pgtype.Bool{Bool: true, Valid: true},
	}

	// Преобразование в модель Post
	post := postgresPost.ToPost()

	// Проверка
	assert.Equal(t, "", post.Desc)
}
