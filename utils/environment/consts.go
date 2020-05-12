package environment

import "time"

const (
	keyS3AccessKey = "s3-access-key"
	keyS3SecretKey = "s3-secret-key"
	keyS3URL       = "s3-url"

	keyGimulatorName                      = "gimulator-name"
	keyGimulatorID                        = "gimulator-id"
	keyGimulatorImage                     = "gimulator-image"
	keyGimulatorType                      = "gimulator-type"
	keyGimulatorCmd                       = "gimulator-command"
	keyGimulatorConfigVolumeName          = "gimulator-config-volume-name"
	keyGimulatorConfigVolumePath          = "gimulator-config-volume-path"
	keyGimulatorConfigMapName             = "gimulator-config-map-name"
	keyGimulatorResourceRequestsCPU       = "gimulator-resource-requests-cpu"
	keyGimulatorResourceRequestsMemory    = "gimulator-resource-requests-memory"
	keyGimulatorResourceRequestsEphemeral = "gimulator-resource-requests-ephemeral"
	keyGimulatorResourceLimitsCPU         = "gimulator-resource-limits-cpu"
	keyGimulatorResourceLimitsMemory      = "gimulator-resource-limits-memory"
	keyGimulatorResourceLimitsEphemeral   = "gimulator-resource-limits-ephemeral"
	keyGimulatorIP                        = "gimulator-ip"
	keyGimulatorEndOfGame                 = "gimulator-end-of-game-key"

	keyLoggerName                      = "logger-name"
	keyLoggerID                        = "logger-id"
	keyLoggerImage                     = "logger-image"
	keyLoggerType                      = "logger-type"
	keyLoggerCmd                       = "logger-command"
	keyLoggerRole                      = "logger-role"
	keyLoggerLogDirName                = "logger-log-dir-name"
	keyLoggerLogDirPath                = "logger-log-dir-path"
	keyLoggerResourceRequestsCPU       = "logger-resource-requests-cpu"
	keyLoggerResourceRequestsMemory    = "logger-resource-requests-memory"
	keyLoggerResourceRequestsEphemeral = "logger-resource-requests-ephemeral"
	keyLoggerResourceLimitsCPU         = "logger-resource-limits-cpu"
	keyLoggerResourceLimitsMemory      = "logger-resource-limits-memory"
	keyLoggerResourceLimitsEphemeral   = "logger-resource-limits-ephemeral"
	keyLoggerS3Bucket                  = "logger-s3-bucket"
	keyLoggerRabbitURI                 = "logger-rabbit-uri"
	keyLoggerRabbitQueue               = "logger-rabbit-queue"
	keyLoggerRecordDir                 = "logger-record-dir"

	keySharedVolumeName = "shared-volume-name"
	keySharedVolumePath = "shared-volume-path"

	keyNamespace = "namespace"

	keyDefaultResourceRequestsCPU       = "default-resource-requests-cpu"
	keyDefaultResourceRequestsMemory    = "default-resource-requests-memory"
	keyDefaultResourceRequestsEphemeral = "default-resource-requests-ephemeral"
	keyDefaultResourceLimitsCPU         = "default-resource-limits-cpu"
	keyDefaultResourceLimitsMemory      = "default-resource-limits-memory"
	keyDefaultResourceLimitsEphemeral   = "default-resource-limits-ephemeral"
)

type ContainerType string

const (
	Finisher ContainerType = "finisher"
	Master   ContainerType = "master"
	Slave    ContainerType = "slave"

	UsernameEnvVarKey = "username"
	PasswordEnvVarKey = "password"
	RoleEnvVarKey     = "role"

	APICallTimeout = time.Second * 5

	CacheExpirationTime  = time.Hour * 6
	CacheCleanupInterval = time.Hour * 12
)

const FinisherArgs = `trap "touch %s" EXIT
%s
`

const SlaveArgs = `%s &
CHILD_PID=$!
while kill -0 $CHILD_PID 2> /dev/null; do
    if [[ %s ]]
    then
        kill $CHILD_PID
        break
    fi
    sleep 1
done &
wait $CHILD_PID
exit 0
`

const MasterArgs = `%s &
CHILD_PID=$!
while kill -0 $CHILD_PID 2> /dev/null; do
    if [[ %s ]]
    then
        kill $CHILD_PID
        break
    fi
    sleep 1
done &
wait $CHILD_PID
STATUS=$?
if [[ %s ]]
then
    exit 0
else
    if [[ "$STATUS" -eq "0" ]]; then
        exit 0
    else
        exit 1
    fi
fi
`
