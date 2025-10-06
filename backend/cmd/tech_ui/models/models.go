package models

import (
	"encoding/json"
	"fmt"
	"time"
)

// Auth
type AuthForm struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type SignUpForm struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
	BirthDate string `json:"birth_date"`
	Sex       int    `json:"sex"`
}

type Session struct {
	ID        string    `json:"id"`
	ExpiresAt time.Time `json:"expires_at"`
}

type SignupResponse struct {
	ID string `json:"id"`
}

// Users
type PublicUserInfoOut struct {
	ID        string `json:"id"`
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
	Username  string `json:"username"`
	AvatarURL string `json:"avatar_url"`
	Online    bool   `json:"online"`
}

type ProfileInfo struct {
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
	Username  string `json:"username"`
	BirthDate string `json:"birth_date"`
	Sex       int    `json:"sex"`
	Bio       string `json:"bio"`
	AvatarURL string `json:"avatar_url"`
	CoverURL  string `json:"cover_url"`
}

type ProfileForm struct {
	ID       string      `json:"id"`
	Profile  ProfileInfo `json:"profile"`
	Online   bool        `json:"online"`
	LastSeen time.Time   `json:"last_seen"`
}

// Friends
type FriendsInfoOut struct {
	ID        string `json:"id"`
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
	Username  string `json:"username"`
	AvatarURL string `json:"avatar_url,omitempty"`
	IsOnline  bool   `json:"is_online"`
}

type FriendRequest struct {
	ReceiverID string `json:"receiver_id"`
}

type FriendRequestStatus struct {
	Status string `json:"status"`
}

// Posts
type CommentForm struct {
	Text  string   `json:"text"`
	Media []string `json:"media"`
	Files []string `json:"files"`
	Audio []string `json:"audio"`
}

type FileOut struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type PostOut struct {
	ID           string            `json:"id"`
	Author       PublicUserInfoOut `json:"author"`
	AuthorType   string            `json:"author_type"`
	Text         string            `json:"text"`
	Media        []FileOut         `json:"media"`
	Files        []FileOut         `json:"files"`
	Audio        []FileOut         `json:"audio"`
	IsLiked      bool              `json:"is_liked"`
	IsRepost     bool              `json:"is_repost"`
	LikeCount    int               `json:"like_count"`
	RepostCount  int               `json:"repost_count"`
	CommentCount int               `json:"comment_count"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

type UpdatePostForm struct {
	Text  string   `json:"text"`
	Media []string `json:"media"`
	Files []string `json:"files"`
	Audio []string `json:"audio"`
}

// Comments
type CommentOut struct {
	ID        string            `json:"id"`
	PostID    string            `json:"post_id"`
	Author    PublicUserInfoOut `json:"author"`
	Text      string            `json:"text"`
	Media     []FileOut         `json:"media"`
	Files     []FileOut         `json:"files"`
	Audio     []FileOut         `json:"audio"`
	IsLiked   bool              `json:"is_liked"`
	LikeCount int               `json:"like_count"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

// Likes
type LikeForm struct {
	TargetID   string `json:"target_id"`
	TargetType string `json:"target_type"`
	UserID     string `json:"user_id"`
}

// Communities
type CommunityInfo struct {
	Name        string `json:"name"`
	Nickname    string `json:"nickname"`
	Description string `json:"description"`
	AvatarURL   string `json:"avatar_url"`
	CoverURL    string `json:"cover_url"`
}

type CommunityForm struct {
	ID        string            `json:"id"`
	Community CommunityInfo     `json:"community"`
	Owner     PublicUserInfoOut `json:"owner"`
	Role      string            `json:"role"`
	CreatedAt time.Time         `json:"created_at"`
}

type CommunityMemberOut struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	CommunityID string    `json:"community_id"`
	Firstname   string    `json:"firstname"`
	Lastname    string    `json:"lastname"`
	Username    string    `json:"username"`
	AvatarURL   string    `json:"avatar_url"`
	Role        string    `json:"role"`
	Online      bool      `json:"online"`
	JoinedAt    time.Time `json:"joined_at"`
}

// Chats
type MessageOut struct {
	ID        string            `json:"id"`
	ChatID    string            `json:"chat_id"`
	Sender    PublicUserInfoOut `json:"sender"`
	Text      string            `json:"text"`
	Media     []FileOut         `json:"media,omitempty"`
	Files     []FileOut         `json:"files,omitempty"`
	Audio     []FileOut         `json:"audio,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

type ChatOut struct {
	ID              string     `json:"id"`
	Name            string     `json:"name"`
	Username        string     `json:"username"`
	AvatarURL       string     `json:"avatar_url"`
	Type            string     `json:"type"`
	Online          bool       `json:"online"`
	LastMessage     MessageOut `json:"last_message"`
	UnreadMessages  int        `json:"unread_messages"`
	LastReadByMe    time.Time  `json:"last_read_by_me"`
	LastReadByOther time.Time  `json:"last_read_by_other"`
	LastSeen        time.Time  `json:"last_seen"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type MessageForm struct {
	ChatID     string   `json:"chat_id"`
	ReceiverID string   `json:"receiver_id"`
	Text       string   `json:"text"`
	Media      []string `json:"media"`
	Files      []string `json:"files"`
	Audio      []string `json:"audio"`
}

// Files
type FileUploadResponse struct {
	Media []string `json:"media"`
	Files []string `json:"files"`
	Audio []string `json:"audio"`
}

// Error
type ErrorForm struct {
	ErrorCode string `json:"error_code"`
	Message   string `json:"message"`
}

func (e *ErrorForm) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode, e.Message)
}

// WebSocket
type WSMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type WSMessagePayload struct {
	Text       string   `json:"text"`
	ChatID     string   `json:"chat_id"`
	ReceiverID string   `json:"receiver_id"`
	Media      []string `json:"media"`
	Audio      []string `json:"audio"`
	File       []string `json:"file"`
}

type WSMessageResponse struct {
	ID        string    `json:"id"`
	Text      string    `json:"text"`
	SenderID  string    `json:"sender_id"`
	ChatID    string    `json:"chat_id"`
	CreatedAt time.Time `json:"created_at"`
	User      struct {
		ID       string `json:"id"`
		Username string `json:"username"`
		Avatar   string `json:"avatar"`
	} `json:"user"`
}

type WSMessageRead struct {
	ChatID    string `json:"chat_id"`
	MessageID string `json:"message_id"`
}

type WSFriendRequest struct {
	ID       string    `json:"id"`
	Username string    `json:"username"`
	Avatar   string    `json:"avatar"`
	LastSeen time.Time `json:"last_seen"`
}
