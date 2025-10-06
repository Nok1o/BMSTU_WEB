package http

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"

	"github.com/google/uuid"
	"github.com/mailru/easyjson"

	"quickflow/gateway/internal/delivery/http/forms"
	"quickflow/gateway/internal/delivery/http/interfaces"
	interfaces2 "quickflow/gateway/internal/delivery/ws/interfaces"
	"quickflow/gateway/internal/errors"
	http2 "quickflow/gateway/utils/http"
	"quickflow/gateway/utils/validation"
	"quickflow/shared/logger"
	"quickflow/shared/models"
)

type FriendHandler struct {
	FriendsUseCase  interfaces.FriendsUseCase
	ConnService     interfaces2.IWebSocketConnectionManager
	friendWSManager interfaces.IFriendsWSManager
}

func NewFriendsHandler(friendsUseCase interfaces.FriendsUseCase, connService interfaces2.IWebSocketConnectionManager, friendWSManager interfaces.IFriendsWSManager) *FriendHandler {
	return &FriendHandler{
		FriendsUseCase:  friendsUseCase,
		ConnService:     connService,
		friendWSManager: friendWSManager,
	}
}

// GetFriends возвращает список друзей
// @Summary Получить друзей
// @Description Возвращает список друзей пользователя
// @Tags Friends
// @Produce json
// @Success 200 {array} forms.FriendsInfoOut "Список друзей"
// @Failure 400 {object} forms.ErrorForm "Некорректные данные"
// @Failure 500 {object} forms.ErrorForm "Ошибка сервера"
// @Router /api/friends [get]
func (f *FriendHandler) GetFriends(w http.ResponseWriter, r *http.Request) {
	ctx := http2.SetRequestId(r.Context())

	user, ok := ctx.Value("user").(models.User)
	if !ok {
		logger.Error(ctx, "Failed to get user from context while fetching friends")
		http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to get user from context", http.StatusInternalServerError))
		return
	}

	// get user_id from route
	userID := mux.Vars(r)["user_id"]
	if userID == "" {
		logger.Error(ctx, "Failed to get user from query params while fetching friends")
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Failed to get user from query params", http.StatusBadRequest))
		return
	}

	limit := r.URL.Query().Get("count")
	offset := r.URL.Query().Get("offset")
	reqType := r.URL.Query().Get("request_type")

	targetUserID := userID

	if reqType == "" {
		reqType = "friend"
	}

	if limit == "" {
		limit = "10"
	}

	if offset == "" {
		offset = "0"
	}

	if !validation.ValidateFriendReqType(reqType) {
		logger.Error(ctx, "Invalid request type: %s", reqType)
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Invalid request type", http.StatusBadRequest))
		return
	}

	logger.Info(ctx, "User %s requested friends", targetUserID)

	friendsInfo, friendsCount, err := f.FriendsUseCase.GetFriendsInfo(ctx, targetUserID, limit, offset, reqType)
	if err != nil {
		err = errors2.FromGRPCError(err)
		logger.Error(ctx, "Unable to get list of friends for user %s: %s", user.Username, err.Error())
		http2.WriteJSONError(w, err)
		return
	}

	logger.Info(ctx, "Successfully got friends info for user %s", user.Username)

	var friendsOnline []bool
	for _, friend := range friendsInfo {
		_, isOnline := f.ConnService.IsConnected(friend.Id)
		friendsOnline = append(friendsOnline, isOnline)
	}

	var friendsInfoOut forms.FriendsInfoOut

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(friendsInfoOut.ToJson(friendsInfo, friendsOnline, friendsCount))
	if err != nil {
		logger.Error(ctx, "Unable to encode friends info to json: %s", err)
		http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Unable to encode friends info to json", http.StatusInternalServerError))
		return
	}

	logger.Info(ctx, "Successfully processed friends request for user: %s", user.Username)
}

