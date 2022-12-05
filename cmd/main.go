package main

import (
	"fmt"
	"github.com/aemakeye/circuit_calculator/internal/config"
	"github.com/aemakeye/circuit_calculator/internal/handlers/storage"
	"github.com/aemakeye/circuit_calculator/internal/shutdown"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"net"
	"net/http"
	"os"
	"syscall"
	"time"
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
	if err != nil {
		logger.Fatal("error instantiating config",
			zap.Error(err),
		)
	}

	router := chi.NewRouter()
	router.Use(middleware.Logger)

	storageHandler := storage.Handler{
		Logger:  logger,
		Storage: cfg.Storage,
	}

	storageHandler.Register(router)

	start(router, logger, cfg)
}

func start(router chi.Router, logger *zap.Logger, cfg *config.CConfig) {
	var server *http.Server
	var listener net.Listener

	logger.Info("calculator API listen:",
		zap.String("ip", cfg.Listen.Addr().String()),
		zap.Uint16("port", cfg.Listen.Port()),
	)

	listener, err := net.Listen("tcp", cfg.Listen.String())
	if err != nil {
		logger.Fatal("Could not start api server",
			zap.Error(err),
		)
	}

	server = &http.Server{
		Addr:         cfg.Listen.String(),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go shutdown.Graceful(logger,
		[]os.Signal{
			syscall.SIGABRT,
			syscall.SIGQUIT,
			syscall.SIGHUP,
			os.Interrupt,
			syscall.SIGTERM,
		}, server)

	if err = server.Serve(listener); err != nil {
		switch {
		case errors.Is(err, http.ErrServerClosed):
			logger.Info("server shutdown")
		default:
			logger.Fatal("fatal error",
				zap.Error(err),
			)
		}
	}
}
