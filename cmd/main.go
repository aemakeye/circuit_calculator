package main

import (
	"fmt"
	"github.com/aemakeye/circuit_calculator/internal/config"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
	"net"
	"net/http"
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

func start(router chi.Router, logger *zap.Logger, cfg *config.CConfig) {
	var server *http.Server
	var listener net.Listener

	logger.Info("calculator API listen:",
		zap.String("ip", cfg.Listen.Addr().String()),
		zap.Uint16("port", cfg.Listen.Port()),
	)

	listener, err := net.Listen("tcp", cfg.Listen.String())
}
