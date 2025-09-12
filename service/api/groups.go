package api

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/GioiaZheng/Wasa_proj/service/models"
	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/gofrs/uuid"
	"github.com/julienschmidt/httprouter"
)

// -------------------------------
// Group: Create
// POST /groups
// -------------------------------

// CreateGroupRequest defines the request body for creating a group.
type CreateGroupRequest struct {
	Name    string   `json:"name"`
	Members []string `json:"members"`
}

// createGroup creates a new group and adds the creator as a member.
// Notes:
// - Keep function signature required by rt.wrap (4 params).
// - Use ctx.UserID for authentication info injected by the wrapper.
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

	// Normalize members (dedupe, strip spaces) and include creator.
	seen := map[string]bool{}
	members := make([]string, 0, len(req.Members)+1)
	push := func(id string) {
		id = strings.TrimSpace(id)
		if id == "" || seen[id] {
			return
		}
		seen[id] = true
		members = append(members, id)
	}
	for _, m := range req.Members {
		push(m)
	}
	if !seen[userID] {
		push(userID)
	}
	sort.Strings(members)

	groupName := strings.TrimSpace(req.Name)
	if groupName == "" {
		groupName = "Group"
	}

	// Generate group ID (schema uses TEXT PRIMARY KEY).
	gid, err := uuid.NewV4()
	if err != nil {
		ctx.Logger.WithError(err).Error("failed to generate group id")
		rt.sendError(w, http.StatusInternalServerError, "Failed to generate group id")
		return
	}
	group := models.Group{ID: gid.String(), Name: groupName}

	// Create group row, then add members.
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

	_ = writeJSON(w, http.StatusCreated, map[string]interface{}{
		"code":    http.StatusCreated,
		"message": "Group created",
		"data": map[string]interface{}{
			"group": map[string]interface{}{
				"id":      group.ID,
				"name":    group.Name,
				"members": members,
			},
		},
	})
}

// -------------------------------
// Group: Get by ID / Detail
// GET /groups/:id
// -------------------------------

