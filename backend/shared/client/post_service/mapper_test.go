//go:build unit
// +build unit

package post_service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"

	time_config "quickflow/config/time"
	shared_models "quickflow/shared/models"
	pb "quickflow/shared/proto/file_service"
	proto "quickflow/shared/proto/post_service"
)

func TestProtoPostToModel(t *testing.T) {
	now := time.Now()
	testFile := &pb.File{
		FileName: "test.jpg",
		Url:      "http://example.com/test.jpg",
	}

	tests := []struct {
		name        string
		input       *proto.Post
		expected    *shared_models.Post
		expectError bool
	}{
		{
			name: "Valid Post",
			input: &proto.Post{
				Id:           uuid.New().String(),
				CreatorId:    uuid.New().String(),
				CreatorType:  "user",
				Description:  "Test post",
				Files:        []*pb.File{testFile},
				CreatedAt:    timestamppb.New(now),
				UpdatedAt:    timestamppb.New(now),
				LikeCount:    10,
				RepostCount:  2,
				CommentCount: 5,
				IsRepost:     true,
				IsLiked:      false,
			},
			expected: &shared_models.Post{
				CreatorType:  "user",
				Desc:         "Test post",
				CreatedAt:    now,
				UpdatedAt:    now,
				LikeCount:    10,
				RepostCount:  2,
				CommentCount: 5,
				IsRepost:     true,
				IsLiked:      false,
			},
			expectError: false,
		},
		{
			name: "Invalid UUID",
			input: &proto.Post{
				Id:        "invalid-uuid",
				CreatorId: uuid.New().String(),
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ProtoPostToModel(tt.input)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.input == nil {
				} else {
					assert.Equal(t, tt.expected.CreatorType, result.CreatorType)
					assert.Equal(t, tt.expected.Desc, result.Desc)
					assert.Equal(t, tt.expected.CreatedAt.Unix(), result.CreatedAt.Unix())
					assert.Equal(t, tt.expected.UpdatedAt.Unix(), result.UpdatedAt.Unix())
					assert.Equal(t, tt.expected.LikeCount, result.LikeCount)
					assert.Equal(t, tt.expected.RepostCount, result.RepostCount)
					assert.Equal(t, tt.expected.CommentCount, result.CommentCount)
					assert.Equal(t, tt.expected.IsRepost, result.IsRepost)
					assert.Equal(t, tt.expected.IsLiked, result.IsLiked)

					if len(tt.input.Files) > 0 {
						assert.Equal(t, tt.input.Files[0].FileName, result.Files[0].Name)
						assert.Equal(t, tt.input.Files[0].Url, result.Files[0].URL)
					}
				}
			}
		})
	}
}

func TestModelPostToProto(t *testing.T) {
	now := time.Now()
	testFile := shared_models.File{
		Name: "test.jpg",
		URL:  "http://example.com/test.jpg",
	}

	post := &shared_models.Post{
		Id:           uuid.New(),
		CreatorId:    uuid.New(),
		CreatorType:  "user",
		Desc:         "Test post",
		Files:        []*shared_models.File{&testFile},
		CreatedAt:    now,
		UpdatedAt:    now,
		LikeCount:    10,
		RepostCount:  2,
		CommentCount: 5,
		IsRepost:     true,
		IsLiked:      false,
	}

	result := ModelPostToProto(post)

	assert.Equal(t, post.Id.String(), result.Id)
	assert.Equal(t, post.CreatorId.String(), result.CreatorId)
	assert.Equal(t, string(post.CreatorType), result.CreatorType)
	assert.Equal(t, post.Desc, result.Description)
	assert.Equal(t, post.CreatedAt.Unix(), result.CreatedAt.AsTime().Unix())
	assert.Equal(t, post.UpdatedAt.Unix(), result.UpdatedAt.AsTime().Unix())
	assert.Equal(t, int64(post.LikeCount), result.LikeCount)
	assert.Equal(t, int64(post.RepostCount), result.RepostCount)
	assert.Equal(t, int64(post.CommentCount), result.CommentCount)
	assert.Equal(t, post.IsRepost, result.IsRepost)
	assert.Equal(t, post.IsLiked, result.IsLiked)

	if len(post.Files) > 0 {
		assert.Equal(t, post.Files[0].Name, result.Files[0].FileName)
		assert.Equal(t, post.Files[0].URL, result.Files[0].Url)
	}
}

func TestProtoPostUpdateToModel(t *testing.T) {
	testFile := &pb.File{
		FileName: "update.jpg",
		Url:      "http://example.com/update.jpg",
	}

	tests := []struct {
		name        string
		input       *proto.PostUpdate
		expected    *shared_models.PostUpdate
		expectError bool
	}{
		{
			name: "Valid PostUpdate",
			input: &proto.PostUpdate{
				Id:          uuid.New().String(),
				Description: "Updated post",
				Files:       []*pb.File{testFile},
			},
			expected: &shared_models.PostUpdate{
				Desc: "Updated post",
			},
			expectError: false,
		},
		{
			name: "Invalid UUID",
			input: &proto.PostUpdate{
				Id: "invalid-uuid",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ProtoPostUpdateToModel(tt.input)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.Desc, result.Desc)

				if len(tt.input.Files) > 0 {
					assert.Equal(t, tt.input.Files[0].FileName, result.Files[0].Name)
					assert.Equal(t, tt.input.Files[0].Url, result.Files[0].URL)
				}
			}
		})
	}
}

