package main

import (
	"fmt"
	"github.com/aemakeye/circuit_calculator/internal/conf"
	"go.uber.org/zap"
	"os"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Println("can't initialize logger")
		os.Exit(1)
	}
	logger.Info("Starting CircuitCalculator")

}
