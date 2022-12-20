package storage

import (
	"bytes"
	"encoding/json"
	"github.com/aemakeye/circuit_calculator/internal/storage"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"go.uber.org/zap"
	"net/http"
	"time"
)

const (
	loadUrl         = "/api/ostorage/load"
	listUrl         = "/api/ostorage/ls"
	uploadUrl       = "/api/ostorage/upload"
	FormFileBody    = "uploadData"
	DeadLineTimeOut = 10 * time.Second
)

type Handler struct {
	Logger  *zap.Logger
	Storage storage.ObjectStorage
}

type LsResponse struct {
	LsItems []string `json:"projects"`
}

func (h *Handler) Register(r chi.Router) {
	r.Use(middleware.Timeout(DeadLineTimeOut))
	r.Route(loadUrl, func(r chi.Router) {
		r.Get("/{filename}", h.LoadFile)
		r.Get("/{filename}/", h.LoadFile)
	})
	r.Route(uploadUrl, func(r chi.Router) {
		r.Post("/", h.UploadFile)
		r.Post("/{project}", h.UploadFile)
		r.Post("/{project}/", h.UploadFile)
	})
	r.Route(listUrl, func(r chi.Router) {
		r.Get("/{project}/", h.ListProjectFiles)
		r.Get("/{project}", h.ListProjectFiles)
		r.Get("/", h.ListProjectFiles)
	})

}

// UploadFile upload file to storage
func (h *Handler) UploadFile(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	err := r.ParseMultipartForm(10 << 20) // 10MB max
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	//w.Header().Set("Content-Type", "application/json")

	project := chi.URLParam(r, "project")
	// do not allow to write to root of the bucket.
	if project == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	file, mpfhandler, err := r.FormFile(FormFileBody)
	if err != nil {
		h.Logger.Error("failed to upload file",
			zap.Error(err),
		)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer file.Close()

	err = h.Storage.UploadTextFile(r.Context(), h.Logger, file, project+"/"+mpfhandler.Filename)
	if err != nil {
		h.Logger.Error("failed to upload file",
			zap.Error(err),
		)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	return
}

// ListProjectFiles lists existing projects or project content
// returns http.StatusNotFound in case project name does not exist in storage
func (h *Handler) ListProjectFiles(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	project := chi.URLParam(r, "project")

	var items LsResponse
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

func (h *Handler) LoadFile(w http.ResponseWriter, r *http.Request) {

}
