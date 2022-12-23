package calculator

import (
	"github.com/aemakeye/circuit_calculator/internal/config"
	"go.uber.org/zap"
	"sync"
)

type Calculator struct {
	Logger      *zap.Logger
	Config      *config.CConfig
	Gstorage    GraphStorage
	TextStorage ObjectStorage
	DiagramSvc  DiagramService
}

var instance *Calculator
var once sync.Once

func NewCalculator(logger *zap.Logger, config *config.CConfig, gs GraphStorage, os ObjectStorage) (*Calculator, error) {
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
