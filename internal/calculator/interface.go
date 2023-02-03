package calculator

import (
	"bytes"
	"context"
	"github.com/aemakeye/circuit_calculator/internal/drawio"
	"go.uber.org/zap"
	"io"
)

type ObjectStorage interface {
	ConfigDump(ctx context.Context, logger *zap.Logger) map[string]string
	DeleteFile(ctx context.Context, logger *zap.Logger, path string) error
	UploadTextFile(ctx context.Context, logger *zap.Logger, r io.Reader, path string) error
	LoadFileByName(ctx context.Context, logger *zap.Logger, path string, version string) (io.Reader, error)
	IsVersioned(ctx context.Context) bool
	Ls(ctx context.Context, path string) <-chan string
	LsVersions(ctx context.Context, path string, logger *zap.Logger) (<-chan string, error)
}

type GraphStorage interface {
	// TODO: decom PushItem
	//PushItem(logger *zap.Logger, item storage.Item) (uuid string, id string, err error)
	PushItems(logger *zap.Logger, items <-chan drawio.Item, pr chan drawio.Item, noMoreItems chan struct{})
}

type DiagramProcessor interface {
	XmlToItems(ctx context.Context, logger *zap.Logger, xmldoc *bytes.Reader, ch chan drawio.Item) (uuid string, err error)
}
