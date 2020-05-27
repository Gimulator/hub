package environment

import "time"

const (
	envvarkeyLoggerS3URL           = "env-var-key_logger-s3-url"
	envvarkeyLoggerS3AccessKey     = "env-var-key_logger-s3-access-key"
	envvarkeyLoggerS3SecretKey     = "env-var-key_logger-s3-secret-key"
	envvarkeyLoggerS3Bucket        = "env-var-key_logger-s3-bucket"
	envvarkeyLoggerRecorderDir     = "env-var-key_logger-recorder-dir"
	envvarkeyLoggerRabbitURI       = "env-var-key_logger-rabbit-uri"
	envvarkeyLoggerRabbitQueue     = "env-var-key_logger-rabbit-queue"
	envvarkeyClientID              = "env-var-key_client-id"
	envvarkeyRoomID                = "env-var-key_room-id"
	envvarkeyRoomEndOfGameKey      = "env-var-key_room-end-of-game"
	envvarkeyGimulatorHost         = "env-var-key_gimulator-host"
	envvarkeyGimulatorRoleFilePath = "env-var-key_gimulator-role-file-path"

	envvarvalLoggerS3Bucket        = "env-var-val_logger-s3-bucket"
	envvarvalLoggerRecorderDir     = "env-var-val_logger-recorder-dir"
	envvarvalLoggerRabbitURI       = "env-var-val_logger-rabbit-uri"
	envvarvalLoggerRabbitQueue     = "env-var-val_logger-rabbit-queue"
	envvarvalRoomEndOfGameKey      = "env-var-val_room-end-of-game"
	envvarvalGimulatorHost         = "env-var-val_gimulator-host"
	envvarvalGimulatorRoleFilePath = "env-var-val_gimulator-roles-file-path"

	s3AccessKey = "S3_ACCESS_KEY"
	s3SecretKey = "S3_SECRET_KEY"
	s3URL       = "S3_URL"

	keyGimulatorName                      = "gimulator-name"
	keyGimulatorID                        = "gimulator-id"
	keyGimulatorImage                     = "gimulator-image"
	keyGimulatorType                      = "gimulator-type"
	keyGimulatorCmd                       = "gimulator-command"
	keyGimulatorConfigVolumeName          = "gimulator-config-volume-name"
	keyGimulatorConfigVolumePath          = "gimulator-config-volume-path"
	keyGimulatorResourceRequestsCPU       = "gimulator-resource-requests-cpu"
	keyGimulatorResourceRequestsMemory    = "gimulator-resource-requests-memory"
	keyGimulatorResourceRequestsEphemeral = "gimulator-resource-requests-ephemeral"
	keyGimulatorResourceLimitsCPU         = "gimulator-resource-limits-cpu"
	keyGimulatorResourceLimitsMemory      = "gimulator-resource-limits-memory"
	keyGimulatorResourceLimitsEphemeral   = "gimulator-resource-limits-ephemeral"

	keyLoggerName                      = "logger-name"
	keyLoggerID                        = "logger-id"
	keyLoggerImage                     = "logger-image"
	keyLoggerType                      = "logger-type"
	keyLoggerCmd                       = "logger-command"
	keyLoggerRole                      = "logger-role"
	keyLoggerResourceRequestsCPU       = "logger-resource-requests-cpu"
	keyLoggerResourceRequestsMemory    = "logger-resource-requests-memory"
	keyLoggerResourceRequestsEphemeral = "logger-resource-requests-ephemeral"
	keyLoggerResourceLimitsCPU         = "logger-resource-limits-cpu"
	keyLoggerResourceLimitsMemory      = "logger-resource-limits-memory"
	keyLoggerResourceLimitsEphemeral   = "logger-resource-limits-ephemeral"

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
	ContainerTypeFinisher ContainerType = "finisher"
	ContainerTypeMaster   ContainerType = "master"
	ContainerTypeSlave    ContainerType = "slave"

	APICallTimeout = time.Second * 5

	CacheExpirationTime  = time.Minute * 6
	CacheCleanupInterval = time.Minute * 6
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
tail -f
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
tail -f
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
