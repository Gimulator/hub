package controllers

import (
	"context"
	"os"
	"strconv"

	"github.com/go-logr/logr"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	hubv1 "github.com/Gimulator/hub/api/v1"
	"github.com/Gimulator/hub/pkg/client"
	"github.com/Gimulator/hub/pkg/config"
	"github.com/Gimulator/hub/pkg/name"
	"github.com/Gimulator/protobuf/go/api"
)

// gimulatorReconciler reconciles Gimulator for a Room
type gimulatorReconciler struct {
	*client.Client
	Log logr.Logger
}

// newGimulatorReconciler returns new instance of GimulatorReconciler
func newGimulatorReconciler(client *client.Client, log logr.Logger) (*gimulatorReconciler, error) {
	return &gimulatorReconciler{
		Log:    log,
		Client: client,
	}, nil
}

func (g *gimulatorReconciler) reconcileGimulator(ctx context.Context, room *hubv1.Room) error {
	logger := g.Log.WithValues("reconciler", "Gimulator", "room", room.Spec.ID)

	logger.Info("starting to reconcile rulse config map")
	if err := g.reconcileRulesConfigMap(ctx, room); err != nil {
		logger.Error(err, "could not reconcile rulse config map")
		return err
	}

	logger.Info("starting to reconcile credentials config map")
	if err := g.reconcileCredentialsConfigMap(ctx, room); err != nil {
		logger.Error(err, "could not reconcile credentials config map")
		return err
	}

	logger.Info("starting to reconcile gimulator's service")
	if err := g.reconcileGimulatorService(ctx, room); err != nil {
		logger.Error(err, "could not reconcile gimulator's service")
		return err
	}

	logger.Info("starting to create gimulator's manifest")
	gimPod, err := g.gimulatorPodManifest(room)
	if err != nil {
		logger.Error(err, "could not create gimulator's manifest")
		return err
	}

	logger.Info("starting to sync gimulator's pod")
	syncedGimPod, err := g.SyncPod(ctx, gimPod, room)
	if err != nil {
		logger.Error(err, "could not sync gimulator's pod")
		return err
	}

	logger.Info("starting to update status of gimulator")
	g.updateGimulatorStatus(room, syncedGimPod)

	return nil
}

func (g *gimulatorReconciler) reconcileRulesConfigMap(ctx context.Context, room *hubv1.Room) error {
	key := types.NamespacedName{
		Name:      name.RulesConfigMapName(room.Spec.ProblemID),
		Namespace: room.Namespace,
	}

	_, err := g.GetConfigMap(ctx, key)
	if err == nil {
		return nil
	}

	rules, err := config.FetchRules(room)
	if err != nil {
		return err
	}

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name.RulesConfigMapName(room.Spec.ProblemID),
			Namespace: room.Namespace,
		},
		Data: map[string]string{
			"data": rules,
		},
	}

	if _, err := g.SyncConfigMap(ctx, configMap, nil); err != nil {
		return err
	}

	return nil
}

func (g *gimulatorReconciler) reconcileCredentialsConfigMap(ctx context.Context, room *hubv1.Room) error {
	// TODO: change to string instead of api enums
	type Cred struct {
		Name      string `yaml:"name"`
		Character string `yaml:"character"`
		Role      string `yaml:"role"`
		Token     string `yaml:"token"`
	}
	creds := make([]Cred, 0)

	creds = append(creds, Cred{
		Name:      room.Spec.Director.Name,
		Character: api.Character_name[int32(api.Character_director)],
		Role:      name.CharacterDirector(),
		Token:     room.Spec.Director.Token,
	})

	for _, actor := range room.Spec.Actors {
		creds = append(creds, Cred{
			Name:      actor.Name,
			Character: api.Character_name[int32(api.Character_actor)],
			Role:      actor.Role,
			Token:     actor.Token,
		})
	}

	token := os.Getenv("HUB_GIMULATOR_TOKEN")
	creds = append(creds, Cred{
		Name:      "hub-manager",
		Character: api.Character_name[int32(api.Character_operator)],
		Role:      name.CharacterOperator(),
		Token:     token,
	})

	bytes, err := yaml.Marshal(creds)
	if err != nil {
		return err
	}

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name.CredConfigMapName(room.Spec.ID),
			Namespace: room.Namespace,
		},
		Data: map[string]string{
			"data": string(bytes),
		},
	}

	if _, err := g.SyncConfigMap(ctx, configMap, room); err != nil {
		return err
	}

	return nil
}

