package v1

type Volume struct {
	ConfigMapVolumes *ConfigMapVolume `json:"config-map,omitempty"`
	EmptyDirVolume   *EmptyDirVolume  `json:"empty-dir,omitempty"`
}

type ConfigMapVolume struct {
	Name          string `json:"name"`
	ConfigMapName string `json:"config-map-name"`
}

type EmptyDirVolume struct {
	Name string `json:"name"`
}
