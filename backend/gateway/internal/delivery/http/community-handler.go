package http

import (
	"encoding/json"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/microcosm-cc/bluemonday"

	"quickflow/gateway/internal/delivery/http/forms"
	"quickflow/gateway/internal/delivery/http/interfaces"
	interfaces2 "quickflow/gateway/internal/delivery/ws/interfaces"
	errors2 "quickflow/gateway/internal/errors"
	"quickflow/gateway/pkg/sanitizer"
	http2 "quickflow/gateway/utils/http"
	"quickflow/shared/logger"
	"quickflow/shared/models"
)

type CommunityHandler struct {
	communityService interfaces.CommunityService
	profileService   interfaces.ProfileUseCase
	connService      interfaces2.IWebSocketConnectionManager
	authService      interfaces.AuthUseCase
	policy           *bluemonday.Policy
}

func NewCommunityHandler(communityService interfaces.CommunityService, profileService interfaces.ProfileUseCase, connService interfaces2.IWebSocketConnectionManager, authService interfaces.AuthUseCase, policy *bluemonday.Policy) *CommunityHandler {
	return &CommunityHandler{
		communityService: communityService,
		profileService:   profileService,
		connService:      connService,
		authService:      authService,
		policy:           policy,
	}
}

func (c *CommunityHandler) CreateCommunity(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := ctx.Value("user").(models.User)
	if !ok {
		logger.Error(ctx, "Failed to get user from context while creating community")
		http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to get user from context", http.StatusInternalServerError))
		return
	}
	logger.Info(ctx, "User %s requested community creation", user.Username)

	err := r.ParseMultipartForm(15 << 20)
	if err != nil {
		logger.Error(ctx, "Failed to parse form: %s", err.Error())
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Failed to parse form", http.StatusBadRequest))
		return
	}

	var communityForm forms.CreateCommunityForm
	communityForm.Nickname = r.FormValue("nickname")
	communityForm.Name = r.FormValue("name")
	communityForm.Description = r.FormValue("description")

	if utf8.RuneCountInString(communityForm.Description) > 500 {
		logger.Error(ctx, "Text length validation failed: length=%d", utf8.RuneCountInString(communityForm.Description))
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Text must be between 1 and 4096 characters", http.StatusBadRequest))
		return
	}

	sanitizer.SanitizeCommunityCreation(&communityForm, c.policy)

	communityForm.Avatar, err = http2.GetFile(r, "avatar")
	if err != nil {
		logger.Error(ctx, "Failed to get avatar: %s", err.Error())
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Failed to get avatar", http.StatusBadRequest))
		return
	}

	communityForm.Cover, err = http2.GetFile(r, "cover")
	if err != nil {
		logger.Error(ctx, "Failed to get cover: %s", err.Error())
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Failed to get cover", http.StatusBadRequest))
		return
	}

	logger.Info(ctx, "Recieved community: %+v", communityForm)

	community := communityForm.CreateFormToModel()
	community.OwnerID = user.Id

	newCommunity, err := c.communityService.CreateCommunity(ctx, &community)
	if err != nil {
		logger.Error(ctx, "Failed to create community: %s", err.Error())
		http2.WriteJSONError(w, err)
		return
	}

	info, err := c.profileService.GetPublicUserInfo(ctx, newCommunity.OwnerID)
	if err != nil {
		logger.Error(ctx, "Failed to get user info: %s", err.Error())
		http2.WriteJSONError(w, err)
		return
	}

	communityOut := forms.ToCommunityForm(*newCommunity, info)
	communityOut.Role = string(models.CommunityRoleOwner)

	w.Header().Set("Content-Type", "application/json")

	out := forms.PayloadWrapper[forms.CommunityForm]{Payload: communityOut}
	js, err := out.MarshalJSON()
	if err != nil {
		logger.Error(ctx, "Failed to marshal json response: %v", err)
		http2.WriteJSONError(w, err)
		return
	}

	if _, err = w.Write(js); err != nil {
		logger.Error(ctx, "Failed to encode community: %s", err.Error())
		http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to encode community", http.StatusInternalServerError))
		return
	}
}

