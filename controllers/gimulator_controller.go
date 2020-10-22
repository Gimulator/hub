package controllers

import (
	"context"

	"github.com/go-logr/logr"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	hubv1 "github.com/Gimulator/hub/api/v1"
	"github.com/Gimulator/hub/pkg/client"
	"github.com/Gimulator/hub/pkg/name"
	"github.com/Gimulator/hub/pkg/s3"
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

	if room.Spec.GameConfig.GimulatorImage == "" {
		logger.Info("this game doesn't need gimulator")
		return nil
	}

	logger.Info("starting to reconcile rolse config map")
	if err := g.reconcileRolesConfigMap(ctx, room); err != nil {
		logger.Error(err, "could not reconcile rolse config map")
		return err
	}

	logger.Info("starting to reconcile credentials config map")
	if err := g.reconcileCredentialsConfigMap(ctx, room); err != nil {
		logger.Error(err, "could not reconcile credentials config map")
		return err
	}

	logger.Info("starting to create gimulator's manifest")
	gimPod := g.gimulatorPodManifest(room)

	logger.Info("starting to sync gimulator's pod")
	syncedGimPod, err := g.SyncPod(ctx, gimPod, room)
	if err != nil {
		logger.Error(err, "could not sync gimulator's pod")
		return err
	}

	logger.Info("starting to reconcile gimulator's service")
	if err := g.reconcileGimulatorService(ctx, room); err != nil {
		logger.Error(err, "could not reconcile gimulator's service")
		return err
	}

	logger.Info("starting to update status of gimulator")
	g.updateGimulatorStatus(room, syncedGimPod)

	return nil
}

func (g *gimulatorReconciler) reconcileRolesConfigMap(ctx context.Context, room *hubv1.Room) error {
	key := types.NamespacedName{
		Name:      name.RolesConfigMapName(room.Spec.ProblemID),
		Namespace: room.Namespace,
	}

	_, err := g.GetConfigMap(ctx, key)
	if err == nil {
		return nil
	}

	bytes, err := s3.GetBytes(name.S3RoleBucket(), name.RolesConfigMapName(room.Spec.ProblemID))
	if err != nil {
		return err
	}

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name.RolesConfigMapName(room.Spec.ProblemID),
			Namespace: room.Namespace,
		},
		Data: map[string]string{
			"data": string(bytes),
		},
	}

	if _, err := g.SyncConfigMap(ctx, configMap, nil); err != nil {
		return err
	}

	return nil
}

func (g *gimulatorReconciler) reconcileCredentialsConfigMap(ctx context.Context, room *hubv1.Room) error {
	roles := make(map[string]string)

	roles[room.Spec.Director.ID] = name.DirectorRoleName()

	for _, actor := range room.Spec.Actors {
		roles[actor.ID] = actor.Role
	}

	bytes, err := yaml.Marshal(roles)
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
					Port: name.GimulatorServicePort(),
				},
			},
			Selector: map[string]string{
				name.PodTypeLabel(): name.PodTypeGimulator(),
				name.RoomIDLabel():  room.Spec.ID,
			},
		},
	}

	_, err := g.SyncService(ctx, service, room)
	return err
}

func (g *gimulatorReconciler) gimulatorPodManifest(room *hubv1.Room) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name.GimulatorPodName(room.Spec.ID),
			Namespace: room.Namespace,
			Labels: map[string]string{
				name.PodTypeLabel(): name.PodTypeGimulator(),
				name.RoomIDLabel():  room.Spec.ID,
			},
		},
		Spec: corev1.PodSpec{
			RestartPolicy: corev1.RestartPolicyNever,
			Containers: []corev1.Container{
				{
					Name:            name.GimulatorContainerName(),
					Image:           room.Spec.GameConfig.GimulatorImage,
					ImagePullPolicy: corev1.PullIfNotPresent,
					Resources:       corev1.ResourceRequirements{},
					Env:             []corev1.EnvVar{},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      name.RolesVolumeName(),
							MountPath: name.RolesVolumeMountPath(),
							ReadOnly:  true,
						},
						{
							Name:      name.CredsVolumeName(),
							MountPath: name.CredsVolumeMountPath(),
							ReadOnly:  true,
						},
					},
				},
			},
			Volumes: []corev1.Volume{
				{
					Name: name.RolesVolumeName(),
					VolumeSource: corev1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: name.RolesConfigMapName(room.Spec.ProblemID),
							},
							Items: []corev1.KeyToPath{
								{
									Key:  "data",
									Path: "roles.yaml",
								},
							},
						},
					},
				},
				{
					Name: name.CredsVolumeName(),
					VolumeSource: corev1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{
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
	}
}

func (g *gimulatorReconciler) updateGimulatorStatus(room *hubv1.Room, pod *corev1.Pod) {
	room.Status.ActorStatuses[name.GimulatorContainerName()] = pod.Status.DeepCopy()
}
