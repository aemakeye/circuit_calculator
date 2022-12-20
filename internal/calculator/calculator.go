package calculator

import (
	"github.com/aemakeye/circuit_calculator/internal/config"
	"github.com/aemakeye/circuit_calculator/internal/storage"
	"go.uber.org/zap"
	"sync"
)

type Calculator struct {
	Logger      *zap.Logger
	Config      *config.CConfig
	Gstorage    storage.GraphStorage
	TextStorage storage.ObjectStorage
	DiagramSvc  DiagramService
}

var instance *Calculator
var once sync.Once

func NewCalculator(logger *zap.Logger, config *config.CConfig, gs storage.GraphStorage, os storage.ObjectStorage) (*Calculator, error) {
	once.Do(func() {
		logger.Info("creating Calculator instance")
		instance = &Calculator{
			Logger:      logger,
			Config:      config,
			Gstorage:    gs,
			TextStorage: os,
			DiagramSvc:  config.DiagramSvc,
		}
	})

	return instance, nil
}