func (c *CommunityHandler) GetCommunityById(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := ctx.Value("user").(models.User)
	if !ok {
		logger.Error(ctx, "Failed to get user from context while fetching community")
		http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to get user from context", http.StatusInternalServerError))
		return
	}
	logger.Info(ctx, "User %s requested community info", user.Username)

	communityIdStr := mux.Vars(r)["id"]
	if len(communityIdStr) == 0 {
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Community ID is required", http.StatusBadRequest))
		return
	}

	communityId, err := uuid.Parse(communityIdStr)
	if err != nil {
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Invalid community ID", http.StatusBadRequest))
		return
	}

	community, err := c.communityService.GetCommunityById(ctx, communityId)
	if err != nil {
		logger.Error(ctx, "Failed to get community: %s", err.Error())
		http2.WriteJSONError(w, err)
		return
	}

	isMember, role, err := c.communityService.IsCommunityMember(ctx, user.Id, communityId)
	if err != nil {
		logger.Error(ctx, "Failed to check community membership: %s", err.Error())
		http2.WriteJSONError(w, err)
		return
	}

	info, err := c.profileService.GetPublicUserInfo(ctx, community.OwnerID)
	if err != nil {
		logger.Error(ctx, "Failed to get user info: %s", err.Error())
		http2.WriteJSONError(w, err)
		return
	}

	communityOut := forms.ToCommunityForm(*community, info)
	if isMember && role != nil {
		communityOut.Role = string(*role)
	}

	w.Header().Set("Content-Type", "application/json")
	out := forms.PayloadWrapper[forms.CommunityForm]{Payload: communityOut}
	js, err := out.MarshalJSON()
	if err != nil {
		logger.Error(ctx, "Failed to marshal json response: %v", err)
		http2.WriteJSONError(w, err)
		return
	}

	if _, err = w.Write(js); err != nil {
		logger.Error(ctx, "Failed to encode community: %s", err.Error())
		http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to encode community", http.StatusInternalServerError))
		return
	}
}

func (c *CommunityHandler) GetCommunityByName(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := ctx.Value("user").(models.User)
	if !ok {
		logger.Error(ctx, "Failed to get user from context while fetching community")
		http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to get user from context", http.StatusInternalServerError))
		return
	}
	logger.Info(ctx, "User %s requested community info", user.Username)

	communityName := mux.Vars(r)["name"]
	if len(communityName) == 0 {
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Community name is required", http.StatusBadRequest))
		return
	}

	community, err := c.communityService.GetCommunityByName(ctx, communityName)
	if err != nil {
		logger.Error(ctx, "Failed to get community: %s", err.Error())
		http2.WriteJSONError(w, err)
		return
	}

	isMember, role, err := c.communityService.IsCommunityMember(ctx, user.Id, community.ID)
	if err != nil {
		logger.Error(ctx, "Failed to check community membership: %s", err.Error())
		http2.WriteJSONError(w, err)
		return
	}

	info, err := c.profileService.GetPublicUserInfo(ctx, community.OwnerID)
	if err != nil {
		logger.Error(ctx, "Failed to get user info: %s", err.Error())
		http2.WriteJSONError(w, err)
		return
	}

	communityOut := forms.ToCommunityForm(*community, info)
	if isMember && role != nil {
		communityOut.Role = string(*role)
	}

	w.Header().Set("Content-Type", "application/json")
	out := forms.PayloadWrapper[forms.CommunityForm]{Payload: communityOut}
	js, err := out.MarshalJSON()
	if err != nil {
		logger.Error(ctx, "Failed to marshal json response: %v", err)
		http2.WriteJSONError(w, err)
		return
	}

	if _, err = w.Write(js); err != nil {
		logger.Error(ctx, "Failed to encode community: %s", err.Error())
		http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to encode community", http.StatusInternalServerError))
		return
	}
}

