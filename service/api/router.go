package api

import (
	"net/http"

	"github.com/gorilla/mux"
)

// _router is the main API router.
type _router struct {
	router *mux.Router
}

// NewRouter creates and returns a new API router.
func NewRouter() *_router {
	r := &_router{
		router: mux.NewRouter().StrictSlash(true),
	}

	r.RegisterHandlers()
	return r
}

// ServeHTTP allows _router to satisfy the http.Handler interface.
func (r *_router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.router.ServeHTTP(w, req)
}
