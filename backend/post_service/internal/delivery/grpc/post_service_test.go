//go:build unit
// +build unit

package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"quickflow/post_service/internal/delivery/grpc/mocks"
	"quickflow/shared/models"
	pb "quickflow/shared/proto/post_service"
)

func TestPostServiceServer_AddPost(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPostUC := mocks.NewMockPostUseCase(ctrl)
	mockUserUC := mocks.NewMockUserUseCase(ctrl)
	server := NewPostServiceServer(mockPostUC, mockUserUC)

	testPost := &pb.Post{
		Id:          uuid.New().String(),
		CreatorId:   uuid.New().String(),
		Description: "Test post",
		CreatedAt:   timestamppb.Now(),
	}

	tests := []struct {
		name        string
		setupMock   func()
		req         *pb.AddPostRequest
		expected    *pb.Post
		expectedErr error
	}{
		{
			name: "successful add post",
			setupMock: func() {
				mockPostUC.EXPECT().AddPost(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, post models.Post) (*models.Post, error) {
						assert.Equal(t, testPost.Description, post.Desc)
						return &models.Post{
							Id:        uuid.MustParse(testPost.Id),
							CreatorId: uuid.MustParse(testPost.CreatorId),
							Desc:      testPost.Description,
						}, nil
					})
			},
			req: &pb.AddPostRequest{
				Post: testPost,
			},
			expected: testPost,
		},
		{
			name:      "invalid post data",
			setupMock: func() {},
			req: &pb.AddPostRequest{
				Post: &pb.Post{
					Id: "invalid-uuid",
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "invalid UUID length: 11"),
		},
		{
			name: "use case error",
			setupMock: func() {
				mockPostUC.EXPECT().AddPost(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("database error"))
			},
			req: &pb.AddPostRequest{
				Post: testPost,
			},
			expectedErr: status.Error(codes.Internal, "Failed to add post:: database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMock != nil {
				tt.setupMock()
			}

			resp, err := server.AddPost(context.Background(), tt.req)

			if tt.expectedErr != nil {
				assert.Error(t, tt.expectedErr, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.Id, resp.Post.Id)
				assert.Equal(t, tt.expected.Description, resp.Post.Description)
			}
		})
	}
}

func TestPostServiceServer_DeletePost(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPostUC := mocks.NewMockPostUseCase(ctrl)
	mockUserUC := mocks.NewMockUserUseCase(ctrl)
	server := NewPostServiceServer(mockPostUC, mockUserUC)

	postId := uuid.New()
	userId := uuid.New()

	tests := []struct {
		name        string
		setupMock   func()
		req         *pb.DeletePostRequest
		expected    bool
		expectedErr error
	}{
		{
			name: "successful delete",
			setupMock: func() {
				mockPostUC.EXPECT().
					DeletePost(gomock.Any(), userId, postId).
					Return(nil)
			},
			req: &pb.DeletePostRequest{
				PostId: postId.String(),
				UserId: userId.String(),
			},
			expected: true,
		},
		{
			name:      "invalid post id",
			setupMock: func() {},
			req: &pb.DeletePostRequest{
				PostId: "invalid",
				UserId: userId.String(),
			},
			expectedErr: status.Error(codes.InvalidArgument, "Invalid post ID::"),
		},
		{
			name: "use case error",
			setupMock: func() {
				mockPostUC.EXPECT().
					DeletePost(gomock.Any(), userId, postId).
					Return(errors.New("not found"))
			},
			req: &pb.DeletePostRequest{
				PostId: postId.String(),
				UserId: userId.String(),
			},
			expectedErr: status.Error(codes.Internal, "Failed to delete post:: not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMock != nil {
				tt.setupMock()
			}

			resp, err := server.DeletePost(context.Background(), tt.req)

			if tt.expectedErr != nil {
				assert.Error(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, resp.Success)
			}
		})
	}
}

