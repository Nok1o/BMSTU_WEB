//go:build unit
// +build unit

package friends_service

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"quickflow/shared/models"
	pb "quickflow/shared/proto/friends_service"
)

func TestFromModelFriendInfoToGrpc(t *testing.T) {
	tests := []struct {
		name     string
		input    models.FriendInfo
		expected *pb.GetFriendInfo
	}{
		{
			name: "complete friend info",
			input: models.FriendInfo{
				Id:         uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
				Username:   "johndoe",
				Firstname:  "John",
				Lastname:   "Doe",
				AvatarURL:  "http://example.com/avatar.jpg",
				University: "Stanford",
			},
			expected: &pb.GetFriendInfo{
				Id:         "123e4567-e89b-12d3-a456-426614174000",
				Username:   "johndoe",
				Firstname:  "John",
				Lastname:   "Doe",
				AvatarUrl:  "http://example.com/avatar.jpg",
				University: "Stanford",
			},
		},
		{
			name: "partial friend info",
			input: models.FriendInfo{
				Id:        uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
				Username:  "johndoe",
				Firstname: "John",
			},
			expected: &pb.GetFriendInfo{
				Id:         "123e4567-e89b-12d3-a456-426614174000",
				Username:   "johndoe",
				Firstname:  "John",
				Lastname:   "",
				AvatarUrl:  "",
				University: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fromModelFriendInfoToGrpc(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFromGrpcToModelFriendInfo(t *testing.T) {
	tests := []struct {
		name     string
		input    *pb.GetFriendInfo
		expected models.FriendInfo
	}{
		{
			name: "complete friend info",
			input: &pb.GetFriendInfo{
				Id:         "123e4567-e89b-12d3-a456-426614174000",
				Username:   "johndoe",
				Firstname:  "John",
				Lastname:   "Doe",
				AvatarUrl:  "http://example.com/avatar.jpg",
				University: "Stanford",
			},
			expected: models.FriendInfo{
				Id:         uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
				Username:   "johndoe",
				Firstname:  "John",
				Lastname:   "Doe",
				AvatarURL:  "http://example.com/avatar.jpg",
				University: "Stanford",
			},
		},
		{
			name: "partial friend info",
			input: &pb.GetFriendInfo{
				Id:        "123e4567-e89b-12d3-a456-426614174000",
				Username:  "johndoe",
				Firstname: "John",
			},
			expected: models.FriendInfo{
				Id:         uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
				Username:   "johndoe",
				Firstname:  "John",
				Lastname:   "",
				AvatarURL:  "",
				University: "",
			},
		},
		{
			name:     "nil input",
			input:    nil,
			expected: models.FriendInfo{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fromGrpcToModelFriendInfo(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFromModelFriendsInfoToGrpc(t *testing.T) {
	tests := []struct {
		name          string
		inputInfos    []models.FriendInfo
		inputCount    int
		expected      *pb.GetFriendsInfoResponse
		expectedCount int32
	}{
		{
			name: "multiple friends",
			inputInfos: []models.FriendInfo{
				{
					Id:        uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
					Username:  "user1",
					Firstname: "John",
				},
				{
					Id:        uuid.MustParse("223e4567-e89b-12d3-a456-426614174000"),
					Username:  "user2",
					Firstname: "Jane",
				},
			},
			inputCount: 2,
			expected: &pb.GetFriendsInfoResponse{
				Friends: []*pb.GetFriendInfo{
					{
						Id:        "123e4567-e89b-12d3-a456-426614174000",
						Username:  "user1",
						Firstname: "John",
					},
					{
						Id:        "223e4567-e89b-12d3-a456-426614174000",
						Username:  "user2",
						Firstname: "Jane",
					},
				},
				TotalCount: 2,
			},
			expectedCount: 2,
		},
		{
			name:          "empty friends list",
			inputInfos:    []models.FriendInfo{},
			inputCount:    0,
			expected:      &pb.GetFriendsInfoResponse{Friends: []*pb.GetFriendInfo{}, TotalCount: 0},
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FromModelFriendsInfoToGrpc(tt.inputInfos, tt.inputCount)
			assert.Equal(t, tt.expectedCount, result.TotalCount)
			assert.Equal(t, len(tt.inputInfos), len(result.Friends))
		})
	}
}

func TestFromGrpcToModelFriendsInfo(t *testing.T) {
	tests := []struct {
		name          string
		input         *pb.GetFriendsInfoResponse
		expectedInfos []models.FriendInfo
		expectedCount int
	}{
		{
			name: "multiple friends",
			input: &pb.GetFriendsInfoResponse{
				Friends: []*pb.GetFriendInfo{
					{
						Id:        "123e4567-e89b-12d3-a456-426614174000",
						Username:  "user1",
						Firstname: "John",
					},
					{
						Id:        "223e4567-e89b-12d3-a456-426614174000",
						Username:  "user2",
						Firstname: "Jane",
					},
				},
				TotalCount: 2,
			},
			expectedInfos: []models.FriendInfo{
				{
					Id:        uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
					Username:  "user1",
					Firstname: "John",
				},
				{
					Id:        uuid.MustParse("223e4567-e89b-12d3-a456-426614174000"),
					Username:  "user2",
					Firstname: "Jane",
				},
			},
			expectedCount: 2,
		},
		{
			name:          "empty friends list",
			input:         &pb.GetFriendsInfoResponse{Friends: []*pb.GetFriendInfo{}, TotalCount: 0},
			expectedInfos: []models.FriendInfo{},
			expectedCount: 0,
		},
		{
			name:          "nil input",
			input:         nil,
			expectedInfos: nil,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			infos, count := FromGrpcToModelFriendsInfo(tt.input)
			assert.Equal(t, tt.expectedInfos, infos)
			assert.Equal(t, tt.expectedCount, count)
		})
	}
}
