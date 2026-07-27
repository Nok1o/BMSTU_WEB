package http

import (
	"net/http"
	"strconv"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/mailru/easyjson"
	"github.com/microcosm-cc/bluemonday"

	"quickflow/gateway/internal/delivery/http/forms"
	"quickflow/gateway/internal/delivery/http/interfaces"
	errors2 "quickflow/gateway/internal/errors"
	"quickflow/gateway/pkg/sanitizer"
	http2 "quickflow/gateway/utils/http"
	"quickflow/shared/logger"
	"quickflow/shared/models"
)

type PostHandler struct {
	postUseCase      interfaces.PostService
	profileUseCase   interfaces.ProfileUseCase
	communityService interfaces.CommunityService
	friendsUseCase   interfaces.FriendsUseCase
	commentUseCase   interfaces.CommentService
	likeWSHandler    interfaces.WSLikeHandler
	policy           *bluemonday.Policy
}

// NewPostHandler creates new post handler.
func NewPostHandler(postUseCase interfaces.PostService, profileUseCase interfaces.ProfileUseCase,
	communityService interfaces.CommunityService, friendsUseCase interfaces.FriendsUseCase,
	commentUseCase interfaces.CommentService, likeHandler interfaces.WSLikeHandler, policy *bluemonday.Policy) *PostHandler {
	return &PostHandler{
		postUseCase:      postUseCase,
		profileUseCase:   profileUseCase,
		communityService: communityService,
		friendsUseCase:   friendsUseCase,
		commentUseCase:   commentUseCase,
		likeWSHandler:    likeHandler,
		policy:           policy,
	}
}

// AddPost добавляет новый пост
// @Summary Добавить пост
// @Description Добавляет новый пост в ленту
// @Tags Feed
// @Accept multipart/form-data
// @Produce json
// @Param text formData string true "Текст поста"
// @Param pics formData file false "Изображения"
// @Success 200 {string} string "OK"
// @Failure 400 {object} forms.ErrorForm "Некорректные данные"
// @Failure 500 {object} forms.ErrorForm "Ошибка сервера"
// @Router /api/post [post]
func (p *PostHandler) AddPost(w http.ResponseWriter, r *http.Request) {
	// extracting user from context
	ctx := r.Context()
	user, ok := ctx.Value("user").(models.User)
	if !ok {
		logger.Error(ctx, "Failed to get user from context while adding post")
		http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to get user from context", http.StatusInternalServerError))
		return
	}

	logger.Info(ctx, "User %s requested to add post", user.Username)

	// parse the post form
	var postForm forms.PostForm
	if err := easyjson.UnmarshalFromReader(r.Body, &postForm); err != nil {
		logger.Error(ctx, "Failed to decode request body for feedback %v", err)
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Bad request body", http.StatusBadRequest))
		return
	}

	// Validate text length
	if utf8.RuneCountInString(postForm.Text) > 4096 {
		logger.Error(ctx, "Text length validation failed: length=%d", utf8.RuneCountInString(postForm.Text))
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Text must be between 1 and 4096 characters", http.StatusBadRequest))
		return
	}

	// Sanitize post content
	sanitizer.SanitizePost(&postForm, p.policy)

	if len(postForm.Text)+len(postForm.Media)+len(postForm.Audio)+len(postForm.File) == 0 {
		logger.Error(ctx, "empty post content")
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "empty post content", http.StatusBadRequest))
		return
	}

	// Convert the post form to a model
	post, err := postForm.ToPostModel(user.Id)
	if err != nil {
		logger.Error(ctx, "Failed to parse post form: %s", err.Error())
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Failed to parse post form", http.StatusBadRequest))
		return
	}

	newPost, err := p.postUseCase.AddPost(ctx, post)
	if err != nil {
		logger.Error(ctx, "Failed to add post: %s", err.Error())
		http2.WriteJSONError(w, err)
		return
	}
	logger.Info(ctx, "Successfully added post")

	// Prepare the post output
	var postOut forms.PostOut
	postOut.FromPost(*newPost)

	// Add creator info to the post
	if newPost.CreatorType == models.PostUser {
		publicUserInfo, err := p.profileUseCase.GetPublicUserInfo(ctx, user.Id)
		if err != nil {
			logger.Error(ctx, "Failed to get public user info: %s", err.Error())
			http2.WriteJSONError(w, err)
			return
		}
		info := forms.PublicUserInfoToOut(publicUserInfo, models.RelationSelf)
		postOut.Creator = &info
	} else {
		community, err := p.communityService.GetCommunityById(ctx, post.CreatorId)
		if err != nil {
			logger.Error(ctx, "Failed to get community: %s", err.Error())
			http2.WriteJSONError(w, err)
			return
		}

		info, err := p.profileUseCase.GetPublicUserInfo(ctx, community.OwnerID)
		if err != nil {
			logger.Error(ctx, "Failed to get user info: %s", err.Error())
			http2.WriteJSONError(w, err)
			return
		}
		postOut.Creator = forms.ToCommunityForm(*community, info)
	}

	// Return the response with the newly created post
	w.Header().Set("Content-Type", "application/json")
	out := forms.PayloadWrapper[forms.PostOut]{Payload: postOut}
	js, err := out.MarshalJSON()
	if err != nil {
		logger.Error(ctx, "Failed to marshal json payload %v", err)
		http2.WriteJSONError(w, err)
		return
	}
	if _, err = w.Write(js); err != nil {
		logger.Error(ctx, "Failed to encode feedback output %v", err)
		http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to encode feedback output", http.StatusInternalServerError))
	}
}

