package minio

import (
	"bytes"
	"context"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/zap"
	"io"
	"io/ioutil"
	"strings"
	"sync"
	"time"
)

var once sync.Once
var instance *minioStorage

type minioStorage struct {
	Url      string
	user     string
	password string
	ssl      bool
	Client   *minio.Client
	Logger   *zap.Logger
	Bucket   *minio.BucketInfo
}

func NewMinioStorage(logger *zap.Logger, url string, bucket string, user string, password string, ssl bool) (*minioStorage, error) {
	var instance *minioStorage
	var err error
	once.Do(func() {
		logger.Info("Creating calculator instance")
		instance = &minioStorage{
			Url:      url,
			user:     user,
			password: password,
			ssl:      ssl,
			Client:   nil,
			Logger:   logger,
			Bucket:   nil,
		}
		instance.Client, err = minio.New(instance.Url, &minio.Options{
			Creds:  credentials.NewStaticV4(instance.user, instance.password, ""),
			Secure: instance.ssl,
		})
		if err != nil {
			instance.Logger.Fatal("can not connect to minio",
				zap.String("Url", instance.Url),
				zap.String("user", instance.user),
				zap.Bool("ssl_enabled", instance.ssl),
			)
		} else {
			instance.Logger.Info("minio backend connected successfully")
		}
		found, errr := instance.Client.BucketExists(context.Background(), bucket)
		err = errr
		if err != nil {
			logger.Error("could not find a bucket",
				zap.Error(err),
			)
			return
		}
		if found {
			instance.Bucket = &minio.BucketInfo{
				Name:         bucket,
				CreationDate: time.Time{},
			}
		} else {
			logger.Error("bucket not found or permissions issue ",
				zap.String("bucket name", bucket),
			)

		}
		return
	})

	return instance, err
}

func (m minioStorage) UploadTextFile(ctx context.Context, logger *zap.Logger, r io.Reader, path string) (err error) {
	//TODO: check if already exists and raise alert
	logger.Info("Diagram upload started",
		zap.String("bucket", m.Bucket.Name),
		zap.String("name", path),
	)

	body, err := ioutil.ReadAll(r)
	if err != nil {
		logger.Error("filed to read from reader",
			zap.Error(err),
		)
		return err
	}

	rn := bytes.NewReader(body)

	info, err := m.Client.PutObject(
		ctx,
		m.Bucket.Name,
		path,
		rn,
		int64(len(body)),
		minio.PutObjectOptions{ContentType: "application/octet-stream"},
	)
	if err != nil {
		logger.Error("Failed to upload file",
			zap.String("bucket", m.Bucket.Name),
			zap.String("path", path),
			zap.Error(err),
		)
		return err
	} else {
		logger.Info("file uploaded",
			zap.String("bucket", info.Bucket),
			zap.String("key", info.Key),
			zap.String("VersionID", info.VersionID),
		)
	}
	return nil
}

// LoadDiagramByName loads latest version of file from minio in case version is empty string,
// in case version is not empty - tries loading provided version
func (m minioStorage) LoadFileByName(ctx context.Context, logger *zap.Logger, path string, version string) (io.Reader, error) {
	_, err := m.Client.StatObject(ctx, m.Bucket.Name, path, minio.StatObjectOptions{})
	if err != nil {
		logger.Error("could not load file",
			zap.String("path", path),
			zap.Error(err),
		)
		return nil, err
	}
	objReader, err := m.Client.GetObject(
		ctx,
		m.Bucket.Name,
		path,
		minio.GetObjectOptions{
			ServerSideEncryption: nil,
			VersionID:            version,
			PartNumber:           0,
			Checksum:             false,
			Internal:             minio.AdvancedGetOptions{},
		},
	)

	if err != nil {
		logger.Error("Could not load file",
			zap.String("bucket", m.Bucket.Name),
			zap.String("path", path),
			zap.String("version", version),
			zap.Error(err),
		)
		return nil, err
	}

	return objReader, nil
}

func (m minioStorage) IsVersioned(ctx context.Context) bool {
	return true
}

// Ls performes list of files actually,  calculator.Diagram has only the name attribute set.
func (m minioStorage) Ls(ctx context.Context, path string) <-chan string {
	if !strings.HasSuffix(path, "/") {
		path = path + "/"
	}
	rChan := make(chan string)
	chanObjInfo := m.Client.ListObjects(ctx, m.Bucket.Name, minio.ListObjectsOptions{
		WithVersions: false,
		WithMetadata: false,
		Prefix:       path,
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
			rChan <- obj.Key
		}
		close(rChan)
	}()
	return rChan
}

// LsVersions shows varsions of provided object, stored in minio object storage.
func (m minioStorage) LsVersions(ctx context.Context, path string, logger *zap.Logger) (<-chan string, error) {
	_, err := m.Client.StatObject(ctx, m.Bucket.Name, path, minio.StatObjectOptions{})
	if err != nil {
		logger.Error("could not load file",
			zap.String("path", path),
			zap.Error(err),
		)
		return nil, err
	}

	rChan := make(chan string)
	chanObjInfo := m.Client.ListObjects(ctx, m.Bucket.Name, minio.ListObjectsOptions{
		WithVersions: true,
		WithMetadata: false,
		Prefix:       path,
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
				zap.String("diagram name", path),
				zap.String("version", obj.VersionID),
			)
			rChan <- obj.VersionID
		}
		close(rChan)
	}()

	return rChan, nil
}

func (m minioStorage) ConfigDump(ctx context.Context, logger *zap.Logger) map[string]string {
	cfg := make(map[string]string)
	cfg["url"] = m.Url
	return cfg
}
