//go:build integration
// +build integration

package usecase_test

import (
	"context"
	"database/sql"
	"github.com/ozontech/allure-go/pkg/framework/asserts_wrapper/require"
	"log"
	"quickflow/config/test"
	postgres "quickflow/friends_service/internal/repository"
	getEnv "quickflow/utils/get-env"
	"testing"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/ozontech/allure-go/pkg/allure"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"

	"quickflow/friends_service/internal/usecase"
	"quickflow/shared/models"
)

type FriendsServiceIntegrationTestSuite struct {
	suite.Suite
	db         *sql.DB
	service    *usecase.FriendsService
	repository usecase.FriendsRepository
	testUser1  uuid.UUID
	testUser2  uuid.UUID
	testUser3  uuid.UUID
	testUser4  uuid.UUID
}

func (s *FriendsServiceIntegrationTestSuite) BeforeAll(t provider.T) {
	t.WithNewStep("Setup test environment", func(ctx provider.StepCtx) {
		// Setup PostgreSQL
		connString := getEnv.GetEnv(test.TestDbConnStringEnvVar, test.DefaultDatabaseTestUrl)
		require.NotEmpty(t, connString, "Connection string must not be empty")

		ctx.WithNewAttachment("connection_string", allure.Text, []byte(connString))

		var err error
		s.db, err = sql.Open("pgx", connString)
		if err != nil {
			log.Fatalf("Failed to connect to test database: %v", err)
		}

		err = s.db.Ping()
		if err != nil {
			log.Fatalf("Failed to ping database: %v", err)
		}

		// Create test tables
		err = s.createTestTables()
		if err != nil {
			log.Fatalf("Failed to create test tables: %v", err)
		}

		// Generate test users
		s.testUser1 = uuid.New()
		s.testUser2 = uuid.New()
		s.testUser3 = uuid.New()
		s.testUser4 = uuid.New()

		// Insert test users and profiles
		err = s.insertTestUsers()
		if err != nil {
			log.Fatalf("Failed to insert test users: %v", err)
		}

		// Initialize repository and service
		s.repository = postgres.NewPostgresFriendsRepository(s.db)
		s.service = usecase.NewFriendsService(s.repository)
	})
}

func (s *FriendsServiceIntegrationTestSuite) AfterAll(t provider.T) {
	t.WithNewStep("Cleanup test environment", func(ctx provider.StepCtx) {
		if s.db != nil {
			s.cleanupTestData()
			s.db.Close()
		}
	})
}

func (s *FriendsServiceIntegrationTestSuite) BeforeEach(t provider.T) {
	t.WithNewStep("Cleanup before each test", func(ctx provider.StepCtx) {
		s.cleanupFriendshipData()
		s.cleanupTestData()
		s.insertTestUsers()
	})
	t.Epic("Integration")
}

func (s *FriendsServiceIntegrationTestSuite) TestGetFriendsInfo_AllFriends(t provider.T) {
	t.WithNewStep("Test get all friends info", func(ctx provider.StepCtx) {
		// Setup: Create friend relationships
		err := s.createFriendRelationship(s.testUser1.String(), s.testUser2.String(), models.RelationFriend)
		if err != nil {
			log.Fatalf("Failed to create friend relationship: %v", err)
		}

		err = s.createFriendRelationship(s.testUser1.String(), s.testUser3.String(), models.RelationFriend)
		if err != nil {
			log.Fatalf("Failed to create friend relationship: %v", err)
		}

		// Execute
		friends, count, err := s.service.GetFriendsInfo(
			context.Background(),
			s.testUser1.String(),
			"10",
			"0",
			"all",
		)

		// Verify
		if err != nil {
			log.Fatalf("GetFriendsInfo failed: %v", err)
		}

		if count != 2 {
			ctx.Errorf("Expected 2 friends, got %d", count)
		}

		if len(friends) != 2 {
			ctx.Errorf("Expected 2 friend info entries, got %d", len(friends))
		}

		// Verify friend info content
		friendIDs := make(map[string]bool)
		for _, friend := range friends {
			friendIDs[friend.Id.String()] = true
		}

		if !friendIDs[s.testUser2.String()] || !friendIDs[s.testUser3.String()] {
			ctx.Errorf("Expected friends %s and %s, got %v", s.testUser2.String(), s.testUser3.String(), friendIDs)
		}
	})
}

func (s *FriendsServiceIntegrationTestSuite) TestGetFriendsInfo_InvalidLimit(t provider.T) {
	t.WithNewStep("Test get friends info with invalid limit", func(ctx provider.StepCtx) {
		// Execute
		friends, count, err := s.service.GetFriendsInfo(
			context.Background(),
			s.testUser1.String(),
			"invalid",
			"0",
			"all",
		)

		// Verify
		if err == nil {
			ctx.Error("Expected error for invalid limit")
		}

		if count != 0 {
			ctx.Errorf("Expected count 0, got %d", count)
		}

		if friends != nil {
			ctx.Errorf("Expected nil friends, got %v", friends)
		}
	})
}

