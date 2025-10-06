//go:build unit
// +build unit

package feedback_service_test

import (
	"errors"
	"quickflow/shared/client/feedback_service"
	feedback "quickflow/shared/proto/feedback_service"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"

	"quickflow/shared/models"
)

func TestNewFeedbackClient(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConn := grpc.ClientConn{}
	client := feedback_service.NewFeedbackClient(&mockConn)

	assert.NotNil(t, client)
}

func TestModelFeedbackToProto(t *testing.T) {
	t.Parallel()

	testUUID := uuid.New()
	now := time.Now()

	tests := []struct {
		name        string
		feedback    *models.Feedback
		expected    *feedback.Feedback
		expectedErr error
	}{
		{
			name: "valid general feedback",
			feedback: &models.Feedback{
				Id:           testUUID,
				Rating:       5,
				RespondentId: testUUID,
				Text:         "Great service!",
				Type:         models.FeedbackGeneral,
				CreatedAt:    now,
			},
			expected: &feedback.Feedback{
				Id:           testUUID.String(),
				Rating:       5,
				RespondentId: testUUID.String(),
				Text:         "Great service!",
				Type:         feedback.FeedbackType_FEEDBACK_GENERAL,
				CreatedAt:    timestamppb.New(now),
			},
			expectedErr: nil,
		},
		{
			name: "invalid feedback type",
			feedback: &models.Feedback{
				Id:           testUUID,
				Rating:       5,
				RespondentId: testUUID,
				Text:         "Great service!",
				Type:         "invalid-type",
				CreatedAt:    now,
			},
			expected:    nil,
			expectedErr: errors.New("unknown feedback type"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := feedback_service.ModelFeedbackToProto(tt.feedback)

			if tt.expectedErr != nil {
				assert.Error(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestProtoFeedbackToModel(t *testing.T) {
	t.Parallel()

	testUUID := uuid.New()
	now := time.Now()

	tests := []struct {
		name        string
		feedback    *feedback.Feedback
		expected    *models.Feedback
		expectedErr error
	}{
		{
			name: "valid general feedback",
			feedback: &feedback.Feedback{
				Id:           testUUID.String(),
				Rating:       5,
				RespondentId: testUUID.String(),
				Text:         "Great service!",
				Type:         feedback.FeedbackType_FEEDBACK_GENERAL,
				CreatedAt:    timestamppb.New(now),
			},
			expected: &models.Feedback{
				Id:           testUUID,
				Rating:       5,
				RespondentId: testUUID,
				Text:         "Great service!",
				Type:         models.FeedbackGeneral,
				CreatedAt:    now,
			},
			expectedErr: nil,
		},
		{
			name: "invalid UUID",
			feedback: &feedback.Feedback{
				Id:           "invalid-uuid",
				Rating:       5,
				RespondentId: testUUID.String(),
				Text:         "Great service!",
				Type:         feedback.FeedbackType_FEEDBACK_GENERAL,
				CreatedAt:    timestamppb.New(now),
			},
			expected:    nil,
			expectedErr: errors.New("invalid UUID length"),
		},
		{
			name: "invalid feedback type",
			feedback: &feedback.Feedback{
				Id:           testUUID.String(),
				Rating:       5,
				RespondentId: testUUID.String(),
				Text:         "Great service!",
				Type:         feedback.FeedbackType(-1), // invalid type
				CreatedAt:    timestamppb.New(now),
			},
			expected:    nil,
			expectedErr: errors.New("unknown feedback type"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := feedback_service.ProtoFeedbackToModel(tt.feedback)

			if tt.expectedErr != nil {
				assert.Error(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.Id, result.Id)
				assert.Equal(t, tt.expected.Type, result.Type)
				assert.Equal(t, tt.expected.Text, result.Text)
			}
		})
	}
}

func TestFeedBackTypeToProto(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       models.FeedbackType
		expected    feedback.FeedbackType
		expectedErr error
	}{
		{"general", models.FeedbackGeneral, feedback.FeedbackType_FEEDBACK_GENERAL, nil},
		{"post", models.FeedbackPost, feedback.FeedbackType_FEEDBACK_POST, nil},
		{"messenger", models.FeedbackMessenger, feedback.FeedbackType_FEEDBACK_MESSENGER, nil},
		{"recommendation", models.FeedbackRecommendation, feedback.FeedbackType_FEEDBACK_RECOMMENDATIONS, nil},
		{"profile", models.FeedbackProfile, feedback.FeedbackType_FEEDBACK_PROFILE, nil},
		{"auth", models.FeedbackAuth, feedback.FeedbackType_FEEDBACK_AUTH, nil},
		{"invalid", "invalid-type", feedback.FeedbackType_FEEDBACK_GENERAL, errors.New("unknown feedback type")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := feedback_service.FeedBackTypeToProto(tt.input)

			if tt.expectedErr != nil {
				assert.Error(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
