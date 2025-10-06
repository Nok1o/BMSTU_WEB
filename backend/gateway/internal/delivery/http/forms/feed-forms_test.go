//go:build unit
// +build unit

package forms

import (
	"errors"
	"net/url"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	time2 "quickflow/config/time"
	"quickflow/shared/models"
)

func TestParseCreatorType(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    models.PostCreatorType
		expectedErr error
	}{
		{
			name:        "user type",
			input:       "user",
			expected:    models.PostUser,
			expectedErr: nil,
		},
		{
			name:        "community type",
			input:       "community",
			expected:    models.PostCommunity,
			expectedErr: nil,
		},
		{
			name:        "invalid type",
			input:       "invalid",
			expected:    "",
			expectedErr: errors.New("invalid creator type"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseCreatorType(tt.input)
			assert.Equal(t, tt.expected, result)
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}

func TestPostForm_ToPostModel(t *testing.T) {
	userId := uuid.New()
	communityId := uuid.New()
	now := time.Now()

	tests := []struct {
		name        string
		form        PostForm
		userId      uuid.UUID
		expected    models.Post
		expectedErr error
	}{
		{
			name: "empty creator type",
			form: PostForm{
				Text:      "test text",
				Media:     []string{"media1", "media2"},
				Audio:     []string{"audio1"},
				File:      []string{"file1"},
				Stickers:  []string{"sticker1"},
				IsRepost:  true,
				CreatorId: communityId,
			},
			userId: userId,
			expected: models.Post{
				Desc:        "test text",
				CreatorType: models.PostUser,
				CreatorId:   userId,
				IsRepost:    true,
				CreatedAt:   now,
				UpdatedAt:   now,
				Files: []*models.File{
					{URL: "media1", DisplayType: models.DisplayTypeMedia},
					{URL: "media2", DisplayType: models.DisplayTypeMedia},
					{URL: "audio1", DisplayType: models.DisplayTypeAudio},
					{URL: "file1", DisplayType: models.DisplayTypeFile},
					{URL: "sticker1", DisplayType: models.DisplayTypeSticker},
				},
			},
			expectedErr: nil,
		},
		{
			name: "user creator type",
			form: PostForm{
				Text:        "test text",
				CreatorType: "user",
				CreatorId:   communityId,
			},
			userId: userId,
			expected: models.Post{
				Desc:        "test text",
				CreatorType: models.PostUser,
				CreatorId:   userId,
				CreatedAt:   now,
				UpdatedAt:   now,
				Files:       []*models.File{},
			},
			expectedErr: nil,
		},
		{
			name: "community creator type",
			form: PostForm{
				Text:        "test text",
				CreatorType: "community",
				CreatorId:   communityId,
			},
			userId: userId,
			expected: models.Post{
				Desc:        "test text",
				CreatorType: models.PostCommunity,
				CreatorId:   communityId,
				CreatedAt:   now,
				UpdatedAt:   now,
				Files:       []*models.File{},
			},
			expectedErr: nil,
		},
		{
			name: "invalid creator type",
			form: PostForm{
				Text:        "test text",
				CreatorType: "invalid",
			},
			userId:      userId,
			expected:    models.Post{},
			expectedErr: errors.New("invalid creator type"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			result, err := tt.form.ToPostModel(tt.userId)
			if tt.expectedErr != nil {
				assert.EqualError(t, err, tt.expectedErr.Error())
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected.Desc, result.Desc)
			assert.Equal(t, tt.expected.CreatorType, result.CreatorType)
			assert.Equal(t, tt.expected.CreatorId, result.CreatorId)
			assert.Equal(t, tt.expected.IsRepost, result.IsRepost)

			assert.Len(t, result.Files, len(tt.expected.Files))
			for i, file := range result.Files {
				assert.Equal(t, tt.expected.Files[i].URL, file.URL)
				assert.Equal(t, tt.expected.Files[i].DisplayType, file.DisplayType)
			}
		})
	}
}

func TestFeedForm_GetParams(t *testing.T) {
	tests := []struct {
		name        string
		values      url.Values
		expected    FeedForm
		expectedErr error
	}{
		{
			name: "valid params",
			values: url.Values{
				"posts_count": []string{"10"},
				"ts":          []string{"12345"},
			},
			expected: FeedForm{
				Posts: 10,
				Ts:    "12345",
			},
			expectedErr: nil,
		},
		{
			name: "missing posts_count",
			values: url.Values{
				"ts": []string{"12345"},
			},
			expected:    FeedForm{},
			expectedErr: errors.New("posts_count parameter missing"),
		},
		{
			name: "invalid posts_count",
			values: url.Values{
				"posts_count": []string{"invalid"},
				"ts":          []string{"12345"},
			},
			expected:    FeedForm{},
			expectedErr: errors.New("failed to parse posts_count"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var form FeedForm
			err := form.GetParams(tt.values)
			if tt.expectedErr != nil {
				assert.EqualError(t, err, tt.expectedErr.Error())
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected.Posts, form.Posts)
			assert.Equal(t, tt.expected.Ts, form.Ts)
		})
	}
}

func TestPublicUserInfoToOut(t *testing.T) {
	userId := uuid.New()

	tests := []struct {
		name     string
		info     models.PublicUserInfo
		relation models.UserRelation
		expected PublicUserInfoOut
	}{
		{
			name: "full info",
			info: models.PublicUserInfo{
				Id:        userId,
				Username:  "testuser",
				Firstname: "Test",
				Lastname:  "User",
				AvatarURL: "http://example.com/avatar.jpg",
			},
			relation: models.RelationFriend,
			expected: PublicUserInfoOut{
				ID:        userId.String(),
				Username:  "testuser",
				FirstName: "Test",
				LastName:  "User",
				AvatarURL: "http://example.com/avatar.jpg",
				Relation:  models.RelationFriend,
			},
		},
		{
			name: "minimal info",
			info: models.PublicUserInfo{
				Id:       userId,
				Username: "testuser",
			},
			relation: models.RelationNone,
			expected: PublicUserInfoOut{
				ID:       userId.String(),
				Username: "testuser",
				Relation: models.RelationNone,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PublicUserInfoToOut(tt.info, tt.relation)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPostOut_FromPost(t *testing.T) {
	postId := uuid.New()
	creatorId := uuid.New()
	createdAt := time.Now()
	updatedAt := createdAt.Add(1 * time.Hour)

	tests := []struct {
		name     string
		post     models.Post
		expected PostOut
	}{
		{
			name: "full post",
			post: models.Post{
				Id:          postId,
				CreatorId:   creatorId,
				CreatorType: models.PostUser,
				Desc:        "test post",
				Files: []*models.File{
					{URL: "media1.jpg", DisplayType: models.DisplayTypeMedia},
					{URL: "audio1.mp3", DisplayType: models.DisplayTypeAudio},
					{URL: "file1.pdf", DisplayType: models.DisplayTypeFile},
					{URL: "sticker1.png", DisplayType: models.DisplayTypeSticker},
				},
				CreatedAt:    createdAt,
				UpdatedAt:    updatedAt,
				LikeCount:    10,
				RepostCount:  5,
				CommentCount: 3,
				IsRepost:     true,
				IsLiked:      false,
			},
			expected: PostOut{
				Id:           postId.String(),
				Creator:      &PublicUserInfoOut{ID: creatorId.String()},
				CreatorType:  string(models.PostUser),
				Desc:         "test post",
				MediaURLs:    []FileOut{{URL: "media1.jpg"}},
				AudioURLs:    []FileOut{{URL: "audio1.mp3"}},
				FileURLs:     []FileOut{{URL: "file1.pdf"}},
				StickerURLs:  []FileOut{{URL: "sticker1.png"}},
				CreatedAt:    createdAt.Format(time2.TimeStampLayout),
				UpdatedAt:    updatedAt.Format(time2.TimeStampLayout),
				LikeCount:    10,
				RepostCount:  5,
				CommentCount: 3,
				IsRepost:     true,
				IsLiked:      false,
			},
		},
		{
			name: "minimal post",
			post: models.Post{
				Id:          postId,
				CreatorId:   creatorId,
				CreatorType: models.PostCommunity,
				Desc:        "minimal post",
				Files:       []*models.File{},
				CreatedAt:   createdAt,
				UpdatedAt:   updatedAt,
			},
			expected: PostOut{
				Id:          postId.String(),
				Creator:     &PublicUserInfoOut{ID: creatorId.String()},
				CreatorType: string(models.PostCommunity),
				Desc:        "minimal post",
				CreatedAt:   createdAt.Format(time2.TimeStampLayout),
				UpdatedAt:   updatedAt.Format(time2.TimeStampLayout),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out PostOut
			out.FromPost(tt.post)

			assert.Equal(t, tt.expected.Id, out.Id)
			assert.Equal(t, tt.expected.Desc, out.Desc)
			assert.Equal(t, tt.expected.CreatorType, out.CreatorType)
			assert.Equal(t, tt.expected.CreatedAt, out.CreatedAt)
			assert.Equal(t, tt.expected.UpdatedAt, out.UpdatedAt)
			assert.Equal(t, tt.expected.LikeCount, out.LikeCount)
			assert.Equal(t, tt.expected.RepostCount, out.RepostCount)
			assert.Equal(t, tt.expected.CommentCount, out.CommentCount)
			assert.Equal(t, tt.expected.IsRepost, out.IsRepost)
			assert.Equal(t, tt.expected.IsLiked, out.IsLiked)

			if creator, ok := out.Creator.(*PublicUserInfoOut); ok {
				assert.Equal(t, tt.expected.Creator.(*PublicUserInfoOut).ID, creator.ID)
			} else {
				t.Error("Creator type assertion failed")
			}

			assert.Len(t, out.MediaURLs, len(tt.expected.MediaURLs))
			assert.Len(t, out.AudioURLs, len(tt.expected.AudioURLs))
			assert.Len(t, out.FileURLs, len(tt.expected.FileURLs))
			assert.Len(t, out.StickerURLs, len(tt.expected.StickerURLs))
		})
	}
}

func TestUpdatePostForm_ToPostUpdateModel(t *testing.T) {
	postId := uuid.New()

	tests := []struct {
		name     string
		form     UpdatePostForm
		postId   uuid.UUID
		expected models.PostUpdate
	}{
		{
			name: "full update",
			form: UpdatePostForm{
				Text:  "updated text",
				Media: []string{"new_media1.jpg", "new_media2.jpg"},
				Audio: []string{"new_audio.mp3"},
				File:  []string{"new_file.pdf"},
			},
			postId: postId,
			expected: models.PostUpdate{
				Id:   postId,
				Desc: "updated text",
				Files: []*models.File{
					{URL: "new_media1.jpg", DisplayType: models.DisplayTypeMedia},
					{URL: "new_media2.jpg", DisplayType: models.DisplayTypeMedia},
					{URL: "new_audio.mp3", DisplayType: models.DisplayTypeAudio},
					{URL: "new_file.pdf", DisplayType: models.DisplayTypeFile},
				},
			},
		},
		{
			name: "text only update",
			form: UpdatePostForm{
				Text: "updated text only",
			},
			postId: postId,
			expected: models.PostUpdate{
				Id:    postId,
				Desc:  "updated text only",
				Files: []*models.File{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.form.ToPostUpdateModel(tt.postId)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected.Id, result.Id)
			assert.Equal(t, tt.expected.Desc, result.Desc)

			assert.Len(t, result.Files, len(tt.expected.Files))
			for i, file := range result.Files {
				assert.Equal(t, tt.expected.Files[i].URL, file.URL)
				assert.Equal(t, tt.expected.Files[i].DisplayType, file.DisplayType)
			}
		})
	}
}
