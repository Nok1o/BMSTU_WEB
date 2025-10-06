//go:build unit
// +build unit

package grpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"quickflow/config/time"
	"quickflow/post_service/internal/delivery/grpc/mocks"
	"quickflow/shared/models"
	pb "quickflow/shared/proto/post_service"
)

func TestCommentServiceServer_AddComment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCommentUC := mocks.NewMockCommentUseCase(ctrl)
	mockUserUC := mocks.NewMockUserUseCase(ctrl)
	server := NewCommentServiceServer(mockCommentUC, mockUserUC)

	testComment := &pb.Comment{
		Id:        uuid.New().String(),
		PostId:    uuid.New().String(),
		UserId:    uuid.New().String(),
		Text:      "Test comment",
		CreatedAt: time.Now().Format(time_config.TimeStampLayout),
		UpdatedAt: time.Now().Format(time_config.TimeStampLayout),
	}

	tests := []struct {
		name        string
		setupMock   func()
		req         *pb.AddCommentRequest
		expected    *pb.Comment
		expectedErr error
	}{
		{
			name: "successful add comment",
			setupMock: func() {
				mockCommentUC.EXPECT().AddComment(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, comment models.Comment) (*models.Comment, error) {
						assert.Equal(t, testComment.Text, comment.Text)
						return &models.Comment{
							Id:     uuid.MustParse(testComment.Id),
							PostId: uuid.MustParse(testComment.PostId),
							UserId: uuid.MustParse(testComment.UserId),
							Text:   testComment.Text,
						}, nil
					})
			},
			req: &pb.AddCommentRequest{
				Comment: testComment,
			},
			expected: testComment,
		},
		{
			name:      "invalid comment data",
			setupMock: func() {},
			req: &pb.AddCommentRequest{
				Comment: &pb.Comment{
					Id: "invalid-uuid",
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "invalid UUID length: 11"),
		},
		{
			name: "use case error",
			setupMock: func() {
				mockCommentUC.EXPECT().AddComment(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("database error"))
			},
			req: &pb.AddCommentRequest{
				Comment: testComment,
			},
			expectedErr: status.Error(codes.Internal, "Failed to add comment: database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMock != nil {
				tt.setupMock()
			}

			resp, err := server.AddComment(context.Background(), tt.req)

			if tt.expectedErr != nil {
				assert.Error(t, tt.expectedErr, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.Id, resp.Comment.Id)
				assert.Equal(t, tt.expected.Text, resp.Comment.Text)
			}
		})
	}
}

func TestCommentServiceServer_DeleteComment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCommentUC := mocks.NewMockCommentUseCase(ctrl)
	mockUserUC := mocks.NewMockUserUseCase(ctrl)
	server := NewCommentServiceServer(mockCommentUC, mockUserUC)

	commentId := uuid.New()
	userId := uuid.New()

	tests := []struct {
		name        string
		setupMock   func()
		req         *pb.DeleteCommentRequest
		expected    bool
		expectedErr error
	}{
		{
			name: "successful delete",
			setupMock: func() {
				mockCommentUC.EXPECT().
					DeleteComment(gomock.Any(), userId, commentId).
					Return(nil)
			},
			req: &pb.DeleteCommentRequest{
				CommentId: commentId.String(),
				UserId:    userId.String(),
			},
			expected: true,
		},
		{
			name:      "invalid comment id",
			setupMock: func() {},
			req: &pb.DeleteCommentRequest{
				CommentId: "invalid",
				UserId:    userId.String(),
			},
			expectedErr: status.Error(codes.InvalidArgument, "Invalid comment ID:"),
		},
		{
			name: "use case error",
			setupMock: func() {
				mockCommentUC.EXPECT().
					DeleteComment(gomock.Any(), userId, commentId).
					Return(errors.New("not found"))
			},
			req: &pb.DeleteCommentRequest{
				CommentId: commentId.String(),
				UserId:    userId.String(),
			},
			expectedErr: status.Error(codes.Internal, "Failed to delete comment: not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMock != nil {
				tt.setupMock()
			}

			resp, err := server.DeleteComment(context.Background(), tt.req)

			if tt.expectedErr != nil {
				assert.Error(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, resp.Success)
			}
		})
	}
}

