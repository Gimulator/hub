package v1

type ConfigMap struct {
	Name   string `json:"name"`
	Bucket string `json:"bucket"`
	Key    string `json:"key"`
}
