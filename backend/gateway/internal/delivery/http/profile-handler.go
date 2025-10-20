package http

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/mailru/easyjson"
	"github.com/microcosm-cc/bluemonday"
	"net/http"
	"quickflow/gateway/internal/delivery/http/forms"
	"quickflow/gateway/internal/delivery/http/interfaces"
	interfaces2 "quickflow/gateway/internal/delivery/ws/interfaces"
	errors2 "quickflow/gateway/internal/errors"
	http2 "quickflow/gateway/utils/http"
	"quickflow/shared/logger"
	"quickflow/shared/models"
	"strconv"
)

type ProfileHandler struct {
	profileUC      interfaces.ProfileUseCase
	friendsUseCase interfaces.FriendsUseCase
	authUseCase    interfaces.AuthUseCase
	chatUseCase    interfaces.ChatUseCase
	connService    interfaces2.IWebSocketConnectionManager
	policy         *bluemonday.Policy
}

func NewProfileHandler(profileUC interfaces.ProfileUseCase, friendUseCase interfaces.FriendsUseCase, authUseCase interfaces.AuthUseCase,
	chatUseCase interfaces.ChatUseCase, connService interfaces2.IWebSocketConnectionManager, policy *bluemonday.Policy) *ProfileHandler {
	return &ProfileHandler{
		profileUC:      profileUC,
		connService:    connService,
		friendsUseCase: friendUseCase,
		authUseCase:    authUseCase,
		chatUseCase:    chatUseCase,
		policy:         policy,
	}
}

const MultipartFormMaxSize = 15 << 20

// GetProfile returns user profile
// @Summary Get user profile
// @Description Get user profile by id
// @Tags Profile
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} forms.ProfileForm "User profile"
// @Failure 400 {object} forms.ErrorForm "Failed to parse user id"
// @Failure 404 {object} forms.ErrorForm "Profile not found"
// @Failure 500 {object} forms.ErrorForm "Failed to get profile"
// @Router /api/profile/{username} [get]
func (p *ProfileHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	// user whose profile is requested
	ctx := r.Context()
	userRequested := mux.Vars(r)["username"]
	logger.Info(ctx, "Request profile of %s", userRequested)

	profileInfo, err := p.profileUC.GetProfileByUsername(ctx, userRequested)
	if err != nil {
		err := errors2.FromGRPCError(err)
		logger.Error(ctx, "Unexpected error: %s", err.Error())
		http2.WriteJSONError(w, err)
		return
	}
	logger.Info(ctx, "Profile of %s was successfully fetched", userRequested)

	_, isOnline := p.connService.IsConnected(profileInfo.UserId)

	var relation = models.RelationNone
	var chatId *uuid.UUID
	if session, err := r.Cookie("session"); err == nil {
		// parse session
		sessionUuid, err := uuid.Parse(session.Value)
		if err != nil {
			logger.Error(ctx, "Failed to parse session: %s", err.Error())
			http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Failed to parse session", http.StatusBadRequest))
			return
		}

		// lookup user by session
		user, err := p.authUseCase.LookupUserSession(r.Context(), models.Session{SessionId: sessionUuid})
		if err != nil {
			err := errors2.FromGRPCError(err)
			logger.Error(ctx, "Failed to lookup user by session: %s", err.Error())
			http2.WriteJSONError(w, err)
			return
		}

		rel, err := p.friendsUseCase.GetUserRelation(ctx, user.Id, profileInfo.UserId)
		if err != nil {
			logger.Error(ctx, "Failed to get user relation: %s", err.Error())
			http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to get user relation", http.StatusInternalServerError))
			return
		}
		relation = rel

		// get chat id
		chat, err := p.chatUseCase.GetPrivateChat(ctx, user.Id, profileInfo.UserId)
		appErr := errors2.FromGRPCError(err)
		if err != nil && appErr.HTTPStatus != http.StatusNotFound {
			logger.Error(ctx, "Failed to get chat id: %s", appErr.Error())
			http2.WriteJSONError(w, appErr)
			return
		} else {
			if err == nil {
				chatId = &chat.ID
			}
		}
	}

	out := forms.ModelToForm(profileInfo, userRequested, isOnline, relation, chatId)

	w.Header().Set("Content-Type", "application/json")
	if _, err := easyjson.MarshalToWriter(out, w); err != nil {
		logger.Error(ctx, "Failed to encode profile: %s", err.Error())
		http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to encode profile", http.StatusInternalServerError))
		return
	}
}

