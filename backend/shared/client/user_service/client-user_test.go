//go:build unit
// +build unit

package userclient

import (
	"context"
	"errors"
	"quickflow/shared/proto/user_service/mocks"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"

	shared_models "quickflow/shared/models"
	pb "quickflow/shared/proto/user_service"
)

func TestNewUserClient(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConn := &grpc.ClientConn{}
	client := NewUserClient(mockConn)

	assert.NotNil(t, client)
	assert.Equal(t, mockConn, client.conn)
	assert.NotNil(t, client.client)
}

func TestClient_CreateUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockUserServiceClient(ctrl)
	client := &Client{client: mockClient}

	ctx := context.Background()
	userID := uuid.New()
	now := time.Now()

	tests := []struct {
		name        string
		setup       func()
		user        shared_models.User
		profile     shared_models.Profile
		expectedID  uuid.UUID
		expectedSes shared_models.Session
		expectError bool
	}{
		{
			name: "success",
			setup: func() {
				mockClient.EXPECT().SignUp(ctx, gomock.Any()).Return(&pb.SignUpResponse{
					Session: &pb.Session{
						Id:     uuid.New().String(),
						Expiry: timestamppb.New(now.Add(time.Hour)),
					},
				}, nil)
			},
			user: shared_models.User{
				Id:       userID,
				Username: "testuser",
				Password: "pass",
				Salt:     "salt",
				LastSeen: now,
			},
			profile:     shared_models.Profile{},
			expectedID:  userID,
			expectedSes: shared_models.Session{ExpireDate: now.Add(time.Hour)},
		},
		{
			name: "grpc error",
			setup: func() {
				mockClient.EXPECT().SignUp(ctx, gomock.Any()).Return(nil, errors.New("grpc error"))
			},
			user:        shared_models.User{Id: userID},
			profile:     shared_models.Profile{},
			expectError: true,
		},
		{
			name: "invalid session id",
			setup: func() {
				mockClient.EXPECT().SignUp(ctx, gomock.Any()).Return(&pb.SignUpResponse{
					Session: &pb.Session{
						Id:     "invalid-uuid",
						Expiry: timestamppb.New(now),
					},
				}, nil)
			},
			user:        shared_models.User{Id: userID},
			profile:     shared_models.Profile{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			id, session, err := client.CreateUser(ctx, tt.user, tt.profile)

			if tt.expectError {
				require.Error(t, err)
				assert.Equal(t, uuid.Nil, id)
				assert.Equal(t, shared_models.Session{}, session)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedID, id)
				assert.NotEqual(t, uuid.Nil, session.SessionId)
			}
		})
	}
}

func TestClient_AuthUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockUserServiceClient(ctrl)
	client := &Client{client: mockClient}

	ctx := context.Background()
	now := time.Now()

	tests := []struct {
		name        string
		setup       func()
		authData    shared_models.LoginData
		expected    shared_models.Session
		expectError bool
	}{
		{
			name: "success",
			setup: func() {
				mockClient.EXPECT().SignIn(ctx, gomock.Any()).Return(&pb.SignInResponse{
					Session: &pb.Session{
						Id:     uuid.New().String(),
						Expiry: timestamppb.New(now),
					},
				}, nil)
			},
			authData: shared_models.LoginData{
				Username: "user",
				Password: "pass",
			},
			expected: shared_models.Session{
				ExpireDate: now,
			},
		},
		{
			name: "grpc error",
			setup: func() {
				mockClient.EXPECT().SignIn(ctx, gomock.Any()).Return(nil, errors.New("grpc error"))
			},
			authData: shared_models.LoginData{
				Username: "user",
				Password: "pass",
			},
			expectError: true,
		},
		{
			name: "invalid session id",
			setup: func() {
				mockClient.EXPECT().SignIn(ctx, gomock.Any()).Return(&pb.SignInResponse{
					Session: &pb.Session{
						Id:     "invalid-uuid",
						Expiry: timestamppb.New(now),
					},
				}, nil)
			},
			authData:    shared_models.LoginData{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			session, err := client.AuthUser(ctx, tt.authData)

			if tt.expectError {
				require.Error(t, err)
				assert.Equal(t, shared_models.Session{}, session)
			} else {
				require.NoError(t, err)
				assert.NotEqual(t, uuid.Nil, session.SessionId)
			}
		})
	}
}

