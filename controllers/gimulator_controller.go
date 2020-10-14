package controllers

import (
	"context"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	hubv1 "github.com/Gimulator/hub/api/v1"
	"github.com/Gimulator/hub/pkg/client"
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

func (g *gimulatorReconciler) reconcileGimulator(ctx context.Context, room *hubv1.Room) error {
	logger := g.Log.WithValues("reconciler", "Gimulator", "room", room.Spec.ID)

	if room.Spec.GameConfig.GimulatorImage == "" {
		logger.Info("this game doesn't need gimulator")
		return nil
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
	labels := map[string]string{
		name.PodTypeLabel(): name.PodTypeGimulator(),
		name.RoomIDLabel():  room.Spec.ID,
	}

	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name.GimulatorPodName(room.Spec.ID),
			Namespace: room.Namespace,
			Labels:    labels,
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
				},
			},
		},
	}
}

func (g *gimulatorReconciler) updateGimulatorStatus(room *hubv1.Room, pod *corev1.Pod) {
	room.Status.ActorStatuses[name.GimulatorContainerName()] = pod.Status.DeepCopy()
}
