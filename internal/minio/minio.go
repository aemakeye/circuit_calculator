package minio

import (
	"context"
	"github.com/aemakeye/circuit_calculator/internal/calculator"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/zap"
	"io/ioutil"
	"strings"
	"sync"
	"time"
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
	Bucket   *minio.BucketInfo
}

func NewMinioStorage(logger *zap.Logger, url string, bucket string, user string, password string, ssl bool) (instance *minioStorage, err error) {
	once.Do(func() {
		logger.Info("Creating calculator instance")
		instance = &minioStorage{
			url:      url,
			user:     user,
			password: password,
			ssl:      ssl,
			Client:   nil,
			Logger:   logger,
			Bucket:   nil,
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
		found, err := instance.Client.BucketExists(context.Background(), bucket)
		if err != nil {
			logger.Fatal("could not find a bucket",
				zap.Error(err),
			)
		}
		if found {
			instance.Bucket = &minio.BucketInfo{
				Name:         bucket,
				CreationDate: time.Time{},
			}
		} else {
			logger.Fatal("bucket not found or permissions issue ",
				zap.String("bucket name", bucket),
			)
		}
	})
	return
}

func (m minioStorage) UploadDiagram(ctx context.Context, logger *zap.Logger, dia *calculator.Diagram) (err error) {
	//TODO: check if already exists and raise alert
	logger.Info("Diagram upload started",
		zap.String("bucket", m.Bucket.Name),
		zap.String("name", dia.Name),
	)

	info, err := m.Client.PutObject(
		ctx,
		m.Bucket.Name,
		"test-diagram.xml",
		strings.NewReader(dia.Body),
		int64(len(dia.Body)),
		minio.PutObjectOptions{ContentType: "application/octet-stream"},
	)
	if err != nil {
		logger.Error("Failed to upload diagram",
			zap.String("bucket", m.Bucket.Name),
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

// LoadDiagramByName loads latest version of file from minio in case version is empty string,
// in case version is not empty - tries loading provided version
func (m minioStorage) LoadDiagramByName(ctx context.Context, logger *zap.Logger, name string, version string) ([]byte, error) {
	objReader, err := m.Client.GetObject(
		ctx,
		m.Bucket.Name,
		name,
		minio.GetObjectOptions{
			ServerSideEncryption: nil,
			VersionID:            version,
			PartNumber:           0,
			Checksum:             false,
			Internal:             minio.AdvancedGetOptions{},
		},
	)

	if err != nil {
		logger.Error("Could not load diagram",
			zap.String("name", name),
			zap.String("version", version),
			zap.Error(err),
		)
		return nil, err
	}

	defer objReader.Close()

	buf, err := ioutil.ReadAll(objReader)

	if err != nil {
		logger.Error("could not read file",
			zap.String("name", name),
			zap.Error(err),
		)
		return nil, err
	}

	return buf, nil
}

func (m minioStorage) IsVersioned(ctx context.Context) bool {
	return true
}

// Ls performes list of files actually,  calculator.Diagram has only the name attribute set.
func (m minioStorage) Ls(ctx context.Context) <-chan calculator.Diagram {
	rChan := make(chan calculator.Diagram)
	chanObjInfo := m.Client.ListObjects(ctx, m.Bucket.Name, minio.ListObjectsOptions{
		WithVersions: false,
		WithMetadata: false,
		Prefix:       "",
		Recursive:    false,
		MaxKeys:      0,
		StartAfter:   "",
		UseV1:        false,
	})

	go func() {
		for obj := range chanObjInfo {
			if obj.Err != nil {
				m.Logger.Error("skipping bad object")
			}
			m.Logger.Debug("New item in list",
				zap.String("diagram name", obj.Key),
			)
			rChan <- calculator.Diagram{Name: obj.Key}
		}
		close(rChan)
	}()
	return rChan
}

//LsVersions shows varsions of provided object, stored in minio object storage.
func (m minioStorage) LsVersions(ctx context.Context, diagram *calculator.Diagram) <-chan calculator.DiagramVersion {
	rChan := make(chan calculator.DiagramVersion)
	chanObjInfo := m.Client.ListObjects(ctx, m.Bucket.Name, minio.ListObjectsOptions{
		WithVersions: true,
		WithMetadata: false,
		Prefix:       diagram.Name,
		Recursive:    false,
		MaxKeys:      0,
		StartAfter:   "",
		UseV1:        false,
	})

	go func() {
		for obj := range chanObjInfo {
			if obj.Err != nil {
				m.Logger.Error("skipping bad object")
			}
			m.Logger.Debug("diagram version found",
				zap.String("diagram name", diagram.Name),
				zap.String("version", obj.VersionID),
			)
			rChan <- calculator.DiagramVersion{Version: obj.VersionID}
		}
		close(rChan)
	}()

	return rChan
}
