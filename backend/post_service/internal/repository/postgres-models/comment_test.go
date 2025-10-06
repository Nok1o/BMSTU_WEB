//go:build unit
// +build unit

package postgres_models

import (
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"

	"quickflow/shared/models"
)

func TestPostgresFile_ToFile(t *testing.T) {
	tests := []struct {
		name     string
		input    PostgresFile
		expected *models.File
	}{
		{
			name: "full data",
			input: PostgresFile{
				URL:         pgtype.Text{String: "http://example.com", Valid: true},
				DisplayType: pgtype.Text{String: "media", Valid: true},
				Name:        pgtype.Text{String: "file.txt", Valid: true},
			},
			expected: &models.File{
				URL:         "http://example.com",
				DisplayType: "media",
				Name:        "file.txt",
			},
		},
		{
			name: "empty display type",
			input: PostgresFile{
				URL:  pgtype.Text{String: "http://example.com", Valid: true},
				Name: pgtype.Text{String: "file.txt", Valid: true},
			},
			expected: &models.File{
				URL:         "http://example.com",
				DisplayType: "file",
				Name:        "file.txt",
			},
		},
		{
			name: "empty name",
			input: PostgresFile{
				URL:         pgtype.Text{String: "http://example.com", Valid: true},
				DisplayType: pgtype.Text{String: "media", Valid: true},
			},
			expected: &models.File{
				URL:         "http://example.com",
				DisplayType: "media",
				Name:        "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.ToFile()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPostgresFilesToModels(t *testing.T) {
	input := []PostgresFile{
		{
			URL: pgtype.Text{String: "url1", Valid: true},
		},
		{
			URL: pgtype.Text{String: "url2", Valid: true},
		},
	}

	result := PostgresFilesToModels(input)
	assert.Len(t, result, 2)
}

func TestFileToPostgres(t *testing.T) {
	input := models.File{
		URL:         "http://example.com",
		DisplayType: "media",
		Name:        "file.txt",
	}

	result := FileToPostgres(input)
	assert.Equal(t, input.URL, result.URL.String)
}

func TestFilesToPostgres(t *testing.T) {
	input := []*models.File{
		{URL: "url1"},
		{URL: "url2"},
	}

	result := FilesToPostgres(input)
	assert.Len(t, result, 2)
}

func TestCommentPostgres_ConvertAndBack(t *testing.T) {
	now := time.Now()
	comment := models.Comment{
		Id:     [16]byte{1, 2, 3},
		PostId: [16]byte{4, 5, 6},
		UserId: [16]byte{7, 8, 9},
		Text:   "test",
		Images: []*models.File{
			{URL: "url1"},
		},
		CreatedAt: now,
		UpdatedAt: now,
		LikeCount: 10,
		IsLiked:   true,
	}

	// Convert to Postgres and back
	pgComment := ConvertCommentToPostgres(comment)
	result := pgComment.ToComment()

	assert.Equal(t, comment.Id, result.Id)
	assert.Len(t, result.Images, 1)
}

func TestCommentPostgres_ToComment(t *testing.T) {
	now := time.Now()
	pgComment := CommentPostgres{
		Id:     pgtype.UUID{Bytes: [16]byte{1, 2, 3}, Valid: true},
		PostId: pgtype.UUID{Bytes: [16]byte{4, 5, 6}, Valid: true},
		UserId: pgtype.UUID{Bytes: [16]byte{7, 8, 9}, Valid: true},
		Text:   pgtype.Text{String: "test", Valid: true},
		Files: []*PostgresFile{
			{URL: pgtype.Text{String: "url1", Valid: true}},
		},
		CreatedAt: pgtype.Timestamptz{Time: now, Valid: true},
		UpdatedAt: pgtype.Timestamptz{Time: now, Valid: true},
		LikeCount: pgtype.Int8{Int64: 10, Valid: true},
		IsLiked:   pgtype.Bool{Bool: true, Valid: true},
	}

	result := pgComment.ToComment()
	assert.Equal(t, "test", result.Text)
}
