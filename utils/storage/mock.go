package storage

import (
	"encoding/json"
	"os"

	"gopkg.in/yaml.v2"
)

type Mock struct {
}

func NewMock() (*Mock, error) {
	return &Mock{}, nil
}

func (m *Mock) GetConfigYamlToString(bucket string, key string) (string, error) {
	return "This is a test config yaml", nil
}

func (m *Mock) GetConfigJsonToString(bucket string, key string) (string, error) {
	return "This is a test config json", nil
}

func (m *Mock) GetConfigYamlToStruct(bucket string, key string, i interface{}) error {
	var f *os.File
	var err error
	if bucket == "foo" {
		f, err = os.Open("/home/ali/Developer/Go/src/gitlab.com/Syfract/Xerac/secretary/storage/test/room.yaml")
		if err != nil {
			return err
		}
	} else {
		f, err = os.Open("/home/ali/Developer/Go/src/gitlab.com/Syfract/Xerac/secretary/storage/test/roles.yaml")
		if err != nil {
			return err
		}
	}
	defer f.Close()

	err = yaml.NewDecoder(f).Decode(i)
	if err != nil {
		return err
	}

	return nil
}

func (m *Mock) GetConfigJsonToStruct(bucket string, key string, i interface{}) error {
	f, err := os.Open("/home/ali/Developer/gitlab.com/Syfract/Xerac/secretary/storage/test/room.yaml")
	if err != nil {
		return err
	}
	defer f.Close()

	err = json.NewDecoder(f).Decode(i)
	if err != nil {
		return err
	}

	return nil
}
