package http

import (
    "errors"
    "net/http"

    "github.com/microcosm-cc/bluemonday"

    "quickflow/gateway/internal/delivery/http/forms"
    `quickflow/gateway/internal/delivery/http/interfaces`
    errors2 "quickflow/gateway/internal/errors"
    http2 "quickflow/gateway/utils/http"
    "quickflow/shared/logger"
    "quickflow/shared/models"
)

type FileHandler struct {
    fileService interfaces.FileService
    policy      *bluemonday.Policy
}

func NewFileHandler(fileService interfaces.FileService, policy *bluemonday.Policy) *FileHandler {
    return &FileHandler{
        fileService: fileService,
        policy:      policy,
    }
}

func (p *FileHandler) AddFiles(w http.ResponseWriter, r *http.Request) {
    // extracting user from context
    ctx := r.Context()
    user, ok := ctx.Value("user").(models.User)
    if !ok {
        logger.Error(ctx, "Failed to get user from context while adding files")
        http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to get user from context", http.StatusInternalServerError))
        return
    }

    logger.Info(ctx, "User %s requested to add files", user.Username)

    // Parse the form data
    err := r.ParseMultipartForm(200 << 20) // 200 MB TODO
    if err != nil {
        logger.Error(ctx, "Failed to parse form: %s", err.Error())
        http2.WriteJSONError(w, err)
        return
    }

    // SendMessage video files
    media, err := http2.GetFiles(r, "media")
    if errors.Is(err, http2.TooManyFilesErr) {
        logger.Error(ctx, "Too many media files requested: %s", err.Error())
        http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Too many media files requested", http.StatusBadRequest))
        return
    }
    if err != nil {
        logger.Error(ctx, "Failed to get video files: %s", err.Error())
        http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to get video files", http.StatusBadRequest))
        return
    }

    // SendMessage audio files
    audios, err := http2.GetFiles(r, "audio")
    if errors.Is(err, http2.TooManyFilesErr) {
        logger.Error(ctx, "Too many audio files requested: %s", err.Error())
        http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Too many audio files requested", http.StatusBadRequest))
        return
    }
    if err != nil {
        logger.Error(ctx, "Failed to get audio files: %s", err.Error())
        http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to get audio files", http.StatusBadRequest))
        return
    }

    stickers, err := http2.GetFiles(r, "stickers")
    if errors.Is(err, http2.TooManyFilesErr) {
        logger.Error(ctx, "Too many sticker files requested: %s", err.Error())
        http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Too many sticker files requested", http.StatusBadRequest))
        return
    } else if err != nil {
        logger.Error(ctx, "Failed to get sticker files: %s", err.Error())
        http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to get sticker files", http.StatusBadRequest))
        return
    }

    // SendMessage other files
    otherFiles, err := http2.GetFiles(r, "files")
    if errors.Is(err, http2.TooManyFilesErr) {
        logger.Error(ctx, "Too many files requested: %s", err.Error())
        http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Too many files requested", http.StatusBadRequest))
        return
    }
    if err != nil {
        logger.Error(ctx, "Failed to get files: %s", err.Error())
        http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to get files", http.StatusBadRequest))
        return
    }

    var res forms.MessageAttachmentForm
    res.MediaURLs, err = p.fileService.UploadManyFiles(ctx, media)
    if err != nil {
        http2.WriteJSONError(w, err)
        return
    }
    res.AudioURLs, err = p.fileService.UploadManyFiles(ctx, audios)
    if err != nil {
        http2.WriteJSONError(w, err)
        return
    }
    res.FileURLs, err = p.fileService.UploadManyFiles(ctx, otherFiles)
    if err != nil {
        http2.WriteJSONError(w, err)
        return
    }

    for i := range stickers {
        stickers[i].DisplayType = models.DisplayTypeSticker
    }
    res.StickerURLs, err = p.fileService.UploadManyFiles(ctx, stickers)
    if err != nil {
        http2.WriteJSONError(w, err)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    out := forms.PayloadWrapper[forms.MessageAttachmentForm]{Payload: res}
    js, err := out.MarshalJSON()
    if err != nil {
        logger.Error(ctx, "Failed to marshal json payload%v", err)
        http2.WriteJSONError(w, err)
        return
    }
    if _, err = w.Write(js); err != nil {
        logger.Error(ctx, "Failed to encode feedback output%v", err)
        http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to encode feedback output", http.StatusInternalServerError))
    }
}
