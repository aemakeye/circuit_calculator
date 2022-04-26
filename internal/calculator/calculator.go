package calculator

import (
	"context"
	"github.com/aemakeye/circuit_calculator/internal/config"
	"go.uber.org/zap"
)

type calculator struct {
	logger *zap.Logger
	config *config.CConfig
}

type Service interface {
	New(logger *zap.Logger, cConfig *config.CConfig) (*calculator, error)
	GetDiagramByUUID(ctx context.Context, logger *zap.Logger, diaUUID string) (*Diagram, error)
	NewDiagram(ctx context.Context, logger *zap.Logger) (uuid string, err error)
	UpdateDiagram(ctx context.Context, logger *zap.Logger, diaUUID string) error
}

func (c *calculator) New(logger *zap.Logger, config *config.CConfig) (*calculator, error) {
	return &calculator{
		logger: logger,
		config: config,
	}, nil
}

func (c *calculator) GetDiagramByUUID(ctx context.Context, logger *zap.Logger, diaUUID string) (*Diagram, error) {
	//TODO implement me
	panic("implement me")
}

func (c *calculator) NewDiagram(ctx context.Context, logger *zap.Logger) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (c *calculator) UpdateDiagram(ctx context.Context, logger *zap.Logger, diaUUID string) error {
	//TODO implement me
	panic("implement me")
}