func (c *CommunityHandler) GetCommunityMembers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := ctx.Value("user").(models.User)
	if !ok {
		logger.Error(ctx, "Failed to get user from context while fetching community members")
		http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to get user from context", http.StatusInternalServerError))
		return
	}
	logger.Info(ctx, "User %s requested community members", user.Username)

	communityIdStr := mux.Vars(r)["id"]
	if len(communityIdStr) == 0 {
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Community ID is required", http.StatusBadRequest))
		return
	}

	communityId, err := uuid.Parse(communityIdStr)
	if err != nil {
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Invalid community ID", http.StatusBadRequest))
		return
	}

	var pagination forms.PaginationForm
	err = pagination.GetParams(r.URL.Query())
	if err != nil {
		logger.Error(ctx, "Failed to parse query params: %v", err)
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Failed to parse query params", http.StatusBadRequest))
		return
	}

	members, err := c.communityService.GetCommunityMembers(ctx, communityId, pagination.Count, pagination.Ts)
	if err != nil {
		logger.Error(ctx, "Failed to get community members: %s", err.Error())
		http2.WriteJSONError(w, err)
		return
	}

	memberIds := make([]uuid.UUID, 0, len(members))
	for _, member := range members {
		memberIds = append(memberIds, member.UserID)
	}

	publicInfos, err := c.profileService.GetPublicUsersInfo(ctx, memberIds)
	if err != nil {
		logger.Error(ctx, "Failed to get public user info: %s", err.Error())
		http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to get public user info", http.StatusInternalServerError))
		return
	}

	publicInfoMap := make(map[uuid.UUID]models.PublicUserInfo)
	for _, info := range publicInfos {
		publicInfoMap[info.Id] = info
	}

	var membersOut []forms.CommunityMemberOut
	for _, member := range members {
		formOut := forms.ToCommunityMemberOut(*member, publicInfoMap[member.UserID])
		if _, isOnline := c.connService.IsConnected(member.UserID); isOnline {
			formOut.IsOnline = &isOnline
		} else {
			formOut.IsOnline = nil
		}
		membersOut = append(membersOut, formOut)
	}

	w.Header().Set("Content-Type", "application/json")
	out := forms.PayloadWrapper[[]forms.CommunityMemberOut]{Payload: membersOut}
	js, err := out.MarshalJSON()
	if err != nil {
		logger.Error(ctx, "Failed to marshal json response: %v", err)
		http2.WriteJSONError(w, err)
		return
	}

	if _, err = w.Write(js); err != nil {
		logger.Error(ctx, "Failed to encode community: %s", err.Error())
		http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to encode community", http.StatusInternalServerError))
		return
	}
}

func (c *CommunityHandler) DeleteCommunity(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := ctx.Value("user").(models.User)
	if !ok {
		logger.Error(ctx, "Failed to get user from context while deleting community")
		http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to get user from context", http.StatusInternalServerError))
		return
	}
	logger.Info(ctx, "User %s requested community deletion", user.Username)

	communityIdStr := mux.Vars(r)["id"]
	if len(communityIdStr) == 0 {
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Community ID is required", http.StatusBadRequest))
		return
	}

	communityId, err := uuid.Parse(communityIdStr)
	if err != nil {
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Invalid community ID", http.StatusBadRequest))
		return
	}

	err = c.communityService.DeleteCommunity(ctx, communityId, user.Id)
	if err != nil {
		logger.Error(ctx, "Failed to delete community: %s", err.Error())
		http2.WriteJSONError(w, err)
		return
	}
	logger.Info(ctx, "Successfully deleted community")
}

