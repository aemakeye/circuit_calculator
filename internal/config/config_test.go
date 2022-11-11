package config

import (
	"bytes"
	"github.com/magiconair/properties/assert"
	"go.uber.org/zap"
	"os"
	"testing"
)

func TestConfig(t *testing.T) {
	var input = []byte(`
			{
              "Loglevel": "INFO",
			  "Neo4j":
			  {
				"host": "localhost",
				"port": "7687",
				"User": "Neo4j",
				"Password": "Password",
				"schema": ""
			
			  },
			  "Minio":
			  {  	
				"host": "localhost:1234",	
				"User": "Minio",
				"Password": "Password",
				"schema": ""
			  }
			}
			`)
	t.Run("check env + CConfig file", func(t *testing.T) {

		env := map[string]string{
			"CALC_LOGLEVEL": "DEBUG",
		}

		logger := zap.NewNop()
		for k, v := range env {
			_ = os.Setenv(k, v)
		}
		cstring := bytes.NewReader(input)
		cfg, _ := NewConfig(logger, cstring)
		assert.Equal(t, cfg.Neo4j.User, "Neo4j")
		assert.Equal(t, cfg.Loglevel, "DEBUG")
		assert.Equal(t, "yes", "yes")
	})

	t.Run("test minio configuration options", func(t *testing.T) {
		logger := zap.NewNop()
		cstring := bytes.NewReader(input)
		cfg, _ := NewConfig(logger, cstring)
		assert.Equal(t, cfg.Minio.EndpointURL().String(), "http://localhost:1234")
	})
}
