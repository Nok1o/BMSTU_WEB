//go:build unit
// +build unit

package post_service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"

	"quickflow/shared/models"
	pb "quickflow/shared/proto/post_service"
)

// MockPostServiceClient is a mock of PostServiceClient interface
type MockPostServiceClient struct {
	mock.Mock
	pb.PostServiceClient
}

func (m *MockPostServiceClient) AddPost(ctx context.Context, in *pb.AddPostRequest, opts ...grpc.CallOption) (*pb.AddPostResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*pb.AddPostResponse), args.Error(1)
}

func (m *MockPostServiceClient) DeletePost(ctx context.Context, in *pb.DeletePostRequest, opts ...grpc.CallOption) (*pb.DeletePostResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*pb.DeletePostResponse), args.Error(1)
}

func (m *MockPostServiceClient) UpdatePost(ctx context.Context, in *pb.UpdatePostRequest, opts ...grpc.CallOption) (*pb.UpdatePostResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*pb.UpdatePostResponse), args.Error(1)
}

func (m *MockPostServiceClient) FetchFeed(ctx context.Context, in *pb.FetchFeedRequest, opts ...grpc.CallOption) (*pb.FetchFeedResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*pb.FetchFeedResponse), args.Error(1)
}

func (m *MockPostServiceClient) FetchRecommendations(ctx context.Context, in *pb.FetchRecommendationsRequest, opts ...grpc.CallOption) (*pb.FetchRecommendationsResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*pb.FetchRecommendationsResponse), args.Error(1)
}

func (m *MockPostServiceClient) FetchUserPosts(ctx context.Context, in *pb.FetchUserPostsRequest, opts ...grpc.CallOption) (*pb.FetchUserPostsResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*pb.FetchUserPostsResponse), args.Error(1)
}

func (m *MockPostServiceClient) LikePost(ctx context.Context, in *pb.LikePostRequest, opts ...grpc.CallOption) (*pb.LikePostResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*pb.LikePostResponse), args.Error(1)
}

func (m *MockPostServiceClient) UnlikePost(ctx context.Context, in *pb.UnlikePostRequest, opts ...grpc.CallOption) (*pb.UnlikePostResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*pb.UnlikePostResponse), args.Error(1)
}

func (m *MockPostServiceClient) GetPost(ctx context.Context, in *pb.GetPostRequest, opts ...grpc.CallOption) (*pb.GetPostResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*pb.GetPostResponse), args.Error(1)
}

