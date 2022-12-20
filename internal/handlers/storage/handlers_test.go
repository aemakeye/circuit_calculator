package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/aemakeye/circuit_calculator/internal/config"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

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

	h := Handler{
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

		var jb LsResponse
		r := chi.NewRouter()

		// trick not to lose context values here
		// as context gets vanished while testing
		// testing /api/ls/{project} here
		// normally {project} goes to context
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
		t.Logf("%s", jb)
		assert.NoError(t, err)
	})

	t.Run("return error  on bad project name", func(t *testing.T) {

		r := chi.NewRouter()

		// trick not to lose context values here
		// as context gets vanished while testing
		// testing /api/ls/{project} here
		// normally {project} goes to context
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("project", "test-not-exist")
		h.Register(r)

		req := httptest.NewRequest(http.MethodGet, "/api/ls/dummypath", nil)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		w := httptest.NewRecorder()
		h.ListProjectFiles(w, req)
		res := w.Result()
		defer res.Body.Close()

		assert.NotEqual(t, http.StatusOK, res.StatusCode)
		assert.Equal(t, http.StatusNotFound, res.StatusCode)

	})

	type testUpload struct {
		urlParams    map[string]string
		expectedFail bool
		expectedCode int
	}

	vtestUpload := []testUpload{
		{urlParams: map[string]string{"project": "upload-test"}, expectedFail: false, expectedCode: http.StatusCreated},
		{urlParams: map[string]string{"project": ""}, expectedFail: true, expectedCode: http.StatusBadRequest},
	}

	for _, tc := range vtestUpload {
		t.Run("try upload to "+tc.urlParams["project"], func(t *testing.T) {

			r := chi.NewRouter()
			rctx := chi.NewRouteContext()
			//rctx.URLParams.Add("uploadData", "test.txt")
			rctx.URLParams.Add("project", tc.urlParams["project"])
			h.Register(r)

			bbuf := &bytes.Buffer{}
			writer := multipart.NewWriter(bbuf)
			fw, err := writer.CreateFormFile(FormFileBody, "test.txt")
			assert.NoError(t, err)
			_, err = io.Copy(fw, strings.NewReader("hello"))
			writer.Close()

			req := httptest.NewRequest(http.MethodPost, "/api/upload/upload-test/test.txt", bytes.NewReader(bbuf.Bytes()))
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			req.Header.Set("Content-Type", writer.FormDataContentType())
			w := httptest.NewRecorder()
			h.UploadFile(w, req)
			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tc.expectedCode, res.StatusCode)
		})

	}

	t.Run("learn json decode/encode", func(t *testing.T) {
		//

		var jb LsResponse
		resp := []byte(`{"projects": ["test/diagram.xml", "test/diagram2.xml", "test/diagram3.xml"]}`)

		json.NewDecoder(bytes.NewReader(resp)).Decode(&jb)

		t.Logf("encode: %v", jb)

		buf := new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(jb)

		assert.NoError(t, err)

		t.Logf("decode: %s", buf)

	})
}
