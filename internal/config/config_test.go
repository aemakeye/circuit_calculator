package config

import (
	"bytes"
	"context"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"os"
	"testing"
)

func TestConfig(t *testing.T) {
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
	t.Run("check env + CConfig file", func(t *testing.T) {

		env := map[string]string{
			"CALC_LOGLEVEL": "DEBUG",
		}

		logger := zap.NewNop()
		for k, v := range env {
			_ = os.Setenv(k, v)
		}
		cstring := bytes.NewReader(input)
		cfg, err := NewConfig(logger, cstring)
		assert.NoError(t, err)
		assert.Equal(t, cfg.Neo4j.User, "Neo4j")
		//assert.Equal(t, cfg.Loglevel, "DEBUG")
	})

	t.Run("test minio configuration options", func(t *testing.T) {
		//TODO: stupid
		tctx := context.Background()
		logger := zap.NewNop()
		cstring := bytes.NewReader(input)
		cfg, _ := NewConfig(logger, cstring)
		cfg_map := cfg.Storage.ConfigDump(tctx, logger)
		_ = cfg_map
		//t.Logf("map %s", cfg_map)
		//assert.Equal(t, "localhost:1234", cfg_map["url"])
	})
	t.Run("learn netip.Addr", func(t *testing.T) {
		logger := zap.NewNop()
		cstring := bytes.NewReader(input)
		cfg, _ := NewConfig(logger, cstring)

		t.Logf("netip.Addr string: %s", cfg.Listen.String())
	})
}