// AddPostCommunity добавляет новый пост в коммьюнити
// @Summary Добавить пост
// @Description Добавляет новый пост в ленту
// @Tags Feed
// @Accept multipart/form-data
// @Produce json
// @Param text formData string true "Текст поста"
// @Param pics formData file false "Изображения"
// @Success 200 {string} string "OK"
// @Failure 400 {object} forms.ErrorForm "Некорректные данные"
// @Failure 500 {object} forms.ErrorForm "Ошибка сервера"
// @Router /api/post [post]
func (p *PostHandler) AddPostCommunity(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, ok := ctx.Value("user").(models.User)
	if !ok {
		logger.Error(ctx, "Failed to get user from context while adding community post")
		http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to get user from context", http.StatusInternalServerError))
		return
	}

	var postForm forms.PostForm
	if err := easyjson.UnmarshalFromReader(r.Body, &postForm); err != nil {
		logger.Error(ctx, "Failed to decode request body for community post %v", err)
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Bad request body", http.StatusBadRequest))
		return
	}

	postForm.CreatorType = "community"
	communityName := mux.Vars(r)["name"]
	if communityName == "" {
		logger.Error(ctx, "name length validation failed: length=%d", utf8.RuneCountInString(postForm.Text))
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "name must be between 1 and 4096 characters", http.StatusBadRequest))
		return
	}

	community, err := p.communityService.GetCommunityByName(ctx, communityName)
	if err != nil {
		logger.Error(ctx, "Could not get community by name %s", err)
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Could not get community by name "+err.Error(), http.StatusBadRequest))
		return
	}

	// Validate text length
	if utf8.RuneCountInString(postForm.Text) > 4096 {
		logger.Error(ctx, "Text length validation failed: length=%d", utf8.RuneCountInString(postForm.Text))
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Text must be between 1 and 4096 characters", http.StatusBadRequest))
		return
	}

	// Sanitize post content
	sanitizer.SanitizePost(&postForm, p.policy)

	if len(postForm.Text)+len(postForm.Media)+len(postForm.Audio)+len(postForm.File) == 0 {
		logger.Error(ctx, "empty post content")
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "empty post content", http.StatusBadRequest))
		return
	}

	// Convert the post form to a model
	post, err := postForm.ToPostModel(community.ID)
	if err != nil {
		logger.Error(ctx, "Failed to parse post form: %s", err.Error())
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Failed to parse post form", http.StatusBadRequest))
		return
	}

	newPost, err := p.postUseCase.AddPost(ctx, post)
	if err != nil {
		logger.Error(ctx, "Failed to add post: %s", err.Error())
		http2.WriteJSONError(w, err)
		return
	}
	logger.Info(ctx, "Successfully added post")

	// Prepare the post output
	var postOut forms.PostOut
	postOut.FromPost(*newPost)

	info, err := p.profileUseCase.GetPublicUserInfo(ctx, community.OwnerID)
	if err != nil {
		logger.Error(ctx, "Failed to get user info: %s", err.Error())
		http2.WriteJSONError(w, err)
		return
	}
	postOut.Creator = forms.ToCommunityForm(*community, info)

	// Return the response with the newly created post
	w.Header().Set("Content-Type", "application/json")
	out := forms.PayloadWrapper[forms.PostOut]{Payload: postOut}
	js, err := out.MarshalJSON()
	if err != nil {
		logger.Error(ctx, "Failed to marshal json payload %v", err)
		http2.WriteJSONError(w, err)
		return
	}
	if _, err = w.Write(js); err != nil {
		logger.Error(ctx, "Failed to encode feedback output %v", err)
		http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to encode feedback output", http.StatusInternalServerError))
	}
}

