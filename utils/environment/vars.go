package environment

import (
	"fmt"

	aiv1 "github.com/Gimulator/hub/apis/ai/v1"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

var gimulatorContainer aiv1.Actor
var loggerContainer aiv1.Actor

func init() {
	viper.BindEnv(s3URL)
	viper.BindEnv(s3AccessKey)
	viper.BindEnv(s3SecretKey)

	viper.SetDefault(keyGimulatorName, "gimulator")
	viper.SetDefault(keyGimulatorID, -1)
	viper.SetDefault(keyGimulatorImage, "gimulator:latest")
	viper.SetDefault(keyGimulatorType, ContainerTypeMaster)
	viper.SetDefault(keyGimulatorCmd, "/app/result")
	viper.SetDefault(keyGimulatorConfigVolumeName, "gimulator-config-path")
	viper.SetDefault(keyGimulatorConfigVolumePath, "/config")
	viper.SetDefault(keyGimulatorResourceRequestsCPU, "200m")
	viper.SetDefault(keyGimulatorResourceRequestsMemory, "500M")
	viper.SetDefault(keyGimulatorResourceRequestsEphemeral, "10M")
	viper.SetDefault(keyGimulatorResourceLimitsCPU, "400m")
	viper.SetDefault(keyGimulatorResourceLimitsMemory, "1G")
	viper.SetDefault(keyGimulatorResourceLimitsEphemeral, "20M")

	viper.SetDefault(keyLoggerName, "logger")
	viper.SetDefault(keyLoggerID, -2)
	viper.SetDefault(keyLoggerImage, "logger:latest")
	viper.SetDefault(keyLoggerType, ContainerTypeFinisher)
	viper.SetDefault(keyLoggerCmd, "/app/logger")
	viper.SetDefault(keyLoggerRole, "logger")
	viper.SetDefault(keyLoggerResourceRequestsCPU, "200m")
	viper.SetDefault(keyLoggerResourceRequestsMemory, "500M")
	viper.SetDefault(keyLoggerResourceRequestsEphemeral, "10M")
	viper.SetDefault(keyLoggerResourceLimitsCPU, "400m")
	viper.SetDefault(keyLoggerResourceLimitsMemory, "1G")
	viper.SetDefault(keyLoggerResourceLimitsEphemeral, "20M")

	viper.SetDefault(keySharedVolumeName, "shared-volume")
	viper.SetDefault(keySharedVolumePath, "/tmp/pod")

	viper.SetDefault(keyNamespace, "default")

	viper.SetDefault(keyDefaultResourceRequestsCPU, "100m")
	viper.SetDefault(keyDefaultResourceRequestsMemory, "100M")
	viper.SetDefault(keyDefaultResourceRequestsEphemeral, "10M")

	viper.SetDefault(keyDefaultResourceLimitsCPU, "200m")
	viper.SetDefault(keyDefaultResourceLimitsMemory, "200M")
	viper.SetDefault(keyDefaultResourceLimitsEphemeral, "20M")

	if err := ReadEnvironments(); err != nil {
		fmt.Println(err)
	}
}

func ReadEnvironments() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/hub")

	err := viper.ReadInConfig()
	if err != nil {
		return err
	}

	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("Config file changed:", e.Name)
	})

	return nil
}

///////////////////////////////// S3

func S3AccessKey() string {
	return viper.GetString(s3AccessKey)
}
func S3SecretKey() string {
	return viper.GetString(s3SecretKey)
}
func S3URL() string {
	return viper.GetString(s3URL)
}

///////////////////////////////// Gimulator

func GimulatorName() string {
	return viper.GetString(keyGimulatorName)
}
func GimulatorID() int {
	return viper.GetInt(keyGimulatorID)
}
func GimulatorImage() string {
	return viper.GetString(keyGimulatorImage)
}
func GimulatorType() ContainerType {
	return ContainerType(viper.GetString(keyGimulatorType))
}
func GimulatorCmd() string {
	return viper.GetString(keyGimulatorCmd)
}
func GimulatorConfigVolumeName() string {
	return viper.GetString(keyGimulatorConfigVolumeName)
}
func GimulatorConfigVolumePath() string {
	return viper.GetString(keyGimulatorConfigVolumePath)
}
func GimulatorResourceRequestsCPU() string {
	return viper.GetString(keyGimulatorResourceRequestsCPU)
}
func GimulatorResourceRequestsMemory() string {
	return viper.GetString(keyGimulatorResourceRequestsMemory)
}
func GimulatorResourceRequestsEphemeral() string {
	return viper.GetString(keyGimulatorResourceRequestsEphemeral)
}
func GimulatorResourceLimitsCPU() string {
	return viper.GetString(keyGimulatorResourceLimitsCPU)
}
func GimulatorResourceLimitsMemory() string {
	return viper.GetString(keyGimulatorResourceLimitsMemory)
}
func GimulatorResourceLimitsEphemeral() string {
	return viper.GetString(keyGimulatorResourceLimitsEphemeral)
}