// UpdateProfile updates user profile
// @Summary Update user profile
// @Description Update user profile by id
// @Tags Profile
// @Accept json
// @Produce json
// @Param firstname formData string true "First name"
// @Param lastname formData string true "Last name"
// @Param birth_date formData string true "Birth date"
// @Param sex formData int true "Sex"
// @Param bio formData string true "Bio"
// @Param avatar formData file false "Avatar"
// @Success 200 {string} string "Profile updated"
// @Failure 400 {object} forms.ErrorForm "Failed to parse form"
// @Failure 500 {object} forms.ErrorForm "Failed to update profile"
// @Router /api/profile [post]
func (p *ProfileHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := ctx.Value("user").(models.User)
	if !ok {
		logger.Error(ctx, "Failed to get user from context while updating profile")
		http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to get user from context", http.StatusInternalServerError))
		return
	}
	logger.Info(ctx, "User %s requested to update profile", user.Username)

	var profileForm forms.ProfileForm
	err := r.ParseMultipartForm(MultipartFormMaxSize)
	if err != nil {
		logger.Error(ctx, "Failed to parse form: %s", err.Error())
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, fmt.Sprintf("Failed to parse form: %v", err), http.StatusBadRequest))
		return
	}

	oldProfile, err := p.profileUC.GetProfileByUsername(ctx, user.Username)
	if err != nil {
		err := errors2.FromGRPCError(err)
		logger.Error(ctx, "Unexpected error: %s", err.Error())
		http2.WriteJSONError(w, err)
		return
	}
	logger.Info(ctx, "Loading pictures")
	// retrieving files if passed
	profileForm.Avatar, err = http2.GetFile(r, "avatar")
	if err != nil {
		logger.Error(ctx, "Failed to get avatar: %s", err.Error())
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, fmt.Sprintf("Failed to get avatar: %v", err), http.StatusBadRequest))
		return
	}
	if profileForm.Avatar != nil {
		logger.Info(ctx, "Loaded avatar: %v size: %v", profileForm.Avatar.Name, profileForm.Avatar.Size)
	}

	profileForm.Background, err = http2.GetFile(r, "cover")
	if err != nil {
		logger.Error(ctx, "Failed to get cover: %s", err.Error())
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, fmt.Sprintf("Failed to get cover: %v", err), http.StatusBadRequest))
		return
	}
	if profileForm.Background != nil {
		logger.Info(ctx, "Loaded cover: %v size: %v", profileForm.Background.Name, profileForm.Background.Size)
	}

	//var recievedValidInfo = profileForm.Avatar != nil || profileForm.Background != nil
	// parsing main profile info
	var profileInfo forms.ProfileInfo

	profileInfo.Bio = r.FormValue("bio")
	if len(profileInfo.Bio) == 0 {
		profileInfo.Bio = oldProfile.BasicInfo.Bio
	}
	profileInfo.Name = r.FormValue("firstname")
	if len(profileInfo.Name) == 0 {
		profileInfo.Name = oldProfile.BasicInfo.Name
	}
	profileInfo.Surname = r.FormValue("lastname")
	if len(profileInfo.Surname) == 0 {
		profileInfo.Surname = oldProfile.BasicInfo.Surname
	}
	sex, err := strconv.Atoi(r.FormValue("sex"))
	if err == nil && sex != models.MALE && sex != models.FEMALE {
		logger.Error(ctx, "Bad sex value")
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Invalid sex value", http.StatusBadRequest))
		return
	} else if err == nil || sex == models.MALE || sex == models.FEMALE {
		profileInfo.Sex = models.Sex(sex)
	} else {
		profileInfo.Sex = oldProfile.BasicInfo.Sex
	}

	profileInfo.DateOfBirth = r.FormValue("birth_date")
	if len(profileInfo.DateOfBirth) == 0 {
		profileInfo.DateOfBirth = oldProfile.BasicInfo.DateOfBirth.Format("2006-01-02")
	}

	profileForm.ProfileInfo = &profileInfo

	// converting form to model
	profile, err := profileForm.FormToModel()
	if err != nil {
		logger.Error(ctx, "Failed to convert form to model: %s", err.Error())
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, fmt.Sprintf("Failed to parse form: %+v", err), http.StatusBadRequest))
		return
	}

	logger.Info(ctx, "Recieved profile update: %v", profile)
	profile.UserId = user.Id
	_, err = p.profileUC.UpdateProfile(ctx, profile)
	if err != nil {
		err := errors2.FromGRPCError(err)
		logger.Error(ctx, "Failed to update profile: %s", err.Error())
		http2.WriteJSONError(w, err)
		return
	}

	logger.Info(ctx, "Profile of %s was successfully updated", user.Username)
}

