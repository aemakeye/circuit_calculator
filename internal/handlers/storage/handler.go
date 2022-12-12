package storage

import (
	"bytes"
	"encoding/json"
	"github.com/aemakeye/circuit_calculator/internal/storage"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
	"net/http"
)

const (
	uploadUrl = "/api/upload"
	listUrl   = "/api/ls"
)

type Handler struct {
	Logger  *zap.Logger
	Storage storage.ObjectStorage
}

type storageLsResponse struct {
	LsItems []string `json:"projects"`
}

func (h *Handler) Register(r chi.Router) {
	r.Route(uploadUrl, func(r chi.Router) {
		r.Post("/", h.UploadFile)
	})
	r.Route(listUrl, func(r chi.Router) {
		r.Get("/{project}/", h.ListProjectFiles)
		r.Get("/{project}", h.ListProjectFiles)
		r.Get("/", h.ListProjectFiles)
	})
}

func (h *Handler) UploadFile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_ = ctx
}

func (h *Handler) ListProjectFiles(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	project := chi.URLParam(r, "project")

	var items storageLsResponse
	for item := range h.Storage.Ls(r.Context(), project) {
		items.LsItems = append(items.LsItems, item)
	}

	if len(items.LsItems) == 0 {
		w.WriteHeader(http.StatusNotFound)
	}

	buf := new(bytes.Buffer)
	err := json.NewEncoder(buf).Encode(items)
	if err != nil {
		h.Logger.Error("error in json encoding of storage path listing",
			zap.String("path", project),
			zap.Error(err),
		)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error listing storage"))
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(buf.Bytes()); err != nil {
		h.Logger.Error("error writing response body",
			zap.Error(err),
		)
	}

}
