package name

// Pods
func ActorPodName(id string) string {
	return "actor-" + id
}

func DirectorPodName(id string) string {
	return "director-" + id
}

func GimulatorPodName(roomID string) string {
	return "gimulator-" + roomID
}

// Containers
func ActorContainerName() string {
	return "actor"
}

func DirectorContainerName() string {
	return "director"
}

func GimulatorContainerName() string {
	return "gimulator"
}

// ConfigMap

func CredConfigMapName(id string) string {
	return "credential-" + id
}

func RolesConfigMapName(id string) string {
	return "roles-" + id
}

// Roles
func DirectorRoleName() string {
	return "director"
}

func MasterRoleName() string {
	return "master"
}

// Services
func GimulatorServiceName(roomID string) string {
	return "gimulator-" + roomID
}

func GimulatorServicePort() int32 {
	return 23579
}

// Volumes
func DataVolumeName() string {
	return "data"
}

func DataVolumeMountPath() string {
	return "/data"
}

func FactVolumeName() string {
	return "fact"
}

func FactVolumeMountPath() string {
	return "/fact"
}

func OutputVolumeName(id string) string {
	return "output-" + id
}

func OutputVolumeMountPath() string {
	return "/output"
}

func ActorOutputVolumeMountPathForDirector(id string) string {
	return "/var/gimulator/" + id
}

func ActorOutputPVCName(id string) string {
	return "actor-output-pvc-" + id
}

func DirectorOutputPVCName(id string) string {
	return "director-output-pvc-" + id
}

func RolesVolumeName() string {
	return "roles-volume"
}

func RolesVolumeMountPath() string {
	return "/etc/gimulator"
}

func CredsVolumeName() string {
	return "credentials-volume"
}

func CredsVolumeMountPath() string {
	return "/etc/gimulator"
}

// Labels
func ActorIDLabel() string {
	return "actorID"
}

func DirectorIDLabel() string {
	return "directorID"
}

func RoomIDLabel() string {
	return "roomID"
}

func PodTypeLabel() string {
	return "podType"
}

// Pod Types
func PodTypeActor() string {
	return "actor"
}

func PodTypeDirector() string {
	return "director"
}

func PodTypeGimulator() string {
	return "gimulator"
}

// S3
func S3GameConfigBucket() string {
	return "game-config"
}

func S3RoleBucket() string {
	return "roles"
}

// Cache
func CacheKeyForGame(game string) string {
	return "game-config-" + game
}
