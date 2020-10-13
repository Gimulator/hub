package s3

import (
	"os"

	"github.com/minio/minio-go"
	"gopkg.in/yaml.v2"
)

var s *minio.Client

func init() {
	S3URL := os.Getenv("S3_URL")
	S3AccessKey := os.Getenv("S3_ACCESS_KEY")
	S3SecretKey := os.Getenv("S3_SECRET_KEY")

	if S3URL == "" || S3AccessKey == "" || S3SecretKey == "" {
		panic("Invalid credential for S3: Set three environment variable S3_URL, S3_ACCESS_KEY, and S3_SECRET_Key for connecting to S3")
	}

	var err error
	s, err = minio.NewV2(S3URL, S3AccessKey, S3SecretKey, false)
	if err != nil {
		panic(err)
	}
}

func GetStruct(bucket, name string, i interface{}) error {
	obj, err := s.GetObject(bucket, name, minio.GetObjectOptions{})
	if err != nil {
		return err
	}

	stat, err := obj.Stat()
	if err != nil {
		return err
	}

	b := make([]byte, stat.Size-1)
	_, err = obj.Read(b)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(b, i)
}