func TestCommentServiceServer_UpdateComment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCommentUC := mocks.NewMockCommentUseCase(ctrl)
	mockUserUC := mocks.NewMockUserUseCase(ctrl)
	server := NewCommentServiceServer(mockCommentUC, mockUserUC)

	commentId := uuid.New()
	userId := uuid.New()

	tests := []struct {
		name        string
		setupMock   func()
		req         *pb.UpdateCommentRequest
		expected    *pb.Comment
		expectedErr error
	}{
		{
			name: "successful update",
			setupMock: func() {
				mockCommentUC.EXPECT().
					UpdateComment(gomock.Any(), gomock.Any(), userId).
					Return(&models.Comment{
						Id:     commentId,
						Text:   "Updated text",
						UserId: userId,
					}, nil)
			},
			req: &pb.UpdateCommentRequest{
				Comment: &pb.CommentUpdate{
					Id:   commentId.String(),
					Text: "Updated text",
				},
				UserId: userId.String(),
			},
			expected: &pb.Comment{
				Id:     commentId.String(),
				Text:   "Updated text",
				UserId: userId.String(),
			},
		},
		{
			name:      "invalid comment id",
			setupMock: func() {},
			req: &pb.UpdateCommentRequest{
				Comment: &pb.CommentUpdate{
					Id: "invalid",
				},
				UserId: userId.String(),
			},
			expectedErr: status.Error(codes.InvalidArgument, "Failed to convert update payload:"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMock != nil {
				tt.setupMock()
			}

			resp, err := server.UpdateComment(context.Background(), tt.req)

			if tt.expectedErr != nil {
				assert.Error(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.Id, resp.Comment.Id)
				assert.Equal(t, tt.expected.Text, resp.Comment.Text)
			}
		})
	}
}

func TestCommentServiceServer_FetchCommentsForPost(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCommentUC := mocks.NewMockCommentUseCase(ctrl)
	mockUserUC := mocks.NewMockUserUseCase(ctrl)
	server := NewCommentServiceServer(mockCommentUC, mockUserUC)

	postId := uuid.New()
	timestamp := time.Now()

	tests := []struct {
		name        string
		setupMock   func()
		req         *pb.FetchCommentsForPostRequest
		expectedLen int
		expectedErr error
	}{
		{
			name: "successful fetch",
			setupMock: func() {
				mockCommentUC.EXPECT().
					FetchCommentsForPost(gomock.Any(), postId, 10, gomock.Any()).
					Return([]models.Comment{
						{Id: uuid.New(), Text: "Comment 1"},
						{Id: uuid.New(), Text: "Comment 2"},
					}, nil)
			},
			req: &pb.FetchCommentsForPostRequest{
				PostId:      postId.String(),
				NumComments: 10,
				Timestamp:   timestamp.Format(time_config.TimeStampLayout),
			},
			expectedLen: 2,
		},
		{
			name:      "invalid post id",
			setupMock: func() {},
			req: &pb.FetchCommentsForPostRequest{
				PostId: "invalid",
			},
			expectedErr: status.Error(codes.InvalidArgument, "Invalid post ID:"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMock != nil {
				tt.setupMock()
			}

			resp, err := server.FetchCommentsForPost(context.Background(), tt.req)

			if tt.expectedErr != nil {
				assert.Error(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedLen, len(resp.Comments))
			}
		})
	}
}

