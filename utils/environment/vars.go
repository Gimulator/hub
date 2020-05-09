package environment

import (
	"fmt"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

func init() {
	viper.SetDefault(keyGimulatorName, "gimulator")
	viper.SetDefault(keyGimulatorID, 123)
	viper.SetDefault(keyGimulatorImage, "gimulator:v1.0")
	viper.SetDefault(keyGimulatorType, Master)
	viper.SetDefault(keyGimulatorCmd, "/app/result")
	viper.SetDefault(keyGimulatorConfigVolumeName, "gimulator-config-path")
	viper.SetDefault(keyGimulatorConfigVolumePath, "/config")
	viper.SetDefault(keyGimulatorConfigMapName, "gimulator-config-map")
	viper.SetDefault(keyGimulatorResourceRequestsCPU, "200m")
	viper.SetDefault(keyGimulatorResourceRequestsMemory, "500M")
	viper.SetDefault(keyGimulatorResourceRequestsEphemeral, "10M")
	viper.SetDefault(keyGimulatorResourceLimitsCPU, "400m")
	viper.SetDefault(keyGimulatorResourceLimitsMemory, "1G")
	viper.SetDefault(keyGimulatorResourceLimitsEphemeral, "20M")

	viper.SetDefault(keyLoggerName, "logger")
	viper.SetDefault(keyLoggerID, 123456789)
	viper.SetDefault(keyLoggerImage, "logger:v1.0")
	viper.SetDefault(keyLoggerType, Finisher)
	viper.SetDefault(keyLoggerCmd, "/app/logger")
	viper.SetDefault(keyLoggerRole, "logger")
	viper.SetDefault(keyLoggerLogDirName, "logger-log-dir")
	viper.SetDefault(keyLoggerLogDirPath, "/tmp")
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
	viper.AddConfigPath(".")

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
	return viper.GetString(keyS3AccessKey)
}
func S3SecretKey() string {
	return viper.GetString(keyS3SecretKey)
}
func S3URL() string {
	return viper.GetString(keyS3URL)
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
	return Master
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
func GimulatorConfigMapName() string {
	return viper.GetString(keyGimulatorConfigMapName)
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
func LoggerType() string {
	return viper.GetString(keyLoggerType)
}
func LoggerCmd() string {
	return viper.GetString(keyLoggerCmd)
}
func LoggerRole() string {
	return viper.GetString(keyLoggerCmd)
}
func LoggerLogDirName() string {
	return viper.GetString(keyLoggerLogDirName)
}
func LoggerLogDirPath() string {
	return viper.GetString(keyLoggerLogDirPath)
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

func ResourceRequestsCPU() string {
	return viper.GetString(keyDefaultResourceRequestsCPU)
}
func ResourceRequestsMemory() string {
	return viper.GetString(keyDefaultResourceRequestsMemory)
}
func ResourceRequestsEphemeral() string {
	return viper.GetString(keyDefaultResourceRequestsEphemeral)
}
func ResourceLimitsCPU() string {
	return viper.GetString(keyDefaultResourceLimitsCPU)
}
func ResourceLimitsMemory() string {
	return viper.GetString(keyDefaultResourceLimitsMemory)
}
func ResourceLimitsEphemeral() string {
	return viper.GetString(keyDefaultResourceLimitsEphemeral)
}
