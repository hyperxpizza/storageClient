package storageClient

import (
	"context"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestStorageClient(t *testing.T) {

	const (
		bucketName = "test-bucket-name"
	)

	ctx := context.Background()
	lgr := logrus.New()
	cfg := StorageClientConfig{}
	cfg.Endpoint = "0.0.0.0:9000"
	cfg.AccessKeyID = "minio"
	cfg.SecretAccessKey = "minio123"
	cfg.Secure = false
	cfg.Token = ""
	sc, err := New(ctx, lgr, cfg)
	if err != nil {
		t.Error(err)
	}

	err = sc.CreateBucket(bucketName)
	if err != nil {
		t.Error(err)
	}

	defer func() {
		err = sc.DeleteBucket(bucketName)
		t.Error(err)
	}()
}
