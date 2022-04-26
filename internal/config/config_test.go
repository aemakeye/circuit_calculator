package config

import (
	"bytes"
	"github.com/magiconair/properties/assert"
	"go.uber.org/zap"
	"os"
	"testing"
)

func TestEnvOverride(t *testing.T) {
	t.Run("check env + CConfig file", func(t *testing.T) {

		env := map[string]string{
			"CALC_LOGLEVEL": "DEBUG",
		}
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
				"host": "localhost",
				"port": "1234",
				"User": "Minio",
				"Password": "Password",
				"schema": ""
			  }
			}
			`)
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

}
