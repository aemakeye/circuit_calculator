package calculator

import (
	"bytes"
	"context"
	"github.com/aemakeye/circuit_calculator/internal/config"
	"github.com/aemakeye/circuit_calculator/internal/storage"
	"go.uber.org/zap"
	"sync"
)

type calculator struct {
	logger      *zap.Logger
	config      *config.CConfig
	gstorage    storage.GraphStorage
	textStorage storage.ObjectStorage
}

var instance *calculator
var once sync.Once

func NewCalculator(logger *zap.Logger, config *config.CConfig, gs storage.GraphStorage, os storage.ObjectStorage) (*calculator, error) {
	once.Do(func() {
		logger.Info("creating calculator instance")
		instance = &calculator{
			logger:      logger,
			config:      config,
			gstorage:    gs,
			textStorage: os,
		}
	})

	return instance, nil
}

type DiagramService interface {
	ReadInDiagram(ctx context.Context, logger *zap.Logger, xmldoc *bytes.Reader) (uuid string, items []storage.Item, err error)
	UpdateDiagram(ctx context.Context, logger *zap.Logger, diaUUID string) error
}
