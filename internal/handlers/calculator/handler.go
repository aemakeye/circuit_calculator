package calculator

import (
	"github.com/aemakeye/circuit_calculator/internal/calculator"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"go.uber.org/zap"
	"net/http"
	"time"
)

const (
	FormFileBody     = "uploadData"
	uploadDiagramUrl = "/api/uploadDiagram"
	uploadFileUrl    = "/api/uploadFile"
	DeadLineTimeOut  = 10 * time.Second
)

type Handler struct {
	Logger     *zap.Logger
	Calculator *calculator.Calculator
}

func (h *Handler) Register(r chi.Router) {
	r.Use(middleware.Timeout(DeadLineTimeOut))
	r.Route(uploadDiagramUrl, func(r chi.Router) {
		r.Post("/{project}", h.UploadDiagram)
		r.Post("/{project}/", h.UploadDiagram)
	})

	r.Route(uploadFileUrl, func(r chi.Router) {
		r.Post("/{project}", h.UploadFile)
		r.Post("/{project}/", h.UploadFile)
	})
}

func (h *Handler) UploadDiagram(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	project := chi.URLParam(r, "project")
	if project == "" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	file, mpfilehandler, err := r.FormFile(FormFileBody)
	_ = file
	_ = mpfilehandler
	if err != nil {
		h.Logger.Error("failed to upload file",
			zap.Error(err),
		)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	dctr := h.Calculator.DiagramSvc
	_ = dctr

	//uuid, err := h.Calculator.Gstorage.PushDiagram(h.Logger, file)

}

func (h *Handler) UploadFile(w http.ResponseWriter, r *http.Request) {

}
