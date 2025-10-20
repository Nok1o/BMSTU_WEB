package interfaces

import (
	"context"
	"time"

	"github.com/google/uuid"

	"quickflow/shared/models"
)

type AuthUseCase interface {
	CreateUser(ctx context.Context, user models.User, profile models.Profile) (uuid.UUID, models.Session, error)
	AuthUser(ctx context.Context, authData models.LoginData) (models.Session, error)
	GetUserByUsername(ctx context.Context, username string) (models.User, error)
	LookupUserSession(ctx context.Context, session models.Session) (models.User, error)
	DeleteUserSession(ctx context.Context, session string) error
}

type ChatUseCase interface {
	GetChatParticipants(ctx context.Context, chatId uuid.UUID) ([]uuid.UUID, error)
	GetUserChats(ctx context.Context, userId uuid.UUID, numChats int, ts time.Time) ([]models.Chat, error)
	GetPrivateChat(ctx context.Context, userId1, userId2 uuid.UUID) (*models.Chat, error)
	DeleteChat(ctx context.Context, chatId uuid.UUID) error
	GetChat(ctx context.Context, chatId uuid.UUID) (*models.Chat, error)
	JoinChat(ctx context.Context, chatId, userId uuid.UUID) error
	LeaveChat(ctx context.Context, chatId, userId uuid.UUID) error
	GetNumUnreadChats(ctx context.Context, userId uuid.UUID) (int, error)
}

type CommentService interface {
	FetchCommentsForPost(ctx context.Context, postId uuid.UUID, numComments int, timestamp time.Time) ([]models.Comment, error)
	AddComment(ctx context.Context, comment models.Comment) (*models.Comment, error)
	DeleteComment(ctx context.Context, userId uuid.UUID, commentId uuid.UUID) error
	UpdateComment(ctx context.Context, update models.CommentUpdate, userId uuid.UUID) (*models.Comment, error)
	LikeComment(ctx context.Context, commentId uuid.UUID, userId uuid.UUID) error
	UnlikeComment(ctx context.Context, commentId uuid.UUID, userId uuid.UUID) error
	GetComment(ctx context.Context, commentId uuid.UUID, userId uuid.UUID) (*models.Comment, error)
	GetLastPostComment(ctx context.Context, postId uuid.UUID) (*models.Comment, error)
	GetCommentLikes(ctx context.Context, commentId uuid.UUID, numLikes, offset int) ([]models.Like, error)
}

type ProfileUseCase interface {
	GetProfileByUsername(ctx context.Context, username string) (models.Profile, error)
	UpdateProfile(ctx context.Context, newProfile models.Profile) (*models.Profile, error)
	GetPublicUserInfo(ctx context.Context, userId uuid.UUID) (models.PublicUserInfo, error)
	GetPublicUsersInfo(ctx context.Context, userIds []uuid.UUID) ([]models.PublicUserInfo, error)
	UpdateLastSeen(ctx context.Context, userId uuid.UUID) error
}

type CommunityService interface {
	CreateCommunity(ctx context.Context, community *models.Community) (*models.Community, error)
	GetCommunityById(ctx context.Context, id uuid.UUID) (*models.Community, error)
	GetCommunityByName(ctx context.Context, name string) (*models.Community, error)
	GetCommunityMembers(ctx context.Context, communityId uuid.UUID, count int, ts time.Time) ([]*models.CommunityMember, error)
	IsCommunityMember(ctx context.Context, userId, communityId uuid.UUID) (bool, *models.CommunityRole, error)
	DeleteCommunity(ctx context.Context, communityId uuid.UUID, userId uuid.UUID) error
	UpdateCommunity(ctx context.Context, community *models.Community, userId uuid.UUID) (*models.Community, error)
	JoinCommunity(ctx context.Context, member *models.CommunityMember) error
	LeaveCommunity(ctx context.Context, userId, communityId uuid.UUID) error
	GetUserCommunities(ctx context.Context, userId uuid.UUID, count int, ts time.Time) ([]*models.Community, error)
	SearchSimilarCommunities(ctx context.Context, name string, count int) ([]*models.Community, error)
	ChangeUserRole(ctx context.Context, userId, communityId uuid.UUID, role models.CommunityRole, requester uuid.UUID) error
	GetControlledCommunities(ctx context.Context, userId uuid.UUID, count int, ts time.Time) ([]*models.Community, error)
}