func TestPostServiceServer_UpdatePost(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPostUC := mocks.NewMockPostUseCase(ctrl)
	mockUserUC := mocks.NewMockUserUseCase(ctrl)
	server := NewPostServiceServer(mockPostUC, mockUserUC)

	postId := uuid.New()
	userId := uuid.New()

	tests := []struct {
		name        string
		setupMock   func()
		req         *pb.UpdatePostRequest
		expected    *pb.Post
		expectedErr error
	}{
		{
			name: "successful update",
			setupMock: func() {
				mockPostUC.EXPECT().
					UpdatePost(gomock.Any(), gomock.Any(), userId).
					Return(&models.Post{
						Id:        postId,
						Desc:      "Updated post",
						CreatorId: userId,
					}, nil)
			},
			req: &pb.UpdatePostRequest{
				Post: &pb.PostUpdate{
					Id:          postId.String(),
					Description: "Updated post",
				},
				UserId: userId.String(),
			},
			expected: &pb.Post{
				Id:          postId.String(),
				Description: "Updated post",
				CreatorId:   userId.String(),
			},
		},
		{
			name:      "invalid post id",
			setupMock: func() {},
			req: &pb.UpdatePostRequest{
				Post: &pb.PostUpdate{
					Id: "invalid",
				},
				UserId: userId.String(),
			},
			expectedErr: status.Error(codes.InvalidArgument, "Failed to convert update payload::"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMock != nil {
				tt.setupMock()
			}

			resp, err := server.UpdatePost(context.Background(), tt.req)

			if tt.expectedErr != nil {
				assert.Error(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.Id, resp.Post.Id)
				assert.Equal(t, tt.expected.Description, resp.Post.Description)
			}
		})
	}
}

func TestPostServiceServer_FetchMethods(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPostUC := mocks.NewMockPostUseCase(ctrl)
	mockUserUC := mocks.NewMockUserUseCase(ctrl)
	server := NewPostServiceServer(mockPostUC, mockUserUC)

	userId := uuid.New()
	timestamp := timestamppb.Now()

	tests := []struct {
		name        string
		method      string
		setupMock   func()
		req         interface{}
		expectedLen int
		expectedErr error
	}{
		{
			name:   "successful fetch feed",
			method: "FetchFeed",
			setupMock: func() {
				mockPostUC.EXPECT().
					FetchFeed(gomock.Any(), userId, 10, timestamp.AsTime()).
					Return([]models.Post{
						{Id: uuid.New(), Desc: "Post 1"},
						{Id: uuid.New(), Desc: "Post 2"},
					}, nil)
			},
			req: &pb.FetchFeedRequest{
				UserId:    userId.String(),
				NumPosts:  10,
				Timestamp: timestamp,
			},
			expectedLen: 2,
		},
		{
			name:   "successful fetch recommendations",
			method: "FetchRecommendations",
			setupMock: func() {
				mockPostUC.EXPECT().
					FetchRecommendations(gomock.Any(), userId, 10, timestamp.AsTime()).
					Return([]models.Post{
						{Id: uuid.New(), Desc: "Rec 1"},
					}, nil)
			},
			req: &pb.FetchRecommendationsRequest{
				UserId:    userId.String(),
				NumPosts:  10,
				Timestamp: timestamp,
			},
			expectedLen: 1,
		},
		{
			name:   "successful fetch user posts",
			method: "FetchUserPosts",
			setupMock: func() {
				mockPostUC.EXPECT().
					FetchUserPosts(gomock.Any(), userId, userId, 10, timestamp.AsTime()).
					Return([]models.Post{
						{Id: uuid.New(), Desc: "User Post 1"},
					}, nil)
			},
			req: &pb.FetchUserPostsRequest{
				UserId:      userId.String(),
				RequesterId: userId.String(),
				NumPosts:    10,
				Timestamp:   timestamp,
			},
			expectedLen: 1,
		},
		{
			name:      "invalid user id in fetch feed",
			method:    "FetchFeed",
			setupMock: func() {},
			req: &pb.FetchFeedRequest{
				UserId: "invalid",
			},
			expectedErr: status.Error(codes.InvalidArgument, "Invalid user ID::"),
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
			case "FetchFeed":
				resp, err = server.FetchFeed(context.Background(), tt.req.(*pb.FetchFeedRequest))
			case "FetchRecommendations":
				resp, err = server.FetchRecommendations(context.Background(), tt.req.(*pb.FetchRecommendationsRequest))
			case "FetchUserPosts":
				resp, err = server.FetchUserPosts(context.Background(), tt.req.(*pb.FetchUserPostsRequest))
			}

			if tt.expectedErr != nil {
				assert.Error(t, tt.expectedErr, err)
			} else {
				assert.NoError(t, err)
				switch r := resp.(type) {
				case *pb.FetchFeedResponse:
					assert.Equal(t, tt.expectedLen, len(r.Posts))
				case *pb.FetchRecommendationsResponse:
					assert.Equal(t, tt.expectedLen, len(r.Posts))
				case *pb.FetchUserPostsResponse:
					assert.Equal(t, tt.expectedLen, len(r.Posts))
				}
			}
		})
	}
}