func (s *FriendsServiceIntegrationTestSuite) TestGetFriendsInfo_InvalidOffset(t provider.T) {
	t.WithNewStep("Test get friends info with invalid offset", func(ctx provider.StepCtx) {
		// Execute
		friends, count, err := s.service.GetFriendsInfo(
			context.Background(),
			s.testUser1.String(),
			"10",
			"invalid",
			"all",
		)

		// Verify
		if err == nil {
			ctx.Error("Expected error for invalid offset")
		}

		if count != 0 {
			ctx.Errorf("Expected count 0, got %d", count)
		}

		if friends != nil {
			ctx.Errorf("Expected nil friends, got %v", friends)
		}
	})
}

func (s *FriendsServiceIntegrationTestSuite) TestSendFriendRequest_Success(t provider.T) {
	t.WithNewStep("Test send friend request successfully", func(ctx provider.StepCtx) {
		// Execute
		err := s.service.SendFriendRequest(
			context.Background(),
			s.testUser1.String(),
			s.testUser2.String(),
		)

		// Verify
		if err != nil {
			log.Fatalf("SendFriendRequest failed: %v", err)
		}

		// Verify request was created
		exists, err := s.checkFriendshipExists(s.testUser1.String(), s.testUser2.String())
		if err != nil {
			log.Fatalf("Failed to check friendship existence: %v", err)
		}

		if !exists {
			ctx.Error("Friend request was not created")
		}

		// Verify correct status
		status, err := s.getFriendshipStatus(s.testUser1.String(), s.testUser2.String())
		if err != nil {
			log.Fatalf("Failed to get friendship status: %v", err)
		}

		if status != models.RelationFollowing && status != models.RelationFollowedBy {
			ctx.Errorf("Expected status Following or FollowedBy, got %s", status)
		}
	})
}

func (s *FriendsServiceIntegrationTestSuite) TestIsExistsFriendRequest_Exists(t provider.T) {
	t.WithNewStep("Test check if friend request exists", func(ctx provider.StepCtx) {
		// Setup: Create friend request
		err := s.createFriendRelationship(s.testUser1.String(), s.testUser2.String(), models.RelationFollowing)
		if err != nil {
			log.Fatalf("Failed to create friend relationship: %v", err)
		}

		// Execute
		exists, err := s.service.IsExistsFriendRequest(
			context.Background(),
			s.testUser1.String(),
			s.testUser2.String(),
		)

		// Verify
		if err != nil {
			log.Fatalf("IsExistsFriendRequest failed: %v", err)
		}

		if !exists {
			ctx.Error("Expected friend request to exist")
		}
	})
}

func (s *FriendsServiceIntegrationTestSuite) TestIsExistsFriendRequest_NotExists(t provider.T) {
	t.WithNewStep("Test check if friend request does not exist", func(ctx provider.StepCtx) {
		// Execute
		exists, err := s.service.IsExistsFriendRequest(
			context.Background(),
			s.testUser1.String(),
			s.testUser2.String(),
		)

		// Verify
		if err != nil {
			log.Fatalf("IsExistsFriendRequest failed: %v", err)
		}

		if exists {
			ctx.Error("Expected friend request to not exist")
		}
	})
}

func (s *FriendsServiceIntegrationTestSuite) TestAcceptFriendRequest_Success(t provider.T) {
	t.WithNewStep("Test accept friend request", func(ctx provider.StepCtx) {
		// Setup: Create incoming request
		err := s.createFriendRelationship(s.testUser2.String(), s.testUser1.String(), models.RelationFollowing)
		if err != nil {
			log.Fatalf("Failed to create friend relationship: %v", err)
		}

		// Execute
		err = s.service.AcceptFriendRequest(
			context.Background(),
			s.testUser2.String(),
			s.testUser1.String(),
		)

		// Verify
		if err != nil {
			log.Fatalf("AcceptFriendRequest failed: %v", err)
		}

		// Verify status changed to friend
		status, err := s.getFriendshipStatus(s.testUser1.String(), s.testUser2.String())
		if err != nil {
			log.Fatalf("Failed to get friendship status: %v", err)
		}

		if status != models.RelationFriend {
			ctx.Errorf("Expected status %s, got %s", models.RelationFriend, status)
		}
	})
}

