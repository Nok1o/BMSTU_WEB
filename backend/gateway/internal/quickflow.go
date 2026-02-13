package internal

import (
	"fmt"
	"net/http"
	service_discovery "quickflow/utils/service-discovery"
	"strings"

	"github.com/gorilla/mux"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/microcosm-cc/bluemonday"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"quickflow/config"
	addr "quickflow/config/micro-addr"
	qfhttp "quickflow/gateway/internal/delivery/http"
	"quickflow/gateway/internal/delivery/http/middleware"
	"quickflow/gateway/internal/delivery/ws"
	"quickflow/metrics"
	"quickflow/shared/client/community_service"
	"quickflow/shared/client/feedback_service"
	"quickflow/shared/client/file_service"
	friendsService "quickflow/shared/client/friends_service"
	"quickflow/shared/client/messenger_service"
	postService "quickflow/shared/client/post_service"
	userService "quickflow/shared/client/user_service"
	"quickflow/shared/interceptors"
)

func BuildHandler(cfg *config.Config) (*mux.Router, error) {
	metrics := metrics.NewMetrics("QuickFlow")

	grpcConnPostService, err := service_discovery.NewGRPCClient(
		addr.DefaultPostServiceName,
		service_discovery.ModeFailover,
		interceptors.RequestIDClientInterceptor(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to post service: %w", err)
	}

	grpcConnUserService, err := service_discovery.NewGRPCClient(
		addr.DefaultUserServiceName,
		service_discovery.ModeFailover,
		interceptors.RequestIDClientInterceptor(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to user service: %w", err)
	}

	grpcConnMessengerService, err := service_discovery.NewGRPCClient(
		addr.DefaultMessengerServiceName,
		service_discovery.ModeFailover,
		interceptors.RequestIDClientInterceptor(),
	)

	grpcConnFeedbackService, err := service_discovery.NewGRPCClient(
		addr.DefaultFeedbackServiceName,
		service_discovery.ModeFailover,
		interceptors.RequestIDClientInterceptor(),
	)

	grpcConnFriendsService, err := service_discovery.NewGRPCClient(
		addr.DefaultFriendsServiceName,
		service_discovery.ModeFailover,
		interceptors.RequestIDClientInterceptor(),
	)

	grcpConnCommunityService, err := service_discovery.NewGRPCClient(
		addr.DefaultCommunityServiceName,
		service_discovery.ModeFailover,
		interceptors.RequestIDClientInterceptor(),
	)

	// services
	fileService, err := getFileService()
	UserService := userService.NewUserClient(grpcConnUserService)
	profileService := userService.NewProfileClient(grpcConnUserService)
	PostService := postService.NewPostServiceClient(grpcConnPostService)
	chatService := messenger_service.NewChatServiceClient(grpcConnMessengerService)
	messageService := messenger_service.NewMessageServiceClient(grpcConnMessengerService)
	feedbackService := feedback_service.NewFeedbackClient(grpcConnFeedbackService)
	FriendsService := friendsService.NewFriendsClient(grpcConnFriendsService)
	communityService := community_service.NewCommunityServiceClient(grcpConnCommunityService)
	commentService := postService.NewCommentClient(grpcConnPostService)
	stickerService := messenger_service.NewStickerServiceClient(grpcConnMessengerService)

	connManager := ws.NewWSConnectionManager()
	wsRouter := ws.NewWebSocketRouter()
	wsMessageHander := ws.NewInternalWSMessageHandler(connManager, messageService, profileService, chatService)
	wsFriendHandler := ws.NewInternalWSFriendsHandler(connManager, profileService)
	wsLikeHandler := ws.NewInternalWSPostHandler(connManager, profileService)
	pingHandler := ws.NewPingHandlerWS()

	sanitizerPolicy := bluemonday.UGCPolicy()

	newAuthHandler := qfhttp.NewAuthHandler(UserService, sanitizerPolicy)
	newFeedHandler := qfhttp.NewFeedHandler(UserService, PostService, profileService, FriendsService, communityService, commentService)
	newPostHandler := qfhttp.NewPostHandler(PostService, profileService, communityService, FriendsService, commentService, wsLikeHandler, sanitizerPolicy)
	newCommentHandler := qfhttp.NewCommentHandler(commentService, profileService, PostService, wsLikeHandler, sanitizerPolicy)
	newProfileHandler := qfhttp.NewProfileHandler(profileService, FriendsService, UserService, chatService, connManager, sanitizerPolicy)
	newMessageHandler := qfhttp.NewMessageHandler(messageService, UserService, profileService, sanitizerPolicy)
	newChatHandler := qfhttp.NewChatHandler(chatService, profileService, messageService, connManager)
	newFriendsHandler := qfhttp.NewFriendsHandler(FriendsService, connManager, wsFriendHandler)
	newSearchHandler := qfhttp.NewSearchHandler(UserService, communityService, profileService)
	newCommunityHandler := qfhttp.NewCommunityHandler(communityService, profileService, connManager, UserService, sanitizerPolicy)
	newFileHandler := qfhttp.NewFileHandler(fileService, sanitizerPolicy)
	newStickerHandler := qfhttp.NewStickerHandler(stickerService, sanitizerPolicy)

	CSRFHandler := qfhttp.NewCSRFHandler()
	FeedbackHandler := qfhttp.NewFeedbackHandler(feedbackService, profileService, sanitizerPolicy)

	// register handlers
	wsRouter.RegisterHandler(ws.MessageEventSend, wsMessageHander.SendMessage)
	wsRouter.RegisterHandler(ws.MessageEventRead, wsMessageHander.MarkMessageRead)
	wsRouter.RegisterHandler(ws.MessageEventDeleted, wsMessageHander.DeleteMessage)
	wsRouter.RegisterHandler(ws.ChatEventDeleted, wsMessageHander.DeleteChat)

	newMessageHandlerWS := ws.NewMessageListenerWS(profileService, connManager, wsRouter, sanitizerPolicy)

	// routing
	r := mux.NewRouter()
	r.Use(middleware.CORSMiddleware(cfg.CORSConfig))
	r.Use(mux.CORSMethodMiddleware(r))
	r.Use(middleware.RecoveryMiddleware)
	r.Use(middleware.MetricsMiddleware(metrics))
	r.Use(middleware.ReadOnlyMiddleware)
	r.MethodNotAllowedHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	r.PathPrefix("/api/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Vary", "Origin")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", strings.Join([]string{"GET", "POST", "DELETE", "PATCH", "PUT", "OPTIONS"}, ", "))
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Cookie")
		w.WriteHeader(http.StatusNoContent)
		return

	}).Methods(http.MethodOptions)

	r.HandleFunc("/api/v1/hello", newAuthHandler.Greet).Methods(http.MethodGet)
	r.HandleFunc("/api/v1/profiles/{username}", newProfileHandler.GetProfile).Methods(http.MethodGet)
	r.Handle("/metrics", promhttp.Handler()).Methods(http.MethodGet)

	apiPostRouter := r.PathPrefix("/").Subrouter()
	apiPostRouter.Use(middleware.ContentTypeMiddleware("application/json", "multipart/form-data", ""))

	apiGetRouter := r.PathPrefix("/").Subrouter()

	// validating that the content type is application/json for every route but /hello

	apiPostRouter.HandleFunc("/api/v1/signup", newAuthHandler.SignUp).Methods(http.MethodPost)
	apiPostRouter.HandleFunc("/api/v1/login", newAuthHandler.Login).Methods(http.MethodPost)
	apiPostRouter.HandleFunc("/api/v1/logout", newAuthHandler.Logout).Methods(http.MethodPost)

	// Subrouter for protected routes
	protectedPost := apiPostRouter.PathPrefix("/").Subrouter()
	protectedPost.Use(middleware.SessionMiddleware(UserService))
	//protectedPost.Use(middleware.CSRFMiddleware)
	protectedPost.HandleFunc("/api/v1/post", newPostHandler.AddPost).Methods(http.MethodPost)
	protectedPost.HandleFunc("/api/v1/posts/{post_id:[0-9a-fA-F-]{36}}", newPostHandler.UpdatePost).Methods(http.MethodPut)

	protectedPost.HandleFunc("/api/v1/comments/{comment_id:[0-9a-fA-F-]{36}}", newCommentHandler.UpdateComment).Methods(http.MethodPut)
	protectedPost.HandleFunc("/api/v1/posts/{post_id:[0-9a-fA-F-]{36}}/like", newPostHandler.LikePost).Methods(http.MethodPost)
	protectedPost.HandleFunc("/api/v1/comments/{comment_id:[0-9a-fA-F-]{36}}/like", newCommentHandler.LikeComment).Methods(http.MethodPost)
	protectedPost.HandleFunc("/api/v1/profile", newProfileHandler.UpdateProfile).Methods(http.MethodPost)
	protectedPost.HandleFunc("/api/v1/follow", newFriendsHandler.SendFriendRequest).Methods(http.MethodPost)
	protectedPost.HandleFunc("/api/v1/followers/accept", newFriendsHandler.AcceptFriendRequest).Methods(http.MethodPost)
	protectedPost.HandleFunc("/api/v1/followers/reject", newFriendsHandler.MarkRead).Methods(http.MethodPost)
	protectedPost.HandleFunc("/api/v1/feedback", FeedbackHandler.SaveFeedback).Methods(http.MethodPost)
	protectedPost.HandleFunc("/api/v1/community", newCommunityHandler.CreateCommunity).Methods(http.MethodPost)
	protectedPost.HandleFunc("/api/v1/communities/{id:[0-9a-fA-F-]{36}}", newCommunityHandler.UpdateCommunity).Methods(http.MethodPut)
	protectedPost.HandleFunc("/api/v1/communities/{id:[0-9a-fA-F-]{36}}/join", newCommunityHandler.JoinCommunity).Methods(http.MethodPost)
	protectedPost.HandleFunc("/api/v1/communities/{id:[0-9a-fA-F-]{36}}/leave", newCommunityHandler.LeaveCommunity).Methods(http.MethodPost)
	protectedPost.HandleFunc("/api/v1/communities/{id:[0-9a-fA-F-]{36}}/members/{user_id:{id:[0-9a-fA-F-]{36}}", newCommunityHandler.ChangeUserRole).Methods(http.MethodPost)
	protectedPost.HandleFunc("/api/v1/upload", newFileHandler.AddFiles).Methods(http.MethodPost)
	protectedPost.HandleFunc("/api/v1/posts/{post_id:[0-9a-fA-F-]{36}}/comment", newCommentHandler.AddComment).Methods(http.MethodPost)
	protectedPost.HandleFunc("/api/v1/sticker_packs/add", newStickerHandler.AddStickerPack).Methods(http.MethodPost)

	protectedGet := apiGetRouter.PathPrefix("/").Subrouter()
	protectedGet.Use(middleware.SessionMiddleware(UserService))
	protectedGet.HandleFunc("/api/v1/feed", newFeedHandler.GetFeed).Methods(http.MethodGet)
	protectedGet.HandleFunc("/api/v1/recommendations", newFeedHandler.GetRecommendations).Methods(http.MethodGet)
	protectedGet.HandleFunc("/api/v1/chats/{chat_id:[0-9a-fA-F-]{36}}/messages", newMessageHandler.GetMessagesForChat).Methods(http.MethodGet)
	protectedGet.HandleFunc("/api/v1/chats", newChatHandler.GetUserChats).Methods(http.MethodGet)
	protectedGet.HandleFunc("/api/v1/friends", newFriendsHandler.GetFriends).Methods(http.MethodGet)
	protectedGet.HandleFunc("/api/v1/csrf", CSRFHandler.GetCSRF).Methods(http.MethodGet)
	protectedGet.HandleFunc("/api/v1/users/search", newSearchHandler.SearchSimilarUsers).Methods(http.MethodGet)
	protectedGet.HandleFunc("/api/v1/communities/search", newSearchHandler.SearchSimilarCommunities).Methods(http.MethodGet)
	protectedGet.HandleFunc("/api/v1/feedback", FeedbackHandler.GetAllFeedbackType).Methods(http.MethodGet)
	protectedGet.HandleFunc("/api/v1/communities/{id:[0-9a-fA-F-]{36}}", newCommunityHandler.GetCommunityById).Methods(http.MethodGet)
	protectedGet.HandleFunc("/api/v1/communities/{name}", newCommunityHandler.GetCommunityByName).Methods(http.MethodGet)
	protectedGet.HandleFunc("/api/v1/communities/{id:[0-9a-fA-F-]{36}}/members", newCommunityHandler.GetCommunityMembers).Methods(http.MethodGet)
	protectedGet.HandleFunc("/api/v1/profiles/{username}/communities", newCommunityHandler.GetUserCommunities).Methods(http.MethodGet)
	protectedGet.HandleFunc("/api/v1/profiles/{username}/controlled", newCommunityHandler.GetControlledCommunities).Methods(http.MethodGet)
	protectedGet.HandleFunc("/api/v1/communities/{name}/posts", newFeedHandler.FetchCommunityPosts).Methods(http.MethodGet)
	protectedGet.HandleFunc("/api/v1/profiles/{username}/posts", newFeedHandler.FetchUserPosts).Methods(http.MethodGet)
	protectedGet.HandleFunc("/api/v1/my_profile", newProfileHandler.GetMyProfile).Methods(http.MethodGet)
	protectedGet.HandleFunc("/api/v1/posts/{post_id:[0-9a-fA-F-]{36}}", newPostHandler.GetPost).Methods(http.MethodGet)
	protectedGet.HandleFunc("/api/v1/posts/{post_id:[0-9a-fA-F-]{36}}/comments", newCommentHandler.FetchCommentsForPost).Methods(http.MethodGet)
	protectedGet.HandleFunc("/api/v1/chats/unread", newChatHandler.GetNumUnreadChats).Methods(http.MethodGet)
	protectedGet.HandleFunc("/api/v1/sticker_packs/{pack_id:[0-9a-fA-F-]{36}}", newStickerHandler.GetStickerPack).Methods(http.MethodGet)
	protectedGet.HandleFunc("/api/v1/sticker_packs", newStickerHandler.GetStickerPacks).Methods(http.MethodGet)
	protectedGet.HandleFunc("/api/v1/sticker_packs/{pack_name}", newStickerHandler.GetStickerPackByName).Methods(http.MethodGet)

	wsProtected := protectedGet.PathPrefix("/").Subrouter()
	wsProtected.Use(middleware.WebSocketMiddleware(connManager, pingHandler))
	wsProtected.HandleFunc("/api/v1/ws", newMessageHandlerWS.HandleMessages).Methods(http.MethodGet)

	apiDeleteRouter := r.PathPrefix("/").Subrouter()
	apiDeleteRouter.Use(middleware.SessionMiddleware(UserService))
	//apiDeleteRouter.Use(middleware.CSRFMiddleware)
	apiDeleteRouter.HandleFunc("/api/v1/posts/{post_id:[0-9a-fA-F-]{36}}", newPostHandler.DeletePost).Methods(http.MethodDelete)
	apiDeleteRouter.HandleFunc("/api/v1/posts/{post_id:[0-9a-fA-F-]{36}}/like", newPostHandler.UnlikePost).Methods(http.MethodDelete)
	apiDeleteRouter.HandleFunc("/api/v1/comments/{comment_id:[0-9a-fA-F-]{36}}/like", newCommentHandler.UnlikeComment).Methods(http.MethodDelete)
	apiDeleteRouter.HandleFunc("/api/v1/friends", newFriendsHandler.DeleteFriend).Methods(http.MethodDelete)
	apiDeleteRouter.HandleFunc("/api/v1/follow", newFriendsHandler.Unfollow).Methods(http.MethodDelete)
	apiDeleteRouter.HandleFunc("/api/v1/communities/{id:[0-9a-fA-F-]{36}}", newCommunityHandler.DeleteCommunity).Methods(http.MethodDelete)
	apiDeleteRouter.HandleFunc("/api/v1/comments/{comment_id:[0-9a-fA-F-]{36}}", newCommentHandler.DeleteComment).Methods(http.MethodDelete)
	apiDeleteRouter.HandleFunc("/api/v1/sticker_packs/{pack_id:[0-9a-fA-F-]{36}}", newStickerHandler.DeleteStickerPack).Methods(http.MethodDelete)

	// v2
	apiPostRouter.HandleFunc("/api/v2/users", newAuthHandler.SignUp).Methods(http.MethodPost)
	apiPostRouter.HandleFunc("/api/v2/sessions", newAuthHandler.Login).Methods(http.MethodPost)
	protectedPost.HandleFunc("/api/v2/sessions/me", newAuthHandler.Logout).Methods(http.MethodDelete)

	protectedPost.HandleFunc("/api/v2/profiles/me", newProfileHandler.UpdateProfile).Methods(http.MethodPatch)
	protectedGet.HandleFunc("/api/v2/profiles/{username}", newProfileHandler.GetProfile).Methods(http.MethodGet)
	protectedGet.HandleFunc("/api/v2/users", newSearchHandler.SearchSimilarUsers).Methods(http.MethodGet)

	protectedGet.HandleFunc("/api/v2/users/{user_id:[0-9a-fA-F-]{36}}/friends", newFriendsHandler.GetFriends).Methods(http.MethodGet)
	apiDeleteRouter.HandleFunc("/api/v2/friends/{friend_id:[0-9a-fA-F-]{36}}", newFriendsHandler.DeleteFriend).Methods(http.MethodDelete)
	protectedPost.HandleFunc("/api/v2/friend_requests", newFriendsHandler.SendFriendRequest).Methods(http.MethodPost)
	protectedPost.HandleFunc("/api/v2/friend_requests/{request_id:[0-9a-fA-F-]{36}}", newFriendsHandler.Unfollow).Methods(http.MethodDelete)
	protectedPost.HandleFunc("/api/v2/friend_requests/{request_id:[0-9a-fA-F-]{36}}", newFriendsHandler.ChangeRequestStatus).Methods(http.MethodPut)

	protectedGet.HandleFunc("/api/v2/users/{username}/communities", newCommunityHandler.GetUserCommunities).Methods(http.MethodGet)
	protectedPost.HandleFunc("/api/v2/communities", newCommunityHandler.CreateCommunity).Methods(http.MethodPost)
	protectedGet.HandleFunc("/api/v2/communities/{id:[0-9a-fA-F-]{36}}", newCommunityHandler.GetCommunityById).Methods(http.MethodGet)
	apiDeleteRouter.HandleFunc("/api/v2/communities/{id:[0-9a-fA-F-]{36}}", newCommunityHandler.DeleteCommunity).Methods(http.MethodDelete)

	protectedGet.HandleFunc("/api/v2/communities", newSearchHandler.SearchSimilarCommunities).Methods(http.MethodGet)
	protectedPost.HandleFunc("/api/v2/communities/{id:[0-9a-fA-F-]{36}}/members", newCommunityHandler.JoinCommunity).Methods(http.MethodPost)
	protectedGet.HandleFunc("/api/v2/communities/{id:[0-9a-fA-F-]{36}}/members", newCommunityHandler.GetCommunityMembers).Methods(http.MethodGet)
	apiDeleteRouter.HandleFunc("/api/v2/communities/{id:[0-9a-fA-F-]{36}}/members/{user_id:{[0-9a-fA-F-]{36}}", newCommunityHandler.LeaveCommunity).Methods(http.MethodDelete)
	protectedPost.HandleFunc("/api/v2/communities/{community_id:[0-9a-fA-F-]{36}}/members/{user_id:[0-9a-fA-F-]{36}}", newCommunityHandler.ChangeUserRole).Methods(http.MethodPatch)
	protectedGet.HandleFunc("/api/v2/communities/{name}", newCommunityHandler.GetCommunityByName).Methods(http.MethodGet)
	protectedGet.HandleFunc("/api/v2/communities/{name}/posts", newFeedHandler.FetchCommunityPosts).Methods(http.MethodGet)
	protectedPost.HandleFunc("/api/v2/communities/{name}/posts", newPostHandler.AddPostCommunity).Methods(http.MethodPost)

	protectedPost.HandleFunc("/api/v2/posts", newPostHandler.AddPost).Methods(http.MethodPost)
	protectedGet.HandleFunc("/api/v2/posts/{post_id:[0-9a-fA-F-]{36}}", newPostHandler.GetPost).Methods(http.MethodGet)
	protectedPost.HandleFunc("/api/v2/posts/{post_id:[0-9a-fA-F-]{36}}", newPostHandler.UpdatePost).Methods(http.MethodPatch)
	apiDeleteRouter.HandleFunc("/api/v2/posts/{post_id:[0-9a-fA-F-]{36}}", newPostHandler.DeletePost).Methods(http.MethodDelete)
	protectedGet.HandleFunc("/api/v2/users/{username}/posts", newFeedHandler.FetchUserPosts).Methods(http.MethodGet)
	protectedGet.HandleFunc("/api/v2/posts", newFeedHandler.GetFeed).Methods(http.MethodGet)

	protectedPost.HandleFunc("/api/v2/posts/{post_id:[0-9a-fA-F-]{36}}/likes", newPostHandler.LikePost).Methods(http.MethodPost)
	protectedPost.HandleFunc("/api/v2/posts/{post_id:[0-9a-fA-F-]{36}}/comments/{comment_id:[0-9a-fA-F-]{36}}/likes", newCommentHandler.LikeComment).Methods(http.MethodPost)
	apiDeleteRouter.HandleFunc("/api/v2/posts/{post_id:[0-9a-fA-F-]{36}}/likes/me", newPostHandler.UnlikePost).Methods(http.MethodDelete)
	apiDeleteRouter.HandleFunc("/api/v2/posts/{post_id:[0-9a-fA-F-]{36}}/comments/{comment_id:[0-9a-fA-F-]{36}}/likes/me", newCommentHandler.UnlikeComment).Methods(http.MethodDelete)
	protectedGet.HandleFunc("/api/v2/posts/{post_id:[0-9a-fA-F-]{36}}/likes", newPostHandler.GetPostLikes).Methods(http.MethodGet)
	protectedGet.HandleFunc("/api/v2/posts/{post_id:[0-9a-fA-F-]{36}}/comments/{comment_id:[0-9a-fA-F-]{36}}/likes", newCommentHandler.GetCommentLikes).Methods(http.MethodGet)

	// comments
	protectedGet.HandleFunc("/api/v2/posts/{post_id:[0-9a-fA-F-]{36}}/comments", newCommentHandler.FetchCommentsForPost).Methods(http.MethodGet)
	protectedPost.HandleFunc("/api/v2/posts/{post_id:[0-9a-fA-F-]{36}}/comments", newCommentHandler.AddComment).Methods(http.MethodPost)
	protectedPost.HandleFunc("/api/v2/posts/{post_id:[0-9a-fA-F-]{36}}/comments/{comment_id:[0-9a-fA-F-]{36}}", newCommentHandler.UpdateComment).Methods(http.MethodPatch)
	apiDeleteRouter.HandleFunc("/api/v2/posts/{post_id:[0-9a-fA-F-]{36}}/comments/{comment_id:[0-9a-fA-F-]{36}}", newCommentHandler.DeleteComment).Methods(http.MethodDelete)

	protectedGet.HandleFunc("/api/v2/chats", newChatHandler.GetUserChats).Methods(http.MethodGet)
	protectedGet.HandleFunc("/api/v2/chats/{chat_id:[0-9a-fA-F-]{36}}/messages", newMessageHandler.GetMessagesForChat).Methods(http.MethodGet)

	r.HandleFunc("/api/v2/health", newAuthHandler.Greet).Methods(http.MethodGet)
	protectedPost.HandleFunc("/api/v2/files", newFileHandler.AddFiles).Methods(http.MethodPost)

	wsProtected.HandleFunc("/api/v2/ws", newMessageHandlerWS.HandleMessages).Methods(http.MethodGet)

	return r, nil
}

func getFileService() (*file_service.FileClient, error) {
	grpcConnFileService, err := service_discovery.NewGRPCClient(
		addr.DefaultFileServiceName,
		service_discovery.ModeFailover,
		interceptors.RequestIDClientInterceptor(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to file service: %w", err)
	}

	return file_service.NewFileClient(grpcConnFileService), nil
}

func Run(cfg *config.Config) error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}

	r, err := BuildHandler(cfg)
	if err != nil {
		return err
	}

	server := http.Server{
		Addr:         cfg.ServerConfig.Addr,
		Handler:      r,
		ReadTimeout:  cfg.ServerConfig.ReadTimeout,
		WriteTimeout: cfg.ServerConfig.WriteTimeout,
	}

	fmt.Printf("starting server at %s\n", cfg.ServerConfig.Addr)
	err = server.ListenAndServe()
	if err != nil {
		return fmt.Errorf("internal.Run: %w", err)
	}

	return nil
}
