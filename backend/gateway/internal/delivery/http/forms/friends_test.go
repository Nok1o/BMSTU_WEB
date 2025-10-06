//go:build unit
// +build unit

package forms

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"quickflow/shared/models"
)

func TestFriendRequest(t *testing.T) {
	tests := []struct {
		name       string
		receiverID string
		expected   string
	}{
		{
			name:       "Valid UUID receiver",
			receiverID: "550e8400-e29b-41d4-a716-446655440000",
			expected:   "550e8400-e29b-41d4-a716-446655440000",
		},
		{
			name:       "Empty receiver",
			receiverID: "",
			expected:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := FriendRequest{
				ReceiverID: tt.receiverID,
			}
			assert.Equal(t, tt.expected, req.ReceiverID)
		})
	}
}

func TestFriendRequestDel(t *testing.T) {
	tests := []struct {
		name     string
		friendID string
		expected string
	}{
		{
			name:     "Valid UUID friend",
			friendID: "550e8400-e29b-41d4-a716-446655440000",
			expected: "550e8400-e29b-41d4-a716-446655440000",
		},
		{
			name:     "Empty friend",
			friendID: "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := FriendRequestDel{
				FriendID: tt.friendID,
			}
			assert.Equal(t, tt.expected, req.FriendID)
		})
	}
}

func TestToFriendsInfoOutForm(t *testing.T) {
	friendID := uuid.New()
	tests := []struct {
		name     string
		info     models.FriendInfo
		isOnline bool
		expected FriendsInfoOut
	}{
		{
			name: "Complete friend info online",
			info: models.FriendInfo{
				Id:         friendID,
				Username:   "johndoe",
				Firstname:  "John",
				Lastname:   "Doe",
				AvatarURL:  "http://example.com/avatar.jpg",
				University: "Stanford",
			},
			isOnline: true,
			expected: FriendsInfoOut{
				ID:         friendID,
				Username:   "johndoe",
				FirstName:  "John",
				LastName:   "Doe",
				AvatarURL:  "http://example.com/avatar.jpg",
				University: "Stanford",
				IsOnline:   true,
			},
		},
		{
			name: "Minimal friend info offline",
			info: models.FriendInfo{
				Id: friendID,
			},
			isOnline: false,
			expected: FriendsInfoOut{
				ID:       friendID,
				IsOnline: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var f FriendsInfoOut
			result := f.toFriendsInfoOutForm(tt.info, tt.isOnline)

			assert.Equal(t, tt.expected.ID, result.ID)
			assert.Equal(t, tt.expected.Username, result.Username)
			assert.Equal(t, tt.expected.FirstName, result.FirstName)
			assert.Equal(t, tt.expected.LastName, result.LastName)
			assert.Equal(t, tt.expected.AvatarURL, result.AvatarURL)
			assert.Equal(t, tt.expected.University, result.University)
			assert.Equal(t, tt.expected.IsOnline, result.IsOnline)
		})
	}
}

func TestToJson(t *testing.T) {
	friendID1 := uuid.New()
	friendID2 := uuid.New()

	tests := []struct {
		name          string
		friendsInfo   []models.FriendInfo
		friendsOnline []bool
		friendsCount  int
		expected      map[string]map[string]interface{}
	}{
		{
			name: "Multiple friends",
			friendsInfo: []models.FriendInfo{
				{
					Id:         friendID1,
					Username:   "johndoe",
					Firstname:  "John",
					Lastname:   "Doe",
					AvatarURL:  "http://example.com/avatar1.jpg",
					University: "Stanford",
				},
				{
					Id:        friendID2,
					Username:  "janedoe",
					Firstname: "Jane",
				},
			},
			friendsOnline: []bool{true, false},
			friendsCount:  2,
			expected: map[string]map[string]interface{}{
				"payload": {
					"friends": []FriendsInfoOut{
						{
							ID:         friendID1,
							Username:   "johndoe",
							FirstName:  "John",
							LastName:   "Doe",
							AvatarURL:  "http://example.com/avatar1.jpg",
							University: "Stanford",
							IsOnline:   true,
						},
						{
							ID:        friendID2,
							Username:  "janedoe",
							FirstName: "Jane",
							IsOnline:  false,
						},
					},
					"total_count": 2,
				},
			},
		},
		{
			name:          "Empty friends list",
			friendsInfo:   []models.FriendInfo{},
			friendsOnline: []bool{},
			friendsCount:  0,
			expected: map[string]map[string]interface{}{
				"payload": {
					"friends":     []FriendsInfoOut{},
					"total_count": 0,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var f FriendsInfoOut
			result := f.ToJson(tt.friendsInfo, tt.friendsOnline, tt.friendsCount)

			assert.Equal(t, tt.expected["payload"]["total_count"], result["payload"]["total_count"])

			expectedFriends := tt.expected["payload"]["friends"].([]FriendsInfoOut)
			actualFriends := result["payload"]["friends"].([]FriendsInfoOut)

			assert.Len(t, actualFriends, len(expectedFriends))

			for i, expectedFriend := range expectedFriends {
				assert.Equal(t, expectedFriend.ID, actualFriends[i].ID)
				assert.Equal(t, expectedFriend.Username, actualFriends[i].Username)
				assert.Equal(t, expectedFriend.FirstName, actualFriends[i].FirstName)
				assert.Equal(t, expectedFriend.LastName, actualFriends[i].LastName)
				assert.Equal(t, expectedFriend.AvatarURL, actualFriends[i].AvatarURL)
				assert.Equal(t, expectedFriend.University, actualFriends[i].University)
				assert.Equal(t, expectedFriend.IsOnline, actualFriends[i].IsOnline)
			}
		})
	}
}
