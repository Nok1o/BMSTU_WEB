//go:build unit
// +build unit

package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"quickflow/gateway/internal/delivery/http/mocks"
	"quickflow/shared/models"
)

func TestFeedHandler(t *testing.T) {
	tests := []struct {
		name         string
		method       string
		path         string
		setupRequest func() *http.Request
		mockSetup    func(*mocks.MockAuthUseCase, *mocks.MockPostService, *mocks.MockProfileUseCase,
			*mocks.MockFriendsUseCase, *mocks.MockCommunityService, *mocks.MockCommentService)
		expectedStatus int
	}{
		// GetFeed tests
		{
			name:   "GetFeed success",
			method: http.MethodGet,
			path:   "/api/feed",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/api/feed?posts_count=10", nil)
				ctx := context.WithValue(req.Context(), "user", models.User{Id: uuid.New()})
				return req.WithContext(ctx)
			},
			mockSetup: func(au *mocks.MockAuthUseCase, ps *mocks.MockPostService, pu *mocks.MockProfileUseCase,
				fu *mocks.MockFriendsUseCase, cs *mocks.MockCommunityService, cu *mocks.MockCommentService) {
				ps.EXPECT().FetchFeed(gomock.Any(), 10, gomock.Any(), gomock.Any()).
					Return([]models.Post{
						{CreatorType: models.PostUser, CreatorId: uuid.New()},
					}, nil)
				pu.EXPECT().GetPublicUsersInfo(gomock.Any(), gomock.Any()).Return([]models.PublicUserInfo{{Id: uuid.New()}}, nil)
				fu.EXPECT().GetUserRelation(gomock.Any(), gomock.Any(), gomock.Any()).Return(models.RelationNone, nil)
				cu.EXPECT().GetLastPostComment(gomock.Any(), gomock.Any()).Return(nil, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "GetFeed no user in context",
			method: http.MethodGet,
			path:   "/api/feed",
			setupRequest: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/api/feed?posts_count=10", nil)
			},
			mockSetup:      nil,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:   "GetFeed missing count param",
			method: http.MethodGet,
			path:   "/api/feed",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/api/feed", nil)
				ctx := context.WithValue(req.Context(), "user", models.User{Id: uuid.New()})
				return req.WithContext(ctx)
			},
			mockSetup:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "GetFeed post service error",
			method: http.MethodGet,
			path:   "/api/feed",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/api/feed?posts_count=10", nil)
				ctx := context.WithValue(req.Context(), "user", models.User{Id: uuid.New()})
				return req.WithContext(ctx)
			},
			mockSetup: func(au *mocks.MockAuthUseCase, ps *mocks.MockPostService, pu *mocks.MockProfileUseCase,
				fu *mocks.MockFriendsUseCase, cs *mocks.MockCommunityService, cu *mocks.MockCommentService) {
				ps.EXPECT().FetchFeed(gomock.Any(), 10, gomock.Any(), gomock.Any()).Return(nil, errors.New("error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},

		// GetRecommendations tests
		{
			name:   "GetRecommendations success",
			method: http.MethodGet,
			path:   "/api/recommendations",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/api/recommendations?posts_count=10", nil)
				ctx := context.WithValue(req.Context(), "user", models.User{Id: uuid.New()})
				return req.WithContext(ctx)
			},
			mockSetup: func(au *mocks.MockAuthUseCase, ps *mocks.MockPostService, pu *mocks.MockProfileUseCase,
				fu *mocks.MockFriendsUseCase, cs *mocks.MockCommunityService, cu *mocks.MockCommentService) {
				ps.EXPECT().FetchRecommendations(gomock.Any(), 10, gomock.Any(), gomock.Any()).
					Return([]models.Post{
						{CreatorType: models.PostUser, CreatorId: uuid.New()},
					}, nil)
				pu.EXPECT().GetPublicUsersInfo(gomock.Any(), gomock.Any()).Return([]models.PublicUserInfo{{Id: uuid.New()}}, nil)
				fu.EXPECT().GetUserRelation(gomock.Any(), gomock.Any(), gomock.Any()).Return(models.RelationNone, nil)
				cu.EXPECT().GetLastPostComment(gomock.Any(), gomock.Any()).Return(nil, nil)
			},
			expectedStatus: http.StatusOK,
		},

		// FetchUserPosts tests
		{
			name:   "FetchUserPosts success",
			method: http.MethodGet,
			path:   "/api/profiles/test/posts",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/api/profiles/test/posts?posts_count=10", nil)
				ctx := context.WithValue(req.Context(), "user", models.User{Id: uuid.New()})
				return req.WithContext(ctx)
			},
			mockSetup: func(au *mocks.MockAuthUseCase, ps *mocks.MockPostService, pu *mocks.MockProfileUseCase,
				fu *mocks.MockFriendsUseCase, cs *mocks.MockCommunityService, cu *mocks.MockCommentService) {
				au.EXPECT().GetUserByUsername(gomock.Any(), "test").Return(models.User{Id: uuid.New()}, nil)
				ps.EXPECT().FetchCreatorPosts(gomock.Any(), gomock.Any(), gomock.Any(), 10, gomock.Any()).
					Return([]models.Post{{CreatorType: models.PostUser, CreatorId: uuid.New()}}, nil)
				pu.EXPECT().GetPublicUserInfo(gomock.Any(), gomock.Any()).Return(models.PublicUserInfo{}, nil)
				cu.EXPECT().GetLastPostComment(gomock.Any(), gomock.Any()).Return(nil, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "FetchUserPosts missing username",
			method: http.MethodGet,
			path:   "/api/profiles//posts",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/api/profiles//posts?posts_count=10", nil)
				ctx := context.WithValue(req.Context(), "user", models.User{Id: uuid.New()})
				return req.WithContext(ctx)
			},
			mockSetup:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "FetchUserPosts user not found",
			method: http.MethodGet,
			path:   "/api/profiles/test/posts",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/api/profiles/test/posts?posts_count=10", nil)
				ctx := context.WithValue(req.Context(), "user", models.User{Id: uuid.New()})
				return req.WithContext(ctx)
			},
			mockSetup: func(au *mocks.MockAuthUseCase, ps *mocks.MockPostService, pu *mocks.MockProfileUseCase,
				fu *mocks.MockFriendsUseCase, cs *mocks.MockCommunityService, cu *mocks.MockCommentService) {
				au.EXPECT().GetUserByUsername(gomock.Any(), "test").Return(models.User{}, errors.New("not found"))
			},
			expectedStatus: http.StatusNotFound,
		},

		// FetchCommunityPosts tests
		{
			name:   "FetchCommunityPosts success",
			method: http.MethodGet,
			path:   "/api/communities/test/posts",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/api/communities/test/posts?posts_count=10", nil)
				ctx := context.WithValue(req.Context(), "user", models.User{Id: uuid.New()})
				return req.WithContext(ctx)
			},
			mockSetup: func(au *mocks.MockAuthUseCase, ps *mocks.MockPostService, pu *mocks.MockProfileUseCase,
				fu *mocks.MockFriendsUseCase, cs *mocks.MockCommunityService, cu *mocks.MockCommentService) {
				cs.EXPECT().GetCommunityByName(gomock.Any(), "test").Return(&models.Community{ID: uuid.New(), OwnerID: uuid.New()}, nil)
				ps.EXPECT().FetchCreatorPosts(gomock.Any(), gomock.Any(), gomock.Any(), 10, gomock.Any()).
					Return([]models.Post{{CreatorType: models.PostCommunity, CreatorId: uuid.New()}}, nil)
				pu.EXPECT().GetPublicUserInfo(gomock.Any(), gomock.Any()).Return(models.PublicUserInfo{}, nil)
				cu.EXPECT().GetLastPostComment(gomock.Any(), gomock.Any()).Return(nil, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "FetchCommunityPosts community not found",
			method: http.MethodGet,
			path:   "/api/communities/test/posts",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/api/communities/test/posts?posts_count=10", nil)
				ctx := context.WithValue(req.Context(), "user", models.User{Id: uuid.New()})
				return req.WithContext(ctx)
			},
			mockSetup: func(au *mocks.MockAuthUseCase, ps *mocks.MockPostService, pu *mocks.MockProfileUseCase,
				fu *mocks.MockFriendsUseCase, cs *mocks.MockCommunityService, cu *mocks.MockCommentService) {
				cs.EXPECT().GetCommunityByName(gomock.Any(), "test").Return(nil, errors.New("not found"))
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockAuthUC := mocks.NewMockAuthUseCase(ctrl)
			mockPostSvc := mocks.NewMockPostService(ctrl)
			mockProfileUC := mocks.NewMockProfileUseCase(ctrl)
			mockFriendsUC := mocks.NewMockFriendsUseCase(ctrl)
			mockCommSvc := mocks.NewMockCommunityService(ctrl)
			mockCommentSvc := mocks.NewMockCommentService(ctrl)

			if tt.mockSetup != nil {
				tt.mockSetup(mockAuthUC, mockPostSvc, mockProfileUC, mockFriendsUC, mockCommSvc, mockCommentSvc)
			}

			handler := NewFeedHandler(
				mockAuthUC,
				mockPostSvc,
				mockProfileUC,
				mockFriendsUC,
				mockCommSvc,
				mockCommentSvc,
			)

			req := tt.setupRequest()
			rr := httptest.NewRecorder()

			// Use router to handle path variables
			router := mux.NewRouter()
			switch tt.path {
			case "/api/feed":
				router.HandleFunc("/api/feed", handler.GetFeed)
			case "/api/recommendations":
				router.HandleFunc("/api/recommendations", handler.GetRecommendations)
			case "/api/profiles/test/posts":
				router.HandleFunc("/api/profiles/{username}/posts", handler.FetchUserPosts)
			case "/api/communities/test/posts":
				router.HandleFunc("/api/communities/{name}/posts", handler.FetchCommunityPosts)
			}

			router.ServeHTTP(rr, req)
		})
	}
}