func (p *ProfileHandler) GetMyProfile(w http.ResponseWriter, r *http.Request) {
	// user whose profile is requested
	ctx := r.Context()
	user, ok := ctx.Value("user").(models.User)
	if !ok {
		logger.Error(ctx, "User not found in context")
		http.NotFound(w, r)
		return
	}
	logger.Info(ctx, "Request profile of himself %s", user.Username)
	logger.Info(ctx, "User: %v", user)

	profileInfo, err := p.profileUC.GetProfileByUsername(ctx, user.Username)
	if err != nil {
		err := errors2.FromGRPCError(err)
		logger.Error(ctx, "Unexpected error: %s", err.Error())
		http2.WriteJSONError(w, err)
		return
	}
	logger.Info(ctx, "Profile of %s was successfully fetched", user.Username)

	_, isOnline := p.connService.IsConnected(profileInfo.UserId)

	var relation = models.RelationNone
	var chatId *uuid.UUID
	if session, err := r.Cookie("session"); err == nil {
		// parse session
		sessionUuid, err := uuid.Parse(session.Value)
		if err != nil {
			logger.Error(ctx, "Failed to parse session: %s", err.Error())
			http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Failed to parse session", http.StatusBadRequest))
			return
		}

		// lookup user by session
		user, err := p.authUseCase.LookupUserSession(r.Context(), models.Session{SessionId: sessionUuid})
		if err != nil {
			err := errors2.FromGRPCError(err)
			logger.Error(ctx, "Failed to lookup user by session: %s", err.Error())
			http2.WriteJSONError(w, err)
			return
		}

		rel, err := p.friendsUseCase.GetUserRelation(ctx, user.Id, profileInfo.UserId)
		if err != nil {
			logger.Error(ctx, "Failed to get user relation: %s", err.Error())
			http2.WriteJSONError(w, err)
			return
		}
		relation = rel

		// get chat id
		chat, err := p.chatUseCase.GetPrivateChat(ctx, user.Id, profileInfo.UserId)
		appErr := errors2.FromGRPCError(err)
		if err != nil && appErr.HTTPStatus != http.StatusNotFound {
			logger.Error(ctx, "Failed to get chat id: %s", appErr.Error())
			http2.WriteJSONError(w, appErr)
			return
		} else {
			if err == nil {
				chatId = &chat.ID
			}
		}
	}

	out := forms.ModelToForm(profileInfo, user.Username, isOnline, relation, chatId)

	w.Header().Set("Content-Type", "application/json")
	if _, err := easyjson.MarshalToWriter(out, w); err != nil {
		logger.Error(ctx, "Failed to encode profile: %s", err.Error())
		http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to encode profile", http.StatusInternalServerError))
		return
	}
}
