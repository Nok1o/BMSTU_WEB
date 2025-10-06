package http

import (
	"net/http"
	"time"

	"github.com/mailru/easyjson"

	"quickflow/gateway/internal/delivery/http/interfaces"
	"quickflow/shared/logger"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	time2 "quickflow/config/time"
	"quickflow/gateway/internal/delivery/http/forms"
	errors2 "quickflow/gateway/internal/errors"
	http2 "quickflow/gateway/utils/http"
	"quickflow/shared/models"
)

type FeedHandler struct {
	authUseCase      interfaces.AuthUseCase
	postService      interfaces.PostService
	profileUseCase   interfaces.ProfileUseCase
	friendUseCase    interfaces.FriendsUseCase
	commentUseCase   interfaces.CommentService
	communityService interfaces.CommunityService
}

// NewFeedHandler creates new feed handler.
func NewFeedHandler(authUseCase interfaces.AuthUseCase, postUseCase interfaces.PostService,
	profileUseCase interfaces.ProfileUseCase, friendUseCase interfaces.FriendsUseCase,
	communityService interfaces.CommunityService, commentService interfaces.CommentService) *FeedHandler {
	return &FeedHandler{
		postService:      postUseCase,
		profileUseCase:   profileUseCase,
		friendUseCase:    friendUseCase,
		authUseCase:      authUseCase,
		communityService: communityService,
		commentUseCase:   commentService,
	}
}

