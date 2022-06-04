package config

import (
	"bytes"
	"fmt"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// CConfig internal structure
type CConfig struct {
	Logger   *zap.Logger
	Loglevel string
	Neo4j    *neo4j
	Minio    *minio
	Filename string
}

// neo4j internal structure
type neo4j struct {
	User     string
	Password string
	Endpoint string
	//play with this and Neo4j structure
	//Timeout  time.Duration
}

// minio internal structure
type minio struct {
	user     string
	password string
	endpoint string
}

// Neo4j structure to unmarshal CConfig file
type Neo4j struct {
	User     string
	Password string
	Host     string
	Port     string
	Schema   string
	// TODO play with this
	//Timeout  time.Duration `yaml:"timeout" json:"timeout"`

}

// Minio structure to unmarshal CConfig file
type Minio struct {
	User     string
	Password string
	Host     string
	Port     string
	Schema   string
}

// FileConfig structure to unmarshal CConfig from file
type FileConfig struct {
	Loglevel string //`mapstructure:"CALC_LOGLEVEL" json:"Loglevel" yaml:"Loglevel"`
	Neo4j    Neo4j
	Minio    Minio
}

// NewConfig function to create CConfig object with viper from file or reader.
// reader should be JSON
func NewConfig(logger *zap.Logger, reader *bytes.Reader) (cfg *CConfig, err error) {

	cfg = &CConfig{
		Logger:   nil,
		Loglevel: "",
		Neo4j:    &neo4j{},
		Minio:    &minio{},
		Filename: "",
	}

	cfg.Logger = logger
	v := viper.New()
	v.SetEnvPrefix("calc")
	v.SetConfigType("env")
	v.AutomaticEnv()
	logger.Info("CConfig from environment variables",
		zap.Any("Loglevel", v.Get("Loglevel")),
	)

	if reader == nil {
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("/")
		v.SetConfigName("CConfig")

		err = v.ReadInConfig()
	} else {
		v.SetConfigType("json")
		err = v.ReadConfig(reader)
	}

	if err != nil {
		return nil, fmt.Errorf("could not read CConfig, error: %s", err)
	}

	var fc FileConfig
	err = v.Unmarshal(&fc)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CConfig with error: %s", err)
	}

	cfg.Loglevel = fc.Loglevel

	cfg.Neo4j = &neo4j{
		User:     fc.Neo4j.User,
		Password: fc.Neo4j.Password,
		Endpoint: fc.Neo4j.Host + ":" + fc.Neo4j.Port,
	}

	cfg.Minio = &minio{
		user:     fc.Minio.User,
		password: fc.Minio.Password,
		endpoint: fc.Minio.Host + ":" + fc.Minio.Port,
	}

	return cfg, nil
}
