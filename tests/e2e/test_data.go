// tests/e2e/test_data.go
package e2e

import (
	"time"
)

// TestData содержит тестовые данные для E2E тестов
var TestData = struct {
	Users []TestUserData
	Posts []PostTestData
}{
	Users: []TestUserData{
		{
			Username: "e2e_test_user_1",
			Email:    "e2e_test1@example.com",
			Password: "TestPassword123!",
			Profile: ProfileTestData{
				FirstName: "E2E",
				LastName:  "TestUser1",
				Sex:       1,
				BirthDate: "1990-01-01",
				Bio:       "E2E Test User 1",
			},
		},
		{
			Username: "e2e_test_user_2",
			Email:    "e2e_test2@example.com",
			Password: "TestPassword123!",
			Profile: ProfileTestData{
				FirstName: "E2E",
				LastName:  "TestUser2",
				Sex:       1,
				BirthDate: "1990-02-02",
				Bio:       "E2E Test User 2",
			},
		},
	},
	Posts: []PostTestData{
		{
			Title:   "E2E Test Post 1",
			Content: "This is the first E2E test post content",
			Privacy: "public",
		},
		{
			Title:   "E2E Test Post 2",
			Content: "This is the second E2E test post content",
			Privacy: "public",
		},
	},
}

type TestUserData struct {
	Username string
	Email    string
	Password string
	Profile  ProfileTestData
}

type ProfileTestData struct {
	FirstName string
	LastName  string
	Sex       int
	BirthDate string
	Bio       string
}

type PostTestData struct {
	Title   string
	Content string
	Privacy string
}

type CommentTestData struct {
	Text string
}

type CommunityTestData struct {
	Nickname    string
	Name        string
	Description string
}

type MessageTestData struct {
	Text string
}

// GetTestTimestamp возвращает тестовую временную метку
func GetTestTimestamp() string {
	return time.Now().Add(-time.Hour).Format("2006-01-02T15:04:05Z")
}
