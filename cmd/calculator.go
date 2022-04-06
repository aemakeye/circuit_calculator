package main

import (
	"fmt"
	"github.com/HeOpuHaMeH9I/CircuitCalculator/internal/config"
	"go.uber.org/zap"
	"os"
	//"github.com/HeOpuHaMeH9I/CirquitCalculator/internal/shutdown/shutdown.go"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Println("can't initialize logger")
		os.Exit(1)
	}
	logger.Info("Starting CircuitCalculator")

	cfg := config.GetConfig()

}
