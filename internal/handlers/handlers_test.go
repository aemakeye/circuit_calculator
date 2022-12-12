package handlers

import (
	"bytes"
	"context"
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

	t.Run("test list project contents, test project name exact match", func(t *testing.T) {
		type ProjectLs []string
		var jb map[string]ProjectLs
		r := chi.NewRouter()

		// trick not to lose context values here
		// as context gets vanished while testing
		// testing /api/ls/{project} here
		// normaly {project} goes to context
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("project", "test")
		h.Register(r)

		req := httptest.NewRequest(http.MethodGet, "/api/ls/dummypath", nil)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		w := httptest.NewRecorder()
		h.ListProjectFiles(w, req)
		res := w.Result()
		defer res.Body.Close()

		err = json.NewDecoder(res.Body).Decode(&jb)
		assert.NoError(t, err)
		assert.Contains(t, jb, "projects")
		assert.Contains(t, jb["projects"], "test/test-diagram0.xml")
	})

	t.Run("learn json decode/encode", func(t *testing.T) {
		//
		type ProjectLs []string

		var jb map[string]ProjectLs
		resp := []byte(`{"projects": ["test/diagram.xml", "test/diagram2.xml", "test/diagram3.xml"]}`)

		json.NewDecoder(bytes.NewReader(resp)).Decode(&jb)
		t.Logf("%v", jb)

		buf := new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(jb)

		assert.NoError(t, err)

		t.Logf("%s", buf)

	})

	t.Run("return error  on bad project name", func(t *testing.T) {

		r := chi.NewRouter()

		// trick not to lose context values here
		// as context gets vanished while testing
		// testing /api/ls/{project} here
		// normaly {project} goes to context
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("project", "test-not-exist")
		h.Register(r)

		req := httptest.NewRequest(http.MethodGet, "/api/ls/dummypath", nil)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		w := httptest.NewRecorder()
		h.ListProjectFiles(w, req)
		res := w.Result()
		defer res.Body.Close()

		assert.NotEqual(t, res.StatusCode, http.StatusOK)

	})
}
