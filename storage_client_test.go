package storageClient

import (
	"context"
	"testing"
)

func getSC() (*StorageClient, error) {
	ctx := context.Background()
	cfg := StorageClientConfig{}
	cfg.Endpoint = "0.0.0.0:9000"
	cfg.AccessKeyID = "minio"
	cfg.SecretAccessKey = "minio123"
	cfg.Secure = false
	cfg.Token = ""
	sc, err := New(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return sc, nil
}

func TestCreateBucket(t *testing.T) {

	const (
		bucketName = "test-bucket-name"
	)

	sc, err := getSC()
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

func TestGetBucketContents(t *testing.T) {
	const (
		bucketName = "test-get-bucket-contents"
	)

	sc, err := getSC()
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
