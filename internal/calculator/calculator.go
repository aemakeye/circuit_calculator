package calculator

import (
	"bytes"
	"context"
	"github.com/aemakeye/circuit_calculator/internal/config"
	"go.uber.org/zap"
	"sync"
)

type calculator struct {
	logger   *zap.Logger
	config   *config.CConfig
	gstorage GraphStorage
	ostorage ObjectStorage
}

var instance *calculator
var once sync.Once

func NewCalculator(logger *zap.Logger, config *config.CConfig, gs GraphStorage, os ObjectStorage) (*calculator, error) {
	once.Do(func() {
		logger.Info("creating calculator instance")
		instance = &calculator{
			logger:   logger,
			config:   config,
			gstorage: gs,
			ostorage: os,
		}
	})

	return instance, nil
}

type DiagramService interface {
	ReadInDiagram(ctx context.Context, logger *zap.Logger, xmldoc *bytes.Reader) (uuid string, items []Item, err error)
	UpdateDiagram(ctx context.Context, logger *zap.Logger, diaUUID string) error
}

type GraphStorage interface {
	PushItem(logger *zap.Logger, item Item) (uuid string, id string, err error)
}

type ObjectStorage interface {
	SaveDiagram(ctx context.Context, logger *zap.Logger, doc *[]byte) error
	LoadDiagramByUUID(ctx context.Context, logger *zap.Logger, uuid string) ([]byte, error)
}