func TestModelPostUpdateToProto(t *testing.T) {
	testFile := shared_models.File{
		Name: "update.jpg",
		URL:  "http://example.com/update.jpg",
	}

	update := &shared_models.PostUpdate{
		Id:    uuid.New(),
		Desc:  "Updated post",
		Files: []*shared_models.File{&testFile},
	}

	result := ModelPostUpdateToProto(update)

	assert.Equal(t, update.Id.String(), result.Id)
	assert.Equal(t, update.Desc, result.Description)

	if len(update.Files) > 0 {
		assert.Equal(t, update.Files[0].Name, result.Files[0].FileName)
		assert.Equal(t, update.Files[0].URL, result.Files[0].Url)
	}
}

func TestProtoCommentToModel(t *testing.T) {
	now := time.Now()
	testFile := &pb.File{
		FileName: "comment.jpg",
		Url:      "http://example.com/comment.jpg",
	}

	tests := []struct {
		name        string
		input       *proto.Comment
		expected    *shared_models.Comment
		expectError bool
	}{
		{
			name: "Valid Comment",
			input: &proto.Comment{
				Id:        uuid.New().String(),
				PostId:    uuid.New().String(),
				UserId:    uuid.New().String(),
				Text:      "Test comment",
				Images:    []*pb.File{testFile},
				CreatedAt: now.Format(time_config.TimeStampLayout),
				UpdatedAt: now.Format(time_config.TimeStampLayout),
				LikeCount: 5,
				IsLiked:   true,
			},
			expected: &shared_models.Comment{
				Text:      "Test comment",
				LikeCount: 5,
				IsLiked:   true,
			},
			expectError: false,
		},
		{
			name: "Invalid UUID",
			input: &proto.Comment{
				Id: "invalid-uuid",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ProtoCommentToModel(tt.input)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.Text, result.Text)
				assert.Equal(t, tt.expected.LikeCount, result.LikeCount)
				assert.Equal(t, tt.expected.IsLiked, result.IsLiked)

				if len(tt.input.Images) > 0 {
					assert.Equal(t, tt.input.Images[0].FileName, result.Images[0].Name)
					assert.Equal(t, tt.input.Images[0].Url, result.Images[0].URL)
				}
			}
		})
	}
}

func TestModelCommentToProto(t *testing.T) {
	now := time.Now()
	testFile := shared_models.File{
		Name: "comment.jpg",
		URL:  "http://example.com/comment.jpg",
	}

	comment := &shared_models.Comment{
		Id:        uuid.New(),
		PostId:    uuid.New(),
		UserId:    uuid.New(),
		Text:      "Test comment",
		Images:    []*shared_models.File{&testFile},
		CreatedAt: now,
		UpdatedAt: now,
		LikeCount: 5,
		IsLiked:   true,
	}

	result := ModelCommentToProto(comment)

	assert.Equal(t, comment.Id.String(), result.Id)
	assert.Equal(t, comment.PostId.String(), result.PostId)
	assert.Equal(t, comment.UserId.String(), result.UserId)
	assert.Equal(t, comment.Text, result.Text)
	assert.Equal(t, comment.CreatedAt.Format(time_config.TimeStampLayout), result.CreatedAt)
	assert.Equal(t, comment.UpdatedAt.Format(time_config.TimeStampLayout), result.UpdatedAt)
	assert.Equal(t, int64(comment.LikeCount), result.LikeCount)
	assert.Equal(t, comment.IsLiked, result.IsLiked)

	if len(comment.Images) > 0 {
		assert.Equal(t, comment.Images[0].Name, result.Images[0].FileName)
		assert.Equal(t, comment.Images[0].URL, result.Images[0].Url)
	}
}

func TestConvertProtoPosts(t *testing.T) {
	now := time.Now()
	protoPosts := []*proto.Post{
		{
			Id:        uuid.New().String(),
			CreatorId: uuid.New().String(),
			CreatedAt: timestamppb.New(now),
		},
		{
			Id:        uuid.New().String(),
			CreatorId: uuid.New().String(),
			CreatedAt: timestamppb.New(now),
		},
	}

	result, err := convertProtoPosts(protoPosts)
	assert.NoError(t, err)
	assert.Equal(t, len(protoPosts), len(result))

	for i := range result {
		assert.Equal(t, protoPosts[i].Id, result[i].Id.String())
		assert.Equal(t, protoPosts[i].CreatorId, result[i].CreatorId.String())
		assert.Equal(t, protoPosts[i].CreatedAt.AsTime().Unix(), result[i].CreatedAt.Unix())
	}
}
