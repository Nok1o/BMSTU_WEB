// tests/e2e/main_test.go
package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
	"github.com/stretchr/testify/assert"
)

type E2ETestSuite struct {
	suite.Suite
	baseURL     string
	client      *http.Client
	testUser1   *TestUser
	testUser2   *TestUser
	communityID string
	postID      string
	commentID   string
	chatID      string
	messageID   uuid.UUID
}

type TestUser struct {
	Username    string
	Email       string
	Password    string
	UserID      string
	Session     string
	CSRFToken   string
	ProfileInfo *ProfileInfo
}

// DTO структуры
type SignUpForm struct {
	Login       string `json:"username"`
	Password    string `json:"password"`
	Name        string `json:"firstname"`
	Surname     string `json:"lastname"`
	Sex         int    `json:"sex"`
	DateOfBirth string `json:"birth_date"`
}

type AuthForm struct {
	Login    string `json:"username"`
	Password string `json:"password"`
}

type PostForm struct {
	Text  string   `json:"text"`
	Media []string `json:"media,omitempty"`
	Audio []string `json:"audio,omitempty"`
	Files []string `json:"files,omitempty"`
}

type CommentForm struct {
	Text     string   `json:"text"`
	Media    []string `json:"media,omitempty"`
	Audio    []string `json:"audio,omitempty"`
	Files    []string `json:"files,omitempty"`
	Stickers []string `json:"stickers,omitempty"`
}

