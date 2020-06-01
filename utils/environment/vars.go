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
	return viper.GetString(gimulatorName)
}
func GimulatorID() int {
	return viper.GetInt(gimulatorID)
}
func GimulatorImage() string {
	return viper.GetString(gimulatorImage)
}
func GimulatorType() ContainerType {
	return ContainerType(viper.GetString(gimulatorType))
}
func GimulatorCmd() string {
	return viper.GetString(gimulatorCmd)
}
func GimulatorConfigVolumeName() string {
	return viper.GetString(gimulatorConfigVolumeName)
}
func GimulatorConfigVolumePath() string {
	return viper.GetString(gimulatorConfigVolumePath)
}
func GimulatorRequestsCPU() string {
	return viper.GetString(gimulatorRequestsCPU)
}
func GimulatorRequestsMemory() string {
	return viper.GetString(gimulatorRequestsMemory)
}
func GimulatorRequestsEphemeral() string {
	return viper.GetString(gimulatorRequestsEphemeral)
}
func GimulatorLimitsCPU() string {
	return viper.GetString(gimulatorLimitsCPU)
}
func GimulatorLimitsMemory() string {
	return viper.GetString(gimulatorLimitsMemory)
}
func GimulatorLimitsEphemeral() string {
	return viper.GetString(gimulatorLimitsEphemeral)
}
func GimulatorHost() string {
	return viper.GetString(gimulatorHost)
}
func GimulatorRoleFileName() string {
	return viper.GetString(gimulatorRoleFileName)
}
func GimulatorHostEnvKey() string {
	return viper.GetString(gimulatorHostEnvKey)
}
func GimulatorRoleFilePathEnvKey() string {
	return viper.GetString(gimulatorRoleFilePathEnvKey)
}

///////////////////////////////// Logger

func LoggerName() string {
	return viper.GetString(loggerName)
}
func LoggerID() int {
	return viper.GetInt(loggerID)
}
func LoggerImage() string {
	return viper.GetString(loggerImage)
}
func LoggerType() ContainerType {
	return ContainerType(viper.GetString(loggerType))
}
func LoggerCmd() string {
	return viper.GetString(loggerCmd)
}
func LoggerRole() string {
	return viper.GetString(loggerRole)
}
func LoggerRequestsCPU() string {
	return viper.GetString(loggerRequestsCPU)
}
func LoggerRequestsMemory() string {
	return viper.GetString(loggerRequestsMemory)
}
func LoggerRequestsEphemeral() string {
	return viper.GetString(loggerRequestsEphemeral)
}
func LoggerLimitsCPU() string {
	return viper.GetString(loggerLimitsCPU)
}
func LoggerLimitsMemory() string {
	return viper.GetString(loggerLimitsMemory)
}
func LoggerLimitsEphemeral() string {
	return viper.GetString(loggerLimitsEphemeral)
}
func LoggerLogVolumeName() string {
	return viper.GetString(loggerLogVolumeName)
}
func LoggerLogVolumePath() string {
	return viper.GetString(loggerLogVolumePath)
}
func LoggerS3Bucket() string {
	return viper.GetString(loggerS3Bucket)
}
func LoggerRecorderDir() string {
	return viper.GetString(loggerRecorderDir)
}
func LoggerRabbitURI() string {
	return viper.GetString(loggerRabbitURI)
}
func LoggerRabbitQueue() string {
	return viper.GetString(loggerRabbitQueue)
}
func LoggerS3URLEnvKey() string {
	return viper.GetString(loggerS3URLEnvKey)
}
func LoggerS3AccessKeyEnvKey() string {
	return viper.GetString(loggerS3AccessKeyEnvKey)
}
func LoggerS3SecretKeyEnvKey() string {
	return viper.GetString(loggerS3SecretKeyEnvKey)
}
func LoggerS3BucketEnvKey() string {
	return viper.GetString(loggerS3BucketEnvKey)
}
func LoggerRecorderDirEnvKey() string {
	return viper.GetString(loggerRecorderDirEnvKey)
}
func LoggerRabbitURIEnvKey() string {
	return viper.GetString(loggerRabbitURIEnvKey)
}
func LoggerRabbitQueueEnvKey() string {
	return viper.GetString(loggerRabbitQueueEnvKey)
}

///////////////////////////////// SharedVolume

func SharedVolumeName() string {
	return viper.GetString(sharedVolumeName)
}
func SharedVolumePath() string {
	return viper.GetString(sharedVolumePath)
}

///////////////////////////////// Namespace

func RoomNamespace() string {
	return viper.GetString(roomNamespace)
}
func RoomEndOfGameKey() string {
	return viper.GetString(roomEndOfGameKey)
}
func ClientIDEnvKey() string {
	return viper.GetString(clientIDEnvKey)
}
func RoomIDEnvKey() string {
	return viper.GetString(roomIDEnvKey)
}
func RoomEndOfGameKeyEnvKey() string {
	return viper.GetString(roomEndOfGameKeyEnvKey)
}

///////////////////////////////// DefaultResources

func DefaultRequestsCPU() string {
	return viper.GetString(defaultRequestsCPU)
}
func DefaultRequestsMemory() string {
	return viper.GetString(defaultRequestsMemory)
}
func DefaultRequestsEphemeral() string {
	return viper.GetString(defaultRequestsEphemeral)
}
func DefaultLimitsCPU() string {
	return viper.GetString(defaultLimitsCPU)
}
func DefaultLimitsMemory() string {
	return viper.GetString(defaultLimitsMemory)
}
func DefaultLimitsEphemeral() string {
	return viper.GetString(defaultLimitsEphemeral)
}

///////////////////////////////// ConfigMap Key for storing configs

func ConfigMapItemKey() string {
	return viper.GetString(configMapItemKey)
}
