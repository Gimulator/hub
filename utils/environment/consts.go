package environment

import "time"

const (
	s3AccessKeyKey = "s3-access-key"
	s3SecretKeyKey = "s3-secret-key"
	s3TokenKey     = "s3-token-key"
	s3URLKey       = "s3-url-key"
)

type ContainerType string

const (
	Finisher ContainerType = "finisher"
	Master   ContainerType = "master"
	Slave    ContainerType = "slave"
)

const (
	gimulatorNameKey             = "gimulator-name"
	gimulatorIDKey               = "gimulator-id"
	gimulatorImageKey            = "gimulator-image"
	gimulatorTypeKey             = "gimulator-type"
	gimulatorCmdKey              = "gimulator-cmd"
	gimulatorConfigVolumeNameKey = "gimulator-config-volume-name"
	gimulatorConfigVolumePathKey = "gimulator-config-volume-path"
	gimulatorConfigMapNameKey    = "gimulator-config-map-name"

	gimulatorResourceRequestsCPUKey       = "gimulator-resource-requests-cpu"
	gimulatorResourceRequestsMemoryKey    = "gimulator-resource-requests-memory"
	gimulatorResourceRequestsEphemeralKey = "gimulator-resource-requests-ephemeral"

	gimulatorResourceLimitsCPUKey       = "gimulator-resource-limits-cpu"
	gimulatorResourceLimitsMemoryKey    = "gimulator-resource-limits-memory"
	gimulatorResourceLimitsEphemeralKey = "gimulator-resource-limits-ephemeral"
)

const (
	loggerNameKey       = "logger-name"
	loggerIDKey         = "logger-id"
	loggerImageKey      = "logger-image"
	loggerTypeKey       = "logger-type"
	loggerCmdKey        = "logger-cmd"
	loggerRoleKey       = "logger-role"
	loggerLogDirNameKey = "logger-log-dir-name"
	loggerLogDirPathKey = "logger-log-dir-path"

	loggerResourceRequestsCPUKey       = "logger-resource-requests-cpu"
	loggerResourceRequestsMemoryKey    = "logger-resource-requests-memory"
	loggerResourceRequestsEphemeralKey = "logger-resource-requests-ephemeral"

	loggerResourceLimitsCPUKey       = "logger-resource-limits-cpu"
	loggerResourceLimitsMemoryKey    = "logger-resource-limits-memory"
	loggerResourceLimitsEphemeralKey = "logger-resource-limits-ephemeral"
)

const (
	UsernameEnvVarKey = "username"
	PasswordEnvVarKey = "password"
	RoleEnvVarKey     = "Role"
)

const (
	sharedVolumeNameKey = "shared-volume-name"
	sharedVolumePathKey = "shared-volume-path"
)

const (
	podNamePrefixKey = "pod-prefix"
	namespaceKey     = "namespace"
	restartPolicyKey = "restart-policy"
)

const (
	APICallTimeout = time.Second * 5
)

const (
	resourceRequestsCPUKey       = "resource-requests-cpu"
	resourceRequestsMemoryKey    = "resource-requests-memory"
	resourceRequestsEphemeralKey = "resource-requests-ephemeral"

	resourceLimitsCPUKey       = "resource-limits-cpu"
	resourceLimitsMemoryKey    = "resource-limits-memory"
	resourceLimitsEphemeralKey = "resource-limits-ephemeral"
)

const (
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
