package environment

import (
	"fmt"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

func init() {
	viper.SetDefault(s3AccessKeyKey, "default-value")
	viper.SetDefault(s3SecretKeyKey, "default-value")
	viper.SetDefault(s3TokenKey, "default-value")
	viper.SetDefault(s3URLKey, "default-value")

	viper.SetDefault(gimulatorNameKey, "gimulator")
	viper.SetDefault(gimulatorImageKey, "gimulator:v1.0")
	viper.SetDefault(gimulatorTypeKey, Master)
	viper.SetDefault(gimulatorCmdKey, "/app/result")
	viper.SetDefault(gimulatorConfigVolumeNameKey, "gimulator-config-path")
	viper.SetDefault(gimulatorConfigVolumePathKey, "/config")
	viper.SetDefault(gimulatorConfigMapNameKey, "gimulator-config-map")

	// viper.SetDefault(resultNameKey, "result")
	// viper.SetDefault(resultImageKey, "result:v1.0")
	// viper.SetDefault(resultTypeKey, Finisher)
	// viper.SetDefault(resultCmdKey, "/app/result")
	// viper.SetDefault(resultRoleKey, "resutl")

	viper.SetDefault(loggerNameKey, "logger")
	viper.SetDefault(loggerImageKey, "logger:v1.0")
	viper.SetDefault(loggerTypeKey, Finisher)
	viper.SetDefault(loggerCmdKey, "/app/logger")
	viper.SetDefault(loggerRoleKey, "logger")
	viper.SetDefault(loggerLogDirNameKey, "logger-log-dir")
	viper.SetDefault(loggerLogDirPathKey, "/tmp")

	viper.SetDefault(sharedVolumeNameKey, "shared-volume")
	viper.SetDefault(sharedVolumePathKey, "/tmp/pod")

	viper.SetDefault(podNamePrefixKey, "room-")
	viper.SetDefault(namespaceKey, "default")
	viper.SetDefault(restartPolicyKey, "OnFailure")

	viper.SetDefault(resourceRequestsCPUKey, "100m")
	viper.SetDefault(resourceRequestsMemoryKey, "100M")
	viper.SetDefault(resourceRequestsEphemeralKey, "10M")

	viper.SetDefault(resourceLimitsCPUKey, "200m")
	viper.SetDefault(resourceLimitsMemoryKey, "200M")
	viper.SetDefault(resourceLimitsEphemeralKey, "20M")
}

func ReadEnvironments() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		//TODO
		panic(err)
	}
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("Config file changed:", e.Name)
	})
}

func S3AccessKey() string { return viper.GetString(s3AccessKeyKey) }
func S3SecretKey() string { return viper.GetString(s3SecretKeyKey) }
func S3URL() string       { return viper.GetString(s3URLKey) }
func S3Token() string     { return viper.GetString(s3TokenKey) }

func GimulatorName() string             { return viper.GetString(gimulatorNameKey) }
func GimulatorImage() string            { return viper.GetString(gimulatorImageKey) }
func GimulatorType() ContainerType      { return Master }
func GimulatorCmd() string              { return viper.GetString(gimulatorCmdKey) }
func GimulatorConfigVolumeName() string { return viper.GetString(gimulatorConfigVolumeNameKey) }
func GimulatorConfigVolumePath() string { return viper.GetString(gimulatorConfigVolumePathKey) }
func GimulatorConfigMapName() string    { return viper.GetString(gimulatorConfigMapNameKey) }

func LoggerName() string        { return viper.GetString(loggerNameKey) }
func LoggerImage() string       { return viper.GetString(loggerImageKey) }
func LoggerType() ContainerType { return Finisher }
func LoggerCmd() string         { return viper.GetString(loggerCmdKey) }
func LoggerRole() string        { return viper.GetString(loggerCmdKey) }
func LoggerLogDirName() string  { return viper.GetString(loggerLogDirNameKey) }
func LoggerLogDirPath() string  { return viper.GetString(loggerLogDirPathKey) }

// func ResultName() string        { return viper.GetString(resultNameKey) }
// func ResultImage() string       { return viper.GetString(resultImageKey) }
// func ResultType() ContainerType { return Finisher }
// func ResultCmd() string         { return viper.GetString(resultCmdKey) }
// func ResultRole() string        { return viper.GetString(resultCmdKey) }

func SharedVolumeName() string { return viper.GetString(sharedVolumeNameKey) }
func SharedVolumePath() string { return viper.GetString(sharedVolumePathKey) }

func PodNamePrefix() string { return viper.GetString(podNamePrefixKey) }
func Namespace() string     { return viper.GetString(namespaceKey) }
func RestartPolicy() string { return viper.GetString(restartPolicyKey) }

func ResourceRequestsCPU() string       { return viper.GetString(resourceRequestsCPUKey) }
func ResourceRequestsMemory() string    { return viper.GetString(resourceRequestsMemoryKey) }
func ResourceRequestsEphemeral() string { return viper.GetString(resourceRequestsEphemeralKey) }
func ResourceLimitsCPU() string         { return viper.GetString(resourceLimitsCPUKey) }
func ResourceLimitsMemory() string      { return viper.GetString(resourceLimitsMemoryKey) }
func ResourceLimitsEphemeral() string   { return viper.GetString(resourceLimitsEphemeralKey) }