// getGroup returns minimal info for a single group.
func (rt *_router) getGroup(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	if strings.TrimSpace(ctx.UserID) == "" {
		rt.sendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	groupID := strings.TrimSpace(ps.ByName("id"))
	if groupID == "" {
		rt.sendError(w, http.StatusBadRequest, "missing group id")
		return
	}

	group, err := rt.db.GetGroup(groupID)
	if err != nil {
		rt.sendError(w, http.StatusNotFound, "group not found")
		return
	}

	_ = writeJSON(w, http.StatusOK, map[string]interface{}{
		"code": http.StatusOK,
		"data": group,
	})
}

// getGroupDetail can return the same payload as getGroup for now.
// This keeps compatibility with routes that expect "detail".
func (rt *_router) getGroupDetail(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	rt.getGroup(w, r, ps, ctx)
}

// -------------------------------
// Group: List mine
// GET /groups
// -------------------------------

func (rt *_router) getGroupsList(
	w http.ResponseWriter,
	r *http.Request,
	_ httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	uid := strings.TrimSpace(ctx.UserID)
	if uid == "" {
		rt.sendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	groups, err := rt.db.GetGroupsList(uid)
	if err != nil {
		rt.sendError(w, http.StatusInternalServerError, "failed to fetch groups")
		return
	}

	_ = writeJSON(w, http.StatusOK, map[string]interface{}{
		"code": http.StatusOK,
		"data": groups,
	})
}

// -------------------------------
// Group: Rename
// PUT /groups/:id/name
// -------------------------------

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
		rt.sendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	groupID := strings.TrimSpace(ps.ByName("id"))
	if groupID == "" {
		rt.sendError(w, http.StatusBadRequest, "missing group id")
		return
	}

	var req UpdateGroupNameRequest
	if err := readJSON(r, &req); err != nil {
		rt.sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		rt.sendError(w, http.StatusBadRequest, "name is required")
		return
	}

	if err := rt.db.UpdateGroupName(groupID, req.Name); err != nil {
		rt.sendError(w, http.StatusInternalServerError, "failed to update group name")
		return
	}

	_ = writeJSON(w, http.StatusOK, map[string]interface{}{
		"code":    http.StatusOK,
		"message": "Group name updated",
	})
}

// updateGroupName is kept for backward compatibility; delegates to setGroupName.
func (rt *_router) updateGroupName(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	rt.setGroupName(w, r, ps, ctx)
}

// -------------------------------
// Group: Leave
// DELETE /groups/:id/members (current user)
// -------------------------------

func (rt *_router) leaveGroup(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	uid := strings.TrimSpace(ctx.UserID)
	if uid == "" {
		rt.sendError(w, http.StatusUnauthorized, "missing or invalid token")
		return
	}
	groupID := strings.TrimSpace(ps.ByName("id"))
	if groupID == "" {
		rt.sendError(w, http.StatusBadRequest, "missing group id")
		return
	}

	if err := rt.db.LeaveGroup(groupID, uid); err != nil {
		rt.sendError(w, http.StatusInternalServerError, "failed to leave group")
		return
	}

	_ = writeJSON(w, http.StatusOK, map[string]interface{}{
		"code":    http.StatusOK,
		"message": "left group successfully",
	})
}

// -------------------------------
// Group: Add members
// POST /groups/:id/members
// -------------------------------

type AddGroupMembersRequest struct {
	MemberIDs []string `json:"member_ids,omitempty"`
	LegacyUID string   `json:"userId,omitempty"` // legacy compatibility
}

func (rt *_router) addToGroup(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	if strings.TrimSpace(ctx.UserID) == "" {
		rt.sendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	groupID := strings.TrimSpace(ps.ByName("id"))
	if groupID == "" {
		rt.sendError(w, http.StatusBadRequest, "missing group id")
		return
	}

	var req AddGroupMembersRequest
	if err := readJSON(r, &req); err != nil {
		rt.sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Normalize + dedupe.
	seen := make(map[string]struct{})
	list := make([]string, 0, len(req.MemberIDs)+1)
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
	if len(list) == 0 && strings.TrimSpace(req.LegacyUID) != "" {
		push(req.LegacyUID)
	}

	if len(list) == 0 {
		rt.sendError(w, http.StatusBadRequest, "member_ids is required (or legacy userId)")
		return
	}
	sort.Strings(list)

	if err := rt.db.AddGroupMembers(groupID, list); err != nil {
		ctx.Logger.WithError(err).Error("failed to add members to group")
		rt.sendError(w, http.StatusInternalServerError, "failed to add member(s) to group")
		return
	}

	_ = writeJSON(w, http.StatusOK, map[string]interface{}{
		"code":    http.StatusOK,
		"message": "members added",
		"data":    map[string]interface{}{"added": list},
	})
}

// addGroupMember is a backward-compatible alias that delegates to addToGroup.
func (rt *_router) addGroupMember(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	rt.addToGroup(w, r, ps, ctx)
}

// -------------------------------
// Group: List members
// GET /groups/:id/members
// -------------------------------

func (rt *_router) getGroupMembers(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	if strings.TrimSpace(ctx.UserID) == "" {
		rt.sendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	groupID := strings.TrimSpace(ps.ByName("id"))
	if groupID == "" {
		rt.sendError(w, http.StatusBadRequest, "missing group id")
		return
	}

	members, err := rt.db.GetGroupMembers(groupID)
	if err != nil {
		rt.sendError(w, http.StatusInternalServerError, "failed to fetch group members")
		return
	}

	_ = writeJSON(w, http.StatusOK, map[string]interface{}{
		"code": http.StatusOK,
		"data": members,
	})
}

// -------------------------------
// Group: Set/Update photo
// PUT /groups/:id/photo
// Support two modes:
//   - Preset:  ?preset=avatar7 -> /uploads/photos/avatar7.jpg
//   - Upload:  multipart/form-data field "upload"
// -------------------------------

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

	// --- 1) PRESET MODE via query param ---
	if preset := strings.TrimSpace(r.URL.Query().Get("preset")); preset != "" {
		if !strings.HasPrefix(strings.ToLower(preset), "avatar") {
			rt.sendError(w, http.StatusBadRequest, "invalid preset name")
			return
		}
		derived := preset + ".jpg"
		publicURL := rt.publicURL(filepath.ToSlash(filepath.Join("uploads", "photos", derived)))
		if err := rt.db.UpdateGroupPhoto(groupID, publicURL); err != nil {
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
					"url":      publicURL,
				},
			},
		})
		return
	}

	// --- 2) UPLOAD MODE ---
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

	publicPath := rt.publicURL(filepath.ToSlash(dstPath))
	if err := rt.db.UpdateGroupPhoto(groupID, publicPath); err != nil {
		_ = os.Remove(dstPath)
		ctx.Logger.WithError(err).Error("failed to persist group photo url")
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
				"url":      publicPath,
			},
		},
	})
}

// setGroupPhotoCompat is a thin wrapper for backward compatibility.
func (rt *_router) setGroupPhotoCompat(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	rt.setGroupPhoto(w, r, ps, ctx)
}

// updateGroupPhoto is a backward-compatible alias to setGroupPhoto.
func (rt *_router) updateGroupPhoto(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	rt.setGroupPhoto(w, r, ps, ctx)
}