// DeletePost удаляет пост
// @Summary Удалить пост
// @Description Удаляет пост из ленты
// @Tags Feed
// @Param post_id path string true "Идентификатор поста"
// @Success 200 {string} string "OK"
// @Failure 400 {object} forms.ErrorForm "Некорректные данные"
// @Failure 403 {object} forms.ErrorForm "Пост не принадлежит пользователю"
// @Failure 404 {object} forms.ErrorForm "Пост не найден"
// @Failure 500 {object} forms.ErrorForm "Ошибка сервера"
// @Router /api/posts/{post_id} [delete]
func (p *PostHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := ctx.Value("user").(models.User)
	if !ok {
		logger.Error(ctx, "Failed to get user from context while deleting post")
		http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to get user from context", http.StatusInternalServerError))
		return
	}

	postIdString := mux.Vars(r)["post_id"]
	if len(postIdString) == 0 {
		logger.Error(ctx, "Failed to get post id from request")
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Failed to get post id", http.StatusBadRequest))
		return
	}
	logger.Info(ctx, "User %s requested to delete post with id %s", user.Username, postIdString)

	postId, err := uuid.Parse(postIdString)
	if err != nil {
		logger.Error(ctx, "Failed to parse post id: %s", err.Error())
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Failed to parse post id", http.StatusBadRequest))
		return
	}

	err = p.postUseCase.DeletePost(ctx, user.Id, postId)
	if err != nil {
		logger.Error(ctx, "Failed to delete post: %s", err.Error())
		http2.WriteJSONError(w, err)
		return
	}
	logger.Info(ctx, "Successfully deleted post %s", postIdString)
	w.WriteHeader(http.StatusNoContent)
}

// UpdatePost updates a post
// @Summary Update post
// @Description Updates a post
// @Tags Post
// @Accept json
// @Produce json
// @Param post_id path string true "Post ID"
// @Param post body forms.UpdatePostForm true "Post data"
// @Success 200 {string} string "OK"
// @Failure 400 {object} forms.ErrorForm "Invalid data"
// @Failure 403 {object} forms.ErrorForm "Post does not belong to user"
// @Failure 404 {object} forms.ErrorForm "Post not found"
// @Failure 500 {object} forms.ErrorForm "Server error"
// @Router /api/post [put]
func (p *PostHandler) UpdatePost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := ctx.Value("user").(models.User)
	if !ok {
		logger.Error(ctx, "Failed to get user from context while updating post")
		http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to get user from context", http.StatusInternalServerError))
		return
	}

	postIdString := mux.Vars(r)["post_id"]
	postId, err := uuid.Parse(postIdString)
	if err != nil {
		logger.Error(ctx, "Failed to parse post id: %s", err.Error())
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Failed to parse post id", http.StatusBadRequest))
		return
	}

	var updatePostForm forms.UpdatePostForm
	if err := easyjson.UnmarshalFromReader(r.Body, &updatePostForm); err != nil {
		logger.Error(ctx, "Failed to decode request body for feedback %v", err)
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Bad request body", http.StatusBadRequest))
		return
	}

	if utf8.RuneCountInString(updatePostForm.Text) > 4000 {
		logger.Error(ctx, "Text length validation failed: length=%d", utf8.RuneCountInString(updatePostForm.Text))
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Text must be between 1 and 4096 characters", http.StatusBadRequest))
		return
	}

	if len(updatePostForm.Text)+len(updatePostForm.Media)+len(updatePostForm.Audio)+len(updatePostForm.File) == 0 {
		logger.Error(ctx, "empty update content")
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "empty update content", http.StatusBadRequest))
		return
	}

	sanitizer.SanitizeUpdatePost(&updatePostForm, p.policy)

	updatePost, err := updatePostForm.ToPostUpdateModel(postId)
	if err != nil {
		logger.Error(ctx, "Failed to parse update post: %s", err.Error())
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Failed to parse update post", http.StatusBadRequest))
		return
	}

	post, err := p.postUseCase.UpdatePost(ctx, updatePost, user.Id)
	if err != nil {
		logger.Error(ctx, "Failed to update post: %s", err.Error())
		http2.WriteJSONError(w, err)
		return
	}

	logger.Info(ctx, "Successfully updated post %s", postIdString)
	var postOut forms.PostOut
	postOut.FromPost(*post)

	publicUserInfo, err := p.profileUseCase.GetPublicUserInfo(ctx, user.Id)
	if err != nil {
		logger.Error(ctx, "Failed to get public user info: %s", err.Error())
		http2.WriteJSONError(w, err)
		return
	}

	lastComment, err := p.commentUseCase.GetLastPostComment(ctx, post.Id)
	appErr := errors2.FromGRPCError(err)
	if appErr != nil && appErr.HTTPStatus != http.StatusNotFound {
		logger.Error(ctx, "Failed to get last comment")
		http2.WriteJSONError(w, appErr)
		return
	}

	if lastComment != nil {
		userInfo, err := p.profileUseCase.GetPublicUserInfo(ctx, lastComment.UserId)
		if err != nil {
			logger.Error(ctx, "Failed to get user info: %s", err.Error())
			http2.WriteJSONError(w, err)
			return
		}

		var commentOut forms.CommentOut
		commentOut.FromComment(*lastComment, userInfo)
		postOut.LastComment = &commentOut
	}

	info := forms.PublicUserInfoToOut(publicUserInfo, models.RelationSelf)
	postOut.Creator = &info

	w.Header().Set("Content-Type", "application/json")
	out := forms.PayloadWrapper[forms.PostOut]{Payload: postOut}
	js, err := out.MarshalJSON()
	if err != nil {
		logger.Error(ctx, "Failed to marshal json payload %v", err)
		http2.WriteJSONError(w, err)
		return
	}
	if _, err = w.Write(js); err != nil {
		logger.Error(ctx, "Failed to encode feedback output %v", err)
		http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to encode feedback output", http.StatusInternalServerError))
	}
}

