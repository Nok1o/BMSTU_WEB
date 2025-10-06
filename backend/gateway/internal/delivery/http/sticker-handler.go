package http

import (
    "net/http"
    "strconv"

    "github.com/google/uuid"
    "github.com/gorilla/mux"
    "github.com/mailru/easyjson"
    "github.com/microcosm-cc/bluemonday"

    "quickflow/gateway/internal/delivery/http/forms"
    `quickflow/gateway/internal/delivery/http/interfaces`
    errors2 "quickflow/gateway/internal/errors"
    http2 "quickflow/gateway/utils/http"
    "quickflow/shared/logger"
    "quickflow/shared/models"
)

type StickerHandler struct {
    stickerUseCase interfaces.StickerUseCase
    policy         *bluemonday.Policy
}

func NewStickerHandler(stickerUseCase interfaces.StickerUseCase, policy *bluemonday.Policy) *StickerHandler {
    return &StickerHandler{
        stickerUseCase: stickerUseCase,
        policy:         policy,
    }
}

// AddStickerPack godoc
// @Summary Add a new Sticker Pack
// @Description Adds a new sticker pack
// @Tags Stickers
// @Accept json
// @Produce json
// @Param sticker_pack body forms.StickerPackForm true "Sticker Pack details"
// @Success 200 {object} forms.StickerPackOut "Created Sticker Pack"
// @Failure 400 {object} forms.ErrorForm "Invalid data"
// @Failure 500 {object} forms.ErrorForm "Server error"
// @Router /api/sticker_packs [post]
func (s *StickerHandler) AddStickerPack(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    logger.Info(ctx, "Received AddStickerPack request")

    user, ok := ctx.Value("user").(models.User)
    if !ok {
        logger.Error(ctx, "Failed to get user from context while adding sticker pack")
        http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to get user from context", http.StatusInternalServerError))
        return
    }

    var form forms.StickerPackForm
    if err := easyjson.UnmarshalFromReader(r.Body, &form); err != nil {
        logger.Error(ctx, "Failed to parse request body: %v", err)
        http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Invalid request body", http.StatusBadRequest))
        return
    }

    form.Name = s.policy.Sanitize(form.Name)

    createdStickerPack, err := s.stickerUseCase.AddStickerPack(ctx, form.ToStickerPackModel(user.Id))
    if err != nil {
        logger.Error(ctx, "Failed to add sticker pack: %v", err)
        http2.WriteJSONError(w, err)
        return
    }

    response := forms.ToStickerPackOut(createdStickerPack)
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)

    out := forms.PayloadWrapper[forms.StickerPackOut]{Payload: response}
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

