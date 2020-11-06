package name

import (
	"fmt"

	"github.com/Gimulator/protobuf/go/api"
)

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
	return CharacterActor()
}

func DirectorContainerName() string {
	return CharacterDirector()
}

func MasterContainerName() string {
	return CharacterMaster()
}

func GimulatorContainerName() string {
	return CharacterGimulator()
}

// ConfigMap

func CredConfigMapName(id string) string {
	return "credential-" + id
}

func RolesConfigMapName(id string) string {
	return "roles-" + id
}

// Gimulator
func GimulatorServiceName(roomID string) string {
	return "gimulator-" + roomID
}

func GimulatorConfigDir() string {
	return "/etc/gimulator"
}

func GimulatorServicePort() int {
	return 23579
}

func GimulatorMemoryLimit() string {
	return "1G"
}

func GimulatorCPULimit() string {
	return "1"
}

func GimulatorEphemeralLimit() string {
	return "1G"
}

func GimulatorHost(roomID string) string {
	return fmt.Sprintf("%s:%d", GimulatorServiceName(roomID), GimulatorServicePort())
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

func GimulatorRulesVolumeName() string {
	return "roles-volume"
}

func GimulatorRulesVolumeMountPath() string {
	return GimulatorConfigDir()
}

func GimulatorCredsVolumeName() string {
	return "credentials-volume"
}

func GimulatorCredsVolumeMountPath() string {
	return GimulatorConfigDir()
}

func ActorOutputVolumeMountPathForDirector(id string) string {
	return "/actors/" + id
}

func OutputPVCName(id string) string {
	return "output-pvc-" + id
}

// Labels
func CharacterLabel() string {
	return "character"
}

func RoleLabel() string {
	return "role"
}

func RoomLabel() string {
	return "room"
}

func ProblemLabel() string {
	return "problem"
}

func IDLabel() string {
	return "id"
}

// character
func CharacterActor() string {
	return api.Character_name[int32(api.Character_actor)]
}

func CharacterDirector() string {
	return api.Character_name[int32(api.Character_director)]
}

func CharacterMaster() string {
	return api.Character_name[int32(api.Character_master)]
}

func CharacterGimulator() string {
	return "gimulator"
}

// S3
func S3ProblemSettingsBucket() string {
	return "problem-settings"
}

func S3ProblemSettingsObjectName(id string) string {
	return id + "-problem-settings.yaml"
}

func S3RulesBucket() string {
	return "roles"
}

func S3RulesObjectName(id string) string {
	return id + "-roles.yaml"
}

// Cache
func CacheKeyForProblemSettings(id string) string {
	return "problem-settings-" + id
}

func CacheKeyForRules(id string) string {
	return "rules-" + id
}
