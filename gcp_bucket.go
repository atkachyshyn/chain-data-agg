package chaindataagg

import (
	"context"
	"io"
	"time"

	"cloud.google.com/go/storage"
)

func UploadToBucket(bucketName, objectName string, data []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*50)
	defer cancel()

	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()

	bucket := client.Bucket(bucketName)
	writer := bucket.Object(objectName).NewWriter(ctx)
	if _, err := writer.Write(data); err != nil {
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}

	return nil
}

func DownloadFromBucket(bucketName, objectName string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*50)
	defer cancel()

	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	bucket := client.Bucket(bucketName)
	reader, err := bucket.Object(objectName).NewReader(ctx)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return data, nil
}
