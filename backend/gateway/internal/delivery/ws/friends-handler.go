package ws

import (
    "context"
    "encoding/json"
    "fmt"

    "github.com/google/uuid"
    "github.com/gorilla/websocket"

    "quickflow/gateway/internal/delivery/http/forms"
    `quickflow/gateway/internal/delivery/http/interfaces`
)

type FriendEvent string

const (
    FriendEventRequestSent     FriendEvent = "fr_received"
    FriendEventRequestAccepted FriendEvent = "fr_accepted"
)

type InternalWSFriendsHandler struct {
    connManager    *WSConnectionManager
    profileService interfaces.ProfileUseCase
}

func NewInternalWSFriendsHandler(wsConnectionManager *WSConnectionManager, profileService interfaces.ProfileUseCase) *InternalWSFriendsHandler {
    return &InternalWSFriendsHandler{
        connManager:    wsConnectionManager,
        profileService: profileService,
    }
}

func (f *InternalWSFriendsHandler) NotifyFriendRequestSent(ctx context.Context, senderId, receiverId uuid.UUID) error {
    _, exists := f.connManager.IsConnected(receiverId)
    if !exists {
        return nil
    }

    senderProfileInfo, err := f.profileService.GetPublicUserInfo(ctx, senderId)
    if err != nil {
        return fmt.Errorf("failed to get sender profile info: %w", err)
    }

    err = f.notifyFriendEvent(ctx, forms.PublicUserInfoToOut(senderProfileInfo, ""), receiverId, FriendEventRequestSent)
    if err != nil {
        return fmt.Errorf("failed to notify friend request sent: %w", err)
    }
    return nil
}

func (f *InternalWSFriendsHandler) NotifyFriendRequestAccepted(ctx context.Context, senderId, receiverId uuid.UUID) error {
    _, exists := f.connManager.IsConnected(receiverId)
    if !exists {
        return nil
    }

    senderProfileInfo, err := f.profileService.GetPublicUserInfo(ctx, senderId)
    if err != nil {
        return fmt.Errorf("failed to get sender profile info: %w", err)
    }

    err = f.notifyFriendEvent(ctx, forms.PublicUserInfoToOut(senderProfileInfo, ""), receiverId, FriendEventRequestAccepted)
    if err != nil {
        return fmt.Errorf("failed to notify friend request accepted: %w", err)
    }
    return nil
}

func (f *InternalWSFriendsHandler) notifyFriendEvent(_ context.Context, read interface{}, receiver uuid.UUID, eventType FriendEvent) error {
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