type CreateCommunityForm struct {
	Nickname    string `json:"nickname"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type ProfileInfo struct {
	Username      string `json:"username,omitempty"`
	Name          string `json:"firstname"`
	Surname       string `json:"lastname"`
	Sex           int    `json:"sex"`
	DateOfBirth   string `json:"birth_date"`
	Bio           string `json:"bio"`
	AvatarUrl     string `json:"avatar_url,omitempty"`
	BackgroundUrl string `json:"cover_url,omitempty"`
}

type MessageForm struct {
	Text       string    `json:"text,omitempty"`
	ChatId     uuid.UUID `json:"chat_id,omitempty"`
	Media      []string  `json:"media,omitempty"`
	Audio      []string  `json:"audio,omitempty"`
	File       []string  `json:"files,omitempty"`
	Stickers   []string  `json:"stickers,omitempty"`
	ReceiverId uuid.UUID `json:"receiver_id,omitempty"`
}

// Response структуры
type AuthResponse struct {
	Message string `json:"message"`
}

type SignUpResponse struct {
	ID        string `json:"id"`
	SessionID string `json:"session_id"`
}

type PostResponse struct {
	Payload struct {
		ID string `json:"id"`
	} `json:"payload"`
}

type CommentResponse struct {
	ID string `json:"id"`
}

type CommunityResponse struct {
	Payload struct {
		ID string `json:"id"`
	} `json:"payload"`
}

type ProfileResponse struct {
	ID   string      `json:"id"`
	Info ProfileInfo `json:"profile"`
}

type CSRFResponse struct {
	CSRFToken string `json:"csrf_token"`
}

type FeedResponse struct {
	Posts []PostOut
}

type SearchUsersResponse struct {
	Users []UserSearchResult `json:"payload"`
}

type UserSearchResult struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Name     string `json:"firstname"`
	Surname  string `json:"lastname"`
}

type MessagesResponse struct {
	Messages []MessageOut `json:"messages"`
}

type MessageOut struct {
	ID        uuid.UUID `json:"id"`
	Text      string    `json:"text"`
	CreatedAt string    `json:"created_at"`
	Sender    UserOut   `json:"sender"`
}

type PostOut struct {
	ID           string `json:"id"`
	Text         string `json:"text"`
	CreatedAt    string `json:"created_at"`
	LikeCount    int    `json:"like_count"`
	CommentCount int    `json:"comment_count"`
	IsLiked      bool   `json:"is_liked"`
}

type UserOut struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Name     string `json:"firstname"`
	Surname  string `json:"lastname"`
}

func TestE2ESuite(t *testing.T) {
	suite.RunSuite(t, new(E2ETestSuite))
}

func (s *E2ETestSuite) BeforeAll(t provider.T) {
	t.Epic("E2E")
	t.Feature("Full User Journey")
	t.Description("Setup test environment and initialize test users")

	t.WithNewStep("Setup test suite", func(ctx provider.StepCtx) {
		s.baseURL = "http://gateway:8080/api/v1"
		s.client = &http.Client{
			Timeout: 30 * time.Second,
		}

		// Тестовые пользователи с уникальными именами
		s.testUser1 = &TestUser{
			Username: "testuser1_" + uuid.New().String()[:8],
			Email:    "test1_" + uuid.New().String()[:8] + "@example.com",
			Password: "Password123!",
		}

		s.testUser2 = &TestUser{
			Username: "testuser2_" + uuid.New().String()[:8],
			Email:    "test2_" + uuid.New().String()[:8] + "@example.com",
			Password: "Password123!",
		}

		ctx.Logf("Initialized test users: %s and %s", s.testUser1.Username, s.testUser2.Username)
	})
}

func (s *E2ETestSuite) TestFullUserJourney(t provider.T) {
	t.Epic("E2E")
	t.Feature("Full User Journey")
	t.Story("Complete user flow from registration to messaging")
	t.Description("Complete end-to-end test covering user registration, authentication, profile management, content creation, social interactions, and real-time messaging")

	t.WithNewStep("=== Starting Full E2E User Journey Test ===", func(ctx provider.StepCtx) {
		// 1. Регистрация и получение user_id из ответа API
		s.registerAndGetUserIDs(t)

		// 2. Логин (для получения новых сессий)
		s.loginUsers(t)

		// 3. Получение CSRF токенов
		s.getCSRFTokens(t)

		// 4. Создание профилей
		s.updateProfiles(t)

		// 5. Создание постов
		s.createPosts(t)

		// 6. Создание комментариев
		s.createComments(t)

		// 7. Лайки
		s.likeContent(t)

		// 8. Создание сообщества
		s.createCommunity(t)

		// 9. Поиск и добавление в друзья
		s.searchAndAddFriends(t)

		// 10. Обмен сообщениями
		s.exchangeMessages(t)

		// 11. Проверка ленты
		s.checkFeed(t)

		// 12. Получение профилей
		s.getProfiles(t)

		t.Log("=== E2E Test Completed Successfully ===")
	})
}

func (s *E2ETestSuite) registerAndGetUserIDs(t provider.T) {
	t.WithNewStep("1. Registering test users and getting user IDs", func(ctx provider.StepCtx) {
		// Регистрация первого пользователя
		signUpData1 := SignUpForm{
			Login:       s.testUser1.Username,
			Password:    s.testUser1.Password,
			Name:        "Nikita",
			Surname:     "Mogilin",
			Sex:         0,
			DateOfBirth: "1990-01-01",
		}

		resp1 := s.makeRequest("POST", "/signup", signUpData1, nil, false, false)
		s.requireStatus(t, resp1, http.StatusCreated)

		s.testUser1.Session = s.getSessionCookie(resp1)
		var signUpResp1 SignUpResponse
		s.parseResponse(resp1, &signUpResp1)
		s.testUser1.UserID = signUpResp1.ID
		t.Require().NotEmpty(s.testUser1.UserID, "Signup should return first user ID")

		// Регистрация второго пользователя
		signUpData2 := SignUpForm{
			Login:       s.testUser2.Username,
			Password:    s.testUser2.Password,
			Name:        "Kikol",
			Surname:     "Ujaja",
			Sex:         1,
			DateOfBirth: "1990-02-02",
		}

		resp2 := s.makeRequest("POST", "/signup", signUpData2, nil, false, false)
		s.requireStatus(t, resp2, http.StatusCreated)

		s.testUser2.Session = s.getSessionCookie(resp2)
		var signUpResp2 SignUpResponse
		s.parseResponse(resp2, &signUpResp2)
		s.testUser2.UserID = signUpResp2.ID
		t.Require().NotEmpty(s.testUser2.UserID, "Signup should return second user ID")

		t.Logf("User1 ID: %s, User2 ID: %s", s.testUser1.UserID, s.testUser2.UserID)
	})
}

func (s *E2ETestSuite) loginUsers(t provider.T) {
	t.WithNewStep("2. Logging in test users to get fresh sessions", func(ctx provider.StepCtx) {
		// Логин первого пользователя для получения новой сессии
		loginData1 := AuthForm{
			Login:    s.testUser1.Username,
			Password: s.testUser1.Password,
		}

		resp1 := s.makeRequest("POST", "/login", loginData1, nil, false, false)
		s.requireStatus(t, resp1, http.StatusCreated)
		s.testUser1.Session = s.getSessionCookie(resp1)

		// Логин второго пользователя для получения новой сессии
		loginData2 := AuthForm{
			Login:    s.testUser2.Username,
			Password: s.testUser2.Password,
		}

		resp2 := s.makeRequest("POST", "/login", loginData2, nil, false, false)
		s.requireStatus(t, resp2, http.StatusCreated)
		s.testUser2.Session = s.getSessionCookie(resp2)

		ctx.Log("Both users successfully logged in")
	})
}

func (s *E2ETestSuite) getCSRFTokens(t provider.T) {
	t.WithNewStep("3. Getting CSRF tokens", func(ctx provider.StepCtx) {
		// CSRF для первого пользователя
		req1, _ := http.NewRequest("GET", s.baseURL+"/csrf", nil)
		req1.Header.Set("Cookie", s.testUser1.Session)
		resp1, _ := s.client.Do(req1)
		s.requireStatus(t, resp1, http.StatusOK)

		csrfTokenHeader := resp1.Header.Get("X-CSRF-Token")
		s.testUser1.CSRFToken = csrfTokenHeader

		// CSRF для второго пользователя
		req2, _ := http.NewRequest("GET", s.baseURL+"/csrf", nil)
		req2.Header.Set("Cookie", s.testUser2.Session)
		resp2, _ := s.client.Do(req2)
		s.requireStatus(t, resp2, http.StatusOK)

		csrfTokenHeader = resp2.Header.Get("X-CSRF-Token")
		s.testUser2.CSRFToken = csrfTokenHeader

		ctx.Log("CSRF tokens obtained for both users")
	})
}

func (s *E2ETestSuite) updateProfiles(t provider.T) {
	t.WithNewStep("4. Updating user profiles", func(ctx provider.StepCtx) {
		// Обновление профиля первого пользователя
		profileData1 := map[string]interface{}{
			"firstname":  "OtherFirstName",
			"lastname":   "OtherLastName",
			"sex":        0,
			"birth_date": "1990-01-01",
			"bio":        "Test bio for user 1",
		}

		resp1 := s.makeRequest("POST", "/profile", profileData1, s.testUser1, true, true)
		s.requireStatus(t, resp1, http.StatusOK)

		// Обновление профиля второго пользователя
		profileData2 := map[string]interface{}{
			"firstname":  "OtherFirstName",
			"lastname":   "OtherLastName",
			"sex":        1,
			"birth_date": "1990-02-02",
			"bio":        "Test bio for user 2",
		}

		resp2 := s.makeRequest("POST", "/profile", profileData2, s.testUser2, true, true)
		s.requireStatus(t, resp2, http.StatusOK)

		ctx.Log("User profiles updated successfully")
	})
}

func (s *E2ETestSuite) createPosts(t provider.T) {
	t.WithNewStep("5. Creating posts", func(ctx provider.StepCtx) {
		// Создание поста первым пользователем
		postData1 := PostForm{
			Text:  "This is the content of the first test post",
			Media: []string{"pict1", "pict2"},
			Audio: []string{"audio1"},
			Files: []string{"file1", "file2"},
		}

		resp1 := s.makeRequest("POST", "/post", postData1, s.testUser1, true, false)
		s.requireStatus(t, resp1, http.StatusOK)

		var postResp1 PostResponse
		s.parseResponse(resp1, &postResp1)
		s.postID = postResp1.Payload.ID

		t.Logf("Created post with ID: %s", s.postID)

		// Создание поста вторым пользователем
		postData2 := PostForm{
			Text:  "This is the content of the second test post",
			Media: []string{"pictA", "pictB"},
			Audio: []string{"audioA"},
			Files: []string{"fileA", "fileB"},
		}

		resp2 := s.makeRequest("POST", "/post", postData2, s.testUser2, true, false)
		s.requireStatus(t, resp2, http.StatusOK)

		ctx.Log("Posts created successfully by both users")
	})
}

func (s *E2ETestSuite) createComments(t provider.T) {
	t.WithNewStep("6. Creating comments", func(ctx provider.StepCtx) {
		// Комментарий второго пользователя к посту первого
		commentData := CommentForm{
			Text: "This is a test comment on the post",
		}

		resp := s.makeRequest("POST", fmt.Sprintf("/posts/%s/comment", s.postID),
			commentData, s.testUser2, true, false)
		s.requireStatus(t, resp, http.StatusOK)

		var commentResp CommentResponse
		s.parseResponse(resp, &commentResp)
		s.commentID = commentResp.ID

		t.Logf("Created comment with ID: %s", s.commentID)
	})
}

func (s *E2ETestSuite) likeContent(t provider.T) {
	t.WithNewStep("7. Liking content", func(ctx provider.StepCtx) {
		// Лайк поста первым пользователем (своему посту)
		resp1 := s.makeRequest("POST", fmt.Sprintf("/posts/%s/like", s.postID),
			nil, s.testUser1, true, false)
		s.requireStatus(t, resp1, http.StatusCreated)

		// Лайк комментария вторым пользователем (своему комментарию)
		resp2 := s.makeRequest("POST", fmt.Sprintf("/comments/%s/like", s.commentID),
			nil, s.testUser2, true, false)
		s.requireStatus(t, resp2, http.StatusCreated)

		ctx.Log("Content liked successfully")
	})
}

func (s *E2ETestSuite) createCommunity(t provider.T) {
	t.WithNewStep("8. Creating community", func(ctx provider.StepCtx) {
		communityData := CreateCommunityForm{
			Nickname:    "community" + uuid.New().String()[:8],
			Name:        "testCommunitynew",
			Description: "A community for testing purposes",
		}

		resp := s.makeRequest("POST", "/community", communityData, s.testUser1, true, true)
		s.requireStatus(t, resp, http.StatusOK)

		var communityResp CommunityResponse
		s.parseResponse(resp, &communityResp)
		s.communityID = communityResp.Payload.ID

		t.Logf("Created community with ID: %s", s.communityID)
	})
}

func (s *E2ETestSuite) searchAndAddFriends(t provider.T) {
	t.WithNewStep("9. Searching and adding friends", func(ctx provider.StepCtx) {
		// Поиск второго пользователя первым пользователем
		searchResp := s.makeRequest("GET",
			fmt.Sprintf("/users/search?to_search=%s&count=1", s.testUser2.Username),
			nil, s.testUser1, false, false)
		s.requireStatus(t, searchResp, http.StatusOK)

		var searchResult SearchUsersResponse
		s.parseResponse(searchResp, &searchResult)

		t.Require().Greater(len(searchResult.Users), 0, "Should find user in search")
		assert.Equal(t, s.testUser2.UserID, searchResult.Users[0].ID, "Found user ID should match")

		// Отправка запроса в друзья
		friendData := map[string]interface{}{
			"receiver_id": s.testUser2.UserID,
		}

		resp2 := s.makeRequest("POST", "/follow", friendData, s.testUser1, true, false)
		s.requireStatus(t, resp2, http.StatusOK)

		ctx.Log("Friend request sent successfully")
	})
}

func (s *E2ETestSuite) exchangeMessages(t provider.T) {
	t.WithNewStep("10. Exchanging messages via WebSocket", func(ctx provider.StepCtx) {
		// Подключаем первого пользователя к WebSocket
		wsConn1, err := s.connectWebSocket(s.testUser1.Session)
		assert.NoError(t, err, "User1 should connect to WebSocket")
		defer wsConn1.Close()

		// Подключаем второго пользователя к WebSocket
		wsConn2, err := s.connectWebSocket(s.testUser2.Session)
		assert.NoError(t, err, "User2 should connect to WebSocket")
		defer wsConn2.Close()

		// Отправка сообщения от первого пользователя второму
		messageData := map[string]interface{}{
			"type": "message",
			"payload": map[string]interface{}{
				"text":        "Hello from user 1!",
				"receiver_id": s.testUser2.UserID,
				"media":       []string{"media.url", "video.url"},
				"audio":       []string{"audio1.file", "audio2.file"},
				"files":       []string{"file1.url", "file2.url"},
			},
		}

		err = s.sendWebSocketMessage(wsConn1, messageData)
		assert.NoError(t, err, "User1 should send message via WebSocket")

		// Ждем получения сообщения вторым пользователем
		receivedMsg, err := s.waitForSpecificMessage(wsConn2, "Hello from user 1!", 10*time.Second)
		assert.NoError(t, err, "User2 should receive message from user1")

		// Проверяем полученное сообщение
		assert.Equal(t, "message", receivedMsg["type"])
		payload, ok := receivedMsg["payload"].(map[string]interface{})
		assert.True(t, ok, "Payload should be a map")
		assert.Equal(t, "Hello from user 1!", payload["text"])

		// Проверяем sender_id (может приходить как строка или как nil, если не установлен)
		if senderID, exists := payload["sender_id"]; exists && senderID != nil {
			assert.Equal(t, s.testUser1.UserID, senderID)
		}

		// Отправка ответного сообщения от второго пользователя первому
		replyData := map[string]interface{}{
			"type": "message",
			"payload": map[string]interface{}{
				"text":        "Hello back from user 2!",
				"receiver_id": s.testUser1.UserID,
				"media":       []string{"reply_media.url"},
				"audio":       []string{"reply_audio.file"},
				"files":       []string{"reply_file.url"},
			},
		}

		err = s.sendWebSocketMessage(wsConn2, replyData)
		assert.NoError(t, err, "User2 should send reply via WebSocket")

		// Ждем получения ответного сообщения первым пользователем
		receivedReply, err := s.waitForSpecificMessage(wsConn1, "Hello back from user 2!", 10*time.Second)
		assert.NoError(t, err, "User1 should receive reply from user2")

		// Проверяем полученное ответное сообщение
		assert.Equal(t, "message", receivedReply["type"])
		replyPayload, ok := receivedReply["payload"].(map[string]interface{})
		assert.True(t, ok, "Reply payload should be a map")
		assert.Equal(t, "Hello back from user 2!", replyPayload["text"])

		// Проверяем sender_id ответного сообщения
		if senderID, exists := replyPayload["sender_id"]; exists && senderID != nil {
			assert.Equal(t, s.testUser2.UserID, senderID)
		}

		t.Log("WebSocket message exchange completed successfully")
	})
}

// Вспомогательные методы для работы с WebSocket
func (s *E2ETestSuite) connectWebSocket(session string) (*websocket.Conn, error) {
	// Заменяем http на ws и добавляем путь к WebSocket endpoint
	wsURL := "ws" + strings.TrimPrefix(s.baseURL, "http") + "/ws"

	header := http.Header{}
	header.Set("Cookie", session)

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, header)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to WebSocket: %v", err)
	}

	return conn, nil
}

func (s *E2ETestSuite) sendWebSocketMessage(conn *websocket.Conn, message map[string]interface{}) error {
	messageJSON, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}

	err = conn.WriteMessage(websocket.TextMessage, messageJSON)
	if err != nil {
		return fmt.Errorf("failed to send WebSocket message: %v", err)
	}

	return nil
}

// waitForSpecificMessage ждет сообщение с определенным текстом
func (s *E2ETestSuite) waitForSpecificMessage(conn *websocket.Conn, expectedText string, timeout time.Duration) (map[string]interface{}, error) {
	deadline := time.Now().Add(timeout)
	conn.SetReadDeadline(deadline)

	for {
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("timeout waiting for message: %s", expectedText)
		}

		_, message, err := conn.ReadMessage()
		if err != nil {
			// Продолжаем попытки при таймаутах
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
				continue
			}
			return nil, err
		}

		var messageData map[string]interface{}
		if err := json.Unmarshal(message, &messageData); err != nil {
			continue
		}

		// Проверяем тип сообщения
		if msgType, ok := messageData["type"].(string); ok && msgType == "message" {
			if payload, ok := messageData["payload"].(map[string]interface{}); ok {
				if text, ok := payload["text"].(string); ok && text == expectedText {
					return messageData, nil
				}
			}
		}
	}
}

// Альтернативный метод - получаем все сообщения в течение таймаута и ищем нужное
func (s *E2ETestSuite) receiveWebSocketMessages(conn *websocket.Conn, timeout time.Duration) ([]map[string]interface{}, error) {
	var messages []map[string]interface{}
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))

		_, message, err := conn.ReadMessage()
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			return nil, err
		}

		var messageData map[string]interface{}
		if err := json.Unmarshal(message, &messageData); err != nil {
			continue
		}

		messages = append(messages, messageData)
	}

	return messages, nil
}

func (s *E2ETestSuite) checkFeed(t provider.T) {
	t.WithNewStep("11. Checking feed", func(ctx provider.StepCtx) {
		// Получение ленты первого пользователя
		resp := s.makeRequest("GET", "/feed?count=100", nil, s.testUser1, false, false)
		s.requireStatus(t, resp, http.StatusOK)

		var feedResp []PostOut
		s.parseResponse(resp, &feedResp)

		// Проверяем, что в ленте есть посты
		assert.Greater(t, len(feedResp), 0, "Feed should contain posts")

		// Ищем наш тестовый пост в ленте
		found := false
		for _, post := range feedResp {
			if post.ID == s.postID {
				found = true
				assert.Equal(t, "This is the content of the first test post", post.Text)
				break
			}
		}
		assert.True(t, found, "Test post should be in feed")

		ctx.Logf("Feed checked successfully, found %d posts", len(feedResp))
	})
}

func (s *E2ETestSuite) getProfiles(t provider.T) {
	t.WithNewStep("12. Getting profiles", func(ctx provider.StepCtx) {
		// Получение собственного профиля первого пользователя
		myProfileResp := s.makeRequest("GET", "/my_profile", nil, s.testUser1, false, false)
		s.requireStatus(t, myProfileResp, http.StatusOK)

		var myProfile ProfileResponse
		s.parseResponse(myProfileResp, &myProfile)
		assert.Equal(t, "OtherFirstName", myProfile.Info.Name)

		// Получение профиля второго пользователя по username
		userProfileResp := s.makeRequest("GET",
			fmt.Sprintf("/profiles/%s", s.testUser2.Username),
			nil, s.testUser1, false, false)
		s.requireStatus(t, userProfileResp, http.StatusOK)

		ctx.Log("Profiles retrieved successfully")
	})
}

func (s *E2ETestSuite) makeRequest(method, endpoint string, data interface{}, user *TestUser, useCSRF bool, useMultipart bool) *http.Response {
	var body io.Reader
	var contentType string

	if data != nil {
		if useMultipart {
			var b bytes.Buffer
			writer := multipart.NewWriter(&b)

			// Допустим, data — это map[string]interface{}
			if m, ok := data.(map[string]interface{}); ok {
				for key, value := range m {
					switch v := value.(type) {
					case string:
						_ = writer.WriteField(key, v)
					default:
						// Всё остальное маршалим в JSON
						jsonVal, _ := json.Marshal(v)
						_ = writer.WriteField(key, string(jsonVal))
					}
				}
			} else {
				jb, err := json.Marshal(data)
				if err != nil {
					return nil
				}

				// обратно в map
				var formMap map[string]interface{}
				if err := json.Unmarshal(jb, &formMap); err != nil {
					return nil
				}

				// кладём каждое поле как form-data
				for key, val := range formMap {
					if val == nil {
						continue
					}
					switch v := val.(type) {
					case string:
						_ = writer.WriteField(key, v)
					default:
						_ = writer.WriteField(key, fmt.Sprint(v))
					}
				}
			}

			_ = writer.Close()
			body = &b
			contentType = writer.FormDataContentType()
		} else {
			// application/json
			jsonData, _ := json.Marshal(data)
			body = bytes.NewReader(jsonData)
			contentType = "application/json"
		}
	}

	req, _ := http.NewRequest(method, s.baseURL+endpoint, body)

	if user != nil {
		if useCSRF {
			req.Header.Set("X-CSRF-Token", user.CSRFToken)
			req.Header.Set("Cookie", user.Session+"; csrf_token="+user.CSRFToken)
		} else {
			req.Header.Set("Cookie", user.Session)
		}
	}

	if data != nil {
		req.Header.Set("Content-Type", contentType)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil
	}

	return resp
}

func (s *E2ETestSuite) requireStatus(t provider.T, resp *http.Response, expected int) {
	t.Helper()
	t.Require().NotNil(resp, "request should return an HTTP response")
	t.Require().Equal(expected, resp.StatusCode)
}

func (s *E2ETestSuite) parseResponse(resp *http.Response, target interface{}) {
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, target)
	if err != nil {
		return
	}
}

func (s *E2ETestSuite) getSessionCookie(resp *http.Response) string {
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "session" || strings.Contains(cookie.Name, "quickflow") {
			return cookie.Name + "=" + cookie.Value
		}
	}
	return ""
}

func (s *E2ETestSuite) AfterAll(t provider.T) {
	t.WithNewStep("Cleanup test data", func(ctx provider.StepCtx) {
		t.Log("Cleaning up test data")
		// Здесь можно добавить очистку тестовых данных
	})
}
