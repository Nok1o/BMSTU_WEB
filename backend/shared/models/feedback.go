package models

import (
    "time"

    "github.com/google/uuid"
)

type FeedbackType string

const (
    FeedbackGeneral        = "general"
    FeedbackPost           = "post"
    FeedbackMessenger      = "messenger"
    FeedbackRecommendation = "recommendation"
    FeedbackProfile        = "profile"
    FeedbackAuth           = "auth"
)

type Feedback struct {
    Id           uuid.UUID
    Rating       int
    RespondentId uuid.UUID
    Text         string
    Type         FeedbackType
    CreatedAt    time.Time
}

type AverageStat struct {
    FeedbackId    uuid.UUID
    AverageRating float32
}
