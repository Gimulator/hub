package storage

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	env "gitlab.com/Syfract/Xerac/hub/utils/environment"
	"gopkg.in/yaml.v2"
)

type S3 struct {
	*s3.S3
}

func NewS3() (*S3, error) {
	creds := credentials.NewStaticCredentials(env.S3AccessKey(), env.S3SecretKey(), env.S3Token())
	_, err := creds.Get()
	if err != nil {
		return nil, err
	}
	cfg := aws.NewConfig().WithCredentials(creds).WithRegion("None").WithEndpoint(env.S3URL())
	newSession, err := session.NewSession()
	if err != nil {
		return nil, err
	}
	svc := s3.New(newSession, cfg)

	return &S3{
		svc,
	}, nil
}

func (s *S3) GetConfigYamlToString(bucket, key string) (string, error) {
	out, err := s.get(bucket, key)
	if err != nil {
		return "", err
	}

	b, err := yaml.Marshal(out.Body)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func (s *S3) GetConfigJsonToString(bucket, key string) (string, error) {
	out, err := s.get(bucket, key)
	if err != nil {
		return "", err
	}

	b, err := json.Marshal(out.Body)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func (s *S3) GetConfigYamlToStruct(bucket, key string, i interface{}) error {
	out, err := s.get(bucket, key)
	if err != nil {
		return err
	}

	err = yaml.NewDecoder(out.Body).Decode(i)
	if err != nil {
		return err
	}
	return nil
}

func (s *S3) GetConfigJsonToStruct(bucket, key string, i interface{}) error {
	out, err := s.get(bucket, key)
	if err != nil {
		return err
	}

	err = json.NewDecoder(out.Body).Decode(i)
	if err != nil {
		return err
	}
	return nil
}

func (s *S3) get(bucket string, key string) (*s3.GetObjectOutput, error) {
	params := &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	}

	resp, err := s.GetObject(params)
	if err != nil {
		fmt.Println("error in get object")
		return nil, err
	}
	return resp, nil
}
