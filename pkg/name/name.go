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

func RulesConfigMapName(id string) string {
	return "rules-" + id
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
	return "200M" // TODO should be set dynamically
}

func GimulatorCPULimit() string {
	return "500m" // TODO should be set dynamically
}

func GimulatorEphemeralLimit() string {
	return "100M" // TODO should be set dynamically
}

func GimulatorHost(roomID string) string {
	return fmt.Sprintf("%s:%d", GimulatorServiceName(roomID), GimulatorServicePort())
}

// Volumes
func DataVolumeName(id string) string {
	return "data-" + id
}

func DataVolumeMountPath(id string) string {
	return "/data/" + id
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

func GimulatorConfigVolumeName() string {
	return "gimulator-config-volume"
}

func GimulatorConfigMountPath() string {
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

func CharacterOperator() string {
	return api.Character_name[int32(api.Character_operator)]
}

func CharacterMaster() string {
	return api.Character_name[int32(api.Character_master)]
}

func CharacterGimulator() string {
	return "gimulator"
}

// S3
func S3LogsBucket() string {
	return "log"
}

func S3LogObjectNameForDirector(runID, directorID string) string {
	return fmt.Sprintf("%s/%s.log", runID, DirectorPodName(directorID))
}

func S3LogObjectNameForActor(runID, actorID string) string {
	return fmt.Sprintf("%s/%s.log", runID, ActorPodName(actorID))
}

func S3SettingBucket() string {
	return "settings"
}

func S3SettingObjectName(id string) string {
	return id + ".yaml"
}

func S3RulesBucket() string {
	return "rules"
}

func S3RulesObjectName(id string) string {
	return id + ".yaml"
}

// Cache
func CacheKeyForSetting(id string) string {
	return "settings-" + id
}

func CacheKeyForRules(id string) string {
	return "rules-" + id
}
