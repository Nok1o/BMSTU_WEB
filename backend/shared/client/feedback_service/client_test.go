//go:build unit
// +build unit

package feedback_service

import (
	"context"
	"errors"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
	feedback2 "quickflow/shared/proto/feedback_service"
	"quickflow/shared/proto/feedback_service/mocks"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"quickflow/shared/models"
)

func TestClient_SaveFeedback_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockFeedbackServiceClient(ctrl)
	client := &Client{client: mockClient}

	testUUID := uuid.New()
	now := time.Now()
	feedback := &models.Feedback{
		Id:           testUUID,
		Rating:       5,
		RespondentId: testUUID,
		Text:         "Great service!",
		Type:         models.FeedbackGeneral,
		CreatedAt:    now,
	}

	// Ожидаем вызов с любыми аргументами
	mockClient.EXPECT().SaveFeedback(gomock.Any(), gomock.Any()).
		Return(&feedback2.SaveFeedbackResponse{}, nil)

	err := client.SaveFeedback(context.Background(), feedback)
	assert.NoError(t, err)
}

func TestClient_SaveFeedback_GRPCError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockFeedbackServiceClient(ctrl)
	client := &Client{client: mockClient}

	testUUID := uuid.New()
	feedback := &models.Feedback{
		Id:           testUUID,
		Rating:       5,
		RespondentId: testUUID,
		Text:         "Great service!",
		Type:         models.FeedbackGeneral,
		CreatedAt:    time.Now(),
	}

	expectedErr := errors.New("grpc error")
	mockClient.EXPECT().SaveFeedback(gomock.Any(), gomock.Any()).
		Return(nil, expectedErr)

	err := client.SaveFeedback(context.Background(), feedback)
	assert.EqualError(t, err, expectedErr.Error())
}

func TestClient_SaveFeedback_InvalidFeedbackType(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockFeedbackServiceClient(ctrl)
	client := &Client{client: mockClient}

	testUUID := uuid.New()
	feedback := &models.Feedback{
		Id:           testUUID,
		Rating:       5,
		RespondentId: testUUID,
		Text:         "Great service!",
		Type:         "invalid-type",
		CreatedAt:    time.Now(),
	}

	// Не ожидаем вызовов к GRPC клиенту
	err := client.SaveFeedback(context.Background(), feedback)
	assert.Error(t, err, "unknown feedback type")
}

func TestClient_GetAllFeedbackType_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockFeedbackServiceClient(ctrl)
	client := &Client{client: mockClient}

	testUUID := uuid.New()
	now := time.Now()
	expectedFeedback := []models.Feedback{
		{
			Id:           testUUID,
			Rating:       5,
			RespondentId: testUUID,
			Text:         "Test feedback",
			Type:         models.FeedbackGeneral,
			CreatedAt:    now,
		},
	}

	mockClient.EXPECT().GetAllFeedbackType(gomock.Any(), gomock.Any()).
		Return(&feedback2.GetAllFeedbackTypeResponse{
			Feedback: []*feedback2.Feedback{
				{
					Id:           testUUID.String(),
					Rating:       5,
					RespondentId: testUUID.String(),
					Text:         "Test feedback",
					Type:         feedback2.FeedbackType_FEEDBACK_GENERAL,
					CreatedAt:    timestamppb.New(now),
				},
			},
		}, nil)

	result, err := client.GetAllFeedbackType(context.Background(), models.FeedbackGeneral, now, 10)
	assert.NoError(t, err)
	assert.Equal(t, len(expectedFeedback), len(result))
}

func TestClient_GetAllFeedbackType_EmptyResult(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockFeedbackServiceClient(ctrl)
	client := &Client{client: mockClient}

	mockClient.EXPECT().GetAllFeedbackType(gomock.Any(), gomock.Any()).
		Return(&feedback2.GetAllFeedbackTypeResponse{
			Feedback: []*feedback2.Feedback{},
		}, nil)

	result, err := client.GetAllFeedbackType(context.Background(), models.FeedbackPost, time.Now(), 5)
	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestClient_GetAllFeedbackType_GRPCError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockFeedbackServiceClient(ctrl)
	client := &Client{client: mockClient}

	expectedErr := errors.New("grpc error")
	mockClient.EXPECT().GetAllFeedbackType(gomock.Any(), gomock.Any()).
		Return(nil, expectedErr)

	result, err := client.GetAllFeedbackType(context.Background(), models.FeedbackMessenger, time.Now(), 5)
	assert.EqualError(t, err, expectedErr.Error())
	assert.Nil(t, result)
}

func TestClient_GetAllFeedbackType_InvalidType(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockFeedbackServiceClient(ctrl)
	client := &Client{client: mockClient}

	// Не ожидаем вызовов к GRPC клиенту
	result, err := client.GetAllFeedbackType(context.Background(), "invalid-type", time.Now(), 5)
	assert.Error(t, err, "unknown feedback type")
	assert.Nil(t, result)
}

func TestClient_GetAllFeedbackType_ConversionError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockFeedbackServiceClient(ctrl)
	client := &Client{client: mockClient}

	mockClient.EXPECT().GetAllFeedbackType(gomock.Any(), gomock.Any()).
		Return(&feedback2.GetAllFeedbackTypeResponse{
			Feedback: []*feedback2.Feedback{
				{
					Id:           "invalid-uuid",
					Rating:       5,
					RespondentId: uuid.New().String(),
					Text:         "feedback2 feedback",
					Type:         feedback2.FeedbackType_FEEDBACK_PROFILE,
					CreatedAt:    timestamppb.New(time.Now()),
				},
			},
		}, nil)

	result, err := client.GetAllFeedbackType(context.Background(), models.FeedbackProfile, time.Now(), 5)
	assert.ErrorContains(t, err, "invalid UUID length")
	assert.Nil(t, result)
}

func TestNewFeedbackClient(t *testing.T) {
	mockConn := &grpc.ClientConn{}
	client := NewFeedbackClient(mockConn)
	assert.NotNil(t, client)
}
