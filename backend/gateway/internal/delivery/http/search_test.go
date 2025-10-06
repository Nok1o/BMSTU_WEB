//go:build unit
// +build unit

package http

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"quickflow/gateway/internal/delivery/http/mocks"
	"quickflow/shared/models"
)

func TestSearchHandler(t *testing.T) {

	uuid_ := uuid.New()
	tests := []struct {
		name           string
		method         string
		path           string
		queryParams    url.Values
		mockSetup      func(*mocks.MockSearchUseCase, *mocks.MockCommunityService, *mocks.MockProfileUseCase)
		expectedStatus int
	}{
		{
			name:   "SearchSimilarUsers success",
			method: http.MethodGet,
			path:   "/search/users",
			queryParams: url.Values{
				"to_search": []string{"test"},
				"count":     []string{"10"},
			},
			mockSetup: func(searchUC *mocks.MockSearchUseCase, commSvc *mocks.MockCommunityService, profileUC *mocks.MockProfileUseCase) {
				searchUC.EXPECT().SearchSimilarUser(gomock.Any(), "test", uint(10)).
					Return([]models.PublicUserInfo{
						{Id: uuid_, Username: "testuser"},
					}, nil).AnyTimes()
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "SearchSimilarUsers missing params",
			method: http.MethodGet,
			path:   "/search/users",
			queryParams: url.Values{
				"count": []string{"10"},
			},
			mockSetup:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "SearchSimilarUsers invalid count",
			method: http.MethodGet,
			path:   "/search/users",
			queryParams: url.Values{
				"to_search": []string{"test"},
				"count":     []string{"invalid"},
			},
			mockSetup:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "SearchSimilarCommunities success",
			method: http.MethodGet,
			path:   "/search/communities",
			queryParams: url.Values{
				"to_search": []string{"test"},
				"count":     []string{"5"},
			},
			mockSetup: func(searchUC *mocks.MockSearchUseCase, commSvc *mocks.MockCommunityService, profileUC *mocks.MockProfileUseCase) {
				commSvc.EXPECT().SearchSimilarCommunities(gomock.Any(), "test", 5).
					Return([]*models.Community{
						{ID: uuid_, OwnerID: uuid_},
					}, nil).AnyTimes()
				profileUC.EXPECT().GetPublicUserInfo(gomock.Any(), "owner1").
					Return(models.PublicUserInfo{Id: uuid_, Username: "owner"}, nil).AnyTimes()
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "SearchSimilarCommunities usecase error",
			method: http.MethodGet,
			path:   "/search/communities",
			queryParams: url.Values{
				"to_search": []string{"test"},
				"count":     []string{"5"},
			},
			mockSetup: func(searchUC *mocks.MockSearchUseCase, commSvc *mocks.MockCommunityService, profileUC *mocks.MockProfileUseCase) {
				commSvc.EXPECT().SearchSimilarCommunities(gomock.Any(), "test", 5).
					Return(nil, errors.New("some error")).AnyTimes()
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSearchUC := mocks.NewMockSearchUseCase(ctrl)
			mockCommSvc := mocks.NewMockCommunityService(ctrl)
			mockProfileUC := mocks.NewMockProfileUseCase(ctrl)

			if tt.mockSetup != nil {
				tt.mockSetup(mockSearchUC, mockCommSvc, mockProfileUC)
			}

			handler := NewSearchHandler(mockSearchUC, mockCommSvc, mockProfileUC)

			req, err := http.NewRequest(tt.method, tt.path, nil)
			require.NoError(t, err)

			req.URL.RawQuery = tt.queryParams.Encode()

			rr := httptest.NewRecorder()

			switch tt.path {
			case "/search/users":
				handler.SearchSimilarUsers(rr, req)
			case "/search/communities":
				handler.SearchSimilarCommunities(rr, req)
			}
		})
	}
}