func TestPostServiceServer_LikeUnlikePost(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPostUC := mocks.NewMockPostUseCase(ctrl)
	mockUserUC := mocks.NewMockUserUseCase(ctrl)
	server := NewPostServiceServer(mockPostUC, mockUserUC)

	postId := uuid.New()
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
			method: "LikePost",
			setupMock: func() {
				mockPostUC.EXPECT().
					LikePost(gomock.Any(), postId, userId).
					Return(nil)
			},
			req: &pb.LikePostRequest{
				PostId: postId.String(),
				UserId: userId.String(),
			},
			expected: true,
		},
		{
			name:   "successful unlike",
			method: "UnlikePost",
			setupMock: func() {
				mockPostUC.EXPECT().
					UnlikePost(gomock.Any(), postId, userId).
					Return(nil)
			},
			req: &pb.UnlikePostRequest{
				PostId: postId.String(),
				UserId: userId.String(),
			},
			expected: true,
		},
		{
			name:      "invalid post id in like",
			method:    "LikePost",
			setupMock: func() {},
			req: &pb.LikePostRequest{
				PostId: "invalid",
				UserId: userId.String(),
			},
			expectedErr: status.Error(codes.InvalidArgument, "Invalid post ID::"),
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
			case "LikePost":
				resp, err = server.LikePost(context.Background(), tt.req.(*pb.LikePostRequest))
			case "UnlikePost":
				resp, err = server.UnlikePost(context.Background(), tt.req.(*pb.UnlikePostRequest))
			}

			if tt.expectedErr != nil {
				assert.Error(t, tt.expectedErr, err)
			} else {
				assert.NoError(t, err)
				switch r := resp.(type) {
				case *pb.LikePostResponse:
					assert.Equal(t, tt.expected, r.Success)
				case *pb.UnlikePostResponse:
					assert.Equal(t, tt.expected, r.Success)
				}
			}
		})
	}
}

func TestPostServiceServer_GetPost(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPostUC := mocks.NewMockPostUseCase(ctrl)
	mockUserUC := mocks.NewMockUserUseCase(ctrl)
	server := NewPostServiceServer(mockPostUC, mockUserUC)

	postId := uuid.New()
	userId := uuid.New()

	tests := []struct {
		name        string
		setupMock   func()
		req         *pb.GetPostRequest
		expected    *pb.Post
		expectedErr error
	}{
		{
			name: "successful get",
			setupMock: func() {
				mockPostUC.EXPECT().
					GetPost(gomock.Any(), postId, userId).
					Return(&models.Post{
						Id:        postId,
						Desc:      "Test post",
						CreatorId: userId,
					}, nil)
			},
			req: &pb.GetPostRequest{
				PostId: postId.String(),
				UserId: userId.String(),
			},
			expected: &pb.Post{
				Id:          postId.String(),
				Description: "Test post",
				CreatorId:   userId.String(),
			},
		},
		{
			name:      "invalid post id",
			setupMock: func() {},
			req: &pb.GetPostRequest{
				PostId: "invalid",
				UserId: userId.String(),
			},
			expectedErr: status.Error(codes.InvalidArgument, "Invalid post ID::"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMock != nil {
				tt.setupMock()
			}

			resp, err := server.GetPost(context.Background(), tt.req)

			if tt.expectedErr != nil {
				assert.Error(t, tt.expectedErr, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.Id, resp.Post.Id)
				assert.Equal(t, tt.expected.Description, resp.Post.Description)
			}
		})
	}
}
