package s3

import (
	"context"
	"io"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"sigs.k8s.io/yaml"
)

var s *minio.Client

func init() {
	s3URL := os.Getenv("HUB_S3_URL")
	s3AccessKey := os.Getenv("HUB_S3_ACCESS_KEY")
	s3SecretKey := os.Getenv("HUB_S3_SECRET_KEY")

	if s3URL == "" || s3AccessKey == "" || s3SecretKey == "" {
		panic("Invalid credential for S3: Set three environment variable HUB_S3_URL, HUB_S3_ACCESS_KEY, and HUB_S3_SECRET_Key for connecting to S3")
	}

	var err error
	s, err = minio.New(s3URL, &minio.Options{
		Creds:  credentials.NewStaticV2(s3AccessKey, s3SecretKey, ""),
		Secure: false,
	})
	if err != nil {
		panic(err)
	}
}

func GetStruct(ctx context.Context, bucket, name string, i interface{}) error {
	reader, err := s.GetObject(ctx, bucket, name, minio.GetObjectOptions{})
	if err != nil {
		return err
	}
	defer reader.Close()

	content, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(content, i); err != nil {
		return err
	}
	return nil
}

func GetBytes(ctx context.Context, bucket, name string) ([]byte, error) {
	obj, err := s.GetObject(ctx, bucket, name, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}

	stat, err := obj.Stat()
	if err != nil {
		return nil, err
	}

	b := make([]byte, stat.Size-1)
	_, err = obj.Read(b)
	return b, err
}

func GetString(ctx context.Context, bucket, name string) (string, error) {
	bytes, err := GetBytes(ctx, bucket, name)
	return string(bytes), err
}
