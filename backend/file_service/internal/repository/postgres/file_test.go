//go:build unit
// +build unit

package postgres_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"

	"quickflow/file_service/internal/repository"
	"quickflow/file_service/internal/repository/postgres"
	"quickflow/shared/models"

	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/runner"
)

func TestPostgresFileRepository_AddFileRecord(t *testing.T) {
	runner.Run(t, "Add File Record Tests", func(t provider.T) {
		t.Epic("Unit")
		t.Feature("AddFileRecord")

		tests := []struct {
			name       string
			file       *models.File
			mockExpect func(mock sqlmock.Sqlmock, file *models.File)
			wantErr    string
		}{
			{
				name: "success",
				file: repository.MotherFile1(),
				mockExpect: func(mock sqlmock.Sqlmock, file *models.File) {
					mock.ExpectExec(`INSERT INTO files \(file_url, filename\) VALUES \(\$1, \$2\)`).
						WithArgs(file.URL, file.Name).
						WillReturnResult(sqlmock.NewResult(1, 1))
				},
				wantErr: "",
			},
			{
				name: "query error",
				file: repository.MotherFile1(),
				mockExpect: func(mock sqlmock.Sqlmock, file *models.File) {
					mock.ExpectExec(`INSERT INTO files \(file_url, filename\) VALUES \(\$1, \$2\)`).
						WithArgs(file.URL, file.Name).
						WillReturnError(errors.New("query error"))
				},
				wantErr: "query error",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t provider.T) {
				t.Epic("Unit")
				t.WithNewStep("Создание mock DB", func(s provider.StepCtx) {
					db, mock, err := sqlmock.New()
					s.Require().NoError(err)
					defer db.Close()

					repo := postgres.NewPostgresFileRepository(db)
					tt.mockExpect(mock, tt.file)

					t.WithNewStep("Вызов AddFileRecord", func(s provider.StepCtx) {
						err = repo.AddFileRecord(context.Background(), tt.file)

						if tt.wantErr != "" {
							assert.EqualError(t, err, tt.wantErr)
						} else {
							assert.NoError(t, err)
						}
						assert.NoError(t, mock.ExpectationsWereMet())
					})
				})
			})
		}
	})
}

func TestPostgresFileRepository_AddFilesRecords(t *testing.T) {
	runner.Run(t, "Add Files Records Tests", func(t provider.T) {

		t.Epic("Unit")
		t.Feature("AddFilesRecords")

		tests := []struct {
			name       string
			files      []*models.File
			mockExpect func(mock sqlmock.Sqlmock, files []*models.File)
			wantErr    string
		}{
			{
				name:  "success",
				files: []*models.File{repository.MotherFile1(), repository.MotherFile2()},
				mockExpect: func(mock sqlmock.Sqlmock, files []*models.File) {
					mock.ExpectBegin()
					mock.ExpectExec(`INSERT INTO files \(file_url, filename\) VALUES \(\$1, \$2\)`).
						WithArgs(files[0].URL, files[0].Name).
						WillReturnResult(sqlmock.NewResult(1, 1))
					mock.ExpectExec(`INSERT INTO files \(file_url, filename\) VALUES \(\$1, \$2\)`).
						WithArgs(files[1].URL, files[1].Name).
						WillReturnResult(sqlmock.NewResult(2, 1))
					mock.ExpectCommit()
				},
				wantErr: "",
			},
			{
				name:  "transaction error",
				files: []*models.File{repository.MotherFile1(), repository.MotherFile2()},
				mockExpect: func(mock sqlmock.Sqlmock, files []*models.File) {
					mock.ExpectBegin()
					mock.ExpectExec(`INSERT INTO files \(file_url, filename\) VALUES \(\$1, \$2\)`).
						WithArgs(files[0].URL, files[0].Name).
						WillReturnResult(sqlmock.NewResult(1, 1))
					mock.ExpectExec(`INSERT INTO files \(file_url, filename\) VALUES \(\$1, \$2\)`).
						WithArgs(files[1].URL, files[1].Name).
						WillReturnError(errors.New("insert error"))
					mock.ExpectRollback()
				},
				wantErr: "insert error",
			},
			{
				name:       "nil files",
				files:      nil,
				mockExpect: func(_ sqlmock.Sqlmock, _ []*models.File) {},
				wantErr:    "",
			},
		}

		for _, tt := range tests {
			t.Epic("Unit")
			t.Run(tt.name, func(t provider.T) {
				t.WithNewStep("Создание mock DB", func(s provider.StepCtx) {
					db, mock, err := sqlmock.New()
					s.Require().NoError(err)
					defer db.Close()

					repo := postgres.NewPostgresFileRepository(db)
					tt.mockExpect(mock, tt.files)

					t.WithNewStep("Вызов AddFilesRecords", func(s provider.StepCtx) {
						err = repo.AddFilesRecords(context.Background(), tt.files)

						if tt.wantErr != "" {
							assert.EqualError(t, err, tt.wantErr)
						} else {
							assert.NoError(t, err)
						}
						assert.NoError(t, mock.ExpectationsWereMet())
					})
				})
			})
		}
	})
}