func TestClient_DeleteUserSession(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockUserServiceClient(ctrl)
	client := &Client{client: mockClient}

	ctx := context.Background()
	sessionID := uuid.New().String()

	tests := []struct {
		name        string
		setup       func()
		session     string
		expectError bool
	}{
		{
			name: "success",
			setup: func() {
				mockClient.EXPECT().SignOut(ctx, gomock.Any()).Return(&pb.SignOutResponse{}, nil)
			},
			session: sessionID,
		},
		{
			name: "grpc error",
			setup: func() {
				mockClient.EXPECT().SignOut(ctx, gomock.Any()).Return(nil, errors.New("grpc error"))
			},
			session:     sessionID,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			err := client.DeleteUserSession(ctx, tt.session)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestClient_GetUserByUsername(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockUserServiceClient(ctrl)
	client := &Client{client: mockClient}

	ctx := context.Background()
	userID := uuid.New()
	now := time.Now()
	username := "testuser"

	tests := []struct {
		name        string
		setup       func()
		username    string
		expected    shared_models.User
		expectError bool
	}{
		{
			name: "success",
			setup: func() {
				mockClient.EXPECT().GetUserByUsername(ctx, &pb.GetUserByUsernameRequest{
					Username: username,
				}).Return(&pb.GetUserByUsernameResponse{
					User: &pb.User{
						Id:       userID.String(),
						Username: username,
						LastSeen: timestamppb.New(now),
					},
				}, nil)
			},
			username: username,
			expected: shared_models.User{
				Id:       userID,
				Username: username,
				LastSeen: now,
			},
		},
		{
			name: "grpc error",
			setup: func() {
				mockClient.EXPECT().GetUserByUsername(ctx, gomock.Any()).Return(nil, errors.New("grpc error"))
			},
			username:    username,
			expectError: true,
		},
		{
			name: "invalid user id",
			setup: func() {
				mockClient.EXPECT().GetUserByUsername(ctx, gomock.Any()).Return(&pb.GetUserByUsernameResponse{
					User: &pb.User{
						Id: "invalid-uuid",
					},
				}, nil)
			},
			username:    username,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			user, err := client.GetUserByUsername(ctx, tt.username)

			if tt.expectError {
				require.Error(t, err)
				assert.Equal(t, shared_models.User{}, user)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected.Id, user.Id)
			}
		})
	}
}

func TestClient_GetUserById(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockUserServiceClient(ctrl)
	client := &Client{client: mockClient}

	ctx := context.Background()
	userID := uuid.New()
	now := time.Now()

	tests := []struct {
		name        string
		setup       func()
		userID      uuid.UUID
		expected    shared_models.User
		expectError bool
	}{
		{
			name: "success",
			setup: func() {
				mockClient.EXPECT().GetUserById(ctx, &pb.GetUserByIdRequest{
					Id: userID.String(),
				}).Return(&pb.GetUserByIdResponse{
					User: &pb.User{
						Id:       userID.String(),
						Username: "testuser",
						LastSeen: timestamppb.New(now),
					},
				}, nil)
			},
			userID: userID,
			expected: shared_models.User{
				Id:       userID,
				Username: "testuser",
				LastSeen: now,
			},
		},
		{
			name: "grpc error",
			setup: func() {
				mockClient.EXPECT().GetUserById(ctx, gomock.Any()).Return(nil, errors.New("grpc error"))
			},
			userID:      userID,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			user, err := client.GetUserById(ctx, tt.userID)

			if tt.expectError {
				require.Error(t, err)
				assert.Equal(t, shared_models.User{}, user)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected.Id, user.Id)
			}
		})
	}
}

