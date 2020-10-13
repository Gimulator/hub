package name

func ActorPodName(id string) string {
	return "actor-" + id
}

func ActorContainerName() string {
	return "actor"
}

func DirectorPodName(id string) string {
	return "director-" + id
}

func DirectorContainerName() string {
	return "director"
}

func GimulatorPodName(roomID string) string {
	return "gimulator-" + roomID
}

func GimulatorServiceName(roomID string) string {
	return "gimulator-" + roomID
}

func GimulatorServicePort() int32 {
	return 23579
}

func GimulatorContainerName() string {
	return "gimulator"
}

func DataVolumeName() string {
	return "data"
}

func DataVolumeMountDir() string {
	return "/data"
}

func FactVolumeName() string {
	return "fact"
}

func FactVolumeMountDir() string {
	return "/fact"
}

func OutputVolumeName() string {
	return "output"
}

func OutputVolumeMountDir() string {
	return "/output"
}

func OutputPVCName(id string) string {
	return "output-pvc-" + id
}

func S3GameConfigBucket() string {
	return "game-config"
}

func CacheKeyForGame(game string) string {
	return "game-config-" + game
}

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

func PodTypeActor() string {
	return "actor"
}

func PodTypeDirector() string {
	return "director"
}

func PodTypeGimulator() string {
	return "gimulator"
}