func (g *gimulatorReconciler) reconcileGimulatorService(ctx context.Context, room *hubv1.Room) error {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name.GimulatorServiceName(room.Spec.ID),
			Namespace: room.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Port: int32(name.GimulatorServicePort()),
				},
			},
			Selector: map[string]string{
				name.CharacterLabel(): name.CharacterGimulator(),
				name.RoomLabel():      room.Spec.ID,
				name.ProblemLabel():   room.Spec.ProblemID,
			},
			Type:      corev1.ServiceTypeClusterIP,
			ClusterIP: corev1.ClusterIPNone,
		},
	}

	_, err := g.SyncService(ctx, service, room)
	return err
}

func (g *gimulatorReconciler) gimulatorPodManifest(room *hubv1.Room) (*corev1.Pod, error) {
	cpu, err := resource.ParseQuantity(name.GimulatorCPULimit())
	if err != nil {
		return nil, err
	}

	memory, err := resource.ParseQuantity(name.GimulatorMemoryLimit())
	if err != nil {
		return nil, err
	}

	ephemeral, err := resource.ParseQuantity(name.GimulatorEphemeralLimit())
	if err != nil {
		return nil, err
	}

	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name.GimulatorPodName(room.Spec.ID),
			Namespace: room.Namespace,
			Labels: map[string]string{
				name.CharacterLabel(): name.CharacterGimulator(),
				name.RoleLabel():      name.CharacterGimulator(),
				name.RoomLabel():      room.Spec.ID,
				name.ProblemLabel():   room.Spec.ProblemID,
			},
		},
		Spec: corev1.PodSpec{
			RestartPolicy: corev1.RestartPolicyNever,
			Containers: []corev1.Container{
				{
					Name:            name.GimulatorContainerName(),
					Image:           room.Spec.Setting.GimulatorImage,
					ImagePullPolicy: corev1.PullIfNotPresent,
					Env: []corev1.EnvVar{
						{
							Name:  "GIMULATOR_HOST",
							Value: "0.0.0.0:" + strconv.Itoa(name.GimulatorServicePort()),
						},
						{
							Name:  "GIMULATOR_CONFIG_DIR",
							Value: name.GimulatorConfigDir(),
						},
						{
							Name: "GIMULATOR_RABBIT_HOST",
							ValueFrom: &corev1.EnvVarSource{
								SecretKeyRef: &corev1.SecretKeySelector{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "rabbit-credentials",
									},
									Key: "host",
								},
							},
						},
						{
							Name: "GIMULATOR_RABBIT_USERNAME",
							ValueFrom: &corev1.EnvVarSource{
								SecretKeyRef: &corev1.SecretKeySelector{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "rabbit-credentials",
									},
									Key: "username",
								},
							},
						},
						{
							Name: "GIMULATOR_RABBIT_PASSWORD",
							ValueFrom: &corev1.EnvVarSource{
								SecretKeyRef: &corev1.SecretKeySelector{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "rabbit-credentials",
									},
									Key: "password",
								},
							},
						},
						{
							Name: "GIMULATOR_RABBIT_RESULT_QUEUE",
							ValueFrom: &corev1.EnvVarSource{
								SecretKeyRef: &corev1.SecretKeySelector{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "rabbit-credentials",
									},
									Key: "result-queue",
								},
							},
						},
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      name.GimulatorConfigVolumeName(),
							MountPath: name.GimulatorConfigMountPath(),
							ReadOnly:  true,
						},
					},
					Resources: corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:              cpu,
							corev1.ResourceMemory:           memory,
							corev1.ResourceEphemeralStorage: ephemeral,
						},
					},
				},
			},
			Volumes: []corev1.Volume{
				{
					Name: name.GimulatorConfigVolumeName(),
					VolumeSource: corev1.VolumeSource{
						Projected: &corev1.ProjectedVolumeSource{
							Sources: []corev1.VolumeProjection{
								{
									ConfigMap: &corev1.ConfigMapProjection{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: name.RulesConfigMapName(room.Spec.ProblemID),
										},
										Items: []corev1.KeyToPath{
											{
												Key:  "data",
												Path: "rules.yaml",
											},
										},
									},
								},
								{
									ConfigMap: &corev1.ConfigMapProjection{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: name.CredConfigMapName(room.Spec.ID),
										},
										Items: []corev1.KeyToPath{
											{
												Key:  "data",
												Path: "credentials.yaml",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}, nil
}

func (g *gimulatorReconciler) updateGimulatorStatus(room *hubv1.Room, pod *corev1.Pod) {
	phase := pod.Status.DeepCopy().Phase

	room.Status.GimulatorStatus = phase
}
