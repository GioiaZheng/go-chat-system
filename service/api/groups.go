// groups.go contains the group management endpoints, including creation,
// membership updates, and media uploads while enforcing OpenAPI constraints.
package api

import (
	"database/sql"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/GioiaZheng/Wasa_proj/service/models"
	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/gofrs/uuid"
	"github.com/julienschmidt/httprouter"
)

// Section: Group create (POST /groups)

// CreateGroupRequest matches the OpenAPI field (memberIds) and also accepts
// a couple of legacy aliases for backward compatibility.
type CreateGroupRequest struct {
	Name            string   `json:"name"`
	MemberIDs       []string `json:"memberIds,omitempty"`  // OpenAPI (camelCase)
	LegacyMembers   []string `json:"members,omitempty"`    // legacy
	LegacyMemberIDs []string `json:"member_ids,omitempty"` // legacy snake_case
}

// groupNamePattern accepts any unicode letters/numbers plus spaces, apostrophes,
// underscores and dashes. This allows names such as "Chinese Group" or O'Connor-Team.
var groupNamePattern = regexp.MustCompile(`^[\p{L}\p{N}\s'_\-]+$`)

// createGroup creates a new group and adds the caller as a member.
//
// Flow:
//  1. Auth check via ctx.UserID (injected by rt.wrap).
//  2. Parse request body and normalize member IDs (dedupe, trim).
//  3. Ensure creator is included.
//  4. Persist the group and the membership.
//  5. Read back the complete group (with members) and return GroupEnvelope.
func (rt *_router) createGroup(
	w http.ResponseWriter,
	r *http.Request,
	_ httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	userID := strings.TrimSpace(ctx.UserID)
	if userID == "" {
		rt.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req CreateGroupRequest
	if err := readJSON(r, &req); err != nil {
		rt.sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate name per OpenAPI (len 1..100 & pattern ^[a-zA-Z0-9\s'-]+$)
	groupName := strings.TrimSpace(req.Name)
	if groupName == "" {
		rt.sendError(w, http.StatusBadRequest, "Name is required")
		return
	}
	if ln := len(groupName); ln < 1 || ln > 100 {
		rt.sendError(w, http.StatusBadRequest, "Name must be 1-100 characters long")
		return
	}
	if !groupNamePattern.MatchString(groupName) {
		rt.sendError(w, http.StatusBadRequest, "Name contains invalid characters")
		return
	}

	// Normalize + dedupe all possible member fields.
	seen := map[string]bool{}
	members := make([]string, 0, len(req.MemberIDs)+len(req.LegacyMembers)+len(req.LegacyMemberIDs)+1)

	push := func(id string) {
		id = strings.TrimSpace(id)
		if id == "" || seen[id] {
			return
		}
		seen[id] = true
		members = append(members, id)
	}

	for _, m := range req.MemberIDs {
		push(m)
	}
	for _, m := range req.LegacyMembers {
		push(m)
	}
	for _, m := range req.LegacyMemberIDs {
		push(m)
	}

	// Ensure the creator is also a member.
	push(userID)
	sort.Strings(members)

	conv, err := rt.db.StartConversation(r.Context(), userID, members, groupName)
	if err != nil {
		ctx.Logger.WithError(err).Error("failed to create conversation for group")
		rt.sendError(w, http.StatusInternalServerError, "Failed to create conversation")
		return
	}

	// Generate group ID (our schema stores TEXT primary keys).
	gid, err := uuid.NewV4()
	if err != nil {
		ctx.Logger.WithError(err).Error("failed to generate group id")
		rt.sendError(w, http.StatusInternalServerError, "Failed to generate group id")
		return
	}
	group := models.Group{ID: gid.String(), Name: groupName, ConversationID: conv.ID}

	// Persist group then its members.
	if err := rt.db.CreateGroup(group); err != nil {
		ctx.Logger.WithError(err).Error("failed to create group")
		rt.sendError(w, http.StatusInternalServerError, "Failed to create group")
		return
	}
	if err := rt.db.AddGroupMembers(group.ID, members); err != nil {
		ctx.Logger.WithError(err).Error("failed to add group members")
		rt.sendError(w, http.StatusInternalServerError, "Failed to add group members")
		return
	}

	// Read back full group (expecting members/conversationId populated by DB layer).
	full, err := rt.db.GetGroup(group.ID)
	if err != nil {
		ctx.Logger.WithError(err).Error("failed to reload created group")
		rt.sendError(w, http.StatusInternalServerError, "Failed to load group")
		return
	}

	// Return GroupEnvelope per OpenAPI.
	_ = writeJSON(w, http.StatusCreated, map[string]interface{}{
		"code":    http.StatusCreated,
		"message": "Group created",
		"data": map[string]interface{}{
			"group": full,
		},
	})
}

// Section: Group detail (GET /groups/:id)

// getGroup returns a single group as GroupEnvelope.
func (rt *_router) getGroup(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	if strings.TrimSpace(ctx.UserID) == "" {
		rt.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	groupID := strings.TrimSpace(ps.ByName("id"))
	if groupID == "" {
		rt.sendError(w, http.StatusBadRequest, "Missing group id")
		return
	}

	group, err := rt.db.GetGroup(groupID)
	if err != nil {
		rt.sendError(w, http.StatusNotFound, "Group not found")
		return
	}

	_ = writeJSON(w, http.StatusOK, map[string]interface{}{
		"code":    http.StatusOK,
		"message": "Group details retrieved",
		"data": map[string]interface{}{
			"group": group,
		},
	})
}

// getGroupDetail is an alias of getGroup to satisfy the registered route.
func (rt *_router) getGroupDetail(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	rt.getGroup(w, r, ps, ctx)
}

// Section: Group listing (GET /groups)

// getGroupsList returns all groups of the current user as GroupEnvelopeCollection.
func (rt *_router) getGroupsList(
	w http.ResponseWriter,
	r *http.Request,
	_ httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	uid := strings.TrimSpace(ctx.UserID)
	if uid == "" {
		rt.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	groups, err := rt.db.GetGroupsList(uid)
	if err != nil {
		rt.sendError(w, http.StatusInternalServerError, "Failed to fetch groups")
		return
	}

	_ = writeJSON(w, http.StatusOK, map[string]interface{}{
		"code":    http.StatusOK,
		"message": "Groups list retrieved",
		"data": map[string]interface{}{
			"items": groups,
		},
	})
}

// Section: Group rename (PUT /groups/:id/name)

type UpdateGroupNameRequest struct {
	Name string `json:"name"`
}

func (rt *_router) setGroupName(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	if strings.TrimSpace(ctx.UserID) == "" {
		rt.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	groupID := strings.TrimSpace(ps.ByName("id"))
	if groupID == "" {
		rt.sendError(w, http.StatusBadRequest, "Missing group id")
		return
	}

	var req UpdateGroupNameRequest
	if err := readJSON(r, &req); err != nil {
		rt.sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		rt.sendError(w, http.StatusBadRequest, "Name is required")
		return
	}
	if ln := len(req.Name); ln < 1 || ln > 100 {
		rt.sendError(w, http.StatusBadRequest, "Name must be 1-100 characters long")
		return
	}
	if !groupNamePattern.MatchString(req.Name) {
		rt.sendError(w, http.StatusBadRequest, "Name contains invalid characters")
		return
	}

	if err := rt.db.UpdateGroupName(groupID, req.Name); err != nil {
		rt.sendError(w, http.StatusInternalServerError, "Failed to update group name")
		return
	}

	_ = writeJSON(w, http.StatusOK, map[string]interface{}{
		"code":    http.StatusOK,
		"message": "Group name updated successfully",
	})
}

// Section: Group leave (DELETE /groups/:id/members)

func (rt *_router) leaveGroup(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	uid := strings.TrimSpace(ctx.UserID)
	if uid == "" {
		rt.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	groupID := strings.TrimSpace(ps.ByName("id"))
	if groupID == "" {
		rt.sendError(w, http.StatusBadRequest, "Missing group id")
		return
	}

	var req struct {
		UserID string `json:"userId,omitempty"`
	}
	if err := readJSON(r, &req); err != nil && !errors.Is(err, io.EOF) {
		rt.sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	targetID := strings.TrimSpace(req.UserID)
	if targetID == "" {
		targetID = uid
	}

	// Ensure caller is at least part of the group before modifying it.
	isMember, err := rt.db.IsGroupMember(uid, groupID)
	if err != nil {
		rt.sendError(w, http.StatusInternalServerError, "Failed to inspect membership")
		return
	}
	if !isMember {
		rt.sendError(w, http.StatusForbidden, "You are not a member of this group")
		return
	}

	if err := rt.db.LeaveGroup(groupID, targetID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			rt.sendError(w, http.StatusNotFound, "User is not a member of this group")
			return
		}
		rt.sendError(w, http.StatusInternalServerError, "Failed to update group membership")
		return
	}

	_ = writeJSON(w, http.StatusOK, map[string]interface{}{
		"code":    http.StatusOK,
		"message": "Left the group",
	})
}

// Section: Group add members (POST /groups/:id/members)

type AddGroupMembersRequest struct {
	MemberIDs       []string `json:"memberIds,omitempty"`  // OpenAPI (camelCase)
	LegacyMemberIDs []string `json:"member_ids,omitempty"` // legacy
	LegacyUserID    string   `json:"userId,omitempty"`     // legacy single user
}

func (rt *_router) addToGroup(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	if strings.TrimSpace(ctx.UserID) == "" {
		rt.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	groupID := strings.TrimSpace(ps.ByName("id"))
	if groupID == "" {
		rt.sendError(w, http.StatusBadRequest, "Missing group id")
		return
	}

	var req AddGroupMembersRequest
	if err := readJSON(r, &req); err != nil {
		rt.sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Normalize + dedupe
	seen := make(map[string]struct{})
	list := make([]string, 0, len(req.MemberIDs)+len(req.LegacyMemberIDs)+1)

	push := func(id string) {
		id = strings.TrimSpace(id)
		if id == "" {
			return
		}
		if _, ok := seen[id]; ok {
			return
		}
		seen[id] = struct{}{}
		list = append(list, id)
	}

	for _, id := range req.MemberIDs {
		push(id)
	}
	for _, id := range req.LegacyMemberIDs {
		push(id)
	}
	if strings.TrimSpace(req.LegacyUserID) != "" {
		push(req.LegacyUserID)
	}

	if len(list) == 0 {
		rt.sendError(w, http.StatusBadRequest, "memberIds is required (or legacy member_ids/userId)")
		return
	}
	sort.Strings(list)

	if err := rt.db.AddGroupMembers(groupID, list); err != nil {
		ctx.Logger.WithError(err).Error("failed to add members to group")
		rt.sendError(w, http.StatusInternalServerError, "Failed to add members to group")
		return
	}

	_ = writeJSON(w, http.StatusOK, map[string]interface{}{
		"code":    http.StatusOK,
		"message": "Members added",
		"data":    map[string]interface{}{"added": list},
	})
}

// Section: Group photo (PUT /groups/:id/photo)
// Two modes:
//   - Preset:  ?preset=avatar7   => /uploads/photos/avatar7.jpg
//   - Upload:  multipart/form-data with file field "upload"
// Response matches FileUploadEnvelope: data.file{ filename, uri, size? }.

func (rt *_router) setGroupPhoto(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	groupID := strings.TrimSpace(ps.ByName("id"))
	if groupID == "" {
		rt.sendError(w, http.StatusBadRequest, "Group ID is required")
		return
	}

	// Mode 1: preset via query param (kept for convenience).
	if preset := strings.TrimSpace(r.URL.Query().Get("preset")); preset != "" {
		if !strings.HasPrefix(strings.ToLower(preset), "avatar") {
			rt.sendError(w, http.StatusBadRequest, "Invalid preset name")
			return
		}
		derived := preset + ".jpg"
		publicURI := rt.publicURL(filepath.ToSlash(filepath.Join("uploads", "photos", derived)))
		if err := rt.db.UpdateGroupPhoto(groupID, publicURI); err != nil {
			ctx.Logger.WithError(err).Error("failed to set preset group photo")
			rt.sendError(w, http.StatusInternalServerError, "Failed to update group photo")
			return
		}
		_ = writeJSON(w, http.StatusOK, map[string]interface{}{
			"code":    http.StatusOK,
			"message": "Group photo updated successfully",
			"data": map[string]interface{}{
				"file": map[string]interface{}{
					"filename": derived,
					"uri":      publicURI, // match OpenAPI's FileUpload.uri
				},
			},
		})
		return
	}

	// Mode 2: uploaded photo via multipart form.
	const maxUploadSizeBytes = 10 << 20 // 10 MiB
	if err := r.ParseMultipartForm(maxUploadSizeBytes); err != nil {
		rt.sendError(w, http.StatusBadRequest, "Invalid multipart form")
		return
	}
	file, header, err := r.FormFile("upload")
	if err != nil {
		rt.sendError(w, http.StatusBadRequest, "Missing file field 'upload'")
		return
	}
	defer file.Close()

	origName := strings.TrimSpace(header.Filename)
	if origName == "" {
		rt.sendError(w, http.StatusBadRequest, "Invalid filename")
		return
	}
	if !allowedExt(origName) {
		rt.sendError(w, http.StatusBadRequest, "Only JPEG/PNG are allowed")
		return
	}
	if ctype, err := detectContentType(file); err == nil {
		if !(strings.HasPrefix(ctype, "image/jpeg") || strings.HasPrefix(ctype, "image/png")) {
			rt.sendError(w, http.StatusBadRequest, "Invalid content type, only JPEG/PNG are allowed")
			return
		}
	}

	// Prepare destination under /uploads/groups
	baseDir := filepath.Join("uploads", "groups")
	if err := ensureDir(baseDir); err != nil {
		ctx.Logger.WithError(err).Error("failed to create uploads dir for groups")
		rt.sendError(w, http.StatusInternalServerError, "Failed to store file")
		return
	}

	ext := strings.ToLower(filepath.Ext(origName))
	safeGroup := strings.NewReplacer("/", "_", "\\", "_", ":", "_").Replace(groupID)
	newName := safeGroup + "_" + time.Now().UTC().Format("20060102T150405Z") + ext
	dstPath := filepath.Join(baseDir, newName)

	dst, err := os.Create(dstPath)
	if err != nil {
		ctx.Logger.WithError(err).Error("failed to create destination file (group photo)")
		rt.sendError(w, http.StatusInternalServerError, "Failed to store file")
		return
	}
	defer dst.Close()

	written, err := io.Copy(dst, io.LimitReader(file, maxUploadSizeBytes+1))
	if err != nil {
		ctx.Logger.WithError(err).Error("failed to write uploaded group photo")
		rt.sendError(w, http.StatusInternalServerError, "Failed to store file")
		return
	}
	if written > maxUploadSizeBytes {
		_ = os.Remove(dstPath)
		rt.sendError(w, http.StatusBadRequest, "File too large (max 10MB)")
		return
	}

	publicURI := rt.publicURL(filepath.ToSlash(dstPath))
	if err := rt.db.UpdateGroupPhoto(groupID, publicURI); err != nil {
		_ = os.Remove(dstPath)
		ctx.Logger.WithError(err).Error("failed to persist group photo uri")
		rt.sendError(w, http.StatusInternalServerError, "Failed to update group photo")
		return
	}

	_ = writeJSON(w, http.StatusOK, map[string]interface{}{
		"code":    http.StatusOK,
		"message": "Group photo updated successfully",
		"data": map[string]interface{}{
			"file": map[string]interface{}{
				"filename": origName,
				"size":     written,
				"uri":      publicURI, // match OpenAPI schema
			},
		},
	})
}
