//go:build unit
// +build unit

package post_service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	time_config "quickflow/config/time"
	"quickflow/shared/models"
	pb "quickflow/shared/proto/post_service"
	"quickflow/shared/proto/post_service/mocks"
)

func TestFetchCommentsForPost(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockCommentServiceClient(ctrl)
	client := &CommentClient{client: mockClient}

	ctx := context.Background()
	postId := uuid.New()
	now := time.Now()
	testComment := &pb.Comment{
		Id:        uuid.New().String(),
		Text:      "Test comment",
		PostId:    postId.String(),
		UserId:    uuid.New().String(),
		CreatedAt: now.Format(time_config.TimeStampLayout),
		UpdatedAt: now.Format(time_config.TimeStampLayout),
	}

	tests := []struct {
		name        string
		mockSetup   func()
		expected    []models.Comment
		expectError bool
	}{
		{
			name: "Success",
			mockSetup: func() {
				mockClient.EXPECT().FetchCommentsForPost(ctx, &pb.FetchCommentsForPostRequest{
					PostId:      postId.String(),
					NumComments: 10,
					Timestamp:   now.Format(time_config.TimeStampLayout),
				}).Return(&pb.FetchCommentsForPostResponse{
					Comments: []*pb.Comment{testComment},
				}, nil)
			},
			expected: []models.Comment{
				{
					Id:     uuid.MustParse(testComment.Id),
					Text:   testComment.Text,
					PostId: uuid.MustParse(testComment.PostId),
					UserId: uuid.MustParse(testComment.UserId),
				},
			},
		},
		{
			name: "GRPC Error",
			mockSetup: func() {
				mockClient.EXPECT().FetchCommentsForPost(ctx, gomock.Any()).
					Return(nil, errors.New("grpc error"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			result, err := client.FetchCommentsForPost(ctx, postId, 10, now)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, len(tt.expected), len(result))
				if len(result) > 0 {
					assert.Equal(t, tt.expected[0].Id, result[0].Id)
					assert.Equal(t, tt.expected[0].Text, result[0].Text)
				}
			}
		})
	}
}

func TestAddComment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockCommentServiceClient(ctrl)
	client := &CommentClient{client: mockClient}

	ctx := context.Background()
	testComment := models.Comment{
		Id:   uuid.New(),
		Text: "Test comment",
	}
	testProtoComment := ModelCommentToProto(&testComment)

	tests := []struct {
		name        string
		mockSetup   func()
		expected    *models.Comment
		expectError bool
	}{
		{
			name: "Success",
			mockSetup: func() {
				mockClient.EXPECT().AddComment(ctx, &pb.AddCommentRequest{
					Comment: testProtoComment,
				}).Return(&pb.AddCommentResponse{
					Comment: testProtoComment,
				}, nil)
			},
			expected: &testComment,
		},
		{
			name: "GRPC Error",
			mockSetup: func() {
				mockClient.EXPECT().AddComment(ctx, gomock.Any()).
					Return(nil, errors.New("grpc error"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			result, err := client.AddComment(ctx, testComment)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.Id, result.Id)
				assert.Equal(t, tt.expected.Text, result.Text)
			}
		})
	}
}