///////////////////////////////// Logger

func LoggerName() string {
	return viper.GetString(keyLoggerName)
}
func LoggerID() int {
	return viper.GetInt(keyLoggerID)
}
func LoggerImage() string {
	return viper.GetString(keyLoggerImage)
}
func LoggerType() ContainerType {
	return ContainerType(viper.GetString(keyLoggerType))
}
func LoggerCmd() string {
	return viper.GetString(keyLoggerCmd)
}
func LoggerRole() string {
	return viper.GetString(keyLoggerRole)
}
func LoggerResourceRequestsCPU() string {
	return viper.GetString(keyLoggerResourceRequestsCPU)
}
func LoggerResourceRequestsMemory() string {
	return viper.GetString(keyLoggerResourceRequestsMemory)
}
func LoggerResourceRequestsEphemeral() string {
	return viper.GetString(keyLoggerResourceRequestsEphemeral)
}
func LoggerResourceLimitsCPU() string {
	return viper.GetString(keyLoggerResourceLimitsCPU)
}
func LoggerResourceLimitsMemory() string {
	return viper.GetString(keyLoggerResourceLimitsMemory)
}
func LoggerResourceLimitsEphemeral() string {
	return viper.GetString(keyLoggerResourceLimitsEphemeral)
}

///////////////////////////////// SharedVolume

func SharedVolumeName() string {
	return viper.GetString(keySharedVolumeName)
}
func SharedVolumePath() string {
	return viper.GetString(keySharedVolumePath)
}

///////////////////////////////// Namespace

func Namespace() string {
	return viper.GetString(keyNamespace)
}

///////////////////////////////// DefaultResources

func ResourceDefaultRequestsCPU() string {
	return viper.GetString(keyDefaultResourceRequestsCPU)
}
func ResourceDefaultRequestsMemory() string {
	return viper.GetString(keyDefaultResourceRequestsMemory)
}
func ResourceDefaultRequestsEphemeral() string {
	return viper.GetString(keyDefaultResourceRequestsEphemeral)
}
func ResourceDefaultLimitsCPU() string {
	return viper.GetString(keyDefaultResourceLimitsCPU)
}
func ResourceDefaultLimitsMemory() string {
	return viper.GetString(keyDefaultResourceLimitsMemory)
}
func ResourceDefaultLimitsEphemeral() string {
	return viper.GetString(keyDefaultResourceLimitsEphemeral)
}

////////////////////////////////// Env Vars
func EnvVarKeyLoggerS3URL() string {
	return viper.GetString(envvarkeyLoggerS3URL)
}
func EnvVarKeyLoggerS3AccessKey() string {
	return viper.GetString(envvarkeyLoggerS3AccessKey)
}
func EnvVarKeyLoggerS3SecretKey() string {
	return viper.GetString(envvarkeyLoggerS3SecretKey)
}
func EnvVarKeyLoggerS3Bucket() string {
	return viper.GetString(envvarkeyLoggerS3Bucket)
}
func EnvVarKeyLoggerRecorderDir() string {
	return viper.GetString(envvarkeyLoggerRecorderDir)
}
func EnvVarKeyLoggerRabbitURI() string {
	return viper.GetString(envvarkeyLoggerRabbitURI)
}
func EnvVarKeyLoggerRabbitQueue() string {
	return viper.GetString(envvarkeyLoggerRabbitQueue)
}
func EnvVarKeyClientID() string {
	return viper.GetString(envvarkeyClientID)
}
func EnvVarKeyRoomID() string {
	return viper.GetString(envvarkeyRoomID)
}
func EnvVarKeyRoomEndOfGameKey() string {
	return viper.GetString(envvarkeyRoomEndOfGameKey)
}
func EnvVarKeyGimulatorHost() string {
	return viper.GetString(envvarkeyGimulatorHost)
}
func EnvVarKeyGimulatorRoleFilePath() string {
	return viper.GetString(envvarkeyGimulatorRoleFilePath)
}

func EnvVarValLoggerS3Bucket() string {
	return viper.GetString(envvarvalLoggerS3Bucket)
}
func EnvVarValLoggerRecorderDir() string {
	return viper.GetString(envvarvalLoggerRecorderDir)
}
func EnvVarValLoggerRabbitURI() string {
	return viper.GetString(envvarvalLoggerRabbitURI)
}
func EnvVarValLoggerRabbitQueue() string {
	return viper.GetString(envvarvalLoggerRabbitQueue)
}
func EnvVarValRoomEndOfGameKey() string {
	return viper.GetString(envvarvalRoomEndOfGameKey)
}
func EnvVarValGimulatorHost() string {
	return viper.GetString(envvarvalGimulatorHost)
}
func EnvVarValGimulatorRoleFilePath() string {
	return viper.GetString(envvarvalGimulatorRoleFilePath)
}