func (c *CommunityHandler) UpdateCommunity(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := ctx.Value("user").(models.User)
	if !ok {
		logger.Error(ctx, "Failed to get user from context while updating community")
		http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to get user from context", http.StatusInternalServerError))
		return
	}
	logger.Info(ctx, "User %s requested community update", user.Username)

	err := r.ParseMultipartForm(15 << 20)
	if err != nil {
		logger.Error(ctx, "Failed to parse form: %s", err.Error())
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Failed to parse form", http.StatusBadRequest))
		return
	}

	communityIdStr := mux.Vars(r)["id"]
	if len(communityIdStr) == 0 {
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Community ID is required", http.StatusBadRequest))
		return
	}

	communityId, err := uuid.Parse(communityIdStr)
	if err != nil {
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Invalid community ID", http.StatusBadRequest))
		return
	}

	var communityForm forms.CreateCommunityForm
	communityForm.Nickname = r.FormValue("nickname")
	communityForm.Name = r.FormValue("name")
	communityForm.Description = r.FormValue("description")

	if utf8.RuneCountInString(communityForm.Description) > 4000 {
		logger.Error(ctx, "Text length validation failed: length=%d", utf8.RuneCountInString(communityForm.Description))
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Text must be between 1 and 4096 characters", http.StatusBadRequest))
		return
	}

	sanitizer.SanitizeCommunityCreation(&communityForm, c.policy)

	communityForm.Avatar, err = http2.GetFile(r, "avatar")
	if err != nil {
		logger.Error(ctx, "Failed to get avatar: %s", err.Error())
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Failed to get avatar", http.StatusBadRequest))
		return
	}

	communityForm.Cover, err = http2.GetFile(r, "cover")
	if err != nil {
		logger.Error(ctx, "Failed to get cover: %s", err.Error())
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Failed to get cover", http.StatusBadRequest))
		return
	}

	community := communityForm.CreateFormToModel()
	community.ID = communityId

	var contactInfo forms.ContactInfo
	err = json.NewDecoder(strings.NewReader(r.FormValue("contact_info"))).Decode(&contactInfo)
	if err == nil {
		community.ContactInfo = forms.ContactInfoFormToModel(&contactInfo)
	}

	logger.Info(ctx, "Recieved community: %+v", communityForm)

	newCommunity, err := c.communityService.UpdateCommunity(ctx, &community, user.Id)
	if err != nil {
		logger.Error(ctx, "Failed to update community: %s", err.Error())
		http2.WriteJSONError(w, err)
		return
	}
	logger.Info(ctx, "Successfully updated community")

	info, err := c.profileService.GetPublicUserInfo(ctx, newCommunity.OwnerID)
	if err != nil {
		logger.Error(ctx, "Failed to get user info: %s", err.Error())
		http2.WriteJSONError(w, err)
		return
	}
	communityOut := forms.ToCommunityForm(*newCommunity, info)

	w.Header().Set("Content-Type", "application/json")
	out := forms.PayloadWrapper[forms.CommunityForm]{Payload: communityOut}
	js, err := out.MarshalJSON()
	if err != nil {
		logger.Error(ctx, "Failed to marshal json response: %v", err)
		http2.WriteJSONError(w, err)
		return
	}

	if _, err = w.Write(js); err != nil {
		logger.Error(ctx, "Failed to encode community: %s", err.Error())
		http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to encode community", http.StatusInternalServerError))
		return
	}
}

func (c *CommunityHandler) JoinCommunity(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := ctx.Value("user").(models.User)
	if !ok {
		logger.Error(ctx, "Failed to get user from context while joining community")
		http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to get user from context", http.StatusInternalServerError))
		return
	}
	logger.Info(ctx, "User %s requested to join community", user.Username)

	communityIdStr := mux.Vars(r)["id"]
	if len(communityIdStr) == 0 {
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Community ID is required", http.StatusBadRequest))
		return
	}

	communityId, err := uuid.Parse(communityIdStr)
	if err != nil {
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Invalid community ID", http.StatusBadRequest))
		return
	}

	member := models.CommunityMember{
		UserID:      user.Id,
		CommunityID: communityId,
		Role:        models.CommunityRoleMember,
	}

	err = c.communityService.JoinCommunity(ctx, &member)
	if err != nil {
		logger.Error(ctx, "Failed to join community: %s", err.Error())
		http2.WriteJSONError(w, err)
		return
	}
	logger.Info(ctx, "Successfully joined community")
}

