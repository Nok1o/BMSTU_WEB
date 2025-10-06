package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/microcosm-cc/bluemonday"

	time2 "quickflow/config/time"
	"quickflow/gateway/internal/delivery/http/forms"
	"quickflow/gateway/internal/delivery/http/interfaces"
	errors2 "quickflow/gateway/internal/errors"
	"quickflow/gateway/pkg/sanitizer"
	http2 "quickflow/gateway/utils/http"
	"quickflow/gateway/utils/validation"
	"quickflow/shared/logger"
	"quickflow/shared/models"
)

type AuthHandler struct {
	authUseCase interfaces.AuthUseCase
	policy      *bluemonday.Policy
}

func NewAuthHandler(authUseCase interfaces.AuthUseCase, policy *bluemonday.Policy) *AuthHandler {
	return &AuthHandler{
		authUseCase: authUseCase,
		policy:      policy,
	}
}

// Greet проверяет доступность API
// @Summary Ping
// @Description Проверяет доступность API, всегда возвращает "Hello, world!"
// @Tags Misc
// @Produce json
// @Success 200 {string} string "Hello, world!"
// @Router /api/hello [get]
func (a *AuthHandler) Greet(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode("Hello World")
	if err != nil {
		http.Error(w, "failed to parse Hello world", http.StatusBadRequest)
		return
	}
}

// SignUp создает нового пользователя
// @Summary Регистрация пользователя
// @Description Создает новую учетную запись пользователя
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body forms.SignUpForm true "Данные для регистрации"
// @Success 200 {object} forms.SignUpResponse "Успешная регистрация"
// @Failure 400 {object} forms.ErrorForm "Некорректные данные"
// @Failure 409 {object} forms.ErrorForm "Пользователь с таким логином уже существует"
// @Router /api/signup [post]
func (a *AuthHandler) SignUp(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	logger.Info(ctx, "Got signup request")

	var form forms.SignUpForm
	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Error(ctx, "Error reading request body: %s", err.Error())
		http2.WriteJSONError(w, errors2.New("BAD_REQUEST", fmt.Sprintf("Unable to read request body: %v", err), http.StatusBadRequest))
		return
	}
	defer r.Body.Close()

	if err = form.UnmarshalJSON(body); err != nil {
		logger.Error(ctx, "Decode error: %s", err.Error())
		http2.WriteJSONError(w, errors2.New("BAD_REQUEST", fmt.Sprintf("Unable to decode request body: %v", err), http.StatusBadRequest))
		return
	}

	sanitizer.SanitizeSignUpData(&form, a.policy)

	// converting transport form to domain model
	user := models.User{
		Username: form.Login,
		Password: form.Password,
	}

	date, err := time.Parse(time2.DateLayout, form.DateOfBirth)
	if err != nil {
		logger.Error(ctx, "Decode error: %s", err.Error())
		http2.WriteJSONError(w, errors2.New("BAD_REQUEST", fmt.Sprintf("Unable to parse date: %v", err), http.StatusBadRequest))
		return
	}

	profile := models.Profile{
		BasicInfo: &models.BasicInfo{
			AvatarUrl:   "",
			Name:        form.Name,
			Surname:     form.Surname,
			Sex:         form.Sex,
			DateOfBirth: date,
		},
	}

	// validation
	if err := validation.ValidateUser(user.Username, user.Password); err != nil {
		http2.WriteJSONError(w, errors2.New("BAD_REQUEST", fmt.Sprintf("Invalid username or password: %v", err), http.StatusBadRequest))
		logger.Error(ctx, "Validation error: %s", err.Error())
		return
	}
	if err = validation.ValidateProfile(profile.BasicInfo.Name, profile.BasicInfo.Surname); err != nil {
		http2.WriteJSONError(w, errors2.New("BAD_REQUEST", fmt.Sprintf("Invalid profile data: %v", err), http.StatusBadRequest))
		logger.Error(ctx, "Validation error: %s", err.Error())
		return
	}

	// process data
	userId, session, err := a.authUseCase.CreateUser(r.Context(), user, profile)
	if err != nil {
		logger.Error(ctx, "Create user error: %s", err.Error())
		http2.WriteJSONError(w, err)
		return
	}

	logger.Info(ctx, "Successfully created new user")

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    session.SessionId.String(),
		Expires:  session.ExpireDate,
		HttpOnly: true,
		Secure:   true,
	})

	w.WriteHeader(http.StatusCreated)
	response := struct {
		Id        string `json:"id"`
		SessionId string `json:"session_id"`
	}{
		Id:        userId.String(),
		SessionId: session.SessionId.String(),
	}
	if err = json.NewEncoder(w).Encode(response); err != nil {
		logger.Error(ctx, "Encode response error: %s", err.Error())
		http2.WriteJSONError(w, errors2.New("INTERNAL_SERVER_ERROR", fmt.Sprintf("Unable to encode response: %v", err), http.StatusInternalServerError))
		return
	}

	logger.Info(ctx, "Successfully processed signup request")
}

