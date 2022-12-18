package storage

import (
	"context"
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
	PushItem(logger *zap.Logger, item Item) (uuid string, id string, err error)
}
