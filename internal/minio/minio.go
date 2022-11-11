package minio

import (
	"context"
	"fmt"
	"github.com/aemakeye/circuit_calculator/internal/calculator"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/zap"
	"strings"
	"sync"
)

var once sync.Once
var instance *minioStorage

type minioStorage struct {
	url      string
	user     string
	password string
	ssl      bool
	Client   *minio.Client
	Logger   *zap.Logger
	bucket   *minio.BucketInfo
}

func NewMinioStorage(logger *zap.Logger, url string, user string, password string, ssl bool) (instance *minioStorage, err error) {
	once.Do(func() {
		logger.Info("Creating calculator instance")
		instance = &minioStorage{
			url:      url,
			user:     user,
			password: password,
			ssl:      ssl,
			Client:   nil,
			Logger:   logger,
		}
		instance.Client, err = minio.New(instance.url, &minio.Options{
			Creds:  credentials.NewStaticV4(instance.user, instance.password, ""),
			Secure: instance.ssl,
		})
		if err != nil {
			instance.Logger.Fatal("can not connect to minio",
				zap.String("url", instance.url),
				zap.String("user", instance.user),
				zap.Bool("ssl_enabled", instance.ssl),
			)
		} else {
			instance.Logger.Info("minio backend connected successfully")
		}
	})
	return
}

func (m minioStorage) UploadDiagram(ctx context.Context, logger *zap.Logger, dia *calculator.Diagram) (err error) {
	//TODO: check if already exists and raise alert
	logger.Info("Diagram upload started",
		zap.String("bucket", m.bucket.Name),
		zap.String("name", dia.Name),
	)

	info, err := m.Client.PutObject(
		ctx,
		m.bucket.Name,
		"test-diagram.xml",
		strings.NewReader(dia.Body),
		int64(len(dia.Body)),
		minio.PutObjectOptions{ContentType: "application/octet-stream"},
	)
	if err != nil {
		logger.Error("Failed to upload diagram",
			zap.String("bucket", m.bucket.Name),
			zap.String("name", dia.Name),
			zap.Error(err),
		)
		return err
	} else {
		logger.Info("Diagram uploaded",
			zap.String("bucket", info.Bucket),
			zap.String("key", info.Key),
			zap.String("VersionID", info.VersionID),
		)
	}
	return nil
}

func (m minioStorage) LoadDiagramByUUID(ctx context.Context, logger *zap.Logger, uuid string, version string) ([]byte, error) {
	logger.Error("Load by UUID (ETAG) is not supported in minio")
	return nil, fmt.Errorf("method not supported by minio storage")
}

func (m minioStorage) LoadDiagramByName(ctx context.Context, logger *zap.Logger, name string, version string) ([]byte, error) {
	if version == "" {
		logger.Info("Loading diagram",
			zap.String("name", name),
		)
	} else {

	}
	return nil, nil
}

func (m minioStorage) IsVersioned(ctx context.Context) bool {
	return true
}

func (m minioStorage) Ls(ctx context.Context) ([]string, error) {
	m.Client.ListObjects(ctx, m.bucket.Name, minio.ListObjectsOptions{
		WithVersions: false,
		WithMetadata: false,
		Prefix:       "",
		Recursive:    false,
		MaxKeys:      0,
		StartAfter:   "",
		UseV1:        false,
	})
	return nil, nil
}

func (m minioStorage) LsVersions(ctx context.Context, diagram *calculator.Diagram) ([]calculator.DiagramVersion, error) {
	m.Client.ListObjects(ctx, m.bucket.Name, minio.ListObjectsOptions{
		WithVersions: true,
		WithMetadata: false,
		Prefix:       diagram.Name,
		Recursive:    false,
		MaxKeys:      0,
		StartAfter:   "",
		UseV1:        false,
	})
	return nil, nil
}
