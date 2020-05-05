package v1

type ConfigMap struct {
	Name   string `json:"name"`
	Bucket string `json:"bucket,omitempty"`
	Key    string `json:"key,omitempty"`
	Data   string `json:"data,omitempty"`
}
