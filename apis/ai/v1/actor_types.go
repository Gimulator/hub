package v1

type AIActorType string

const (
	AIActorTypeMaster   AIActorType = "master"
	AIActorTypeSlave    AIActorType = "slave"
	AIActorTypeFinisher AIActorType = "finisher"
)

type Resource struct {
	CPU       string `json:"cpu,omitempty"`
	Memory    string `json:"memory,omitempty"`
	Ephemeral string `json:"ephemeral,omitempty"`
}

type Resources struct {
	Requests Resource `json:"requests,omitempty"`
	Limits   Resource `json:"limits,omitempty"`
}

type EnvVar struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type VolumeMount struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type Actor struct {
	ID           int           `json:"id"`
	Name         string        `json:"name"`
	Role         string        `json:"role"`
	Image        string        `json:"image"`
	Command      string        `json:"command"`
	Args         []string      `json:"args,omitempty"`
	Type         AIActorType   `json:"type,omitempty"`
	EnvVars      []EnvVar      `json:"env-vars,omitempty"`
	VolumeMounts []VolumeMount `json:"volume-mounts,omitempty"`
	Resources    Resources     `json:"resources,omitempty"`
}
