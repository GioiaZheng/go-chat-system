package api

import (
	"net/http"

	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// GET /api/v1/messages
func (rt *_router) getMessages(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	rt.getConversation(w, r, ps, ctx)
}
