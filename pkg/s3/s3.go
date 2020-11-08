package s3

import (
	"os"

	"github.com/minio/minio-go"
	"gopkg.in/yaml.v2"
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
	s, err = minio.NewV2(s3URL, s3AccessKey, s3SecretKey, false)
	if err != nil {
		panic(err)
	}
}

func GetStruct(bucket, name string, i interface{}) error {
	obj, err := s.GetObject(bucket, name, minio.GetObjectOptions{})
	if err != nil {
		return err
	}
	defer obj.Close()

	if err := yaml.NewDecoder(obj).Decode(i); err != nil {
		return err
	}
	return nil
}

func GetBytes(bucket, name string) ([]byte, error) {
	obj, err := s.GetObject(bucket, name, minio.GetObjectOptions{})
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

func GetString(bucket, name string) (string, error) {
	bytes, err := GetBytes(bucket, name)
	return string(bytes), err
}
