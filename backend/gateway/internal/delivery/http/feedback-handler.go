package http

import (
    "net/http"

    "github.com/google/uuid"
    "github.com/mailru/easyjson"
    "github.com/microcosm-cc/bluemonday"

    "quickflow/gateway/internal/delivery/http/forms"
    `quickflow/gateway/internal/delivery/http/interfaces`
    errors2 "quickflow/gateway/internal/errors"
    "quickflow/gateway/pkg/sanitizer"
    http2 "quickflow/gateway/utils/http"
    "quickflow/shared/logger"
    "quickflow/shared/models"
)

type FeedbackHandler struct {
    feedbackUseCase interfaces.FeedbackUseCase
    profileService  interfaces.ProfileUseCase
    policy          *bluemonday.Policy
}

func NewFeedbackHandler(feedbackUseCase interfaces.FeedbackUseCase, profileService interfaces.ProfileUseCase, policy *bluemonday.Policy) *FeedbackHandler {
    return &FeedbackHandler{
        feedbackUseCase: feedbackUseCase,
        profileService:  profileService,
        policy:          policy,
    }
}

func (f *FeedbackHandler) SaveFeedback(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    user, ok := ctx.Value("user").(models.User)
    if !ok {
        logger.Error(ctx, "Failed to get user from context while saving feedback")
        http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to get user from context", http.StatusInternalServerError))
        return
    }

    var form forms.FeedbackForm
    err := easyjson.UnmarshalFromReader(r.Body, &form)
    if err != nil {
        logger.Error(ctx, "Failed to decode request body for feedback %v", err)
        http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Bad request body", http.StatusBadRequest))
        return
    }

    sanitizer.SanitizeFeedbackText(&form, f.policy)

    feedback, err := form.ToFeedback(user.Id)
    if err != nil {
        logger.Error(ctx, "Invalid feedback form %v", err)
        http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Invalid feedback form", http.StatusBadRequest))
        return
    }

    if err := f.feedbackUseCase.SaveFeedback(ctx, feedback); err != nil {
        appErr := errors2.FromGRPCError(err)
        logger.Error(ctx, "Failed to save feedback %v", err)
        http2.WriteJSONError(w, appErr)
        return
    }
}

func (f *FeedbackHandler) GetAllFeedbackType(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    var chatForm forms.GetFeedbackForm
    if err := chatForm.GetParams(r.URL.Query()); err != nil {
        logger.Error(ctx, "Failed to parse query params for feedback list %v", err)
        http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Invalid query parameters", http.StatusBadRequest))
        return
    }

    feedbacks, err := f.feedbackUseCase.GetAllFeedbackType(ctx, chatForm.Type, chatForm.Ts, chatForm.Count)
    if err != nil {
        appErr := errors2.FromGRPCError(err)
        logger.Error(ctx, "Failed to fetch feedback list %v", err)
        http2.WriteJSONError(w, appErr)
        return
    }

    profileInfos := make(map[uuid.UUID]models.PublicUserInfo)
    var avg float64
    var feedbackOutput []forms.FeedbackFormOut

    for _, feedback := range feedbacks {
        info, found := profileInfos[feedback.RespondentId]
        if !found {
            info, err = f.profileService.GetPublicUserInfo(ctx, feedback.RespondentId)
            if err != nil {
                appErr := errors2.FromGRPCError(err)
                logger.Error(ctx, "Failed to load respondent info %v", err)
                http2.WriteJSONError(w, appErr)
                return
            }
            profileInfos[feedback.RespondentId] = info
        }
        feedbackOutput = append(feedbackOutput, forms.FromFeedBack(feedback, info))
        avg += float64(feedback.Rating)
    }

    if len(feedbackOutput) > 0 {
        avg /= float64(len(feedbackOutput))
    }

    result := forms.FeedbackOutAverage{Feedbacks: feedbackOutput, Average: avg}
    out := forms.PayloadWrapper[forms.FeedbackOutAverage]{Payload: result}
    w.Header().Set("Content-Type", "application/json")
    js, err := out.MarshalJSON()
    if err != nil {
        logger.Error(ctx, "Failed to marshal json payload %v", err)
        http2.WriteJSONError(w, err)
        return
    }
    if _, err = w.Write(js); err != nil {
        logger.Error(ctx, "Failed to encode feedback output %v", err)
        http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to encode feedback output", http.StatusInternalServerError))
    }
}