// GetStickerPack godoc
// @Summary Get a specific Sticker Pack
// @Description Fetches a sticker pack by ID
// @Tags Stickers
// @Accept json
// @Produce json
// @Param id path string true "Sticker Pack ID"
// @Success 200 {object} forms.StickerPackOut "Sticker Pack details"
// @Failure 404 {object} forms.ErrorForm "Sticker Pack not found"
// @Failure 500 {object} forms.ErrorForm "Server error"
// @Router /api/sticker_packs/{id} [get]
func (s *StickerHandler) GetStickerPack(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    logger.Info(ctx, "Received GetStickerPack request")

    packIdString := mux.Vars(r)["pack_id"]
    packId, err := uuid.Parse(packIdString)
    if err != nil {
        logger.Error(ctx, "Invalid Sticker Pack ID:  %v", err)
        http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Invalid Sticker Pack ID", http.StatusBadRequest))
        return
    }

    stickerPack, err := s.stickerUseCase.GetStickerPack(ctx, packId)
    if err != nil {
        logger.Error(ctx, "Failed to get StickerPack with ID %s: %v", packId.String(), err)
        http2.WriteJSONError(w, err)
        return
    }

    // Mapping model to DTO
    response := forms.ToStickerPackOut(stickerPack)
    w.Header().Set("Content-Type", "application/json")
    out := forms.PayloadWrapper[forms.StickerPackOut]{Payload: response}
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

// GetStickerPackByName godoc
// @Summary Get a Sticker Pack by name
// @Description Fetches a sticker pack by its name
// @Tags Stickers
// @Accept json
// @Produce json
// @Param pack_name path string true "Sticker Pack name"
// @Success 200 {object} forms.StickerPackOut "Sticker Pack details"
// @Failure 400 {object} forms.ErrorForm "Invalid data"
// @Failure 404 {object} forms.ErrorForm "Sticker Pack not found"
// @Failure 500 {object} forms.ErrorForm "Server error"
// @Router /api/sticker_packs/{pack_name} [get]
func (s *StickerHandler) GetStickerPackByName(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    logger.Info(ctx, "Received GetStickerPackByName request")

    packString := mux.Vars(r)["pack_name"]
    if len(packString) == 0 {
        logger.Error(ctx, "Empty Sticker Pack name provided")
        http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Empty Sticker Pack name", http.StatusBadRequest))
        return
    }

    stickerPack, err := s.stickerUseCase.GetStickerPackByName(ctx, packString)
    if err != nil {
        logger.Error(ctx, "Failed to get StickerPack with name %s: %v", packString, err)
        http2.WriteJSONError(w, err)
        return
    }

    // Mapping model to DTO
    response := forms.ToStickerPackOut(stickerPack)
    w.Header().Set("Content-Type", "application/json")
    out := forms.PayloadWrapper[forms.StickerPackOut]{Payload: response}
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

// GetStickerPacks godoc
// @Summary Get all Sticker Packs for a user
// @Description Fetches all sticker packs for a given user
// @Tags Stickers
// @Accept json
// @Produce json
// @Param user_id query string true "User ID"
// @Param count query int true "Number of sticker packs"
// @Param offset query int true "Pagination offset"
// @Success 200 {array} forms.StickerPackOut "List of sticker packs"
// @Failure 400 {object} forms.ErrorForm "Invalid data"
// @Failure 500 {object} forms.ErrorForm "Server error"
// @Router /api/sticker_packs [get]
func (s *StickerHandler) GetStickerPacks(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    logger.Info(ctx, "Received GetStickerPacks request")

    user, ok := ctx.Value("user").(models.User)
    if !ok {
        logger.Error(ctx, "Failed to get user from context while fetching sticker packs")
        http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to get user from context", http.StatusInternalServerError))
        return
    }

    count, err := strconv.Atoi(r.URL.Query().Get("count"))
    if err != nil {
        logger.Error(ctx, "Invalid count parameter:  %v", err)
        http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Invalid count parameter", http.StatusBadRequest))
        return
    }

    offsetStr := r.URL.Query().Get("offset")
    if len(offsetStr) == 0 {
        offsetStr = "0"
    }
    offset, err := strconv.Atoi(offsetStr)
    if err != nil {
        logger.Error(ctx, "Invalid offset parameter:  %v", err)
        http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Invalid offset parameter", http.StatusBadRequest))
        return
    }

    stickerPacks, err := s.stickerUseCase.GetStickerPacks(ctx, user.Id, count, offset)
    if err != nil {
        logger.Error(ctx, "Failed to fetch StickerPacks: %v", err)
        http2.WriteJSONError(w, err)
        return
    }

    // Mapping list of StickerPacks to DTOs
    stickerPacksOut := forms.ToStickerPacksOut(stickerPacks)

    w.Header().Set("Content-Type", "application/json")
    out := forms.PayloadWrapper[[]forms.StickerPackOut]{Payload: stickerPacksOut}
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

// DeleteStickerPack godoc
// @Summary Delete a sticker pack
// @Description Deletes a sticker pack by ID
// @Tags Stickers
// @Accept json
// @Produce json
// @Param id path string true "Sticker Pack ID"
// @Success 200 {object} forms.SuccessResponse "Sticker Pack deleted"
// @Failure 400 {object} forms.ErrorForm "Invalid data"
// @Failure 500 {object} forms.ErrorForm "Server error"
// @Router /api/sticker_packs/{id} [delete]
func (s *StickerHandler) DeleteStickerPack(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    logger.Info(ctx, "Received DeleteStickerPack request")

    packIdStr := mux.Vars(r)["pack_id"]
    packId, err := uuid.Parse(packIdStr)
    if err != nil {
        logger.Error(ctx, "Invalid Sticker Pack ID:  %v", err)
        http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Invalid Sticker Pack ID", http.StatusBadRequest))
        return
    }

    // Call DeleteStickerPack from the usecase
    user, ok := ctx.Value("user").(models.User)
    if !ok {
        logger.Error(ctx, "Failed to get user from context while deleting sticker pack")
        http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to get user from context", http.StatusInternalServerError))
        return
    }

    err = s.stickerUseCase.DeleteStickerPack(ctx, user.Id, packId)
    if err != nil {
        logger.Error(ctx, "Failed to delete StickerPack %s: %v", packId.String(), err)
        http2.WriteJSONError(w, err)
        return
    }

    w.WriteHeader(http.StatusOK)
}