type PostService interface {
	FetchFeed(ctx context.Context, numPosts int, timestamp time.Time, userId uuid.UUID) ([]models.Post, error)
	FetchRecommendations(ctx context.Context, numPosts int, timestamp time.Time, userId uuid.UUID) ([]models.Post, error)
	FetchCreatorPosts(ctx context.Context, creatorId uuid.UUID, requesterId uuid.UUID, numPosts int, timestamp time.Time) ([]models.Post, error)
	AddPost(ctx context.Context, post models.Post) (*models.Post, error)
	DeletePost(ctx context.Context, userId uuid.UUID, postId uuid.UUID) error
	UpdatePost(ctx context.Context, update models.PostUpdate, userId uuid.UUID) (*models.Post, error)
	LikePost(ctx context.Context, postId, userId uuid.UUID) error
	UnlikePost(ctx context.Context, postId, userId uuid.UUID) error
	GetPost(ctx context.Context, postId, userId uuid.UUID) (*models.Post, error)
	GetPostLikes(ctx context.Context, postId uuid.UUID, numLikes, offset int) ([]models.Like, error)
}

type FeedbackUseCase interface {
	SaveFeedback(ctx context.Context, feedback *models.Feedback) error
	GetAllFeedbackType(ctx context.Context, feedbackType models.FeedbackType, ts time.Time, count int) ([]models.Feedback, error)
}

type FileService interface {
	UploadManyFiles(ctx context.Context, files []*models.File) ([]string, error)
	DeleteFile(ctx context.Context, filename string) error
}

type FriendsUseCase interface {
	GetFriendsInfo(ctx context.Context, userID string, limit string, offset string, reqType string) ([]models.FriendInfo, int, error)
	SendFriendRequest(ctx context.Context, senderID string, receiverID string) error
	AcceptFriendRequest(ctx context.Context, senderID string, receiverID string) error
	Unfollow(ctx context.Context, userID string, friendID string) error
	DeleteFriend(ctx context.Context, user string, friend string) error
	// GetUserRelation IsExistsFriendRequest(ctx context.Context, senderID string, receiverID string) (bool, error)
	GetUserRelation(ctx context.Context, user1 uuid.UUID, user2 uuid.UUID) (models.UserRelation, error)
	MarkRead(ctx context.Context, userID string, friendID string) error
}

type IFriendsWSManager interface {
	NotifyFriendRequestSent(ctx context.Context, senderId, receiverId uuid.UUID) error
	NotifyFriendRequestAccepted(ctx context.Context, senderId, receiverId uuid.UUID) error
}

type WSLikeHandler interface {
	NotifyPostLiked(ctx context.Context, senderId, receiverId uuid.UUID, post *models.Post) error
	NotifyCommentLiked(ctx context.Context, senderId, receiverId uuid.UUID, comment *models.Comment) error
	NotifyPostCommented(ctx context.Context, senderId, receiverId uuid.UUID, post *models.Post, comment *models.Comment) error
}
type SearchUseCase interface {
	SearchSimilarUser(ctx context.Context, toSearch string, postsCount uint) ([]models.PublicUserInfo, error)
}

type StickerUseCase interface {
	AddStickerPack(ctx context.Context, stickerPack *models.StickerPack) (*models.StickerPack, error)
	GetStickerPack(ctx context.Context, packId uuid.UUID) (*models.StickerPack, error)
	GetStickerPackByName(ctx context.Context, packName string) (*models.StickerPack, error)
	GetStickerPacks(ctx context.Context, userId uuid.UUID, count, offset int) ([]*models.StickerPack, error)
	DeleteStickerPack(ctx context.Context, userId, packId uuid.UUID) error
}

type MessageService interface {
	GetMessageById(ctx context.Context, messageId uuid.UUID) (*models.Message, error)
	GetMessagesForChat(ctx context.Context, chatId uuid.UUID, numMessages int, timestamp time.Time, userId uuid.UUID) ([]*models.Message, error)
	SendMessage(ctx context.Context, message *models.Message, userId uuid.UUID) (*models.Message, error)
	DeleteMessage(ctx context.Context, messageId uuid.UUID) error
	GetLastReadTs(ctx context.Context, chatId uuid.UUID, userId uuid.UUID) (time.Time, error)
	UpdateLastReadTs(ctx context.Context, chatId uuid.UUID, userId uuid.UUID, timestamp time.Time, userAuthId uuid.UUID) error
	GetNumUnreadMessages(ctx context.Context, chatId, userId uuid.UUID) (int, error)
}
