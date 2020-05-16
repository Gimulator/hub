package storage

import (
	env "github.com/Gimulator/hub/utils/environment"
	"github.com/minio/minio-go"
	"gopkg.in/yaml.v2"
)

var s *minio.Client

func init() {
	var err error
	s, err = minio.NewV2(env.S3URL(), env.S3AccessKey(), env.S3SecretKey(), false)
	if err != nil {
		panic(err)
	}
}

func Get(bucket, name string) (string, error) {
	obj, err := s.GetObject(bucket, name, minio.GetObjectOptions{})
	if err != nil {
		return "", err
	}
	defer obj.Close()

	stat, err := obj.Stat()
	if err != nil {
		return "", err
	}

	b := make([]byte, stat.Size-1)
	_, err = obj.Read(b)
	if err != nil {
		return "", err
	}
	return string(b), nil
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
