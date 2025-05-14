package api

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/GioiaZheng/Wasa_proj/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

func (rt *_router) setGroupPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	groupID := ps.ByName("groupId")

	err := r.ParseMultipartForm(10 << 20) // 10MB
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid multipart form")
		return
	}

	file, handler, err := r.FormFile("photo")
	if err != nil {
		writeError(w, http.StatusBadRequest, "Photo file is required")
		return
	}
	defer file.Close()

	savePath := filepath.Join("uploads", "group_photos", fmt.Sprintf("%s_%s", groupID, handler.Filename))

	dst, err := os.Create(savePath)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to save photo")
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to write photo")
		return
	}

	err = rt.db.SetGroupPhoto(groupID, "/static/group_photos/"+filepath.Base(savePath))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to update group photo")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "Group photo updated successfully",
	})
}
