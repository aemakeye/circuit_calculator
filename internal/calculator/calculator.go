package calculator

import (
	"bytes"
	"context"
	"github.com/aemakeye/circuit_calculator/internal/config"
	"go.uber.org/zap"
	"sync"
)

type calculator struct {
	logger      *zap.Logger
	config      *config.CConfig
	gstorage    GraphStorage
	textStorage TextStorage
}

var instance *calculator
var once sync.Once

func NewCalculator(logger *zap.Logger, config *config.CConfig, gs GraphStorage, os TextStorage) (*calculator, error) {
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
	ReadInDiagram(ctx context.Context, logger *zap.Logger, xmldoc *bytes.Reader) (uuid string, items []Item, err error)
	UpdateDiagram(ctx context.Context, logger *zap.Logger, diaUUID string) error
}

type GraphStorage interface {
	PushItem(logger *zap.Logger, item Item) (uuid string, id string, err error)
}

type TextStorage interface {
	UploadDiagram(ctx context.Context, logger *zap.Logger, dia *Diagram) error
	LoadDiagramByUUID(ctx context.Context, logger *zap.Logger, uuid string, version string) ([]byte, error)
	LoadDiagramByName(ctx context.Context, logger *zap.Logger, name string, version string) ([]byte, error)
	IsVersioned(ctx context.Context) bool
	Ls(ctx context.Context) ([]string, error)
	LsVersions(ctx context.Context, diagram *Diagram) ([]DiagramVersion, error)
}
