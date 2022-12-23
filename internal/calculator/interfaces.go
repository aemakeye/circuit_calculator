package calculator

import (
	"bytes"
	"context"
	"github.com/aemakeye/circuit_calculator/internal/diagram"
	"go.uber.org/zap"
)

type DiagramService interface {
	ReadInDiagram(ctx context.Context, logger *zap.Logger, xmldoc *bytes.Reader) (uuid string, _ <-chan diagram.Item, err error)
	UpdateDiagram(ctx context.Context, logger *zap.Logger, diaUUID string) error
}
