package config

import (
	"fmt"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// internal structure for config
type config struct {
	logger   *zap.Logger
	loglevel string
	neo4j    neo4j
	minio    minio
}

type neo4j struct {
	user     string
	password string
	endpoint string
}

type minio struct {
	user     string
	password string
	endpoint string
}

// FileConfig structure to unmarshal config from file
type FileConfig struct {
	LogLevel string `yaml:"log_level"`
	Neo4j    struct {
		user     string `yaml:"user"`
		password string `yaml:"password"`
		host     string `yaml:"host"`
		port     string `yaml:"port"`
		schema   string `yaml:"schema"`
	} `yaml:"neo4j"`
	Minio struct {
		user     string `yaml:"user"`
		password string `yaml:"password"`
		host     string `yaml:"host"`
		port     string `yaml:"port"`
		schema   string `yaml:"schema"`
	} `yaml:"minio"`
}

func NewConfig(logger *zap.Logger) (cfg *config, err error) {
	cfg.logger = logger
	logger.Info("Configuring")
	v := viper.New()
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.SetConfigName("config")
	err = v.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("could not read config, error: %s", err)
	}

	var fc FileConfig
	err = v.Unmarshal(&fc)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config with error: %s", err)
	}

	cfg.loglevel = fc.LogLevel

	cfg.neo4j = neo4j{
		user:     fc.Neo4j.user,
		password: fc.Neo4j.password,
		endpoint: fc.Neo4j.host + ":" + fc.Neo4j.port,
	}

	cfg.minio = minio{
		user:     fc.Minio.user,
		password: fc.Minio.password,
		endpoint: fc.Minio.host + ":" + fc.Minio.port,
	}

	return cfg, nil
}