func (p *PostHandler) LikePost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := ctx.Value("user").(models.User)
	if !ok {
		logger.Error(ctx, "Failed to get user from context while liking post")
		http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to get user from context", http.StatusInternalServerError))
		return
	}

	postIdStr := mux.Vars(r)["post_id"]
	postId, err := uuid.Parse(postIdStr)
	if err != nil {
		logger.Error(ctx, "Failed to parse post ID: %s", err.Error())
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Failed to parse post ID", http.StatusBadRequest))
		return
	}

	logger.Info(ctx, "User %s liked post %s", user.Username, postId.String())

	err = p.postUseCase.LikePost(ctx, postId, user.Id)
	if err != nil {
		logger.Error(ctx, "Failed to like post: %s", err.Error())
		http2.WriteJSONError(w, err)
		return
	}

	post, err := p.postUseCase.GetPost(ctx, postId, user.Id)
	if err != nil {
		logger.Error(ctx, "Failed to get post: %s", err.Error())
		http2.WriteJSONError(w, err)
		return
	}

	// notify
	err = p.likeWSHandler.NotifyPostLiked(ctx, user.Id, post.CreatorId, post)
	if err != nil {
		logger.Error(ctx, "Failed to like post: %s", err.Error())
		http2.WriteJSONError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (p *PostHandler) UnlikePost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := ctx.Value("user").(models.User)
	if !ok {
		logger.Error(ctx, "Failed to get user from context while unliking post")
		http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to get user from context", http.StatusInternalServerError))
		return
	}

	postIdStr := mux.Vars(r)["post_id"]
	postId, err := uuid.Parse(postIdStr)
	if err != nil {
		logger.Error(ctx, "Failed to parse post ID: %s", err.Error())
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Failed to parse post ID", http.StatusBadRequest))
		return
	}

	logger.Info(ctx, "User %s unliked post %s", user.Username, postId.String())

	err = p.postUseCase.UnlikePost(ctx, postId, user.Id)
	if err != nil {
		logger.Error(ctx, "Failed to unlike post: %s", err.Error())
		http2.WriteJSONError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (p *PostHandler) GetPost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	postIdStr := mux.Vars(r)["post_id"]
	postId, err := uuid.Parse(postIdStr)
	if err != nil {
		logger.Error(ctx, "Failed to parse post ID: %s", err.Error())
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Failed to parse post ID", http.StatusBadRequest))
		return
	}

	user, ok := ctx.Value("user").(models.User)
	if !ok {
		logger.Error(ctx, "Failed to get user from context while getting post")
		http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to get user from context", http.StatusInternalServerError))
		return
	}

	logger.Info(ctx, "User requested post %s", postId.String())

	post, err := p.postUseCase.GetPost(ctx, postId, user.Id)
	if err != nil {
		logger.Error(ctx, "Failed to get post: %s", err.Error())
		http2.WriteJSONError(w, err)
		return
	}

	lastComment, err := p.commentUseCase.GetLastPostComment(ctx, post.Id)
	appErr := errors2.FromGRPCError(err)
	if appErr != nil && appErr.HTTPStatus != http.StatusNotFound {
		logger.Error(ctx, "Failed to get last comment")
		http2.WriteJSONError(w, appErr)
		return
	}

	var postOut forms.PostOut
	postOut.FromPost(*post)

	if lastComment != nil {
		userInfo, err := p.profileUseCase.GetPublicUserInfo(ctx, lastComment.UserId)
		if err != nil {
			logger.Error(ctx, "Failed to get user info: %s", err.Error())
			http2.WriteJSONError(w, err)
			return
		}

		var commentOut forms.CommentOut
		commentOut.FromComment(*lastComment, userInfo)
		postOut.LastComment = &commentOut
	}

	if post.CreatorType == models.PostUser {
		publicAuthorInfo, err := p.profileUseCase.GetPublicUserInfo(ctx, post.CreatorId)
		if err != nil {
			logger.Error(ctx, "Failed to get public user info: %s", err.Error())
			http2.WriteJSONError(w, err)
			return
		}

		relation, err := p.friendsUseCase.GetUserRelation(ctx, post.CreatorId, post.CreatorId)
		if err != nil {
			logger.Error(ctx, "Failed to get user relation: %s", err.Error())
			http2.WriteJSONError(w, err)
			return
		}

		info := forms.PublicUserInfoToOut(publicAuthorInfo, relation)
		postOut.Creator = &info
	} else if post.CreatorType == models.PostCommunity {
		community, err := p.communityService.GetCommunityById(ctx, post.CreatorId)
		if err != nil {
			logger.Error(ctx, "Failed to get community: %s", err.Error())
			http2.WriteJSONError(w, err)
			return
		}

		publicAuthorInfo, err := p.profileUseCase.GetPublicUserInfo(ctx, community.OwnerID)
		if err != nil {
			logger.Error(ctx, "Failed to get public user info: %s", err.Error())
			http2.WriteJSONError(w, err)
			return
		}

		info := forms.ToCommunityForm(*community, publicAuthorInfo)
		postOut.Creator = &info
	}

	w.Header().Set("Content-Type", "application/json")
	out := forms.PayloadWrapper[forms.PostOut]{Payload: postOut}
	js, err := out.MarshalJSON()
	if err != nil {
		logger.Error(ctx, "Failed to marshal json payload %v", err)
		http2.WriteJSONError(w, err)
		return
	}
	if _, err = w.Write(js); err != nil {
		logger.Error(ctx, "Failed to encode feedback output %v", err)
		http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to encode feedback output", http.StatusInternalServerError))
	}
}

func (p *PostHandler) GetPostLikes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	postIdStr := mux.Vars(r)["post_id"]
	postId, err := uuid.Parse(postIdStr)
	if err != nil {
		logger.Error(ctx, "Failed to parse post ID: %s", err.Error())
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Failed to parse post ID", http.StatusBadRequest))
		return
	}

	numLikesStr := r.URL.Query().Get("count")
	numLikes, err := strconv.Atoi(numLikesStr)
	if err != nil || numLikes <= 0 {
		numLikes = 10 // значение по умолчанию
	}

	offsetStr := r.URL.Query().Get("offset")
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0 // значение по умолчанию
	}

	logger.Info(ctx, "User requested likes for post %s, num_likes: %d, offset: %d", postId.String(), numLikes, offset)

	likes, err := p.postUseCase.GetPostLikes(ctx, postId, numLikes, offset)
	if err != nil {
		logger.Error(ctx, "Failed to get post likes: %s", err.Error())
		http2.WriteJSONError(w, err)
		return
	}

	var likesOut []forms.LikeOut
	for _, like := range likes {
		userInfo, err := p.profileUseCase.GetPublicUserInfo(ctx, like.UserId)
		if err != nil {
			logger.Error(ctx, "Failed to get user info for like: %s", err.Error())
			http2.WriteJSONError(w, err)
			return
		}

		likeOut := forms.LikeFromModel(like, forms.PublicUserInfoToOut(userInfo, ""))
		likesOut = append(likesOut, likeOut)
	}

	w.Header().Set("Content-Type", "application/json")
	out := forms.PayloadWrapper[[]forms.LikeOut]{Payload: likesOut}
	js, err := out.MarshalJSON()
	if err != nil {
		logger.Error(ctx, "Failed to marshal json payload %v", err)
		http2.WriteJSONError(w, err)
		return
	}
	if _, err = w.Write(js); err != nil {
		logger.Error(ctx, "Failed to encode likes output %v", err)
		http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to encode likes output", http.StatusInternalServerError))
	}
}
