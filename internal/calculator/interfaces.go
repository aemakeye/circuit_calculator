package calculator

import (
	"bytes"
	"context"
	"go.uber.org/zap"
)

type DiagramService interface {
	ReadInDiagram(ctx context.Context, logger *zap.Logger, xmldoc *bytes.Reader) (uuid string, _ <-chan Item, err error)
	UpdateDiagram(ctx context.Context, logger *zap.Logger, diaUUID string) error
}