// SendFriendRequest отправляет заявку в друзья
// @Summary Добавить в друзья
// @Description Отправляет запрос в друзья конкретному пользователю
// @Tags Friends
// @Produce json
// @Success 200 {array} forms.FriendsInfoOut "Список друзей"
// @Failure 400 {object} forms.ErrorForm "Некорректные данные"
// @Failure 409 {object} forms.ErrorForm "Отношение между пользователями (подписчик/друг) уже существует
// @Failure 500 {object} forms.ErrorForm "Ошибка сервера"
// @Router /api/friends [post]
func (f *FriendHandler) SendFriendRequest(w http.ResponseWriter, r *http.Request) {
	ctx := http2.SetRequestId(r.Context())

	user, ok := ctx.Value("user").(models.User)
	if !ok {
		logger.Error(ctx, "Failed to get user from context while sending friend request")
		http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to get user from context", http.StatusInternalServerError))
		return
	}

	logger.Info(ctx, "Trying to parse request body")

	var req forms.FriendRequest
	err := easyjson.UnmarshalFromReader(r.Body, &req)
	if err != nil {
		logger.Error(ctx, "Unable to decode request body: %s", err)
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Unable to decode request body", http.StatusBadRequest))
		return
	}

	logger.Info(ctx, "User %s trying to add friend %s", user.Username, req.ReceiverID)

	recId, err := uuid.Parse(req.ReceiverID)
	if err != nil {
		logger.Error(ctx, "Invalid receiver ID format: %s", req.ReceiverID)
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Invalid receiver ID format", http.StatusBadRequest))
		return
	}

	if err = f.FriendsUseCase.SendFriendRequest(ctx, user.Id.String(), req.ReceiverID); err != nil {
		err := errors2.FromGRPCError(err)
		logger.Error(ctx, "Unable to send friend request: %s", err)
		http2.WriteJSONError(w, err)
		return
	}

	// Notify the WebSocket manager about the friend request
	if err = f.friendWSManager.NotifyFriendRequestSent(ctx, user.Id, recId); err != nil {
		return
	}

	logger.Info(ctx, "Successfully processed friend request to user %s", req.ReceiverID)
	w.WriteHeader(http.StatusOK)
	res := struct {
		RequestId uuid.UUID `json:"request_id"`
	}{
		RequestId: user.Id,
	}
	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		logger.Error(ctx, "Unable to encode response to json: %s", err)
		http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Unable to encode response to json", http.StatusInternalServerError))
		return
	}
}

func (f *FriendHandler) AcceptFriendRequest(w http.ResponseWriter, r *http.Request) {
	ctx := http2.SetRequestId(r.Context())

	user, ok := ctx.Value("user").(models.User)
	if !ok {
		logger.Error(ctx, "Failed to get user from context while accepting friend request")
		http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to get user from context", http.StatusInternalServerError))
		return
	}

	logger.Info(ctx, "Trying to parse request body")

	var req forms.FriendRequest
	err := easyjson.UnmarshalFromReader(r.Body, &req)
	if err != nil {
		logger.Error(ctx, "Unable to decode request body: %s", err)
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Unable to decode request body", http.StatusBadRequest))
		return
	}

	recId, err := uuid.Parse(req.ReceiverID)
	if err != nil {
		logger.Error(ctx, "Invalid receiver ID format: %s", req.ReceiverID)
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Invalid receiver ID format", http.StatusBadRequest))
		return
	}

	logger.Info(ctx, "User %s trying to accept friend %s", user.Username, req.ReceiverID)

	if err = f.FriendsUseCase.AcceptFriendRequest(ctx, user.Id.String(), req.ReceiverID); err != nil {
		err := errors2.FromGRPCError(err)
		logger.Error(ctx, "Unable to accept friend request: %s", err)
		http2.WriteJSONError(w, err)
		return
	}

	// Notify the WebSocket manager about the friend request acceptance
	if err = f.friendWSManager.NotifyFriendRequestAccepted(ctx, user.Id, recId); err != nil {
		logger.Error(ctx, "Failed to notify about friend request acceptance: %s", err)
		http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to notify about friend request acceptance", http.StatusInternalServerError))
		return
	}

	logger.Info(ctx, "Successfully accepted friend request from user %s", req.ReceiverID)
}

func (f *FriendHandler) DeleteFriend(w http.ResponseWriter, r *http.Request) {
	ctx := http2.SetRequestId(r.Context())

	user, ok := ctx.Value("user").(models.User)
	if !ok {
		logger.Error(ctx, "Failed to get user from context while deleting friend")
		http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to get user from context", http.StatusInternalServerError))
		return
	}

	logger.Info(ctx, "Trying to parse request body")

	var req forms.FriendRequestDel

	req.FriendID, ok = mux.Vars(r)["friend_id"]
	if !ok || req.FriendID == "" {
		err := easyjson.UnmarshalFromReader(r.Body, &req)
		if err != nil {
			logger.Error(ctx, "Failed to get friend_id from route params while deleting friend")
			http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Failed to get friend_id from route params", http.StatusBadRequest))
			return
		}
	}

	logger.Info(ctx, "User %s trying to delete friend %s", user.Username, req.FriendID)

	if err := f.FriendsUseCase.DeleteFriend(ctx, user.Id.String(), req.FriendID); err != nil {
		err := errors2.FromGRPCError(err)
		logger.Error(ctx, "Unable to delete friend: %s", err)
		http2.WriteJSONError(w, err)
		return
	}

	logger.Info(ctx, "Successfully deleted friend from user %s", req.FriendID)
	w.WriteHeader(http.StatusNoContent)
}

