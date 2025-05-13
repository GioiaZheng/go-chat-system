package api

import (
	"net/http"

	"github.com/GioiaZheng/Wasa_proj/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// httpRouterHandler 定义
type httpRouterHandler func(http.ResponseWriter, *http.Request, httprouter.Params, reqcontext.RequestContext)

// wrap 包装函数，使 handler 可以使用 RequestContext
func (rt *_router) wrap(fn httpRouterHandler) func(http.ResponseWriter, *http.Request, httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		// 创建 RequestContext
		ctx := reqcontext.RequestContext{
			UserID: GetUserIDFromContext(r.Context()),
		}

		// 调用实际的 handler
		fn(w, r, ps, ctx)
	}
}