func (s *FriendsServiceIntegrationTestSuite) TestDeleteFriend_Success(t provider.T) {
	t.WithNewStep("Test delete friend", func(ctx provider.StepCtx) {
		// Setup: Create friend relationship
		err := s.createFriendRelationship(s.testUser1.String(), s.testUser2.String(), models.RelationFriend)
		if err != nil {
			log.Fatalf("Failed to create friend relationship: %v", err)
		}

		// Execute
		err = s.service.DeleteFriend(
			context.Background(),
			s.testUser1.String(),
			s.testUser2.String(),
		)

		// Verify
		if err != nil {
			log.Fatalf("DeleteFriend failed: %v", err)
		}

		// Verify status changed to following/followed_by
		status, err := s.getFriendshipStatus(s.testUser1.String(), s.testUser2.String())
		if err != nil {
			log.Fatalf("Failed to get friendship status: %v", err)
		}

		if status != models.RelationFollowing && status != models.RelationFollowedBy {
			ctx.Errorf("Expected status to be Following or FollowedBy, got %s", status)
		}
	})
}

func (s *FriendsServiceIntegrationTestSuite) TestUnfollow_Success(t provider.T) {
	t.WithNewStep("Test unfollow user", func(ctx provider.StepCtx) {
		// Setup: Create following relationship
		err := s.createFriendRelationship(s.testUser1.String(), s.testUser2.String(), models.RelationFollowing)
		if err != nil {
			log.Fatalf("Failed to create friend relationship: %v", err)
		}

		// Execute
		err = s.service.Unfollow(
			context.Background(),
			s.testUser1.String(),
			s.testUser2.String(),
		)

		// Verify
		if err != nil {
			log.Fatalf("Unfollow failed: %v", err)
		}

		// Verify relationship was deleted
		exists, err := s.checkFriendshipExists(s.testUser1.String(), s.testUser2.String())
		if err != nil {
			log.Fatalf("Failed to check friendship existence: %v", err)
		}

		if exists {
			ctx.Error("Relationship should have been deleted")
		}
	})
}

func (s *FriendsServiceIntegrationTestSuite) TestGetUserRelation_Friend(t provider.T) {
	t.WithNewStep("Test get user relation - friend", func(ctx provider.StepCtx) {
		// Setup: Create friend relationship
		err := s.createFriendRelationship(s.testUser1.String(), s.testUser2.String(), models.RelationFriend)
		if err != nil {
			log.Fatalf("Failed to create friend relationship: %v", err)
		}

		// Execute
		relation, err := s.service.GetUserRelation(
			context.Background(),
			s.testUser1,
			s.testUser2,
		)

		// Verify
		if err != nil {
			log.Fatalf("GetUserRelation failed: %v", err)
		}

		if relation != models.RelationFriend {
			ctx.Errorf("Expected relation Friend, got %s", relation)
		}
	})
}

func (s *FriendsServiceIntegrationTestSuite) TestGetUserRelation_Stranger(t provider.T) {
	t.WithNewStep("Test get user relation - stranger", func(ctx provider.StepCtx) {
		// Execute
		relation, err := s.service.GetUserRelation(
			context.Background(),
			s.testUser1,
			s.testUser2,
		)

		// Verify
		if err != nil {
			log.Fatalf("GetUserRelation failed: %v", err)
		}

		if relation != models.RelationStranger {
			ctx.Errorf("Expected relation Stranger, got %s", relation)
		}
	})
}

func (s *FriendsServiceIntegrationTestSuite) TestGetUserRelation_Self(t provider.T) {
	t.WithNewStep("Test get user relation - self", func(ctx provider.StepCtx) {
		// Execute
		relation, err := s.service.GetUserRelation(
			context.Background(),
			s.testUser1,
			s.testUser1,
		)

		// Verify
		if err != nil {
			log.Fatalf("GetUserRelation failed: %v", err)
		}

		if relation != models.RelationSelf {
			ctx.Errorf("Expected relation Self, got %s", relation)
		}
	})
}

func (s *FriendsServiceIntegrationTestSuite) TestGetUserRelation_NilUser(t provider.T) {
	t.WithNewStep("Test get user relation with nil user", func(ctx provider.StepCtx) {
		// Execute
		relation, err := s.service.GetUserRelation(
			context.Background(),
			uuid.Nil,
			s.testUser2,
		)

		// Verify
		if err == nil {
			ctx.Error("Expected error for nil user")
		}

		if relation != models.RelationStranger {
			ctx.Errorf("Expected relation Stranger, got %s", relation)
		}
	})
}

func (s *FriendsServiceIntegrationTestSuite) TestMarkRead_Success(t provider.T) {
	t.WithNewStep("Test mark friend request as read", func(ctx provider.StepCtx) {
		// Setup: Create unread friend request
		err := s.createUnreadFriendRequest(s.testUser2.String(), s.testUser1.String(), models.RelationFollowing)
		if err != nil {
			log.Fatalf("Failed to create unread friend request: %v", err)
		}

		// Execute
		err = s.service.MarkRead(
			context.Background(),
			s.testUser1.String(),
			s.testUser2.String(),
		)

		// Verify
		if err != nil {
			log.Fatalf("MarkRead failed: %v", err)
		}

		// Verify request is marked as read
		isRead, err := s.checkRequestIsRead(s.testUser1.String(), s.testUser2.String())
		if err != nil {
			log.Fatalf("Failed to check read status: %v", err)
		}

		if !isRead {
			ctx.Error("Request should be marked as read")
		}
	})
}

