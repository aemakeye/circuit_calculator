package storage

import (
	"github.com/aemakeye/circuit_calculator/internal/storage"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
	"net/http"
)

const (
	uploadUrl = "/api/upload"
	listUrl   = "/api/ls/{project}"
)

type Handler struct {
	Logger  *zap.Logger
	Storage storage.ObjectStorage
}

func (h *Handler) Register(r chi.Router) {
	r.Route(uploadUrl, func(r chi.Router) {
		r.Post("/", uploadFile)
	})
	r.Route(listUrl, func(r chi.Router) {
		r.Get("/", listProjectFiles)
	})
}

func uploadFile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_ = ctx

}

func listProjectFiles(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_ = ctx
	project := chi.URLParam(r, "project")
	_ = project

}
