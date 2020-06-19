package environment

import "time"

const (
	s3AccessKey = "S3_ACCESS_KEY"
	s3SecretKey = "S3_SECRET_KEY"
	s3URL       = "S3_URL"

	rabbitURI   = "RABBIT_URI"
	rabbitQueue = "RABBIT_QUEUE"

	gimulatorName               = "gimulator-name"
	gimulatorID                 = "gimulator-id"
	gimulatorImage              = "gimulator-image"
	gimulatorType               = "gimulator-type"
	gimulatorCmd                = "gimulator-command"
	gimulatorConfigVolumeName   = "gimulator-config-volume-name"
	gimulatorConfigVolumePath   = "gimulator-config-volume-path"
	gimulatorRequestsCPU        = "gimulator-requests-cpu"
	gimulatorRequestsMemory     = "gimulator-requests-memory"
	gimulatorRequestsEphemeral  = "gimulator-requests-ephemeral"
	gimulatorLimitsCPU          = "gimulator-limits-cpu"
	gimulatorLimitsMemory       = "gimulator-limits-memory"
	gimulatorLimitsEphemeral    = "gimulator-limits-ephemeral"
	gimulatorHost               = "gimulator-host"
	gimulatorRoleFileName       = "gimulator-roles-file-name"
	gimulatorHostEnvKey         = "gimulator-host-env-key"
	gimulatorRoleFilePathEnvKey = "gimulator-role-file-path-env-key"

	loggerName              = "logger-name"
	loggerID                = "logger-id"
	loggerImage             = "logger-image"
	loggerType              = "logger-type"
	loggerCmd               = "logger-command"
	loggerRole              = "logger-role"
	loggerRequestsCPU       = "logger-requests-cpu"
	loggerRequestsMemory    = "logger-requests-memory"
	loggerRequestsEphemeral = "logger-requests-ephemeral"
	loggerLimitsCPU         = "logger-limits-cpu"
	loggerLimitsMemory      = "logger-limits-memory"
	loggerLimitsEphemeral   = "logger-limits-ephemeral"
	loggerLogVolumeName     = "logger-log-volume-name"
	loggerLogVolumePath     = "logger-log-volume-Path"
	loggerS3Bucket          = "logger-s3-bucket"
	loggerRecorderDir       = "logger-recorder-dir"
	loggerS3URLEnvKey       = "logger-s3-url-env-key"
	loggerS3AccessKeyEnvKey = "logger-s3-access-key-env-key"
	loggerS3SecretKeyEnvKey = "logger-s3-secret-key-env-key"
	loggerS3BucketEnvKey    = "logger-s3-bucket-env-key"
	loggerRecorderDirEnvKey = "logger-recorder-dir-env-key"
	loggerRabbitURIEnvKey   = "logger-rabbit-uri-env-key"
	loggerRabbitQueueEnvKey = "logger-rabbit-queue-env-key"

	sharedVolumeName = "shared-volume-name"
	sharedVolumePath = "shared-volume-path"

	roomNamespace          = "room-namespace"
	roomEndOfGameKey       = "room-end-of-game"
	roomIDEnvKey           = "room-id-env-key"
	roomEndOfGameKeyEnvKey = "room-end-of-game-env-key"
	clientIDEnvKey         = "client-id-env-key"

	defaultRequestsCPU       = "default-requests-cpu"
	defaultRequestsMemory    = "default-requests-memory"
	defaultRequestsEphemeral = "default-requests-ephemeral"
	defaultLimitsCPU         = "default-limits-cpu"
	defaultLimitsMemory      = "default-limits-memory"
	defaultLimitsEphemeral   = "default-limits-ephemeral"

	configMapItemKey = "config-map-items-key"
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

const FinisherArgs = `
while ! [[ -f "%s/start" ]]; do
	echo ">>>>> waitting for start-file in shared volume"
	sleep 1
done

echo ">>>>> starting finisher app"

trap "touch %s" EXIT
%s

echo ">>>>> end of finisher app"
`

const SlaveArgs = `while ! [[ -f "%s/start" ]]; do
	echo ">>>>> waitting for start file in shared volume"
	sleep 1
done

echo ">>>>> starting slave app"

%s &> /dev/null
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
echo ">>>>> end of slave app"
exit 0
`

const MasterArgs = `echo ">>>>> start..."
while ! [[ -f "%s/start" ]]; do
	echo ">>>>> waitting for start file in shared volume"
	sleep 1
done

echo ">>>>> starting master app"

%s &
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
echo ">>>>> end of master app"
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
fi`

const GimulatorArgs = `%s &
CHILD_PID=$!
echo ">>>>> starting gimulator app"
sleep 3

echo ">>>>> creating start file"
touch %s/start

while kill -0 $CHILD_PID 2> /dev/null; do
    if [[ %s ]]
    then
        kill $CHILD_PID
        break
    fi
    sleep 1
done &
wait $CHILD_PID
echo ">>>>> end of gimulator app"
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
fi`
