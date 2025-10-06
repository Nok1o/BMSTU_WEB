//go:build e2e
// +build e2e

package e2e_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"testing"
	"time"
)

type TestClient struct {
	t      *testing.T
	client *http.Client
	base   string
	cookie *http.Cookie
}

// NewTestClient создает клиент для подключения к существующему серверу
func NewTestClient(t *testing.T, baseURL string) *TestClient {
	return &TestClient{
		t: t,
		client: &http.Client{
			Timeout: 30 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		base: baseURL,
	}
}

func (c *TestClient) doRequest(req *http.Request, expected int) *http.Response {
	if c.cookie != nil {
		req.AddCookie(c.cookie)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		c.t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != expected {
		b, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		c.t.Fatalf("unexpected status: expected %d, got %d, body: %s", expected, resp.StatusCode, string(b))
	}

	// сохраняем cookie после логина
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "session_id" {
			c.cookie = cookie
		}
	}
	return resp
}

func (c *TestClient) post(path string, body any, expected int) *http.Response {
	buf := &bytes.Buffer{}
	_ = json.NewEncoder(buf).Encode(body)
	req, _ := http.NewRequest(http.MethodPost, c.base+path, buf)
	req.Header.Set("Content-Type", "application/json")
	return c.doRequest(req, expected)
}

func (c *TestClient) put(path string, body any, expected int) *http.Response {
	buf := &bytes.Buffer{}
	_ = json.NewEncoder(buf).Encode(body)
	req, _ := http.NewRequest(http.MethodPut, c.base+path, buf)
	req.Header.Set("Content-Type", "application/json")
	return c.doRequest(req, expected)
}

func (c *TestClient) delete(path string, expected int) *http.Response {
	req, _ := http.NewRequest(http.MethodDelete, c.base+path, nil)
	return c.doRequest(req, expected)
}

func (c *TestClient) get(path string, expected int) *http.Response {
	req, _ := http.NewRequest(http.MethodGet, c.base+path, nil)
	return c.doRequest(req, expected)
}

// Генерация случайных данных для изоляции тестов
func generateRandomEmail(prefix string) string {
	return fmt.Sprintf("%s%d@test.com", prefix, rand.Intn(100000))
}

func generateRandomUsername(prefix string) string {
	return fmt.Sprintf("%s%d", prefix, rand.Intn(100000))
}

// getServerURL возвращает URL сервера из переменной окружения или значение по умолчанию
func getServerURL() string {
	url := os.Getenv("TEST_SERVER_URL")
	if url == "" {
		return "http://localhost:8080"
	}
	return url
}

func TestE2E_FullScenario(t *testing.T) {
	// Инициализация генератора случайных чисел
	rand.Seed(time.Now().UnixNano())

	// Получаем URL существующего сервера
	serverURL := getServerURL()
	t.Logf("Connecting to server: %s", serverURL)

	// Проверяем доступность сервера
	if !isServerAvailable(serverURL) {
		t.Fatalf("Server is not available at %s", serverURL)
	}

	// Создаем клиентов с уникальными данными для изоляции тестов
	timestamp := time.Now().UnixNano()
	clientA := NewTestClient(t, serverURL)
	clientB := NewTestClient(t, serverURL)

	// Генерируем уникальные данные для этого запуска теста
	aliceUsername := generateRandomUsername("alice")
	aliceEmail := generateRandomEmail("alice")
	bobUsername := generateRandomUsername("bob")
	bobEmail := generateRandomEmail("bob")

	// === Сценарий ===

	// 1. Регистрация пользователя A
	t.Log("Step 1: Registering user A")
	clientA.post("/signup", map[string]string{
		"username": aliceUsername,
		"email":    aliceEmail,
		"password": "123456",
	}, http.StatusOK)

	// 2. Логаут A
	t.Log("Step 2: Logout user A")
	clientA.post("/logout", nil, http.StatusOK)

	// 3. Логин A
	t.Log("Step 3: Login user A")
	clientA.post("/login", map[string]string{
		"email":    aliceEmail,
		"password": "123456",
	}, http.StatusOK)

	// 4. Создание поста
	t.Log("Step 4: Creating post")
	resp := clientA.post("/post", map[string]string{
		"content": fmt.Sprintf("Привет, это мой первый пост! Тест %d", timestamp),
	}, http.StatusOK)

	var postResp struct {
		PostID string `json:"post_id"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&postResp)
	resp.Body.Close()

	// 5. Добавление комментария
	t.Log("Step 5: Adding comment")
	clientA.post("/posts/"+postResp.PostID+"/comment", map[string]string{
		"text": fmt.Sprintf("Мой первый коммент %d", timestamp),
	}, http.StatusOK)

	// 6. Изменение поста
	t.Log("Step 6: Updating post")
	clientA.put("/posts/"+postResp.PostID, map[string]string{
		"content": fmt.Sprintf("Обновлённый текст поста %d", timestamp),
	}, http.StatusOK)

	// 7. Удаление поста
	t.Log("Step 7: Deleting post")
	clientA.delete("/posts/"+postResp.PostID, http.StatusOK)

	// 8. Регистрация пользователя B
	t.Log("Step 8: Registering user B")
	clientB.post("/signup", map[string]string{
		"username": bobUsername,
		"email":    bobEmail,
		"password": "qwerty",
	}, http.StatusOK)

	// 9. A добавляет B в друзья
	t.Log("Step 9: User A follows user B")
	clientA.post("/follow", map[string]string{
		"username": bobUsername,
	}, http.StatusOK)

	// 10. Логин B
	t.Log("Step 10: Login user B")
	clientB.post("/login", map[string]string{
		"email":    bobEmail,
		"password": "qwerty",
	}, http.StatusOK)

	// 11. B пишет сообщение A
	t.Log("Step 11: User B sends message to user A")
	clientB.post("/chats", map[string]string{
		"receiver": aliceUsername,
		"text":     fmt.Sprintf("Привет, я Боб! Тест %d", timestamp),
	}, http.StatusOK)

	t.Log("E2E test completed successfully")
}

func TestE2E_AuthenticationFlow(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	serverURL := getServerURL()

	if !isServerAvailable(serverURL) {
		t.Skipf("Server is not available at %s, skipping test", serverURL)
	}

	client := NewTestClient(t, serverURL)

	// Генерируем уникальные данные
	username := generateRandomUsername("testuser")
	email := generateRandomEmail("testuser")
	password := "testpassword123"

	// Тест регистрации
	t.Log("Testing registration")
	client.post("/signup", map[string]string{
		"username": username,
		"email":    email,
		"password": password,
	}, http.StatusOK)

	// Тест логина
	t.Log("Testing login")
	client.post("/login", map[string]string{
		"email":    email,
		"password": password,
	}, http.StatusOK)

	// Тест доступа к защищенному ресурсу
	t.Log("Testing protected resource access")
	client.get("/profile", http.StatusOK)

	// Тест логаута
	t.Log("Testing logout")
	client.post("/logout", nil, http.StatusOK)

	// Проверяем, что после логаута доступ запрещен
	t.Log("Testing access after logout")
	resp, err := client.client.Get(serverURL + "/profile")
	if err == nil {
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			t.Error("Expected access denied after logout, but got OK")
		}
	}
}

func TestE2E_PostCRUD(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	serverURL := getServerURL()

	if !isServerAvailable(serverURL) {
		t.Skipf("Server is not available at %s, skipping test", serverURL)
	}

	client := NewTestClient(t, serverURL)

	// Регистрируем и логиним пользователя
	username := generateRandomUsername("postuser")
	email := generateRandomEmail("postuser")

	client.post("/signup", map[string]string{
		"username": username,
		"email":    email,
		"password": "password123",
	}, http.StatusOK)

	client.post("/login", map[string]string{
		"email":    email,
		"password": "password123",
	}, http.StatusOK)

	timestamp := time.Now().UnixNano()

	// Создание поста
	t.Log("Creating post")
	resp := client.post("/post", map[string]string{
		"content": fmt.Sprintf("Test post content %d", timestamp),
	}, http.StatusOK)

	var postResp struct {
		PostID string `json:"post_id"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&postResp)
	resp.Body.Close()

	// Чтение поста
	t.Log("Reading post")
	client.get("/posts/"+postResp.PostID, http.StatusOK)

	// Обновление поста
	t.Log("Updating post")
	client.put("/posts/"+postResp.PostID, map[string]string{
		"content": fmt.Sprintf("Updated post content %d", timestamp),
	}, http.StatusOK)

	// Удаление поста
	t.Log("Deleting post")
	client.delete("/posts/"+postResp.PostID, http.StatusOK)
}

// isServerAvailable проверяет доступность сервера
func isServerAvailable(url string) bool {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url + "/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// TestMain позволяет настроить глобальные настройки для тестов
func TestMain(m *testing.M) {
	// Инициализация глобального генератора случайных чисел
	rand.Seed(time.Now().UnixNano())

	// Проверяем доступность сервера перед запуском тестов
	serverURL := getServerURL()
	fmt.Printf("Testing against server: %s\n", serverURL)

	if !isServerAvailable(serverURL) {
		fmt.Printf("Warning: Server is not available at %s\n", serverURL)
		fmt.Println("Some tests may be skipped")
	}

	// Запускаем тесты
	code := m.Run()
	os.Exit(code)
}
