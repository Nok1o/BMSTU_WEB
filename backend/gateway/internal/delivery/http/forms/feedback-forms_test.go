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

func TestConvertTypeToModel(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    models.FeedbackType
		expectError bool
	}{
		{"General", FeedbackGeneral, models.FeedbackGeneral, false},
		{"Recommendation", FeedbackRecommendation, models.FeedbackRecommendation, false},
		{"Post", FeedbackPost, models.FeedbackPost, false},
		{"Profile", FeedbackProfile, models.FeedbackProfile, false},
		{"Auth", FeedbackAuth, models.FeedbackAuth, false},
		{"Messenger", FeedbackMessenger, models.FeedbackMessenger, false},
		{"Invalid", "invalid", models.FeedbackGeneral, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertTypeToModel(tt.input)

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, InvalidTypeError, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFeedbackForm_ToFeedback(t *testing.T) {
	respondentID := uuid.New()
	now := time.Now()

	tests := []struct {
		name        string
		form        FeedbackForm
		respondent  uuid.UUID
		expectError bool
	}{
		{
			name: "Valid general feedback",
			form: FeedbackForm{
				Type:   FeedbackGeneral,
				Text:   "Great service!",
				Rating: 5,
			},
			respondent:  respondentID,
			expectError: false,
		},
		{
			name: "Valid post feedback",
			form: FeedbackForm{
				Type:   FeedbackPost,
				Text:   "Interesting post",
				Rating: 4,
			},
			respondent:  respondentID,
			expectError: false,
		},
		{
			name: "Invalid type",
			form: FeedbackForm{
				Type:   "invalid",
				Text:   "Bad",
				Rating: 1,
			},
			respondent:  respondentID,
			expectError: true,
		},
		{
			name: "Empty text",
			form: FeedbackForm{
				Type:   FeedbackGeneral,
				Text:   "",
				Rating: 3,
			},
			respondent:  respondentID,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.form.ToFeedback(tt.respondent)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.NotEqual(t, uuid.Nil, result.Id)
				assert.Equal(t, tt.form.Text, result.Text)
				assert.Equal(t, tt.form.Rating, result.Rating)
				assert.Equal(t, tt.respondent, result.RespondentId)
				assert.WithinDuration(t, now, result.CreatedAt, time.Second)
			}
		})
	}
}

func TestFromFeedBack(t *testing.T) {
	feedbackID := uuid.New()
	respondentID := uuid.New()
	now := time.Now()

	tests := []struct {
		name     string
		feedback models.Feedback
		info     models.PublicUserInfo
		expected FeedbackFormOut
	}{
		{
			name: "Full feedback with user info",
			feedback: models.Feedback{
				Id:           feedbackID,
				Rating:       5,
				RespondentId: respondentID,
				Text:         "Excellent!",
				Type:         models.FeedbackGeneral,
				CreatedAt:    now,
			},
			info: models.PublicUserInfo{
				Username:  "johndoe",
				Firstname: "John",
				Lastname:  "Doe",
			},
			expected: FeedbackFormOut{
				Type:      string(models.FeedbackGeneral),
				Text:      "Excellent!",
				Rating:    5,
				Username:  "johndoe",
				Firstname: "John",
				Lastname:  "Doe",
			},
		},
		{
			name: "Minimal feedback",
			feedback: models.Feedback{
				Id:           feedbackID,
				Rating:       3,
				RespondentId: respondentID,
				Text:         "Average",
				Type:         models.FeedbackPost,
				CreatedAt:    now,
			},
			info: models.PublicUserInfo{},
			expected: FeedbackFormOut{
				Type:      string(models.FeedbackPost),
				Text:      "Average",
				Rating:    3,
				Username:  "",
				Firstname: "",
				Lastname:  "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FromFeedBack(tt.feedback, tt.info)

			assert.Equal(t, tt.expected.Type, result.Type)
			assert.Equal(t, tt.expected.Text, result.Text)
			assert.Equal(t, tt.expected.Rating, result.Rating)
			assert.Equal(t, tt.expected.Username, result.Username)
			assert.Equal(t, tt.expected.Firstname, result.Firstname)
			assert.Equal(t, tt.expected.Lastname, result.Lastname)
		})
	}
}

func TestGetFeedbackForm_GetParams(t *testing.T) {
	now := time.Now()
	nowStr := now.Format(time2.TimeStampLayout)

	tests := []struct {
		name        string
		values      url.Values
		expected    GetFeedbackForm
		expectError bool
		errMessage  string
	}{
		{
			name: "Valid params with type",
			values: url.Values{
				"feedback_count": []string{"10"},
				"ts":             []string{nowStr},
				"type":           []string{"general"},
			},
			expected: GetFeedbackForm{
				Count: 10,
				Ts:    now,
				Type:  models.FeedbackGeneral,
			},
			expectError: false,
		},
		{
			name: "Valid params without type",
			values: url.Values{
				"feedback_count": []string{"5"},
				"ts":             []string{nowStr},
			},
			expected: GetFeedbackForm{
				Count: 5,
				Ts:    now,
				Type:  "",
			},
			expectError: false,
		},
		{
			name: "Missing feedback_count",
			values: url.Values{
				"ts": []string{nowStr},
			},
			expectError: true,
			errMessage:  "chats_count parameter missing",
		},
		{
			name: "Invalid feedback_count",
			values: url.Values{
				"feedback_count": []string{"invalid"},
				"ts":             []string{nowStr},
			},
			expectError: true,
			errMessage:  "failed to parse feedback_count",
		},
		{
			name: "Invalid timestamp",
			values: url.Values{
				"feedback_count": []string{"10"},
				"ts":             []string{"invalid"},
			},
			expected: GetFeedbackForm{
				Count: 10,
				Ts:    time.Now(), // Will be set to current time
				Type:  "",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var form GetFeedbackForm
			err := form.GetParams(tt.values)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMessage)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.Count, form.Count)
				assert.Equal(t, tt.expected.Type, form.Type)

				// For timestamp, we can't compare directly due to potential slight differences
				if tt.values.Get("ts") == "invalid" {
					assert.WithinDuration(t, time.Now(), form.Ts, time.Second)
				} else {
					assert.Equal(t, tt.expected.Ts.Format(time2.TimeStampLayout), form.Ts.Format(time2.TimeStampLayout))
				}
			}
		})
	}
}

func TestFeedbackForm_ZeroRating(t *testing.T) {
	form := FeedbackForm{
		Type:   FeedbackGeneral,
		Text:   "Zero rating",
		Rating: 0,
	}
	respondent := uuid.New()

	result, err := form.ToFeedback(respondent)

	assert.NoError(t, err)
	assert.Equal(t, 0, result.Rating)
}

func TestFromFeedBack_EmptyFeedback(t *testing.T) {
	emptyFeedback := models.Feedback{}
	userInfo := models.PublicUserInfo{
		Username: "testuser",
	}

	result := FromFeedBack(emptyFeedback, userInfo)

	assert.Equal(t, "", result.Type)
	assert.Equal(t, "", result.Text)
	assert.Equal(t, 0, result.Rating)
	assert.Equal(t, "testuser", result.Username)
}

func TestGetFeedbackForm_DefaultValues(t *testing.T) {
	values := url.Values{
		"feedback_count": []string{"10"},
		// No ts provided
	}

	var form GetFeedbackForm
	err := form.GetParams(values)

	assert.NoError(t, err)
	assert.Equal(t, 10, form.Count)
	assert.WithinDuration(t, time.Now(), form.Ts, time.Second)
	assert.Equal(t, models.FeedbackType(""), form.Type)
}
