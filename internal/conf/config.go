package conf

import (
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"os"
)

type Config struct {
}

func GetConfig(logger *zap.Logger) *Config {
	logger.Info("configuring application")
	v := viper.New()
	v.SetConfigName("config")
	v.AddConfigPath(".")
	v.SetConfigType("yaml")

	err := v.ReadInConfig()
	if err != nil {
		logger.Error("could not read config",
			zap.Error(err),
		)
		os.Exit(1)
	}
	logger.Info("configuration file",
		zap.String("file", v.ConfigFileUsed()),
	)
	return &Config{}
}
