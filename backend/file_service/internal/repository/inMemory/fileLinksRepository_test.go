//go:build unit
// +build unit

package inMemory_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"quickflow/file_service/internal/errors"
	"quickflow/file_service/internal/repository/inMemory"
	"quickflow/shared/models"

	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/runner"
)

func TestFileUrlsRepository_AddFileRecordAndGetFileURL(t *testing.T) {
	runner.Run(t, "AddFileRecord and GetFileURL Tests", func(t provider.T) {
		t.Epic("Unit")
		t.Feature("InMemory FileUrlsRepository")

		tests := []struct {
			name      string
			file      *models.File
			queryName string
			wantURL   string
			wantErr   error
		}{
			{
				name:      "success - add and get file url",
				file:      &models.File{Name: "test.txt", URL: "http://example.com/test.txt"},
				queryName: "test.txt",
				wantURL:   "http://example.com/test.txt",
				wantErr:   nil,
			},
			{
				name:      "error - file not found",
				file:      &models.File{Name: "test.txt", URL: "http://example.com/test.txt"},
				queryName: "not_exist.txt",
				wantURL:   "",
				wantErr:   errors.FileNotFound,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t provider.T) {
				t.Epic("Unit")
				t.WithNewStep("Создание репозитория", func(s provider.StepCtx) {
					repo := inMemory.NewFileUrlsRepository()

					s.WithNewStep("Добавление файла", func(s provider.StepCtx) {
						err := repo.AddFileRecord(context.Background(), tt.file)
						assert.NoError(t, err)
					})

					t.WithNewStep("Получение URL файла", func(s provider.StepCtx) {
						gotURL, err := repo.GetFileURL(context.Background(), tt.queryName)

						if tt.wantErr != nil {
							assert.ErrorIs(t, err, tt.wantErr)
							assert.Equal(t, tt.wantURL, gotURL)
						} else {
							assert.NoError(t, err)
							assert.Equal(t, tt.wantURL, gotURL)
						}
					})
				})
			})
		}
	})
}

func TestFileUrlsRepository_AddFilesRecords(t *testing.T) {
	runner.Run(t, "AddFilesRecords Tests", func(t provider.T) {
		t.Epic("Unit")
		t.Feature("InMemory FileUrlsRepository")

		tests := []struct {
			name    string
			files   []*models.File
			queries map[string]string
		}{
			{
				name: "success - add multiple files",
				files: []*models.File{
					{Name: "a.txt", URL: "http://example.com/a.txt"},
					{Name: "b.txt", URL: "http://example.com/b.txt"},
				},
				queries: map[string]string{
					"a.txt": "http://example.com/a.txt",
					"b.txt": "http://example.com/b.txt",
				},
			},
			{
				name:    "empty list - no files added",
				files:   []*models.File{},
				queries: map[string]string{},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t provider.T) {
				t.Epic("Unit")
				t.WithNewStep("Создание репозитория", func(s provider.StepCtx) {
					repo := inMemory.NewFileUrlsRepository()

					s.WithNewStep("Добавление файлов", func(s provider.StepCtx) {
						err := repo.AddFilesRecords(context.Background(), tt.files)
						assert.NoError(t, err)
					})

					t.WithNewStep("Проверка добавленных файлов", func(s provider.StepCtx) {
						for name, wantURL := range tt.queries {
							gotURL, err := repo.GetFileURL(context.Background(), name)
							assert.NoError(t, err)
							assert.Equal(t, wantURL, gotURL)
						}
					})
				})
			})
		}
	})
}
