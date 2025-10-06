//go:build unit
// +build unit

package forms

import (
	"net/url"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	time2 "quickflow/config/time"
	"quickflow/shared/models"
)

func TestCommentForm_ToCommentModel(t *testing.T) {
	tests := []struct {
		name     string
		form     CommentForm
		expected models.Comment
	}{
		{
			name: "Full form",
			form: CommentForm{
				Text:     "Test comment",
				Media:    []string{"media1.jpg", "media2.jpg"},
				Audio:    []string{"audio1.mp3"},
				Files:    []string{"file1.pdf"},
				Stickers: []string{"sticker1.png"},
			},
			expected: models.Comment{
				Text:      "Test comment",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				LikeCount: 0,
				IsLiked:   false,
			},
		},
		{
			name: "Only text",
			form: CommentForm{
				Text: "Simple comment",
			},
			expected: models.Comment{
				Text:      "Simple comment",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				LikeCount: 0,
				IsLiked:   false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.form.ToCommentModel()

			assert.NotEqual(t, uuid.Nil, result.Id)
			assert.Equal(t, tt.expected.Text, result.Text)
			assert.Equal(t, tt.expected.LikeCount, result.LikeCount)
			assert.Equal(t, tt.expected.IsLiked, result.IsLiked)
			assert.WithinDuration(t, tt.expected.CreatedAt, result.CreatedAt, time.Second)
			assert.WithinDuration(t, tt.expected.UpdatedAt, result.UpdatedAt, time.Second)

			// Check attachments
			if len(tt.form.Media) > 0 {
				assert.Len(t, result.Images, len(tt.form.Media)+len(tt.form.Audio)+len(tt.form.Files)+len(tt.form.Stickers))
			} else {
				assert.Len(t, result.Images, 0)
			}
		})
	}
}

func TestCommentUpdateForm_ToCommentUpdateModel(t *testing.T) {
	commentID := uuid.New()
	tests := []struct {
		name     string
		form     CommentUpdateForm
		expected models.CommentUpdate
	}{
		{
			name: "With all attachments",
			form: CommentUpdateForm{
				CommentForm: CommentForm{
					Text:  "Updated comment",
					Media: []string{"new_media.jpg"},
					Files: []string{"new_file.pdf"},
					Audio: []string{"new_audio.mp3"},
				},
			},
			expected: models.CommentUpdate{
				Id:   commentID,
				Text: "Updated comment",
			},
		},
		{
			name: "Only text update",
			form: CommentUpdateForm{
				CommentForm: CommentForm{
					Text: "Text only update",
				},
			},
			expected: models.CommentUpdate{
				Id:   commentID,
				Text: "Text only update",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.form.ToCommentUpdateModel(commentID)

			assert.Equal(t, tt.expected.Id, result.Id)
			assert.Equal(t, tt.expected.Text, result.Text)

			// Check attachments
			expectedAttachments := len(tt.form.Media) + len(tt.form.Audio) + len(tt.form.Files)
			assert.Len(t, result.Files, expectedAttachments)
		})
	}
}

