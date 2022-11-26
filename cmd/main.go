package main

import (
	"fmt"
	"github.com/aemakeye/circuit_calculator/internal/config"
	"go.uber.org/zap"
	"os"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Println("can't initialize logger")
		os.Exit(1)
	}
	defer func() {
		err := logger.Sync()
		if err != nil {
			fmt.Println("can't do final logger sync")
		}
	}()

	logger.Info("Starting Circuit Calculator")

	cfg, err := config.NewConfig(logger, nil)

	_ = cfg
}
