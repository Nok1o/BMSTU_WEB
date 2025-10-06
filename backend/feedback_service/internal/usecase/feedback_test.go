//go:build unit
// +build unit

package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"

	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"

	"quickflow/feedback_service/internal/usecase/mocks"
	"quickflow/shared/models"
)

type FeedbackSuite struct {
	suite.Suite
}

func (s *FeedbackSuite) TestSaveFeedback(t provider.T) {
	type args struct {
		ctx      context.Context
		feedback *models.Feedback
	}

	tests := []struct {
		name         string
		args         args
		mockBehavior func(r *mocks.MockFeedbackRepository, args args)
		wantErr      bool
		expectedErr  error
	}{
		{
			name: "success - valid feedback",
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
			mockBehavior: func(r *mocks.MockFeedbackRepository, args args) {
				r.EXPECT().SaveFeedback(args.ctx, args.feedback).Return(nil)
			},
		},
		{
			name: "error - validation failed",
			args: args{
				ctx: context.Background(),
				feedback: &models.Feedback{
					Id:           uuid.New(),
					Rating:       6,
					RespondentId: uuid.New(),
					Text:         "Great service!",
					Type:         models.FeedbackGeneral,
					CreatedAt:    time.Now(),
				},
			},
			mockBehavior: func(r *mocks.MockFeedbackRepository, args args) {},
			wantErr:      true,
			expectedErr:  errors.New("invalid rating"),
		},
		{
			name: "error - repository error",
			args: args{
				ctx: context.Background(),
				feedback: &models.Feedback{
					Id:           uuid.New(),
					Rating:       4,
					RespondentId: uuid.New(),
					Text:         "Good service",
					Type:         models.FeedbackGeneral,
					CreatedAt:    time.Now(),
				},
			},
			mockBehavior: func(r *mocks.MockFeedbackRepository, args args) {
				r.EXPECT().SaveFeedback(args.ctx, args.feedback).Return(errors.New("database error"))
			},
			wantErr:     true,
			expectedErr: errors.New("database error"),
		},
		{
			name: "success - different feedback types",
			args: args{
				ctx: context.Background(),
				feedback: &models.Feedback{
					Id:           uuid.New(),
					Rating:       3,
					RespondentId: uuid.New(),
					Text:         "Bug report",
					Type:         models.FeedbackAuth,
					CreatedAt:    time.Now(),
				},
			},
			mockBehavior: func(r *mocks.MockFeedbackRepository, args args) {
				r.EXPECT().SaveFeedback(args.ctx, args.feedback).Return(nil)
			},
		},
		{
			name: "success - empty text allowed",
			args: args{
				ctx: context.Background(),
				feedback: &models.Feedback{
					Id:           uuid.New(),
					Rating:       5,
					RespondentId: uuid.New(),
					Text:         "",
					Type:         models.FeedbackGeneral,
					CreatedAt:    time.Now(),
				},
			},
			mockBehavior: func(r *mocks.MockFeedbackRepository, args args) {
				r.EXPECT().SaveFeedback(args.ctx, args.feedback).Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(pt provider.T) {
			t.Epic("Unit")
			pt.WithNewStep("Prepare mocks", func(s provider.StepCtx) {
				ctrl := gomock.NewController(t)
				defer ctrl.Finish()

				mockRepo := mocks.NewMockFeedbackRepository(ctrl)
				tt.mockBehavior(mockRepo, tt.args)

				uc := NewFeedBackUseCase(mockRepo)

				s.WithNewStep("Call SaveFeedback", func(ss provider.StepCtx) {
					err := uc.SaveFeedback(tt.args.ctx, tt.args.feedback)

					if tt.wantErr {
						ss.Assert().Error(err)
						if tt.expectedErr != nil {
							ss.Assert().Equal(tt.expectedErr.Error(), err.Error())
						}
					} else {
						ss.Assert().NoError(err)
					}
				})
			})
		})
	}
}

func (s *FeedbackSuite) TestGetAllFeedbackType(t provider.T) {
	type args struct {
		ctx          context.Context
		feedbackType models.FeedbackType
		ts           time.Time
		count        int
	}

	tests := []struct {
		name         string
		args         args
		mockBehavior func(r *mocks.MockFeedbackRepository, args args, result []models.Feedback, err error)
		want         []models.Feedback
		wantErr      bool
		expectedErr  string
	}{
		{
			name: "success - get feedback with results",
			args: args{
				ctx:          context.Background(),
				feedbackType: models.FeedbackGeneral,
				ts:           time.Now(),
				count:        5,
			},
			mockBehavior: func(r *mocks.MockFeedbackRepository, args args, result []models.Feedback, err error) {
				r.EXPECT().GetAllFeedbackType(args.ctx, args.feedbackType, args.ts, args.count).Return(result, err)
			},
			want: []models.Feedback{
				{
					Id:           uuid.New(),
					Rating:       5,
					RespondentId: uuid.New(),
					Text:         "Excellent service",
					Type:         models.FeedbackGeneral,
					CreatedAt:    time.Now(),
				},
				{
					Id:           uuid.New(),
					Rating:       4,
					RespondentId: uuid.New(),
					Text:         "Good experience",
					Type:         models.FeedbackGeneral,
					CreatedAt:    time.Now().Add(-time.Hour),
				},
			},
		},
		{
			name: "success - empty results",
			args: args{
				ctx:          context.Background(),
				feedbackType: models.FeedbackPost,
				ts:           time.Now(),
				count:        10,
			},
			mockBehavior: func(r *mocks.MockFeedbackRepository, args args, result []models.Feedback, err error) {
				r.EXPECT().GetAllFeedbackType(args.ctx, args.feedbackType, args.ts, args.count).Return(result, err)
			},
			want: []models.Feedback{},
		},
		{
			name: "error - repository error",
			args: args{
				ctx:          context.Background(),
				feedbackType: models.FeedbackGeneral,
				ts:           time.Now(),
				count:        5,
			},
			mockBehavior: func(r *mocks.MockFeedbackRepository, args args, result []models.Feedback, err error) {
				r.EXPECT().GetAllFeedbackType(args.ctx, args.feedbackType, args.ts, args.count).Return(result, err)
			},
			want:        nil,
			wantErr:     true,
			expectedErr: "get all feedback type: database error",
		},
		{
			name: "success - different feedback types",
			args: args{
				ctx:          context.Background(),
				feedbackType: models.FeedbackGeneral,
				ts:           time.Now().Add(-24 * time.Hour),
				count:        3,
			},
			mockBehavior: func(r *mocks.MockFeedbackRepository, args args, result []models.Feedback, err error) {
				r.EXPECT().GetAllFeedbackType(args.ctx, args.feedbackType, args.ts, args.count).Return(result, err)
			},
			want: []models.Feedback{
				{
					Id:           uuid.New(),
					Rating:       4,
					RespondentId: uuid.New(),
					Text:         "Add dark mode feature",
					Type:         models.FeedbackGeneral,
					CreatedAt:    time.Now().Add(-2 * time.Hour),
				},
			},
		},
		{
			name: "success - zero count",
			args: args{
				ctx:          context.Background(),
				feedbackType: models.FeedbackGeneral,
				ts:           time.Now(),
				count:        0,
			},
			mockBehavior: func(r *mocks.MockFeedbackRepository, args args, result []models.Feedback, err error) {
			},
			want: []models.Feedback{},
		},
		{
			name: "error - not found error from repository",
			args: args{
				ctx:          context.Background(),
				feedbackType: models.FeedbackGeneral,
				ts:           time.Now(),
				count:        5,
			},
			mockBehavior: func(r *mocks.MockFeedbackRepository, args args, result []models.Feedback, err error) {
				r.EXPECT().GetAllFeedbackType(args.ctx, args.feedbackType, args.ts, args.count).Return(result, err)
			},
			want:        nil,
			wantErr:     true,
			expectedErr: "get all feedback type: not found error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(pt provider.T) {
			t.Epic("Unit")
			pt.WithNewStep("Prepare mocks", func(s provider.StepCtx) {
				ctrl := gomock.NewController(t)
				defer ctrl.Finish()

				mockRepo := mocks.NewMockFeedbackRepository(ctrl)

				var mockErr error
				if tt.wantErr && tt.expectedErr != "" {
					mockErr = errors.New(tt.expectedErr)
				}
				tt.mockBehavior(mockRepo, tt.args, tt.want, mockErr)

				uc := NewFeedBackUseCase(mockRepo)

				s.WithNewStep("Call GetAllFeedbackType", func(ss provider.StepCtx) {
					if tt.args.count == 0 {
						result, err := uc.GetAllFeedbackType(tt.args.ctx, tt.args.feedbackType, tt.args.ts, tt.args.count)
						ss.Assert().NoError(err)
						ss.Assert().Empty(result)
						return
					}

					result, err := uc.GetAllFeedbackType(tt.args.ctx, tt.args.feedbackType, tt.args.ts, tt.args.count)

					if tt.wantErr {
						ss.Assert().Error(err)
						ss.Assert().Contains(err.Error(), tt.expectedErr)
						ss.Assert().Nil(result)
					} else {
						ss.Assert().NoError(err)
						ss.Assert().Len(result, len(tt.want))
						for i, expected := range tt.want {
							ss.Assert().Equal(expected.Type, result[i].Type)
							ss.Assert().Equal(expected.Rating, result[i].Rating)
							ss.Assert().Equal(expected.Text, result[i].Text)
						}
					}
				})
			})
		})
	}
}

func (s *FeedbackSuite) TestNewFeedBackUseCase(t provider.T) {
	t.Epic("Unit")
	t.WithNewStep("Create usecase", func(s provider.StepCtx) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockFeedbackRepository(ctrl)
		uc := NewFeedBackUseCase(mockRepo)

		s.Assert().NotNil(uc)
		s.Assert().Equal(mockRepo, uc.feedbackRepo)
	})
}

func TestFeedbackSuiteRunner(t *testing.T) {
	suite.RunSuite(t, new(FeedbackSuite))
}
