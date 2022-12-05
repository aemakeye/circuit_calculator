package handlers

import (
	"bytes"
	"encoding/json"
	"github.com/aemakeye/circuit_calculator/internal/config"
	"github.com/aemakeye/circuit_calculator/internal/handlers/storage"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLearnBasics(t *testing.T) {
	t.Run("learn JSON marshal", func(t *testing.T) {
		type target struct {
			Names []string `json:"names"`
		}

		trg := target{Names: []string{"name1", "name2", "name3"}}

		jt, err := json.Marshal(trg)
		assert.NoError(t, err)

		t.Logf("json: \n%s", jt)
	})
}

func TestHandlers(t *testing.T) {
	var input = []byte(`
			{
              "Loglevel": "INFO",
              "Listen": "0.0.0.0:8099",
			  "Neo4j":
			  {
				"host": "localhost",
				"port": "7687",
				"User": "Neo4j",
				"Password": "Password",
				"schema": ""
			
			  },
              "ObjectStorage": 
				{
				  "Minio":
				  {  	
					"host": "localhost:9000",	
					"User": "calculator",
					"Password": "c@1cu1@t0r",
					"secure": "",
					"bucket": "calculator"
				  }
				}
			}
			`)

	logger := zap.NewNop()
	cfg, err := config.NewConfig(logger, bytes.NewReader(input))
	assert.NoError(t, err)

	h := storage.Handler{
		Logger:  logger,
		Storage: cfg.Storage,
	}

	t.Run("test list projects (top level directories)", func(t *testing.T) {
		r := chi.NewRouter()
		h.Register(r)

		req := httptest.NewRequest(http.MethodGet, "/api/ls", nil)
		w := httptest.NewRecorder()
		h.ListProjectFiles(w, req)
		res := w.Result()
		defer res.Body.Close()

		b, err := io.ReadAll(res.Body)
		assert.NoError(t, err)
		t.Logf("listing %s", b)
	})

	t.Run("test list project contents", func(t *testing.T) {
		r := chi.NewRouter()
		h.Register(r)
		req := httptest.NewRequest(http.MethodGet, "/api/ls/tdst", nil)
		w := httptest.NewRecorder()
		h.ListProjectFiles(w, req)
		res := w.Result()
		defer res.Body.Close()

		b, err := io.ReadAll(res.Body)
		assert.NoError(t, err)
		t.Logf("listing %s", b)
	})
}