func (f *FriendHandler) Unfollow(w http.ResponseWriter, r *http.Request) {
	ctx := http2.SetRequestId(r.Context())

	user, ok := ctx.Value("user").(models.User)
	if !ok {
		logger.Error(ctx, "Failed to get user from context while unfollowing user")
		http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to get user from context", http.StatusInternalServerError))
		return
	}

	logger.Info(ctx, "Trying to parse request body")

	var req forms.FriendRequestDel
	req.FriendID, ok = mux.Vars(r)["request_id"]
	if !ok || req.FriendID == "" {
		err := easyjson.UnmarshalFromReader(r.Body, &req)
		if err != nil {
			logger.Error(ctx, "Unable to decode request_id: %s", err)
			http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Unable to decode request body", http.StatusBadRequest))
			return
		}
	}

	logger.Info(ctx, "User %s trying to unfollow user %s", user.Username, req.FriendID)

	if err := f.FriendsUseCase.Unfollow(ctx, user.Id.String(), req.FriendID); err != nil {
		err := errors2.FromGRPCError(err)
		logger.Error(ctx, "Unable to unfollow user: %s", err)
		http2.WriteJSONError(w, err)
		return
	}

	logger.Info(ctx, "Successfully unfollowed user %s", req.FriendID)
	w.WriteHeader(http.StatusNoContent)
}

func (f *FriendHandler) ChangeRequestStatus(w http.ResponseWriter, r *http.Request) {
	ctx := http2.SetRequestId(r.Context())

	user, ok := ctx.Value("user").(models.User)
	if !ok {
		logger.Error(ctx, "Failed to get user from context while unfollowing user")
		http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to get user from context", http.StatusInternalServerError))
		return
	}

	logger.Info(ctx, "Trying to parse request body")

	var req forms.FriendRequestDel
	req.FriendID, ok = mux.Vars(r)["request_id"]
	if !ok || req.FriendID == "" {
		err := easyjson.UnmarshalFromReader(r.Body, &req)
		if err != nil {
			logger.Error(ctx, "Unable to decode request_id: %s", err)
			http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Unable to decode request body", http.StatusBadRequest))
			return
		}
	}

	logger.Info(ctx, "User %s trying to unfollow user %s", user.Username, req.FriendID)

	type StatusStruct struct {
		Status string `json:"status"`
	}
	statusStruct := new(StatusStruct)
	err := json.NewDecoder(r.Body).Decode(statusStruct)
	if err != nil {
		logger.Error(ctx, "Unable to decode status: %s", err)
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Unable to decode status", http.StatusBadRequest))
		return
	}

	if statusStruct.Status == "accepted" {
		if err = f.FriendsUseCase.AcceptFriendRequest(ctx, user.Id.String(), req.FriendID); err != nil {
			err := errors2.FromGRPCError(err)
			logger.Error(ctx, "Unable to accept friend request: %s", err)
			http2.WriteJSONError(w, err)
			return
		}
	} else if statusStruct.Status == "rejected" {
		if err := f.FriendsUseCase.Unfollow(ctx, req.FriendID, user.Id.String()); err != nil {
			err := errors2.FromGRPCError(err)
			logger.Error(ctx, "Unable to unfollow user: %s", err)
			http2.WriteJSONError(w, err)
			return
		}
		logger.Info(ctx, "Successfully unfollowed user %s", req.FriendID)
	}

}

func (f *FriendHandler) MarkRead(w http.ResponseWriter, r *http.Request) {
	ctx := http2.SetRequestId(r.Context())

	user, ok := ctx.Value("user").(models.User)
	if !ok {
		logger.Error(ctx, "Failed to get user from context while unfollowing user")
		http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to get user from context", http.StatusInternalServerError))
		return
	}

	logger.Info(ctx, "Trying to parse request body")

	var req forms.FriendRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		logger.Error(ctx, "Unable to decode request body: %s", err)
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Unable to decode request body", http.StatusBadRequest))
		return
	}

	logger.Info(ctx, "User %s trying to mark read friend request from user %s", user.Username, req.ReceiverID)

	if err = f.FriendsUseCase.MarkRead(ctx, user.Id.String(), req.ReceiverID); err != nil {
		err = errors2.FromGRPCError(err)
		logger.Error(ctx, "Unable to mark read friend request: %s", err)
		http2.WriteJSONError(w, err)
		return
	}

	logger.Info(ctx, "Successfully marked read friend request from user %s", req.ReceiverID)
}
