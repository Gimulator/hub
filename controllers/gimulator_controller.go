package controllers

import (
	"context"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	hubv1 "github.com/Gimulator/hub/api/v1"
	"github.com/Gimulator/hub/pkg/client"
	"github.com/Gimulator/hub/pkg/config"
	"github.com/Gimulator/hub/pkg/name"
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

func (g *gimulatorReconciler) reconcileGimulator(ctx context.Context, room *hubv1.Room, gameConfig config.GameConfig) error {
	logger := g.Log.WithValues("reconciler", "Gimulator", "room", room.Spec.ID)

	if gameConfig.GimulatorImage == "" {
		logger.Info("no need to gimulator")
		return nil
	}

	key := types.NamespacedName{
		Name:      name.GimulatorPodName(room.Spec.ID),
		Namespace: room.Namespace,
	}

	logger.Info("starting to get gimulator's Pod")
	pod, err := g.GetPod(ctx, key)
	if err != nil && !errors.IsNotFound(err) {
		logger.Error(err, "could not get gimulator's Pod")
		return err
	} else if err == nil {
		logger.Info("starting to update status of gimulator because pod is present")
		g.updateGimulatorStatus(room, pod)
		return nil
	} else {
		logger.Info("gimulator's pod is not found")
	}

	logger.Info("starting to create gimulator's manifest")
	pod = g.gimulatorPodManifest(gameConfig)

	logger.Info("starting to sync actor's pod")
	syncedPod, err := g.SyncPod(ctx, pod)
	if err != nil {
		logger.Error(err, "could not sync gimulator's pod")
		return err
	}

	logger.Info("starting to update status of gimulator")
	g.updateGimulatorStatus(room, syncedPod)

	return nil
}

func (g *gimulatorReconciler) reconcileGimulatorService(ctx context.Context, gameConfig config.GameConfig) error {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name.GimulatorServiceName(gameConfig.RoomID),
			Namespace: gameConfig.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Port: name.GimulatorServicePort(),
				},
			},
			Selector: map[string]string{
				name.PodTypeLabel(): name.PodTypeGimulator(),
				name.RoomIDLabel():  gameConfig.RoomID,
			},
		},
	}

	_, err := g.SyncService(ctx, service)
	return err
}

func (g *gimulatorReconciler) gimulatorPodManifest(gameConfig config.GameConfig) *corev1.Pod {
	labels := map[string]string{
		name.PodTypeLabel(): name.PodTypeGimulator(),
		name.RoomIDLabel():  gameConfig.RoomID,
	}

	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name.GimulatorPodName(gameConfig.RoomID),
			Namespace: gameConfig.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			RestartPolicy: corev1.RestartPolicyNever,
			Containers: []corev1.Container{
				{
					Name:            name.GimulatorContainerName(),
					Image:           gameConfig.GimulatorImage,
					ImagePullPolicy: corev1.PullIfNotPresent,
					Resources:       corev1.ResourceRequirements{},
					Env:             []corev1.EnvVar{},
				},
			},
		},
	}
}

func (g *gimulatorReconciler) updateGimulatorStatus(room *hubv1.Room, pod *corev1.Pod) {
	room.Status.ActorStatuses[name.GimulatorContainerName()] = pod.Status.DeepCopy()
}