func TestDeleteComment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockCommentServiceClient(ctrl)
	client := &CommentClient{client: mockClient}

	ctx := context.Background()
	userId := uuid.New()
	commentId := uuid.New()

	tests := []struct {
		name        string
		mockSetup   func()
		expectError bool
	}{
		{
			name: "Success",
			mockSetup: func() {
				mockClient.EXPECT().DeleteComment(ctx, &pb.DeleteCommentRequest{
					CommentId: commentId.String(),
					UserId:    userId.String(),
				}).Return(&pb.DeleteCommentResponse{}, nil)
			},
		},
		{
			name: "GRPC Error",
			mockSetup: func() {
				mockClient.EXPECT().DeleteComment(ctx, gomock.Any()).
					Return(nil, errors.New("grpc error"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			err := client.DeleteComment(ctx, userId, commentId)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUpdateComment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockCommentServiceClient(ctrl)
	client := &CommentClient{client: mockClient}

	ctx := context.Background()
	userId := uuid.New()
	commentId := uuid.New()
	commentUpdate := models.CommentUpdate{
		Id:   commentId,
		Text: "Updated comment",
	}
	now := time.Now()
	postId := uuid.New()
	testProtoComment := &pb.Comment{
		Id:        commentId.String(),
		Text:      "Updated comment",
		PostId:    postId.String(),
		UserId:    uuid.New().String(),
		CreatedAt: now.Format(time_config.TimeStampLayout),
		UpdatedAt: now.Format(time_config.TimeStampLayout),
	}

	tests := []struct {
		name        string
		mockSetup   func()
		expected    *models.Comment
		expectError bool
	}{
		{
			name: "Success",
			mockSetup: func() {
				mockClient.EXPECT().UpdateComment(ctx, &pb.UpdateCommentRequest{
					Comment: ModelCommentUpdateToProto(&commentUpdate),
					UserId:  userId.String(),
				}).Return(&pb.UpdateCommentResponse{
					Comment: testProtoComment,
				}, nil)
			},
			expected: &models.Comment{
				Id:     commentUpdate.Id,
				Text:   commentUpdate.Text,
				PostId: uuid.MustParse(postId.String()),
				UserId: uuid.MustParse(userId.String()),
			},
		},
		{
			name: "GRPC Error",
			mockSetup: func() {
				mockClient.EXPECT().UpdateComment(ctx, gomock.Any()).
					Return(nil, errors.New("grpc error"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			result, err := client.UpdateComment(ctx, commentUpdate, userId)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.Id, result.Id)
				assert.Equal(t, tt.expected.Text, result.Text)
			}
		})
	}
}

func TestLikeComment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockCommentServiceClient(ctrl)
	client := &CommentClient{client: mockClient}

	ctx := context.Background()
	userId := uuid.New()
	commentId := uuid.New()

	tests := []struct {
		name        string
		mockSetup   func()
		expectError bool
	}{
		{
			name: "Success",
			mockSetup: func() {
				mockClient.EXPECT().LikeComment(ctx, &pb.LikeCommentRequest{
					CommentId: commentId.String(),
					UserId:    userId.String(),
				}).Return(&pb.LikeCommentResponse{}, nil)
			},
		},
		{
			name: "GRPC Error",
			mockSetup: func() {
				mockClient.EXPECT().LikeComment(ctx, gomock.Any()).
					Return(nil, errors.New("grpc error"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			err := client.LikeComment(ctx, commentId, userId)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUnlikeComment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockCommentServiceClient(ctrl)
	client := &CommentClient{client: mockClient}

	ctx := context.Background()
	userId := uuid.New()
	commentId := uuid.New()

	tests := []struct {
		name        string
		mockSetup   func()
		expectError bool
	}{
		{
			name: "Success",
			mockSetup: func() {
				mockClient.EXPECT().UnlikeComment(ctx, &pb.UnlikeCommentRequest{
					CommentId: commentId.String(),
					UserId:    userId.String(),
				}).Return(&pb.UnlikeCommentResponse{}, nil)
			},
		},
		{
			name: "GRPC Error",
			mockSetup: func() {
				mockClient.EXPECT().UnlikeComment(ctx, gomock.Any()).
					Return(nil, errors.New("grpc error"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			err := client.UnlikeComment(ctx, commentId, userId)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetComment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockCommentServiceClient(ctrl)
	client := &CommentClient{client: mockClient}

	ctx := context.Background()
	userId := uuid.New()
	commentId := uuid.New()
	postId := uuid.New()
	now := time.Now()
	testComment := &pb.Comment{
		Id:        commentId.String(),
		Text:      "Test comment",
		PostId:    postId.String(),
		UserId:    userId.String(),
		CreatedAt: now.Format(time_config.TimeStampLayout),
		UpdatedAt: now.Format(time_config.TimeStampLayout),
	}

	tests := []struct {
		name        string
		mockSetup   func()
		expected    *models.Comment
		expectError bool
	}{
		{
			name: "Success",
			mockSetup: func() {
				mockClient.EXPECT().GetComment(ctx, &pb.GetCommentRequest{
					CommentId: commentId.String(),
					UserId:    userId.String(),
				}).Return(&pb.GetCommentResponse{
					Comment: testComment,
				}, nil)
			},
			expected: &models.Comment{
				Id:   commentId,
				Text: "Test comment",
			},
		},
		{
			name: "GRPC Error",
			mockSetup: func() {
				mockClient.EXPECT().GetComment(ctx, gomock.Any()).
					Return(nil, errors.New("grpc error"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			result, err := client.GetComment(ctx, commentId, userId)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.Id, result.Id)
				assert.Equal(t, tt.expected.Text, result.Text)
			}
		})
	}
}

func TestGetLastPostComment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockCommentServiceClient(ctrl)
	client := &CommentClient{client: mockClient}

	ctx := context.Background()
	userId := uuid.New()
	commentId := uuid.New()
	postId := uuid.New()
	now := time.Now()
	testComment := &pb.Comment{
		Id:        commentId.String(),
		Text:      "Last comment",
		PostId:    postId.String(),
		UserId:    userId.String(),
		CreatedAt: now.Format(time_config.TimeStampLayout),
		UpdatedAt: now.Format(time_config.TimeStampLayout),
	}

	tests := []struct {
		name        string
		mockSetup   func()
		expected    *models.Comment
		expectError bool
	}{
		{
			name: "Success",
			mockSetup: func() {
				mockClient.EXPECT().GetLastPostComment(ctx, &pb.GetLastPostCommentRequest{
					PostId: postId.String(),
				}).Return(&pb.GetLastPostCommentResponse{
					Comment: testComment,
				}, nil)
			},
			expected: &models.Comment{
				Id:   uuid.MustParse(testComment.Id),
				Text: "Last comment",
			},
		},
		{
			name: "GRPC Error",
			mockSetup: func() {
				mockClient.EXPECT().GetLastPostComment(ctx, gomock.Any()).
					Return(nil, errors.New("grpc error"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			result, err := client.GetLastPostComment(ctx, postId)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.Id, result.Id)
				assert.Equal(t, tt.expected.Text, result.Text)
			}
		})
	}
}