func TestCommentOut_FromComment(t *testing.T) {
	now := time.Now()
	commentID := uuid.New()
	postID := uuid.New()
	userInfo := models.PublicUserInfo{
		Id:        uuid.New(),
		Firstname: "John",
		Lastname:  "Doe",
		AvatarURL: "avatar.jpg",
	}

	tests := []struct {
		name     string
		comment  models.Comment
		expected CommentOut
	}{
		{
			name: "Full comment",
			comment: models.Comment{
				Id:        commentID,
				Text:      "Test comment",
				CreatedAt: now,
				UpdatedAt: now,
				Images: []*models.File{
					{URL: "media1.jpg", DisplayType: models.DisplayTypeMedia},
					{URL: "audio1.mp3", DisplayType: models.DisplayTypeAudio},
					{URL: "file1.pdf", DisplayType: models.DisplayTypeFile},
					{URL: "sticker1.png", DisplayType: models.DisplayTypeSticker},
				},
				PostId:    postID,
				LikeCount: 10,
				IsLiked:   true,
			},
			expected: CommentOut{
				ID:        commentID.String(),
				Text:      "Test comment",
				CreatedAt: now.Format(time2.TimeStampLayout),
				UpdatedAt: now.Format(time2.TimeStampLayout),
				Media:     []FileOut{{URL: "media1.jpg"}},
				Audio:     []FileOut{{URL: "audio1.mp3"}},
				Files:     []FileOut{{URL: "file1.pdf"}},
				Stickers:  []FileOut{{URL: "sticker1.png"}},
				PostId:    postID,
				LikeCount: 10,
				IsLiked:   true,
			},
		},
		{
			name: "Minimal comment",
			comment: models.Comment{
				Id:        commentID,
				Text:      "Simple comment",
				CreatedAt: now,
				UpdatedAt: now,
				PostId:    postID,
				LikeCount: 0,
				IsLiked:   false,
			},
			expected: CommentOut{
				ID:        commentID.String(),
				Text:      "Simple comment",
				CreatedAt: now.Format(time2.TimeStampLayout),
				UpdatedAt: now.Format(time2.TimeStampLayout),
				PostId:    postID,
				LikeCount: 0,
				IsLiked:   false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result CommentOut
			result.FromComment(tt.comment, userInfo)

			assert.Equal(t, tt.expected.ID, result.ID)
			assert.Equal(t, tt.expected.Text, result.Text)
			assert.Equal(t, tt.expected.CreatedAt, result.CreatedAt)
			assert.Equal(t, tt.expected.UpdatedAt, result.UpdatedAt)
			assert.Equal(t, tt.expected.PostId, result.PostId)
			assert.Equal(t, tt.expected.LikeCount, result.LikeCount)
			assert.Equal(t, tt.expected.IsLiked, result.IsLiked)

			// Check attachments
			assert.Len(t, result.Media, len(tt.expected.Media))
			assert.Len(t, result.Audio, len(tt.expected.Audio))
			assert.Len(t, result.Files, len(tt.expected.Files))
			assert.Len(t, result.Stickers, len(tt.expected.Stickers))

			// Check user info
			assert.Equal(t, userInfo.Id.String(), result.Creator.ID)
			assert.Equal(t, userInfo.Firstname, result.Creator.FirstName)
			assert.Equal(t, userInfo.Lastname, result.Creator.LastName)
			assert.Equal(t, userInfo.AvatarURL, result.Creator.AvatarURL)
		})
	}
}

func TestCommentFetchForm_GetParams(t *testing.T) {
	tests := []struct {
		name        string
		values      url.Values
		expected    CommentFetchForm
		expectError bool
		errMessage  string
	}{
		{
			name: "Valid params",
			values: url.Values{
				"count": []string{"10"},
				"ts":    []string{"2023-01-01T00:00:00Z"},
			},
			expected: CommentFetchForm{
				Count: 10,
				Ts:    "2023-01-01T00:00:00Z",
			},
			expectError: false,
		},
		{
			name: "Missing count",
			values: url.Values{
				"ts": []string{"2023-01-01T00:00:00Z"},
			},
			expectError: true,
			errMessage:  "count parameter missing",
		},
		{
			name: "Invalid count",
			values: url.Values{
				"count": []string{"invalid"},
				"ts":    []string{"2023-01-01T00:00:00Z"},
			},
			expectError: true,
			errMessage:  "failed to parse count",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var form CommentFetchForm
			err := form.GetParams(tt.values)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMessage)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.Count, form.Count)
				assert.Equal(t, tt.expected.Ts, form.Ts)
			}
		})
	}
}

func TestEmptyCommentForm(t *testing.T) {
	form := CommentForm{}
	model := form.ToCommentModel()

	assert.NotEqual(t, uuid.Nil, model.Id)
	assert.Empty(t, model.Text)
	assert.Len(t, model.Images, 0)
}

func TestCommentUpdateForm_EmptyAttachments(t *testing.T) {
	commentID := uuid.New()
	form := CommentUpdateForm{
		CommentForm: CommentForm{
			Text: "No attachments",
		},
	}
	model := form.ToCommentUpdateModel(commentID)

	assert.Equal(t, commentID, model.Id)
	assert.Equal(t, "No attachments", model.Text)
	assert.Len(t, model.Files, 0)
}

func TestCommentOut_FromCommentWithUnknownFileType(t *testing.T) {
	now := time.Now()
	comment := models.Comment{
		Id:   uuid.New(),
		Text: "Test",
		Images: []*models.File{
			{URL: "unknown.type", DisplayType: "unknown"},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
	userInfo := models.PublicUserInfo{Id: uuid.New()}

	var out CommentOut
	out.FromComment(comment, userInfo)

	// Unknown file types should be treated as regular files
	assert.Len(t, out.Files, 1)
	assert.Equal(t, "unknown.type", out.Files[0].URL)
}