// Helper methods
func (s *FriendsServiceIntegrationTestSuite) createTestTables() error {
	// Drop existing tables
	return nil
}

func (s *FriendsServiceIntegrationTestSuite) insertTestUsers() error {
	// Insert test users
	users := []struct {
		id       string
		username string
	}{
		{s.testUser1.String(), "testuser" + uuid.New().String()},
		{s.testUser2.String(), "testuser2" + uuid.New().String()},
		{s.testUser3.String(), "testuser3" + uuid.New().String()},
		{s.testUser4.String(), "testuser4" + uuid.New().String()},
	}

	for _, user := range users {
		_, err := s.db.Exec(`INSERT INTO "user" (id, username, psw_hash, salt) VALUES ($1, $2, 'some_hash', 'salt')`, user.id, user.username)
		if err != nil {
			return err
		}

		_, err = s.db.Exec(`INSERT INTO profile (id, firstname, lastname) VALUES ($1, $2, $3)`,
			user.id, "Test", "User")
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *FriendsServiceIntegrationTestSuite) createFriendRelationship(user1, user2 string, status models.UserRelation) error {
	var u1, u2 string
	if user1 < user2 {
		u1, u2 = user1, user2
	} else {
		u1, u2 = user2, user1
	}

	_, err := s.db.Exec(
		`INSERT INTO friendship (user1_id, user2_id, status, is_read) VALUES ($1, $2, $3, $4) 
		 ON CONFLICT (user1_id, user2_id) DO UPDATE SET status = $3, is_read = $4`,
		u1, u2, status, true,
	)
	return err
}

func (s *FriendsServiceIntegrationTestSuite) createUnreadFriendRequest(user1, user2 string, status models.UserRelation) error {
	var u1, u2 string
	if user1 < user2 {
		u1, u2 = user1, user2
	} else {
		u1, u2 = user2, user1
	}

	_, err := s.db.Exec(
		`INSERT INTO friendship (user1_id, user2_id, status, is_read) VALUES ($1, $2, $3, $4)`,
		u1, u2, status, false,
	)
	return err
}

func (s *FriendsServiceIntegrationTestSuite) checkFriendshipExists(user1, user2 string) (bool, error) {
	var u1, u2 string
	if user1 < user2 {
		u1, u2 = user1, user2
	} else {
		u1, u2 = user2, user1
	}

	var count int
	err := s.db.QueryRow(
		`SELECT COUNT(*) FROM friendship WHERE user1_id = $1 AND user2_id = $2`,
		u1, u2,
	).Scan(&count)

	return count > 0, err
}

func (s *FriendsServiceIntegrationTestSuite) getFriendshipStatus(user1, user2 string) (models.UserRelation, error) {
	var u1, u2 string
	if user1 < user2 {
		u1, u2 = user1, user2
	} else {
		u1, u2 = user2, user1
	}

	var status models.UserRelation
	err := s.db.QueryRow(
		`SELECT status FROM friendship WHERE user1_id = $1 AND user2_id = $2`,
		u1, u2,
	).Scan(&status)

	return status, err
}

func (s *FriendsServiceIntegrationTestSuite) checkRequestIsRead(user1, user2 string) (bool, error) {
	var u1, u2 string
	if user1 < user2 {
		u1, u2 = user1, user2
	} else {
		u1, u2 = user2, user1
	}

	var isRead bool
	err := s.db.QueryRow(
		`SELECT is_read FROM friendship WHERE user1_id = $1 AND user2_id = $2`,
		u1, u2,
	).Scan(&isRead)

	return isRead, err
}

func (s *FriendsServiceIntegrationTestSuite) cleanupFriendshipData() error {
	_, err := s.db.Exec(`DELETE FROM friendship`)
	return err
}

func (s *FriendsServiceIntegrationTestSuite) cleanupTestData() error {
	queries := []string{
		`DELETE FROM friendship`,
		`DELETE FROM education`,
		`DELETE FROM profile`,
		`DELETE FROM "user"`,
		`DELETE FROM faculty`,
		`DELETE FROM university`,
	}

	for _, query := range queries {
		_, err := s.db.Exec(query)
		if err != nil {
			return err
		}
	}
	return nil
}

func TestFriendsServiceIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	suite.RunSuite(t, new(FriendsServiceIntegrationTestSuite))
}