func (c *CommunityHandler) LeaveCommunity(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := ctx.Value("user").(models.User)
	if !ok {
		logger.Error(ctx, "Failed to get user from context while leaving community")
		http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to get user from context", http.StatusInternalServerError))
		return
	}
	logger.Info(ctx, "User %s requested to leave community", user.Username)

	communityIdStr := mux.Vars(r)["id"]
	if len(communityIdStr) == 0 {
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Community ID is required", http.StatusBadRequest))
		return
	}

	communityId, err := uuid.Parse(communityIdStr)
	if err != nil {
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Invalid community ID", http.StatusBadRequest))
		return
	}

	err = c.communityService.LeaveCommunity(ctx, user.Id, communityId)
	if err != nil {
		logger.Error(ctx, "Failed to leave community: %s", err.Error())
		http2.WriteJSONError(w, err)
		return
	}
	logger.Info(ctx, "Successfully left community")
}

func (c *CommunityHandler) GetUserCommunities(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	username := mux.Vars(r)["username"]
	if username == "" {
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Failed to get username from URL", http.StatusBadRequest))
		return
	}

	user, err := c.authService.GetUserByUsername(ctx, username)
	if err != nil {
		http2.WriteJSONError(w, err)
		return
	}

	var pagination forms.PaginationForm
	err = pagination.GetParams(r.URL.Query())
	if err != nil {
		logger.Error(ctx, "Failed to parse query params: %v", err)
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Failed to parse query params", http.StatusBadRequest))
		return
	}

	communities, err := c.communityService.GetUserCommunities(ctx, user.Id, pagination.Count, pagination.Ts)
	if err != nil {
		logger.Error(ctx, "Failed to get user communities: %s", err.Error())
		http2.WriteJSONError(w, err)
		return
	}

	requiredRole := r.URL.Query().Get("role")

	var communitiesOut []forms.CommunityForm
	for _, community := range communities {
		info, err := c.profileService.GetPublicUserInfo(ctx, community.OwnerID)
		if err != nil {
			logger.Error(ctx, "Failed to get user info: %s", err.Error())
			http2.WriteJSONError(w, err)
			return
		}

		communityOut := forms.ToCommunityForm(*community, info)
		isMember, role, err := c.communityService.IsCommunityMember(ctx, user.Id, community.ID)
		if err != nil {
			logger.Error(ctx, "Failed to check community membership: %s", err.Error())
			http2.WriteJSONError(w, err)
			return
		}
		if isMember && role != nil {
			communityOut.Role = string(*role)

			if requiredRole != "" && requiredRole == string(*role) || requiredRole == "" {
				communitiesOut = append(communitiesOut, communityOut)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	out := forms.PayloadWrapper[[]forms.CommunityForm]{Payload: communitiesOut}
	js, err := out.MarshalJSON()
	if err != nil {
		logger.Error(ctx, "Failed to marshal json response: %v", err)
		http2.WriteJSONError(w, err)
		return
	}

	if _, err = w.Write(js); err != nil {
		logger.Error(ctx, "Failed to encode community: %s", err.Error())
		http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to encode community", http.StatusInternalServerError))
		return
	}
}

func (c *CommunityHandler) ChangeUserRole(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := ctx.Value("user").(models.User)
	if !ok {
		logger.Error(ctx, "Failed to get user from context while changing community role")
		http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to get user from context", http.StatusInternalServerError))
		return
	}
	logger.Info(ctx, "User %s requested to change community role", user.Username)

	communityIdStr := mux.Vars(r)["id"]
	if len(communityIdStr) == 0 {
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Community ID is required", http.StatusBadRequest))
		return
	}

	communityId, err := uuid.Parse(communityIdStr)
	if err != nil {
		logger.Error(ctx, "Invalid community ID: %s", err.Error())
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Invalid community ID", http.StatusBadRequest))
		return
	}

	userId := mux.Vars(r)["user_id"]
	if len(userId) == 0 {
		logger.Error(ctx, "User ID is required")
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "User ID is required", http.StatusBadRequest))
		return
	}

	userIdParsed, err := uuid.Parse(userId)
	if err != nil {
		logger.Error(ctx, "Invalid user ID: %s", err.Error())
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Invalid user ID", http.StatusBadRequest))
		return
	}

	role := r.URL.Query().Get("role")
	if len(role) == 0 {
		logger.Error(ctx, "Role is required")
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Role is required", http.StatusBadRequest))
		return
	}

	switch role {
	case "member":
		role = string(models.CommunityRoleMember)
	case "admin":
		role = string(models.CommunityRoleAdmin)
	case "owner":
		role = string(models.CommunityRoleOwner)
	default:
		logger.Error(ctx, "Invalid role provided")
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Invalid role", http.StatusBadRequest))
		return
	}

	roleParsed := models.CommunityRole(role)

	err = c.communityService.ChangeUserRole(ctx, userIdParsed, communityId, roleParsed, user.Id)
	if err != nil {
		logger.Error(ctx, "Failed to change community role: %s", err.Error())
		http2.WriteJSONError(w, err)
		return
	}
	logger.Info(ctx, "Successfully changed community role")
}

func (c *CommunityHandler) GetControlledCommunities(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	username := mux.Vars(r)["username"]
	if username == "" {
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Failed to get username from URL", http.StatusBadRequest))
		return
	}

	user, err := c.authService.GetUserByUsername(ctx, username)
	if err != nil {
		http2.WriteJSONError(w, err)
		return
	}

	var pagination forms.PaginationForm
	err = pagination.GetParams(r.URL.Query())
	if err != nil {
		logger.Error(ctx, "Failed to parse query params: %v", err)
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Failed to parse query params", http.StatusBadRequest))
		return
	}

	communities, err := c.communityService.GetControlledCommunities(ctx, user.Id, pagination.Count, pagination.Ts)
	if err != nil {
		logger.Error(ctx, "Failed to get user communities: %s", err.Error())
		http2.WriteJSONError(w, err)
		return
	}

	communitiesOut := make([]forms.CommunityForm, len(communities))
	for i, community := range communities {
		info, err := c.profileService.GetPublicUserInfo(ctx, community.OwnerID)
		if err != nil {
			logger.Error(ctx, "Failed to get user info: %s", err.Error())
			http2.WriteJSONError(w, err)
			return
		}

		communitiesOut[i] = forms.ToCommunityForm(*community, info)
		isMember, role, err := c.communityService.IsCommunityMember(ctx, user.Id, community.ID)
		if err != nil {
			logger.Error(ctx, "Failed to check community membership: %s", err.Error())
			http2.WriteJSONError(w, err)
			return
		}
		if isMember && role != nil {
			communitiesOut[i].Role = string(*role)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	out := forms.PayloadWrapper[[]forms.CommunityForm]{Payload: communitiesOut}
	js, err := out.MarshalJSON()
	if err != nil {
		logger.Error(ctx, "Failed to marshal json response: %v", err)
		http2.WriteJSONError(w, err)
		return
	}

	if _, err = w.Write(js); err != nil {
		logger.Error(ctx, "Failed to encode community: %s", err.Error())
		http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to encode community", http.StatusInternalServerError))
		return
	}
}