func TestClient_LookupUserSession(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockUserServiceClient(ctrl)
	client := &Client{client: mockClient}

	ctx := context.Background()
	sessionID := uuid.New()
	userID := uuid.New()
	username := "testuser"

	tests := []struct {
		name        string
		setup       func()
		session     shared_models.Session
		expected    shared_models.User
		expectError bool
	}{
		{
			name: "success",
			setup: func() {
				mockClient.EXPECT().LookupUserSession(ctx, &pb.LookupUserSessionRequest{
					SessionId: sessionID.String(),
				}).Return(&pb.LookupUserSessionResponse{
					UserId:   userID.String(),
					Username: username,
				}, nil)
			},
			session: shared_models.Session{
				SessionId: sessionID,
			},
			expected: shared_models.User{
				Id:       userID,
				Username: username,
			},
		},
		{
			name: "grpc error",
			setup: func() {
				mockClient.EXPECT().LookupUserSession(ctx, gomock.Any()).Return(nil, errors.New("grpc error"))
			},
			session:     shared_models.Session{SessionId: sessionID},
			expectError: true,
		},
		{
			name: "invalid user id",
			setup: func() {
				mockClient.EXPECT().LookupUserSession(ctx, gomock.Any()).Return(&pb.LookupUserSessionResponse{
					UserId: "invalid-uuid",
				}, nil)
			},
			session:     shared_models.Session{SessionId: sessionID},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			user, err := client.LookupUserSession(ctx, tt.session)

			if tt.expectError {
				require.Error(t, err)
				assert.Equal(t, shared_models.User{}, user)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, user)
			}
		})
	}
}

func TestClient_SearchSimilarUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockUserServiceClient(ctrl)
	client := &Client{client: mockClient}

	ctx := context.Background()
	userID := uuid.New()
	searchTerm := "test"
	count := uint(5)

	tests := []struct {
		name        string
		setup       func()
		toSearch    string
		usersCount  uint
		expected    []shared_models.PublicUserInfo
		expectError bool
	}{
		{
			name: "success with results",
			setup: func() {
				mockClient.EXPECT().SearchSimilarUser(ctx, &pb.SearchSimilarUserRequest{
					ToSearch: searchTerm,
					NumUsers: int32(count),
				}).Return(&pb.SearchSimilarUserResponse{
					UsersInfo: []*pb.PublicUserInfo{
						{
							Id:       userID.String(),
							Username: "testuser1",
						},
						{
							Id:       uuid.New().String(),
							Username: "testuser2",
						},
					},
				}, nil)
			},
			toSearch:   searchTerm,
			usersCount: count,
			expected: []shared_models.PublicUserInfo{
				{
					Id:       userID,
					Username: "testuser1",
				},
				{
					Id:       uuid.Nil, // will be overwritten
					Username: "testuser2",
				},
			},
		},
		{
			name: "empty results",
			setup: func() {
				mockClient.EXPECT().SearchSimilarUser(ctx, gomock.Any()).Return(&pb.SearchSimilarUserResponse{
					UsersInfo: []*pb.PublicUserInfo{},
				}, nil)
			},
			toSearch:   searchTerm,
			usersCount: count,
			expected:   []shared_models.PublicUserInfo{},
		},
		{
			name: "grpc error",
			setup: func() {
				mockClient.EXPECT().SearchSimilarUser(ctx, gomock.Any()).Return(nil, errors.New("grpc error"))
			},
			toSearch:    searchTerm,
			usersCount:  count,
			expectError: true,
		},
		{
			name: "invalid user id in response",
			setup: func() {
				mockClient.EXPECT().SearchSimilarUser(ctx, gomock.Any()).Return(&pb.SearchSimilarUserResponse{
					UsersInfo: []*pb.PublicUserInfo{
						{
							Id:       "invalid-uuid",
							Username: "testuser",
						},
					},
				}, nil)
			},
			toSearch:    searchTerm,
			usersCount:  count,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			users, err := client.SearchSimilarUser(ctx, tt.toSearch, tt.usersCount)

			if tt.expectError {
				require.Error(t, err)
				assert.Nil(t, users)
			} else {
				require.NoError(t, err)
				require.Equal(t, len(tt.expected), len(users))

				for i, expected := range tt.expected {
					assert.Equal(t, expected.Username, users[i].Username)
					if expected.Id != uuid.Nil {
						assert.Equal(t, expected.Id, users[i].Id)
					} else {
						assert.NotEqual(t, uuid.Nil, users[i].Id)
					}
				}
			}
		})
	}
}
