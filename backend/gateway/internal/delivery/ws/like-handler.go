package ws

import (
    "context"
    "encoding/json"
    "fmt"

    "quickflow/gateway/internal/delivery/http/forms"
    `quickflow/gateway/internal/delivery/http/interfaces`
    "quickflow/shared/models"

    "github.com/google/uuid"
    "github.com/gorilla/websocket"
)

type PostEvent string

const (
    PostLiked     PostEvent = "post_liked"
    CommentLiked  PostEvent = "comment_liked"
    PostCommented PostEvent = "post_commented"
)

type InternalWSPostHandler struct {
    connManager    *WSConnectionManager
    profileService interfaces.ProfileUseCase
}

func NewInternalWSPostHandler(wsConnectionManager *WSConnectionManager, profileService interfaces.ProfileUseCase) *InternalWSPostHandler {
    return &InternalWSPostHandler{
        connManager:    wsConnectionManager,
        profileService: profileService,
    }
}

func (f *InternalWSPostHandler) notifyLikeEvent(_ context.Context, read interface{}, receiver uuid.UUID, eventType PostEvent) error {
    conn, exists := f.connManager.IsConnected(receiver)
    if !exists {
        return nil
    }

    out := struct {
        Type string      `json:"type"`
        Data interface{} `json:"payload"`
    }{string(eventType), read}

    msgJSON, err := json.Marshal(out)
    if err != nil {
        return fmt.Errorf("failed to marshal message: %w", err)
    }

    err = conn.WriteMessage(websocket.TextMessage, msgJSON)
    if err != nil {
        return fmt.Errorf("failed to send message: %w", err)
    }
    return nil
}

func (f *InternalWSPostHandler) NotifyPostLiked(ctx context.Context, senderId, receiverId uuid.UUID, post *models.Post) error {
    _, exists := f.connManager.IsConnected(receiverId)
    if !exists {
        return nil
    }

    var postOut forms.PostOut
    postOut.FromPost(*post)

    senderProfileInfo, err := f.profileService.GetPublicUserInfo(ctx, senderId)
    if err != nil {
        return fmt.Errorf("failed to get sender profile info: %w", err)
    }

    out := struct {
        Post forms.PostOut           `json:"post"`
        User forms.PublicUserInfoOut `json:"user"`
    }{
        Post: postOut,
        User: forms.PublicUserInfoToOut(senderProfileInfo, ""),
    }

    err = f.notifyLikeEvent(ctx, out, receiverId, PostLiked)
    if err != nil {
        return fmt.Errorf("failed to notify friend request sent: %w", err)
    }
    return nil
}

func (f *InternalWSPostHandler) NotifyCommentLiked(ctx context.Context, senderId, receiverId uuid.UUID, comment *models.Comment) error {
    _, exists := f.connManager.IsConnected(receiverId)
    if !exists {
        return nil
    }

    senderProfileInfo, err := f.profileService.GetPublicUserInfo(ctx, senderId)
    if err != nil {
        return fmt.Errorf("failed to get sender profile info: %w", err)
    }

    authorProfileInfo, err := f.profileService.GetPublicUserInfo(ctx, comment.UserId)
    if err != nil {
        return fmt.Errorf("failed to get author profile info: %w", err)
    }

    var commentOut forms.CommentOut
    commentOut.FromComment(*comment, authorProfileInfo)

    out := struct {
        Comment forms.CommentOut        `json:"comment"`
        User    forms.PublicUserInfoOut `json:"user"`
    }{
        Comment: commentOut,
        User:    forms.PublicUserInfoToOut(senderProfileInfo, ""),
    }

    err = f.notifyLikeEvent(ctx, out, receiverId, CommentLiked)
    if err != nil {
        return fmt.Errorf("failed to notify friend request accepted: %w", err)
    }
    return nil
}

func (f *InternalWSPostHandler) NotifyPostCommented(ctx context.Context, senderId, receiverId uuid.UUID, post *models.Post, comment *models.Comment) error {
    _, exists := f.connManager.IsConnected(receiverId)
    if !exists {
        return nil
    }

    senderProfileInfo, err := f.profileService.GetPublicUserInfo(ctx, senderId)
    if err != nil {
        return fmt.Errorf("failed to get sender profile info: %w", err)
    }

    var postOut forms.PostOut
    postOut.FromPost(*post)

    var commentOut forms.CommentOut
    commentOut.FromComment(*comment, senderProfileInfo)

    out := struct {
        Post    forms.PostOut    `json:"post"`
        Comment forms.CommentOut `json:"comment"`
    }{
        Post:    postOut,
        Comment: commentOut,
    }

    err = f.notifyLikeEvent(ctx, out, receiverId, PostCommented)
    if err != nil {
        return fmt.Errorf("failed to notify friend request accepted: %w", err)
    }
    return nil
}
