package handlers

import (
	"github.com/go-chi/chi"
)

type Handler interface {
	Register(router chi.Router)
}

// TODO: "github.com/go-chi/cors"
