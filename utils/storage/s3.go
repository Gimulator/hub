package storage

import (
	"github.com/minio/minio-go"
	env "gitlab.com/Syfract/Xerac/hub/utils/environment"
	"gopkg.in/yaml.v2"
)

var s *minio.Client

func init() {
	var err error
	s, err = minio.New(env.S3URL(), env.S3AccessKey(), env.S3SecretKey(), false)
	if err != nil {
		panic(err)
	}
}

func Get(bucket, name string) (string, error) {
	obj, err := s.GetObject(bucket, name, minio.GetObjectOptions{})
	if err != nil {
		return "", err
	}

	var b []byte
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

	var b []byte
	_, err = obj.Read(b)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(b, i)
}
