//go:build unit
// +build unit

package userclient

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"

	"quickflow/shared/models"
	pb "quickflow/shared/proto/user_service"
)

func TestDTOMappings(t *testing.T) {
	userID := uuid.New()
	lastSeen := time.Now()
	sessionID := uuid.New()
	expiry := time.Now().Add(24 * time.Hour)

	t.Run("MapUserToUserDTO", func(t *testing.T) {
		user := &models.User{
			Id:       userID,
			Username: "testuser",
			Password: "password",
			Salt:     "salt",
			LastSeen: lastSeen,
		}

		dto := MapUserToUserDTO(user)

		assert.Equal(t, userID.String(), dto.Id)
		assert.Equal(t, "testuser", dto.Username)
		assert.Equal(t, "password", dto.Password)
		assert.Equal(t, "salt", dto.Salt)
	})

	t.Run("MapUserDTOToUser", func(t *testing.T) {
		dto := &pb.User{
			Id:       userID.String(),
			Username: "testuser",
			Password: "password",
			Salt:     "salt",
			LastSeen: timestamppb.New(lastSeen),
		}

		user, err := MapUserDTOToUser(dto)

		assert.NoError(t, err)
		assert.Equal(t, userID, user.Id)
		assert.Equal(t, "testuser", user.Username)
		assert.Equal(t, "password", user.Password)
		assert.Equal(t, "salt", user.Salt)
	})

	t.Run("MapSessionToDTO", func(t *testing.T) {
		session := models.Session{
			SessionId:  sessionID,
			ExpireDate: expiry,
		}

		dto := MapSessionToDTO(session)

		assert.Equal(t, sessionID.String(), dto.Id)
	})

	t.Run("MapSignInToSignInDTO", func(t *testing.T) {
		signIn := &pb.SignIn{
			Username: "testuser",
			Password: "password",
		}

		dto := MapSignInToSignInDTO(signIn)

		assert.Equal(t, "testuser", dto.Username)
		assert.Equal(t, "password", dto.Password)
	})
}
