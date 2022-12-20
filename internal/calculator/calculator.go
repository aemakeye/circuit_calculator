package calculator

import (
	"bytes"
	"context"
	"github.com/aemakeye/circuit_calculator/internal/config"
	"github.com/aemakeye/circuit_calculator/internal/drawio"
	"github.com/aemakeye/circuit_calculator/internal/storage"
	"go.uber.org/zap"
	"sync"
)

type Calculator struct {
	Logger           *zap.Logger
	Config           *config.CConfig
	Gstorage         storage.GraphStorage
	TextStorage      storage.ObjectStorage
	DiagramConverter *drawio.Controller
}

var instance *Calculator
var once sync.Once

func NewCalculator(logger *zap.Logger, config *config.CConfig, gs storage.GraphStorage, os storage.ObjectStorage) (*Calculator, error) {
	once.Do(func() {
		logger.Info("creating Calculator instance")
		instance = &Calculator{
			Logger:           logger,
			Config:           config,
			Gstorage:         gs,
			TextStorage:      os,
			DiagramConverter: drawio.NewController(logger),
		}
	})

	return instance, nil
}

type DiagramService interface {
	ReadInDiagram(ctx context.Context, logger *zap.Logger, xmldoc *bytes.Reader) (uuid string, items []Item, err error)
	UpdateDiagram(ctx context.Context, logger *zap.Logger, diaUUID string) error
}
