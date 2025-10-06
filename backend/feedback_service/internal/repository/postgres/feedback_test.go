//go:build unit
// +build unit

package postgres

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"

	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/runner"

	"quickflow/shared/models"
)

func TestFeedbackRepository_SaveFeedback(t *testing.T) {
	runner.Run(t, "Save Feedback Tests", func(t provider.T) {

		t.Epic("Unit")
		t.Feature("SaveFeedback")

		type args struct {
			ctx      context.Context
			feedback *models.Feedback
		}

		type mockBehavior func(mock sqlmock.Sqlmock, args args)

		tests := []struct {
			name         string
			args         args
			mockBehavior mockBehavior
			wantErr      bool
			expectedErr  string
		}{
			{
				name: "successful save",
				args: args{
					ctx: context.Background(),
					feedback: &models.Feedback{
						Id:           uuid.New(),
						Rating:       5,
						RespondentId: uuid.New(),
						Text:         "Great service!",
						Type:         models.FeedbackGeneral,
						CreatedAt:    time.Now(),
					},
				},
				mockBehavior: func(mock sqlmock.Sqlmock, args args) {
					mock.ExpectExec("insert into feedback").
						WithArgs(
							args.feedback.Rating,
							args.feedback.RespondentId,
							args.feedback.Text,
							args.feedback.Type,
							args.feedback.CreatedAt,
						).
						WillReturnResult(sqlmock.NewResult(1, 1))
				},
				wantErr: false,
			},
			{
				name: "database error",
				args: args{
					ctx: context.Background(),
					feedback: &models.Feedback{
						Id:           uuid.New(),
						Rating:       5,
						RespondentId: uuid.New(),
						Text:         "Great service!",
						Type:         models.FeedbackGeneral,
						CreatedAt:    time.Now(),
					},
				},
				mockBehavior: func(mock sqlmock.Sqlmock, args args) {
					mock.ExpectExec("insert into feedback").
						WithArgs(
							args.feedback.Rating,
							args.feedback.RespondentId,
							args.feedback.Text,
							args.feedback.Type,
							args.feedback.CreatedAt,
						).
						WillReturnError(errors.New("database error"))
				},
				wantErr:     true,
				expectedErr: "save feedback: database error",
			},
			{
				name: "zero rating",
				args: args{
					ctx: context.Background(),
					feedback: &models.Feedback{
						Id:           uuid.New(),
						Rating:       0,
						RespondentId: uuid.New(),
						Text:         "Zero rating test",
						Type:         models.FeedbackGeneral,
						CreatedAt:    time.Now(),
					},
				},
				mockBehavior: func(mock sqlmock.Sqlmock, args args) {
					mock.ExpectExec("insert into feedback").
						WithArgs(
							args.feedback.Rating,
							args.feedback.RespondentId,
							args.feedback.Text,
							args.feedback.Type,
							args.feedback.CreatedAt,
						).
						WillReturnResult(sqlmock.NewResult(1, 1))
				},
				wantErr: false,
			},
			{
				name: "empty text",
				args: args{
					ctx: context.Background(),
					feedback: &models.Feedback{
						Id:           uuid.New(),
						Rating:       4,
						RespondentId: uuid.New(),
						Text:         "",
						Type:         models.FeedbackGeneral,
						CreatedAt:    time.Now(),
					},
				},
				mockBehavior: func(mock sqlmock.Sqlmock, args args) {
					mock.ExpectExec("insert into feedback").
						WithArgs(
							args.feedback.Rating,
							args.feedback.RespondentId,
							args.feedback.Text,
							args.feedback.Type,
							args.feedback.CreatedAt,
						).
						WillReturnResult(sqlmock.NewResult(1, 1))
				},
				wantErr: false,
			},
			{
				name: "different feedback types",
				args: args{
					ctx: context.Background(),
					feedback: &models.Feedback{
						Id:           uuid.New(),
						Rating:       3,
						RespondentId: uuid.New(),
						Text:         "Post feedback",
						Type:         models.FeedbackPost,
						CreatedAt:    time.Now(),
					},
				},
				mockBehavior: func(mock sqlmock.Sqlmock, args args) {
					mock.ExpectExec("insert into feedback").
						WithArgs(
							args.feedback.Rating,
							args.feedback.RespondentId,
							args.feedback.Text,
							args.feedback.Type,
							args.feedback.CreatedAt,
						).
						WillReturnResult(sqlmock.NewResult(1, 1))
				},
				wantErr: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t provider.T) {
				t.Epic("Unit")
				t.Epic("Unit")
				t.WithNewStep("Создание mock DB", func(s provider.StepCtx) {
					db, mock, err := sqlmock.New()
					s.Require().NoError(err)
					defer db.Close()

					repo := NewFeedbackRepository(db)
					tt.mockBehavior(mock, tt.args)

					t.WithNewStep("Вызов SaveFeedback", func(s provider.StepCtx) {
						err = repo.SaveFeedback(tt.args.ctx, tt.args.feedback)

						if tt.wantErr {
							assert.Error(t, err)
							assert.Contains(t, err.Error(), tt.expectedErr)
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

func TestFeedbackRepository_GetAllFeedbackType(t *testing.T) {
	runner.Run(t, "GetAllFeedbackType Tests", func(t provider.T) {
		t.Epic("Unit")
		t.Feature("GetAllFeedbackType")

		type args struct {
			ctx          context.Context
			feedbackType models.FeedbackType
			ts           time.Time
			count        int
		}

		type mockBehavior func(mock sqlmock.Sqlmock, args args, feedbacks []*models.Feedback)

		tests := []struct {
			name         string
			args         args
			mockBehavior mockBehavior
			want         []models.Feedback
			wantErr      bool
			expectedErr  string
		}{
			{
				name: "successful get feedback",
				args: args{
					ctx:          context.Background(),
					feedbackType: models.FeedbackGeneral,
					ts:           time.Now(),
					count:        5,
				},
				mockBehavior: func(mock sqlmock.Sqlmock, args args, feedbacks []*models.Feedback) {
					rows := sqlmock.NewRows([]string{"rating", "respondent_id", "text", "type", "created_at"})
					for _, f := range feedbacks {
						rows.AddRow(f.Rating, f.RespondentId, f.Text, f.Type, f.CreatedAt)
					}
					mock.ExpectQuery("select rating, respondent_id, text, type, created_at").
						WithArgs(
							pgtype.Timestamptz{Time: args.ts, Valid: true},
							pgtype.Text{String: string(args.feedbackType), Valid: true},
							args.count,
						).
						WillReturnRows(rows)
				},
				want: []models.Feedback{
					{
						Rating:       5,
						RespondentId: uuid.New(),
						Text:         "Great service!",
						Type:         models.FeedbackGeneral,
						CreatedAt:    time.Now(),
					},
				},
				wantErr: false,
			},
			{
				name: "query error",
				args: args{
					ctx:          context.Background(),
					feedbackType: models.FeedbackGeneral,
					ts:           time.Now(),
					count:        5,
				},
				mockBehavior: func(mock sqlmock.Sqlmock, args args, feedbacks []*models.Feedback) {
					mock.ExpectQuery("select rating, respondent_id, text, type, created_at").
						WithArgs(
							pgtype.Timestamptz{Time: args.ts, Valid: true},
							pgtype.Text{String: string(args.feedbackType), Valid: true},
							args.count,
						).
						WillReturnError(errors.New("query error"))
				},
				want:        nil,
				wantErr:     true,
				expectedErr: "get feedback: query error",
			},
			{
				name: "no rows found",
				args: args{
					ctx:          context.Background(),
					feedbackType: models.FeedbackGeneral,
					ts:           time.Now(),
					count:        5,
				},
				mockBehavior: func(mock sqlmock.Sqlmock, args args, feedbacks []*models.Feedback) {
					mock.ExpectQuery("select rating, respondent_id, text, type, created_at").
						WithArgs(
							pgtype.Timestamptz{Time: args.ts, Valid: true},
							pgtype.Text{String: string(args.feedbackType), Valid: true},
							args.count,
						).
						WillReturnError(sql.ErrNoRows)
				},
				want:        nil,
				wantErr:     true,
				expectedErr: "not found",
			},
			{
				name: "multiple feedback items",
				args: args{
					ctx:          context.Background(),
					feedbackType: models.FeedbackPost,
					ts:           time.Now().Add(-24 * time.Hour),
					count:        10,
				},
				mockBehavior: func(mock sqlmock.Sqlmock, args args, feedbacks []*models.Feedback) {
					rows := sqlmock.NewRows([]string{"rating", "respondent_id", "text", "type", "created_at"})
					for _, f := range feedbacks {
						rows.AddRow(f.Rating, f.RespondentId, f.Text, f.Type, f.CreatedAt)
					}
					mock.ExpectQuery("select rating, respondent_id, text, type, created_at").
						WithArgs(
							pgtype.Timestamptz{Time: args.ts, Valid: true},
							pgtype.Text{String: string(args.feedbackType), Valid: true},
							args.count,
						).
						WillReturnRows(rows)
				},
				want: []models.Feedback{
					{
						Rating:       2,
						RespondentId: uuid.New(),
						Text:         "Post feedback",
						Type:         models.FeedbackPost,
						CreatedAt:    time.Now().Add(-2 * time.Hour),
					},
					{
						Rating:       1,
						RespondentId: uuid.New(),
						Text:         "Auth feedback",
						Type:         models.FeedbackAuth,
						CreatedAt:    time.Now().Add(-1 * time.Hour),
					},
				},
				wantErr: false,
			},
			{
				name: "zero count",
				args: args{
					ctx:          context.Background(),
					feedbackType: models.FeedbackGeneral,
					ts:           time.Now(),
					count:        0,
				},
				mockBehavior: func(mock sqlmock.Sqlmock, args args, feedbacks []*models.Feedback) {},
				want:         []models.Feedback{},
				wantErr:      true,
			},
			{
				name: "scan error",
				args: args{
					ctx:          context.Background(),
					feedbackType: models.FeedbackGeneral,
					ts:           time.Now(),
					count:        5,
				},
				mockBehavior: func(mock sqlmock.Sqlmock, args args, feedbacks []*models.Feedback) {
					rows := sqlmock.NewRows([]string{"rating", "respondent_id", "text", "type", "created_at"}).
						AddRow("invalid", uuid.New(), "test", models.FeedbackGeneral, time.Now())
					mock.ExpectQuery("select rating, respondent_id, text, type, created_at").
						WithArgs(
							pgtype.Timestamptz{Time: args.ts, Valid: true},
							pgtype.Text{String: string(args.feedbackType), Valid: true},
							args.count,
						).
						WillReturnRows(rows)
				},
				want:        nil,
				wantErr:     true,
				expectedErr: "get feedback:",
			},
			{
				name: "different feedback types mixed",
				args: args{
					ctx:          context.Background(),
					feedbackType: models.FeedbackGeneral,
					ts:           time.Now(),
					count:        3,
				},
				mockBehavior: func(mock sqlmock.Sqlmock, args args, feedbacks []*models.Feedback) {
					rows := sqlmock.NewRows([]string{"rating", "respondent_id", "text", "type", "created_at"})
					for _, f := range feedbacks {
						rows.AddRow(f.Rating, f.RespondentId, f.Text, f.Type, f.CreatedAt)
					}
					mock.ExpectQuery("select rating, respondent_id, text, type, created_at").
						WithArgs(
							pgtype.Timestamptz{Time: args.ts, Valid: true},
							pgtype.Text{String: string(args.feedbackType), Valid: true},
							args.count,
						).
						WillReturnRows(rows)
				},
				want: []models.Feedback{
					{
						Rating:       4,
						RespondentId: uuid.New(),
						Text:         "Add dark mode",
						Type:         models.FeedbackGeneral,
						CreatedAt:    time.Now().Add(-1 * time.Hour),
					},
				},
				wantErr: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t provider.T) {
				t.Epic("Unit")
				t.Epic("Unit")
				t.WithNewStep("Создание mock DB", func(s provider.StepCtx) {
					db, mock, err := sqlmock.New()
					s.Require().NoError(err)
					defer db.Close()

					repo := NewFeedbackRepository(db)

					var feedbackPtrs []*models.Feedback
					for i := range tt.want {
						feedbackPtrs = append(feedbackPtrs, &tt.want[i])
					}
					tt.mockBehavior(mock, tt.args, feedbackPtrs)

					if tt.args.count == 0 {
						result, err := repo.GetAllFeedbackType(tt.args.ctx, tt.args.feedbackType, tt.args.ts, tt.args.count)
						assert.NoError(t, err)
						assert.Empty(t, result)
						return
					}

					t.WithNewStep("Вызов GetAllFeedbackType", func(s provider.StepCtx) {
						result, err := repo.GetAllFeedbackType(tt.args.ctx, tt.args.feedbackType, tt.args.ts, tt.args.count)

						if tt.wantErr {
							assert.Error(t, err)
							if tt.expectedErr != "" {
								assert.Contains(t, err.Error(), tt.expectedErr)
							}
							assert.Nil(t, result)
						} else {
							assert.NoError(t, err)
							assert.Len(t, result, len(tt.want))
							for i, expected := range tt.want {
								assert.Equal(t, expected.Rating, result[i].Rating)
								assert.Equal(t, expected.Type, result[i].Type)
								assert.Equal(t, expected.Text, result[i].Text)
							}
						}
						assert.NoError(t, mock.ExpectationsWereMet())
					})
				})
			})
		}
	})
}

func TestFeedbackRepository_Close(t *testing.T) {
	runner.Run(t, "Close Repository Tests", func(t provider.T) {
		t.Epic("Unit")
		t.Feature("Close")

		t.WithNewStep("Создание mock DB", func(s provider.StepCtx) {
			db, mock, err := sqlmock.New()
			s.Require().NoError(err)

			repo := NewFeedbackRepository(db)
			mock.ExpectClose()

			t.WithNewStep("Закрытие соединения", func(s provider.StepCtx) {
				repo.Close()
				assert.NoError(t, mock.ExpectationsWereMet())
			})
		})
	})
}