func TestCommentServiceServer_LikeUnlikeComment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCommentUC := mocks.NewMockCommentUseCase(ctrl)
	mockUserUC := mocks.NewMockUserUseCase(ctrl)
	server := NewCommentServiceServer(mockCommentUC, mockUserUC)

	commentId := uuid.New()
	userId := uuid.New()

	tests := []struct {
		name        string
		method      string
		setupMock   func()
		req         interface{}
		expected    bool
		expectedErr error
	}{
		{
			name:   "successful like",
			method: "LikeComment",
			setupMock: func() {
				mockCommentUC.EXPECT().
					LikeComment(gomock.Any(), commentId, userId).
					Return(nil)
			},
			req: &pb.LikeCommentRequest{
				CommentId: commentId.String(),
				UserId:    userId.String(),
			},
			expected: true,
		},
		{
			name:   "successful unlike",
			method: "UnlikeComment",
			setupMock: func() {
				mockCommentUC.EXPECT().
					UnlikeComment(gomock.Any(), commentId, userId).
					Return(nil)
			},
			req: &pb.UnlikeCommentRequest{
				CommentId: commentId.String(),
				UserId:    userId.String(),
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMock != nil {
				tt.setupMock()
			}

			var resp interface{}
			var err error

			switch tt.method {
			case "LikeComment":
				resp, err = server.LikeComment(context.Background(), tt.req.(*pb.LikeCommentRequest))
			case "UnlikeComment":
				resp, err = server.UnlikeComment(context.Background(), tt.req.(*pb.UnlikeCommentRequest))
			}

			if tt.expectedErr != nil {
				assert.Equal(t, tt.expectedErr, err)
			} else {
				assert.NoError(t, err)
				switch r := resp.(type) {
				case *pb.LikeCommentResponse:
					assert.Equal(t, tt.expected, r.Success)
				case *pb.UnlikeCommentResponse:
					assert.Equal(t, tt.expected, r.Success)
				}
			}
		})
	}
}

func TestCommentServiceServer_GetComment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCommentUC := mocks.NewMockCommentUseCase(ctrl)
	mockUserUC := mocks.NewMockUserUseCase(ctrl)
	server := NewCommentServiceServer(mockCommentUC, mockUserUC)

	commentId := uuid.New()
	userId := uuid.New()

	tests := []struct {
		name        string
		setupMock   func()
		req         *pb.GetCommentRequest
		expected    *pb.Comment
		expectedErr error
	}{
		{
			name: "successful get",
			setupMock: func() {
				mockCommentUC.EXPECT().
					GetComment(gomock.Any(), commentId, userId).
					Return(&models.Comment{
						Id:     commentId,
						Text:   "Test comment",
						UserId: userId,
					}, nil)
			},
			req: &pb.GetCommentRequest{
				CommentId: commentId.String(),
				UserId:    userId.String(),
			},
			expected: &pb.Comment{
				Id:     commentId.String(),
				Text:   "Test comment",
				UserId: userId.String(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMock != nil {
				tt.setupMock()
			}

			resp, err := server.GetComment(context.Background(), tt.req)

			if tt.expectedErr != nil {
				assert.Equal(t, tt.expectedErr, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.Id, resp.Comment.Id)
				assert.Equal(t, tt.expected.Text, resp.Comment.Text)
			}
		})
	}
}

func TestCommentServiceServer_GetLastPostComment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCommentUC := mocks.NewMockCommentUseCase(ctrl)
	mockUserUC := mocks.NewMockUserUseCase(ctrl)
	server := NewCommentServiceServer(mockCommentUC, mockUserUC)

	postId := uuid.New()
	commentId := uuid.New()

	tests := []struct {
		name        string
		setupMock   func()
		req         *pb.GetLastPostCommentRequest
		expected    *pb.Comment
		expectedErr error
	}{
		{
			name: "successful get last comment",
			setupMock: func() {
				mockCommentUC.EXPECT().
					GetLastPostComment(gomock.Any(), postId).
					Return(&models.Comment{
						Id:     commentId,
						PostId: postId,
						Text:   "Last comment",
					}, nil)
			},
			req: &pb.GetLastPostCommentRequest{
				PostId: postId.String(),
			},
			expected: &pb.Comment{
				Id:     commentId.String(),
				PostId: postId.String(),
				Text:   "Last comment",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMock != nil {
				tt.setupMock()
			}

			resp, err := server.GetLastPostComment(context.Background(), tt.req)

			if tt.expectedErr != nil {
				assert.Equal(t, tt.expectedErr, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.Id, resp.Comment.Id)
				assert.Equal(t, tt.expected.Text, resp.Comment.Text)
			}
		})
	}
}
