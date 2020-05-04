package storage

type Storage interface {
	GetConfigYamlToString(string, string) (string, error)
	GetConfigJsonToString(string, string) (string, error)
	GetConfigYamlToStruct(string, string, interface{}) error
	GetConfigJsonToStruct(string, string, interface{}) error
}
