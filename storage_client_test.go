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
	sc, err := New(ctx, lgr, cfg)
	if err != nil {
		t.Error(err)
	}

	err = sc.CreateBucket(bucketName)
	if err != nil {

	}
}