// GetFeed возвращает ленту постов
// @Summary Получить ленту
// @Description Возвращает список постов, опубликованных до указанного времени
// @Tags Feed
// @Produce json
// @Param posts_count query int true "Количество постов"
// @Param ts query string false "Временная метка"
// @Success 200 {array} forms.PostOut "Список постов"
// @Failure 400 {object} forms.ErrorForm "Некорректные данные"
// @Failure 500 {object} forms.ErrorForm "Ошибка сервера"
// @Router /api/feed [get]
func (f *FeedHandler) GetFeed(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := ctx.Value("user").(models.User)
	if !ok {
		logger.Error(ctx, "Failed to get user from context while fetching feed")
		http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to get user from context", http.StatusInternalServerError))
		return
	}

	feedType := r.URL.Query().Get("type")
	if feedType == "recommendations" {
		f.GetRecommendations(w, r)
		return
	}

	var feedForm forms.FeedForm
	err := feedForm.GetParams(r.URL.Query())
	if err != nil {
		logger.Error(ctx, "Failed to parse query params")
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Failed to parse query params", http.StatusBadRequest))
		return
	}

	ts, err := time.Parse(time2.TimeStampLayout, feedForm.Ts)
	if err != nil {
		ts = time.Now()
	}

	posts, err := f.postService.FetchFeed(ctx, feedForm.Posts, ts, user.Id)
	appErr := errors2.FromGRPCError(err)
	if appErr != nil && appErr.HTTPStatus != http.StatusNotFound {
		logger.Error(ctx, "Failed to load feed")
		http2.WriteJSONError(w, appErr)
		return
	}

	var postsOut forms.PostsOut
	var authors, communities []uuid.UUID
	for _, post := range posts {
		var postOut forms.PostOut
		postOut.FromPost(post)

		// getLastComment
		lastComment, err := f.commentUseCase.GetLastPostComment(ctx, post.Id)
		appErr := errors2.FromGRPCError(err)
		if appErr != nil && appErr.HTTPStatus != http.StatusNotFound {
			logger.Error(ctx, "Failed to get last comment")
			http2.WriteJSONError(w, appErr)
			return
		}

		if lastComment != nil && (appErr == nil || appErr.HTTPStatus != http.StatusNotFound) {
			commentAuthor, err := f.profileUseCase.GetPublicUserInfo(ctx, lastComment.UserId)
			if err != nil {
				logger.Error(ctx, "Failed to get comment author info")
				http2.WriteJSONError(w, err)
				return
			}
			var commentOut forms.CommentOut
			commentOut.FromComment(*lastComment, commentAuthor)

			postOut.LastComment = &commentOut
		}

		postsOut = append(postsOut, postOut)
		if post.CreatorType == models.PostUser {
			authors = append(authors, post.CreatorId)
		} else {
			communities = append(communities, post.CreatorId)
		}
	}

	publicAuthorsInfo, err := f.profileUseCase.GetPublicUsersInfo(ctx, authors)
	if err != nil {
		logger.Error(ctx, "Failed to load authors info")
		http2.WriteJSONError(w, errors2.FromGRPCError(err))
		return
	}

	infosMap := make(map[uuid.UUID]models.PublicUserInfo)
	for _, info := range publicAuthorsInfo {
		infosMap[info.Id] = info
	}

	infosCommunityMap := make(map[uuid.UUID]*models.Community)
	for _, communityId := range communities {
		infosCommunityMap[communityId], err = f.communityService.GetCommunityById(ctx, communityId)
		if err != nil {
			logger.Error(ctx, "Failed to load community info")
			http2.WriteJSONError(w, errors2.FromGRPCError(err))
			return
		}

		infosMap[infosCommunityMap[communityId].OwnerID], err = f.profileUseCase.GetPublicUserInfo(ctx, infosCommunityMap[communityId].OwnerID)
		if err != nil {
			logger.Error(ctx, "Failed to load user info")
			http2.WriteJSONError(w, errors2.FromGRPCError(err))
			return
		}
	}

	var numUser, numComm int
	for i := range postsOut {
		if postsOut[i].CreatorType == "user" {
			rel, err := f.friendUseCase.GetUserRelation(ctx, user.Id, authors[i-numComm])
			if err != nil {
				logger.Error(ctx, "Failed to get user relation")
				http2.WriteJSONError(w, errors2.FromGRPCError(err))
				return
			}
			info := forms.PublicUserInfoToOut(infosMap[authors[i-numComm]], rel)
			postsOut[i].Creator = &info
			numUser++
		} else {
			postsOut[i].Creator = forms.ToCommunityForm(*infosCommunityMap[communities[i-numUser]], infosMap[infosCommunityMap[communities[i-numUser]].OwnerID])
			numComm++
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err = easyjson.MarshalToWriter(postsOut, w); err != nil {
		logger.Error(ctx, "Failed to encode feed")
		http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to encode feed", http.StatusInternalServerError))
	}
}

// GetRecommendations возвращает рекомендации
// @Summary Получить рекомендации
// @Description Возвращает список постов, опубликованных до указанного времени
// @Tags Feed
// @Produce json
// @Param posts_count query int true "Количество постов"
// @Param ts query string false "Временная метка"
// @Success 200 {array} forms.PostOut "Список постов"
// @Failure 400 {object} forms.ErrorForm "Некорректные данные"
// @Failure 500 {object} forms.ErrorForm "Ошибка сервера"
// @Router /api/recommendations [get]
// @Security Session
func (f *FeedHandler) GetRecommendations(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := ctx.Value("user").(models.User)
	if !ok {
		logger.Error(ctx, "Failed to get user from context while getting recommendations")
		http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to get user from context", http.StatusInternalServerError))
		return
	}

	logger.Info(ctx, "Loading recommendations")

	var feedForm forms.FeedForm
	err := feedForm.GetParams(r.URL.Query())
	if err != nil {
		logger.Error(ctx, "Failed to parse query params%v", err)
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Invalid query parameters", http.StatusBadRequest))
		return
	}

	ts, err := time.Parse(time2.TimeStampLayout, feedForm.Ts)
	if err != nil {
		ts = time.Now()
	}

	posts, err := f.postService.FetchRecommendations(ctx, feedForm.Posts, ts, user.Id)
	appErr := errors2.FromGRPCError(err)
	if appErr != nil && appErr.HTTPStatus != http.StatusNotFound {
		logger.Error(ctx, "Failed to fetch recommendations%v", err)
		http2.WriteJSONError(w, appErr)
		return
	}

	var postsOut forms.PostsOut
	var authors, communities []uuid.UUID
	for _, post := range posts {
		var postOut forms.PostOut
		postOut.FromPost(post)

		// getLastComment
		lastComment, err := f.commentUseCase.GetLastPostComment(ctx, post.Id)
		appErr := errors2.FromGRPCError(err)
		if appErr != nil && appErr.HTTPStatus != http.StatusNotFound {
			logger.Error(ctx, "Failed to get last comment")
			http2.WriteJSONError(w, appErr)
			return
		}

		if lastComment != nil {
			commentAuthor, err := f.profileUseCase.GetPublicUserInfo(ctx, lastComment.UserId)
			if err != nil {
				logger.Error(ctx, "Failed to get comment author info")
				http2.WriteJSONError(w, err)
				return
			}
			var commentOut forms.CommentOut
			commentOut.FromComment(*lastComment, commentAuthor)

			postOut.LastComment = &commentOut
		}

		postsOut = append(postsOut, postOut)
		if post.CreatorType == models.PostUser {
			authors = append(authors, post.CreatorId)
		} else {
			communities = append(communities, post.CreatorId)
		}
	}

	publicAuthorsInfo, err := f.profileUseCase.GetPublicUsersInfo(ctx, authors)
	if err != nil {
		err := errors2.FromGRPCError(err)
		logger.Error(ctx, "Failed to get public author info%v", err)
		http2.WriteJSONError(w, err)
		return
	}

	infosMap := make(map[uuid.UUID]models.PublicUserInfo)
	for _, info := range publicAuthorsInfo {
		infosMap[info.Id] = info
	}

	infosCommunityMap := make(map[uuid.UUID]*models.Community)
	for _, communityId := range communities {
		infosCommunityMap[communityId], err = f.communityService.GetCommunityById(ctx, communityId)
		if err != nil {
			err := errors2.FromGRPCError(err)
			logger.Error(ctx, "Failed to load community info%v", err)
			http2.WriteJSONError(w, err)
			return
		}

		infosMap[infosCommunityMap[communityId].OwnerID], err = f.profileUseCase.GetPublicUserInfo(ctx, infosCommunityMap[communityId].OwnerID)
		if err != nil {
			err := errors2.FromGRPCError(err)
			logger.Error(ctx, "Failed to load user info for community owner%v", err)
			http2.WriteJSONError(w, err)
			return
		}
	}

	var numUser, numComm int
	for i := range postsOut {
		if postsOut[i].CreatorType == "user" {
			rel, err := f.friendUseCase.GetUserRelation(ctx, user.Id, authors[i-numComm])
			if err != nil {
				err := errors2.FromGRPCError(err)
				logger.Error(ctx, "Failed to get user relation%v", err)
				http2.WriteJSONError(w, err)
				return
			}
			info := forms.PublicUserInfoToOut(infosMap[authors[i-numComm]], rel)
			postsOut[i].Creator = &info
			numUser++
		} else {
			postsOut[i].Creator = forms.ToCommunityForm(*infosCommunityMap[communities[i-numUser]], infosMap[infosCommunityMap[communities[i-numUser]].OwnerID])
			numComm++
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err = easyjson.MarshalToWriter(postsOut, w); err != nil {
		logger.Error(ctx, "Failed to encode feed")
		http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to encode feed", http.StatusInternalServerError))
	}
}

// FetchUserPosts возвращает посты пользователя
// @Summary Получить посты пользователя
// @Description Возвращает список постов пользователя, опубликованных до указанного времени
// @Tags Feed
// @Produce json
// @Param posts_count query int true "Количество постов"
// @Param ts query string false "Временная метка"
// @Success 200 {array} forms.PostOut "Список постов"
// @Failure 400 {object} forms.ErrorForm "Некорректные данные"
// @Failure 500 {object} forms.ErrorForm "Ошибка сервера"
// @Router /api/profiles/{username}/posts [get]
// @Security Session
func (f *FeedHandler) FetchUserPosts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requester, ok := ctx.Value("user").(models.User)
	if !ok {
		logger.Error(ctx, "Failed to get user from context while fetching user posts")
		http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to get user from context", http.StatusInternalServerError))
		return
	}

	username := mux.Vars(r)["username"]
	if username == "" {
		logger.Error(ctx, "Username is missing in URL while fetching user posts")
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Username is required", http.StatusBadRequest))
		return
	}

	user, err := f.authUseCase.GetUserByUsername(ctx, username)
	if err != nil {
		appErr := errors2.FromGRPCError(err)
		logger.Error(ctx, "Failed to fetch user by username%v", err)
		http2.WriteJSONError(w, appErr)
		return
	}

	var feedForm forms.FeedForm
	err = feedForm.GetParams(r.URL.Query())
	if err != nil {
		logger.Error(ctx, "Failed to parse query params while fetching user posts%v", err)
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Invalid query parameters", http.StatusBadRequest))
		return
	}

	ts, err := time.Parse(time2.TimeStampLayout, feedForm.Ts)
	if err != nil {
		ts = time.Now()
	}

	posts, err := f.postService.FetchCreatorPosts(ctx, user.Id, requester.Id, feedForm.Posts, ts)
	appErr := errors2.FromGRPCError(err)
	if appErr != nil && appErr.HTTPStatus != http.StatusNotFound {
		logger.Error(ctx, "Failed to fetch user posts%v", err)
		http2.WriteJSONError(w, appErr)
		return
	}

	publicUserInfo, err := f.profileUseCase.GetPublicUserInfo(ctx, user.Id)
	if err != nil {
		appErr := errors2.FromGRPCError(err)
		logger.Error(ctx, "Failed to load public user info%v", err)
		http2.WriteJSONError(w, appErr)
		return
	}

	var postsOut forms.PostsOut
	for _, post := range posts {
		var postOut forms.PostOut
		postOut.FromPost(post)

		// getLastComment
		lastComment, err := f.commentUseCase.GetLastPostComment(ctx, post.Id)
		appErr := errors2.FromGRPCError(err)
		if appErr != nil && appErr.HTTPStatus != http.StatusNotFound {
			logger.Error(ctx, "Failed to get last comment")
			http2.WriteJSONError(w, appErr)
			return
		}

		if lastComment != nil {
			commentAuthor, err := f.profileUseCase.GetPublicUserInfo(ctx, lastComment.UserId)
			if err != nil {
				logger.Error(ctx, "Failed to get comment author info")
				http2.WriteJSONError(w, err)
				return
			}
			var commentOut forms.CommentOut
			commentOut.FromComment(*lastComment, commentAuthor)

			postOut.LastComment = &commentOut
		}

		info := forms.PublicUserInfoToOut(publicUserInfo, models.RelationSelf)
		postOut.Creator = &info
		postsOut = append(postsOut, postOut)
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err = easyjson.MarshalToWriter(postsOut, w); err != nil {
		logger.Error(ctx, "Failed to encode feed")
		http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to encode feed", http.StatusInternalServerError))
	}
}

func (f *FeedHandler) FetchCommunityPosts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requester, ok := ctx.Value("user").(models.User)
	if !ok {
		logger.Error(ctx, "Failed to get user from context while fetching community posts")
		http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to get user from context", http.StatusInternalServerError))
		return
	}

	commName := mux.Vars(r)["name"]
	if commName == "" {
		logger.Error(ctx, "Community name is missing in URL")
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Community name is required", http.StatusBadRequest))
		return
	}

	community, err := f.communityService.GetCommunityByName(ctx, commName)
	if err != nil {
		appErr := errors2.FromGRPCError(err)
		logger.Error(ctx, "Failed to fetch community by name%v", err)
		http2.WriteJSONError(w, appErr)
		return
	}

	var feedForm forms.FeedForm
	err = feedForm.GetParams(r.URL.Query())
	if err != nil {
		logger.Error(ctx, "Failed to parse query params for community posts%v", err)
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Invalid query parameters", http.StatusBadRequest))
		return
	}

	ts, err := time.Parse(time2.TimeStampLayout, feedForm.Ts)
	if err != nil {
		ts = time.Now()
	}

	posts, err := f.postService.FetchCreatorPosts(ctx, community.ID, requester.Id, feedForm.Posts, ts)
	appErr := errors2.FromGRPCError(err)
	if appErr != nil && appErr.HTTPStatus != http.StatusNotFound {
		logger.Error(ctx, "Failed to fetch community posts%v", err)
		http2.WriteJSONError(w, appErr)
		return
	}

	ownerInfo, err := f.profileUseCase.GetPublicUserInfo(ctx, community.OwnerID)
	if err != nil {
		appErr := errors2.FromGRPCError(err)
		logger.Error(ctx, "Failed to load community owner info%v", err)
		http2.WriteJSONError(w, appErr)
		return
	}

	var postsOut forms.PostsOut
	var commentOut forms.CommentOut
	for _, post := range posts {
		var postOut forms.PostOut
		postOut.FromPost(post)

		// getLastComment
		lastComment, err := f.commentUseCase.GetLastPostComment(ctx, post.Id)
		appErr := errors2.FromGRPCError(err)
		if appErr != nil && appErr.HTTPStatus != http.StatusNotFound {
			logger.Error(ctx, "Failed to get last comment")
			http2.WriteJSONError(w, appErr)
			return
		}

		if lastComment != nil {
			commentAuthor, err := f.profileUseCase.GetPublicUserInfo(ctx, lastComment.UserId)
			if err != nil {
				logger.Error(ctx, "Failed to get comment author info")
				http2.WriteJSONError(w, err)
				return
			}
			commentOut.FromComment(*lastComment, commentAuthor)

			postOut.LastComment = &commentOut
		}

		postOut.Creator = forms.ToCommunityForm(*community, ownerInfo)
		postsOut = append(postsOut, postOut)
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err = easyjson.MarshalToWriter(postsOut, w); err != nil {
		logger.Error(ctx, "Failed to encode feed")
		http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to encode feed", http.StatusInternalServerError))
	}
}