func TestPostServiceClient_AddPost(t *testing.T) {
	now := time.Now()
	postID := uuid.New()
	creatorID := uuid.New()

	tests := []struct {
		name        string
		inputPost   models.Post
		mockResp    *pb.AddPostResponse
		mockErr     error
		expected    *models.Post
		expectedErr bool
	}{
		{
			name: "successful post creation",
			inputPost: models.Post{
				CreatorId:   creatorID,
				CreatorType: models.PostUser,
				Desc:        "Test post",
				CreatedAt:   now,
				UpdatedAt:   now,
			},
			mockResp: &pb.AddPostResponse{
				Post: &pb.Post{
					Id:           postID.String(),
					CreatorId:    creatorID.String(),
					CreatorType:  string(models.PostUser),
					Description:  "Test post",
					CreatedAt:    timestamppb.New(now),
					UpdatedAt:    timestamppb.New(now),
					LikeCount:    0,
					RepostCount:  0,
					CommentCount: 0,
					IsRepost:     false,
					IsLiked:      false,
				},
			},
			expected: &models.Post{
				Id:           postID,
				CreatorId:    creatorID,
				CreatorType:  models.PostUser,
				Desc:         "Test post",
				CreatedAt:    now,
				UpdatedAt:    now,
				LikeCount:    0,
				RepostCount:  0,
				CommentCount: 0,
				IsRepost:     false,
				IsLiked:      false,
			},
		},
		{
			name: "empty description",
			inputPost: models.Post{
				CreatorId:   creatorID,
				CreatorType: models.PostUser,
				Desc:        "",
				CreatedAt:   now,
				UpdatedAt:   now,
			},
			mockResp: &pb.AddPostResponse{
				Post: &pb.Post{
					Id:           postID.String(),
					CreatorId:    creatorID.String(),
					CreatorType:  string(models.PostUser),
					Description:  "",
					CreatedAt:    timestamppb.New(now),
					UpdatedAt:    timestamppb.New(now),
					LikeCount:    0,
					RepostCount:  0,
					CommentCount: 0,
					IsRepost:     false,
					IsLiked:      false,
				},
			},
			expected: &models.Post{
				Id:           postID,
				CreatorId:    creatorID,
				CreatorType:  models.PostUser,
				Desc:         "",
				CreatedAt:    now,
				UpdatedAt:    now,
				LikeCount:    0,
				RepostCount:  0,
				CommentCount: 0,
				IsRepost:     false,
				IsLiked:      false,
			},
		},
		{
			name: "grpc error",
			inputPost: models.Post{
				CreatorId:   creatorID,
				CreatorType: models.PostUser,
				Desc:        "Test post",
			},
			mockErr:     errors.New("grpc error"),
			expectedErr: true,
		},
		{
			name: "invalid uuid in response",
			inputPost: models.Post{
				CreatorId:   creatorID,
				CreatorType: models.PostUser,
				Desc:        "Test post",
			},
			mockResp: &pb.AddPostResponse{
				Post: &pb.Post{
					Id:          "invalid-uuid",
					CreatorId:   creatorID.String(),
					CreatorType: string(models.PostUser),
					Description: "Test post",
					CreatedAt:   timestamppb.New(now),
					UpdatedAt:   timestamppb.New(now),
				},
			},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockPostServiceClient)
			client := &PostServiceClient{client: mockClient}

			if tt.mockErr != nil || tt.mockResp != nil {
				mockClient.On("AddPost", mock.Anything, mock.Anything, mock.Anything).Return(tt.mockResp, tt.mockErr)
			}

			result, err := client.AddPost(context.Background(), tt.inputPost)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.Id, result.Id)
				assert.Equal(t, tt.expected.CreatorId, result.CreatorId)
				assert.Equal(t, tt.expected.CreatorType, result.CreatorType)
				assert.Equal(t, tt.expected.Desc, result.Desc)
				assert.Equal(t, tt.expected.IsRepost, result.IsRepost)
				assert.Equal(t, tt.expected.IsLiked, result.IsLiked)
				assert.Equal(t, tt.expected.RepostCount, result.RepostCount)
				assert.Equal(t, tt.expected.CommentCount, result.CommentCount)
				assert.Equal(t, tt.expected.IsLiked, result.IsLiked)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestPostServiceClient_DeletePost(t *testing.T) {
	postID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name        string
		postID      uuid.UUID
		userID      uuid.UUID
		mockErr     error
		expectedErr bool
	}{
		{
			name:   "successful deletion",
			postID: postID,
			userID: userID,
		},
		{
			name:        "grpc error",
			postID:      postID,
			userID:      userID,
			mockErr:     errors.New("grpc error"),
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockPostServiceClient)
			client := &PostServiceClient{client: mockClient}

			mockClient.On("DeletePost", mock.Anything, &pb.DeletePostRequest{
				PostId: tt.postID.String(),
				UserId: tt.userID.String(),
			}, mock.Anything).Return(&pb.DeletePostResponse{}, tt.mockErr)

			err := client.DeletePost(context.Background(), tt.userID, tt.postID)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestPostServiceClient_UpdatePost(t *testing.T) {
	postID := uuid.New()
	userID := uuid.New()
	now := time.Now()

	tests := []struct {
		name        string
		update      models.PostUpdate
		userID      uuid.UUID
		mockResp    *pb.UpdatePostResponse
		mockErr     error
		expected    *models.Post
		expectedErr bool
	}{
		{
			name: "successful update",
			update: models.PostUpdate{
				Id:   postID,
				Desc: "Updated description",
			},
			userID: userID,
			mockResp: &pb.UpdatePostResponse{
				Post: &pb.Post{
					Id:           postID.String(),
					CreatorId:    userID.String(),
					CreatorType:  string(models.PostUser),
					Description:  "Updated description",
					CreatedAt:    timestamppb.New(now.Add(-time.Hour)),
					UpdatedAt:    timestamppb.New(now),
					LikeCount:    10,
					RepostCount:  2,
					CommentCount: 5,
					IsRepost:     false,
					IsLiked:      true,
				},
			},
			expected: &models.Post{
				Id:           postID,
				CreatorId:    userID,
				CreatorType:  models.PostUser,
				Desc:         "Updated description",
				CreatedAt:    now.Add(-time.Hour),
				UpdatedAt:    now,
				LikeCount:    10,
				RepostCount:  2,
				CommentCount: 5,
				IsRepost:     false,
				IsLiked:      true,
			},
		},
		{
			name: "grpc error",
			update: models.PostUpdate{
				Id:   postID,
				Desc: "Updated description",
			},
			userID:      userID,
			mockErr:     errors.New("grpc error"),
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockPostServiceClient)
			client := &PostServiceClient{client: mockClient}

			mockClient.On("UpdatePost", mock.Anything, &pb.UpdatePostRequest{
				Post:   ModelPostUpdateToProto(&tt.update),
				UserId: tt.userID.String(),
			}, mock.Anything).Return(tt.mockResp, tt.mockErr)

			result, err := client.UpdatePost(context.Background(), tt.update, tt.userID)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.Id, result.Id)
				assert.Equal(t, tt.expected.CreatorId, result.CreatorId)
				assert.Equal(t, tt.expected.CreatorType, result.CreatorType)
				assert.Equal(t, tt.expected.Desc, result.Desc)
				assert.Equal(t, tt.expected.IsRepost, result.IsRepost)
				assert.Equal(t, tt.expected.IsLiked, result.IsLiked)
				assert.Equal(t, tt.expected.RepostCount, result.RepostCount)
				assert.Equal(t, tt.expected.CommentCount, result.CommentCount)
				assert.Equal(t, tt.expected.IsLiked, result.IsLiked)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestPostServiceClient_FetchFeed(t *testing.T) {
	userID := uuid.New()
	postID1 := uuid.New()
	postID2 := uuid.New()
	creatorID2 := uuid.New()
	now := time.Now()

	tests := []struct {
		name        string
		numPosts    int
		timestamp   time.Time
		userID      uuid.UUID
		mockResp    *pb.FetchFeedResponse
		mockErr     error
		expected    []models.Post
		expectedErr bool
	}{
		{
			name:      "successful fetch",
			numPosts:  10,
			timestamp: now,
			userID:    userID,
			mockResp: &pb.FetchFeedResponse{
				Posts: []*pb.Post{
					{
						Id:           postID1.String(),
						CreatorId:    userID.String(),
						CreatorType:  string(models.PostUser),
						Description:  "Post 1",
						CreatedAt:    timestamppb.New(now.Add(-time.Hour)),
						UpdatedAt:    timestamppb.New(now.Add(-time.Hour)),
						LikeCount:    5,
						RepostCount:  1,
						CommentCount: 2,
						IsRepost:     false,
						IsLiked:      true,
					},
					{
						Id:           postID2.String(),
						CreatorId:    creatorID2.String(),
						CreatorType:  string(models.PostUser),
						Description:  "Post 2",
						CreatedAt:    timestamppb.New(now.Add(-30 * time.Minute)),
						UpdatedAt:    timestamppb.New(now.Add(-30 * time.Minute)),
						LikeCount:    10,
						RepostCount:  3,
						CommentCount: 5,
						IsRepost:     true,
						IsLiked:      false,
					},
				},
			},
			expected: []models.Post{
				{
					Id:           postID1,
					CreatorId:    userID,
					CreatorType:  models.PostUser,
					Desc:         "Post 1",
					CreatedAt:    now.Add(-time.Hour),
					UpdatedAt:    now.Add(-time.Hour),
					LikeCount:    5,
					RepostCount:  1,
					CommentCount: 2,
					IsRepost:     false,
					IsLiked:      true,
				},
				{
					Id:           postID2,
					CreatorId:    creatorID2,
					CreatorType:  models.PostUser,
					Desc:         "Post 2",
					CreatedAt:    now.Add(-30 * time.Minute),
					UpdatedAt:    now.Add(-30 * time.Minute),
					LikeCount:    10,
					RepostCount:  3,
					CommentCount: 5,
					IsRepost:     true,
					IsLiked:      false,
				},
			},
		},
		{
			name:        "grpc error",
			numPosts:    10,
			timestamp:   now,
			userID:      userID,
			mockErr:     errors.New("grpc error"),
			expectedErr: true,
		},
		{
			name:      "zero posts",
			numPosts:  0,
			timestamp: now,
			userID:    userID,
			mockResp: &pb.FetchFeedResponse{
				Posts: []*pb.Post{},
			},
			expected: []models.Post{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockPostServiceClient)
			client := &PostServiceClient{client: mockClient}

			mockClient.On("FetchFeed", mock.Anything, &pb.FetchFeedRequest{
				NumPosts:  int32(tt.numPosts),
				Timestamp: ToTimestamp(tt.timestamp),
				UserId:    tt.userID.String(),
			}, mock.Anything).Return(tt.mockResp, tt.mockErr)

			result, err := client.FetchFeed(context.Background(), tt.numPosts, tt.timestamp, tt.userID)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, len(tt.expected), len(result))
				for i := range tt.expected {
					assert.Equal(t, tt.expected[i].Id, result[i].Id)
					assert.Equal(t, tt.expected[i].Desc, result[i].Desc)
					assert.Equal(t, tt.expected[i].CreatorId, result[i].CreatorId)
					assert.Equal(t, tt.expected[i].LikeCount, result[i].LikeCount)
				}
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestPostServiceClient_FetchRecommendations(t *testing.T) {
	userID := uuid.New()
	postID := uuid.New()
	now := time.Now()

	tests := []struct {
		name        string
		numPosts    int
		timestamp   time.Time
		userID      uuid.UUID
		mockResp    *pb.FetchRecommendationsResponse
		mockErr     error
		expected    []models.Post
		expectedErr bool
	}{
		{
			name:      "successful fetch",
			numPosts:  5,
			timestamp: now,
			userID:    userID,
			mockResp: &pb.FetchRecommendationsResponse{
				Posts: []*pb.Post{
					{
						Id:           postID.String(),
						CreatorId:    userID.String(),
						CreatorType:  string(models.PostUser),
						Description:  "Recommended post",
						CreatedAt:    timestamppb.New(now.Add(-time.Hour)),
						UpdatedAt:    timestamppb.New(now.Add(-time.Hour)),
						LikeCount:    20,
						RepostCount:  5,
						CommentCount: 8,
						IsRepost:     false,
						IsLiked:      false,
					},
				},
			},
			expected: []models.Post{
				{
					Id:           postID,
					CreatorId:    userID,
					CreatorType:  models.PostUser,
					Desc:         "Recommended post",
					CreatedAt:    now.Add(-time.Hour),
					UpdatedAt:    now.Add(-time.Hour),
					LikeCount:    20,
					RepostCount:  5,
					CommentCount: 8,
					IsRepost:     false,
					IsLiked:      false,
				},
			},
		},
		{
			name:        "grpc error",
			numPosts:    5,
			timestamp:   now,
			userID:      userID,
			mockErr:     errors.New("grpc error"),
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockPostServiceClient)
			client := &PostServiceClient{client: mockClient}

			mockClient.On("FetchRecommendations", mock.Anything, &pb.FetchRecommendationsRequest{
				NumPosts:  int32(tt.numPosts),
				Timestamp: ToTimestamp(tt.timestamp),
				UserId:    tt.userID.String(),
			}, mock.Anything).Return(tt.mockResp, tt.mockErr)

			result, err := client.FetchRecommendations(context.Background(), tt.numPosts, tt.timestamp, tt.userID)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, len(tt.expected), len(result))
				if len(result) > 0 {
					assert.Equal(t, tt.expected[0].Id, result[0].Id)
					assert.Equal(t, tt.expected[0].Desc, result[0].Desc)
				}
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestPostServiceClient_FetchCreatorPosts(t *testing.T) {
	userID := uuid.New()
	requesterID := uuid.New()
	postID := uuid.New()
	now := time.Now()

	tests := []struct {
		name        string
		userID      uuid.UUID
		requesterID uuid.UUID
		numPosts    int
		timestamp   time.Time
		mockResp    *pb.FetchUserPostsResponse
		mockErr     error
		expected    []models.Post
		expectedErr bool
	}{
		{
			name:        "successful fetch",
			userID:      userID,
			requesterID: requesterID,
			numPosts:    10,
			timestamp:   now,
			mockResp: &pb.FetchUserPostsResponse{
				Posts: []*pb.Post{
					{
						Id:           postID.String(),
						CreatorId:    userID.String(),
						CreatorType:  string(models.PostUser),
						Description:  "User post",
						CreatedAt:    timestamppb.New(now.Add(-time.Hour)),
						UpdatedAt:    timestamppb.New(now.Add(-time.Hour)),
						LikeCount:    15,
						RepostCount:  2,
						CommentCount: 3,
						IsRepost:     false,
						IsLiked:      true,
					},
				},
			},
			expected: []models.Post{
				{
					Id:           postID,
					CreatorId:    userID,
					CreatorType:  models.PostUser,
					Desc:         "User post",
					CreatedAt:    now.Add(-time.Hour),
					UpdatedAt:    now.Add(-time.Hour),
					LikeCount:    15,
					RepostCount:  2,
					CommentCount: 3,
					IsRepost:     false,
					IsLiked:      true,
				},
			},
		},
		{
			name:        "empty requester ID",
			userID:      userID,
			requesterID: uuid.Nil,
			numPosts:    10,
			timestamp:   now,
			mockResp: &pb.FetchUserPostsResponse{
				Posts: []*pb.Post{
					{
						Id:          postID.String(),
						CreatorId:   userID.String(),
						CreatorType: string(models.PostUser),
						Description: "User post",
						CreatedAt:   timestamppb.New(now.Add(-time.Hour)),
						UpdatedAt:   timestamppb.New(now.Add(-time.Hour)),
					},
				},
			},
			expected: []models.Post{
				{
					Id:          postID,
					CreatorId:   userID,
					CreatorType: models.PostUser,
					Desc:        "User post",
					CreatedAt:   now.Add(-time.Hour),
					UpdatedAt:   now.Add(-time.Hour),
				},
			},
		},
		{
			name:        "grpc error",
			userID:      userID,
			requesterID: requesterID,
			numPosts:    10,
			timestamp:   now,
			mockErr:     errors.New("grpc error"),
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockPostServiceClient)
			client := &PostServiceClient{client: mockClient}

			mockClient.On("FetchUserPosts", mock.Anything, &pb.FetchUserPostsRequest{
				UserId:      tt.userID.String(),
				RequesterId: tt.requesterID.String(),
				NumPosts:    int32(tt.numPosts),
				Timestamp:   ToTimestamp(tt.timestamp),
			}, mock.Anything).Return(tt.mockResp, tt.mockErr)

			result, err := client.FetchCreatorPosts(context.Background(), tt.userID, tt.requesterID, tt.numPosts, tt.timestamp)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, len(tt.expected), len(result))
				if len(result) > 0 {
					assert.Equal(t, tt.expected[0].Id, result[0].Id)
					assert.Equal(t, tt.expected[0].CreatorId, result[0].CreatorId)
				}
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestPostServiceClient_LikePost(t *testing.T) {
	postID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name        string
		postID      uuid.UUID
		userID      uuid.UUID
		mockErr     error
		expectedErr bool
	}{
		{
			name:   "successful like",
			postID: postID,
			userID: userID,
		},
		{
			name:        "grpc error",
			postID:      postID,
			userID:      userID,
			mockErr:     errors.New("grpc error"),
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockPostServiceClient)
			client := &PostServiceClient{client: mockClient}

			mockClient.On("LikePost", mock.Anything, &pb.LikePostRequest{
				PostId: tt.postID.String(),
				UserId: tt.userID.String(),
			}, mock.Anything).Return(&pb.LikePostResponse{}, tt.mockErr)

			err := client.LikePost(context.Background(), tt.postID, tt.userID)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestPostServiceClient_UnlikePost(t *testing.T) {
	postID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name        string
		postID      uuid.UUID
		userID      uuid.UUID
		mockErr     error
		expectedErr bool
	}{
		{
			name:   "successful unlike",
			postID: postID,
			userID: userID,
		},
		{
			name:        "grpc error",
			postID:      postID,
			userID:      userID,
			mockErr:     errors.New("grpc error"),
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockPostServiceClient)
			client := &PostServiceClient{client: mockClient}

			mockClient.On("UnlikePost", mock.Anything, &pb.UnlikePostRequest{
				PostId: tt.postID.String(),
				UserId: tt.userID.String(),
			}, mock.Anything).Return(&pb.UnlikePostResponse{}, tt.mockErr)

			err := client.UnlikePost(context.Background(), tt.postID, tt.userID)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestPostServiceClient_GetPost(t *testing.T) {
	postID := uuid.New()
	userID := uuid.New()
	now := time.Now()

	tests := []struct {
		name        string
		postID      uuid.UUID
		userID      uuid.UUID
		mockResp    *pb.GetPostResponse
		mockErr     error
		expected    *models.Post
		expectedErr bool
	}{
		{
			name:   "successful get",
			postID: postID,
			userID: userID,
			mockResp: &pb.GetPostResponse{
				Post: &pb.Post{
					Id:           postID.String(),
					CreatorId:    userID.String(),
					CreatorType:  string(models.PostUser),
					Description:  "Single post",
					CreatedAt:    timestamppb.New(now.Add(-time.Hour)),
					UpdatedAt:    timestamppb.New(now.Add(-time.Hour)),
					LikeCount:    25,
					RepostCount:  4,
					CommentCount: 7,
					IsRepost:     false,
					IsLiked:      true,
				},
			},
			expected: &models.Post{
				Id:           postID,
				CreatorId:    userID,
				CreatorType:  models.PostUser,
				Desc:         "Single post",
				CreatedAt:    now.Add(-time.Hour),
				UpdatedAt:    now.Add(-time.Hour),
				LikeCount:    25,
				RepostCount:  4,
				CommentCount: 7,
				IsRepost:     false,
				IsLiked:      true,
			},
		},
		{
			name:        "grpc error",
			postID:      postID,
			userID:      userID,
			mockErr:     errors.New("grpc error"),
			expectedErr: true,
		},
		{
			name:        "post not found",
			postID:      postID,
			userID:      userID,
			mockErr:     errors.New("not found"),
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockPostServiceClient)
			client := &PostServiceClient{client: mockClient}

			mockClient.On("GetPost", mock.Anything, &pb.GetPostRequest{
				PostId: tt.postID.String(),
				UserId: tt.userID.String(),
			}, mock.Anything).Return(tt.mockResp, tt.mockErr)

			result, err := client.GetPost(context.Background(), tt.postID, tt.userID)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.Id, result.Id)
				assert.Equal(t, tt.expected.CreatorId, result.CreatorId)
				assert.Equal(t, tt.expected.CreatorType, result.CreatorType)
				assert.Equal(t, tt.expected.Desc, result.Desc)
				assert.Equal(t, tt.expected.IsRepost, result.IsRepost)
				assert.Equal(t, tt.expected.IsLiked, result.IsLiked)
				assert.Equal(t, tt.expected.RepostCount, result.RepostCount)
				assert.Equal(t, tt.expected.CommentCount, result.CommentCount)
				assert.Equal(t, tt.expected.IsLiked, result.IsLiked)
			}

			mockClient.AssertExpectations(t)
		})
	}
}
