package config

import (
	"bytes"
	"fmt"
	"github.com/aemakeye/circuit_calculator/internal/calculator"
	"github.com/aemakeye/circuit_calculator/internal/drawio"
	"github.com/aemakeye/circuit_calculator/internal/minio"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"net/netip"
)

const (
	defaultApiListen = "127.0.0.1:8099"
)

// CConfig internal structure
type CConfig struct {
	DiagramSvc calculator.DiagramProcessor
	Listen     netip.AddrPort
	Logger     *zap.Logger
	Loglevel   string
	Neo4j      *neo4j
	Storage    calculator.ObjectStorage
	Filename   string
}

// neo4j internal structure
type neo4j struct {
	User     string
	Password string
	Endpoint string
	//play with this and Neo4j structure
	//Timeout  time.Duration
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
	User     string `yaml:"User" json:"user"`
	Password string `yaml:"Password" json:"password"`
	Host     string `yaml:"Host" json:"host"`
	Secure   bool   `yaml:"Secure" json:"secure"`
	Bucket   string `yaml:"Bucket" json:"bucket"`
}

// FileConfig structure to unmarshal CConfig from file
type FileConfig struct {
	Loglevel      string //`mapstructure:"CALC_LOGLEVEL" json:"Loglevel" yaml:"Loglevel"`
	Listen        string `yaml:"listen" json:"Listen"`
	Neo4j         Neo4j  `yaml:"neo4j" json:"Neo4J"`
	ObjectStorage struct {
		Minio *Minio `json:"minio,omitempty"`
		//	maybe another type of storage here
	} `json:"objectStorage"`
}

// NewConfig function to create CConfig object with viper from file or reader.
// reader should be JSON
func NewConfig(logger *zap.Logger, reader *bytes.Reader) (cfg *CConfig, err error) {

	cfg = &CConfig{
		DiagramSvc: drawio.NewController(logger),
		Logger:     nil,
		Loglevel:   "",
		Neo4j:      &neo4j{},
		Storage:    nil,
		Filename:   "",
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
		v.SetConfigName("config")

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

	cfg.Listen, err = netip.ParseAddrPort(fc.Listen)

	if err != nil {
		logger.Error("bad listen field in config. ip:port expected. Setting default.")
		cfg.Listen, _ = netip.ParseAddrPort(defaultApiListen)
	}

	cfg.Neo4j = &neo4j{
		User:     fc.Neo4j.User,
		Password: fc.Neo4j.Password,
		Endpoint: fc.Neo4j.Host + ":" + fc.Neo4j.Port,
	}
	switch {
	case fc.ObjectStorage.Minio != nil:
		strg, err := minio.NewMinioStorage(
			cfg.Logger,
			fc.ObjectStorage.Minio.Host,
			fc.ObjectStorage.Minio.Bucket,
			fc.ObjectStorage.Minio.User,
			fc.ObjectStorage.Minio.Password,
			fc.ObjectStorage.Minio.Secure,
		)
		if err != nil {
			logger.Error("could not initialize minio storage",
				zap.Error(err),
			)
			return nil, err
		}
		cfg.Storage = strg
	default:
		logger.Fatal("no object storage defined")
	}

	return cfg, nil
}