// Login аутентифицирует пользователя
// @Summary Авторизация
// @Description Аутентифицирует пользователя и устанавливает сессию
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body forms.AuthForm true "Данные для авторизации"
// @Success 200 {object} forms.AuthResponse "Успешная авторизация"
// @Failure 400 {object} forms.ErrorForm "Некорректные данные"
// @Failure 401 {object} forms.ErrorForm "Пользователь не авторизован"
// @Router /api/login [post]
func (a *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	logger.Info(ctx, "Got login request")

	var form forms.AuthForm

	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Error(ctx, "Error reading request body: %s", err.Error())
		http2.WriteJSONError(w, errors2.New("BAD_REQUEST", fmt.Sprintf("Unable to read request body: %v", err), http.StatusBadRequest))
		return
	}
	defer r.Body.Close()

	if err := form.UnmarshalJSON(body); err != nil {
		logger.Error(ctx, "Decode error: %s", err.Error())
		http2.WriteJSONError(w, errors2.New("BAD_REQUEST", fmt.Sprintf("Unable to decode request body: %v", err), http.StatusBadRequest))
		return
	}

	sanitizer.SanitizeLoginData(&form, a.policy)

	// converting transport form to domain model
	loginData := models.LoginData{
		Username: form.Login,
		Password: form.Password,
	}

	// process data
	session, err := a.authUseCase.AuthUser(r.Context(), loginData)
	if err != nil {
		logger.Error(ctx, "Get User error: %s", err.Error())
		http2.WriteJSONError(w, err)
		return
	}

	logger.Info(ctx, "Successfully got User with SessionID: %s", session.SessionId.String())

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    session.SessionId.String(),
		Expires:  session.ExpireDate,
		HttpOnly: true,
		Secure:   true,
	})

	w.WriteHeader(http.StatusCreated)
	response := struct {
		Id        string `json:"id"`
		ExpiresAt string `json:"expires_at"`
	}{
		Id:        session.SessionId.String(),
		ExpiresAt: session.ExpireDate.Format(time2.TimeStampLayout),
	}
	if err = json.NewEncoder(w).Encode(response); err != nil {
		logger.Error(ctx, "Encode response error: %s", err.Error())
		http2.WriteJSONError(w, errors2.New("INTERNAL_SERVER_ERROR", fmt.Sprintf("Unable to encode response: %v", err), http.StatusInternalServerError))
		return
	}

	logger.Info(ctx, "Successfully processed login request")
}

// Logout завершает сессию пользователя
// @Summary Выход из системы
// @Description Завершает сессию пользователя
// @Tags Auth
// @Accept json
// @Produce json
// @Success 200 {string} string "Успешный выход"
// @Failure 401 {object} forms.ErrorForm "Пользователь не авторизован"
// @Router /api/logout [post]
func (a *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	logger.Info(ctx, "Got logout request")

	cookie, err := r.Cookie("session")
	if err != nil {
		logger.Error(ctx, "Get cookie error: %s", err.Error())
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	sessionIdQuery := r.URL.Query().Get("session_id")
	if sessionIdQuery != "" && cookie.Value != sessionIdQuery {
		logger.Error(ctx, "Session ID in query does not match cookie")
		http.Error(w, "Session not owned by the user", http.StatusForbidden)
		return
	}

	cookieUUID, err := uuid.Parse(cookie.Value)
	if err != nil {
		logger.Error(ctx, "Parse cookie error: %s", err.Error())
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if _, err = a.authUseCase.LookupUserSession(r.Context(), models.Session{SessionId: cookieUUID}); err != nil {
		logger.Error(ctx, "Couldn't find user session: %s", err.Error())
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err = a.authUseCase.DeleteUserSession(r.Context(), cookie.Value); err != nil {
		logger.Error(ctx, "Delete session error: %s", err.Error())
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	cookie.Expires = time.Now().AddDate(0, 0, -1)
	http.SetCookie(w, cookie)
	w.WriteHeader(http.StatusNoContent)

	logger.Info(ctx, "Successfully processed logout request")
}
