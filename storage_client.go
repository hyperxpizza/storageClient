package storageClient

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"sync"

	minio "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/sirupsen/logrus"
)

type StorageClientConfig struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	Secure          bool
	Token           string
}

type Client interface {
}

type StorageClient struct {
	ctx context.Context
	lgr logrus.FieldLogger
	mc  *minio.Client
}

type BulkObjectResponse struct {
	Objects map[string]*bytes.Buffer
	mutex   sync.Mutex
}

func newBulkObjectResponse() *BulkObjectResponse {
	return &BulkObjectResponse{
		Objects: make(map[string]*bytes.Buffer),
		mutex:   sync.Mutex{},
	}
}

func (r *BulkObjectResponse) addFile(name string, buf *bytes.Buffer) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.Objects[name] = buf
}

func New(ctx context.Context, lgr logrus.FieldLogger, cfg StorageClientConfig) (*StorageClient, error) {
	mc, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, cfg.Token),
		Secure: cfg.Secure,
	})
	if err != nil {
		return nil, err
	}

	return &StorageClient{
		ctx: ctx,
		lgr: lgr.WithField("module", "storage-client"),
		mc:  mc,
	}, nil
}

func (s *StorageClient) GetBucketContents(bucket string, filenames []string) (*BulkObjectResponse, error) {
	bucketExists, err := s.mc.BucketExists(s.ctx, bucket)
	if err != nil {
		return nil, err
	}

	if !bucketExists {
		return nil, fmt.Errorf("bucket '%s' does not exist", bucket)
	}

	wg := sync.WaitGroup{}
	response := newBulkObjectResponse()
	getErr := newBulkFileError("failed to get following files:")
	for _, name := range filenames {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			buf, err := s.handleBulkFile(bucket, name)
			if err != nil {
				getErr.addFailedFile(name, err)
				return
			}

			response.addFile(name, buf)
		}(name)
	}

	wg.Wait()

	return response, nil
}

func (s *StorageClient) BulkUploadFromMultipart(files []*multipart.FileHeader, bucket string) error {

	//check if bucket exists
	bucketExists, err := s.mc.BucketExists(s.ctx, bucket)
	if err != nil {
		return err
	}

	//if bucket does not exist, create one
	if !bucketExists {
		err := s.mc.MakeBucket(s.ctx, bucket, minio.MakeBucketOptions{})
		if err != nil {
			return err
		}
	}

	//handle each file on a separate goroutine
	wg := sync.WaitGroup{}
	uploadErr := newBulkFileError("failed to upload files:")
	for _, file := range files {
		wg.Add(1)
		go func(file *multipart.FileHeader) {
			defer wg.Done()
			if err := s.handleMultipartFile(file, bucket); err != nil {
				uploadErr.addFailedFile(file.Filename, err)
				return
			}
		}(file)
	}

	wg.Wait()

	if len(uploadErr.FailedFiles) == 0 {
		return nil
	}

	return uploadErr
}

func (s *StorageClient) CreateBucket(name string) error {
	return s.mc.MakeBucket(s.ctx, name, minio.MakeBucketOptions{})
}

func (s *StorageClient) GetFile(bucket, file string) (*minio.Object, error) {
	return s.mc.GetObject(s.ctx, bucket, file, minio.GetObjectOptions{})
}

func (s *StorageClient) handleMultipartFile(file *multipart.FileHeader, bucketName string) error {
	buf := bytes.NewBuffer([]byte{})
	src, err := file.Open()
	if err != nil {
		return err
	}
	_, err = io.Copy(buf, src)
	if err != nil {
		return err
	}

	_, err = s.mc.PutObject(s.ctx, bucketName, file.Filename, buf, file.Size, minio.PutObjectOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (s *StorageClient) handleBulkFile(bucket, filename string) (*bytes.Buffer, error) {
	obj, err := s.mc.GetObject(s.ctx, bucket, filename, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer([]byte{})
	_, err = io.Copy(buf, obj)
	if err != nil {
		return nil, err
	}

	return buf, nil
}
