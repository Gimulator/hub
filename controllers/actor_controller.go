package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	hubv1 "github.com/Gimulator/hub/api/v1"
	"github.com/Gimulator/hub/pkg/client"
	"github.com/Gimulator/hub/pkg/name"
)

// actorReconciler reconciles an actor of a Room object
type actorReconciler struct {
	*client.Client
	Log logr.Logger
}

// newActorReconciler returns new instance of ActorReconciler
func newActorReconciler(client *client.Client, log logr.Logger) (*actorReconciler, error) {
	return &actorReconciler{
		Log:    log,
		Client: client,
	}, nil
}

func (a *actorReconciler) reconcileActor(ctx context.Context, room *hubv1.Room, actor *hubv1.Actor) error {
	logger := a.Log.WithValues("reconciler", "Actor", "actor", actor.ID, "room", room.Spec.ID)

	logger.Info("starting to reconcile actor's output PVC")
	if err := a.reconcileOutputPVC(ctx, actor, room); err != nil {
		logger.Error(err, "could not reconcile actor's output PVC")
		return err
	}

	logger.Info("starting to create actor's manifest")
	actorPod, err := a.actorPodManifest(actor, room)
	if err != nil {
		logger.Error(err, "could not create actor's manifest")
		return err
	}

	logger.Info("starting to sync actor's pod")
	syncedActorPod, err := a.SyncPod(ctx, actorPod, room)
	if err != nil {
		logger.Error(err, "could not sync actor's pod")
		return err
	}

	logger.Info("starting to update status of actor")
	a.updateActorStatus(room, actor, syncedActorPod)

	return nil
}

func (a *actorReconciler) reconcileOutputPVC(ctx context.Context, actor *hubv1.Actor, room *hubv1.Room) error {
	quantity, err := resource.ParseQuantity(room.Spec.ProblemSettings.OutputVolumeSize)
	if err != nil {
		return err
	}

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name.ActorOutputPVCName(actor.ID),
			Namespace: room.Namespace,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			Resources: corev1.ResourceRequirements{
				Requests: map[corev1.ResourceName]resource.Quantity{
					corev1.ResourceStorage: quantity,
				},
			},
		},
	}

	_, err = a.SyncPVC(ctx, pvc, room)
	return err
}

func (a *actorReconciler) updateActorStatus(room *hubv1.Room, actor *hubv1.Actor, pod *corev1.Pod) {
	room.Status.ActorStatuses[actor.ID] = *pod.Status.DeepCopy()
}

func (a *actorReconciler) actorPodManifest(actor *hubv1.Actor, room *hubv1.Room) (*corev1.Pod, error) {
	volumes := make([]corev1.Volume, 0)
	volumeMounts := make([]corev1.VolumeMount, 0)

	volumes = append(volumes, corev1.Volume{
		Name: name.OutputVolumeName(actor.ID),
		VolumeSource: corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: name.ActorOutputPVCName(actor.ID),
			},
		},
	})

	volumeMounts = append(volumeMounts, corev1.VolumeMount{
		Name:      name.OutputVolumeName(actor.ID),
		MountPath: name.OutputVolumeMountPath(),
	})

	if room.Spec.ProblemSettings.DataPVCName != "" {
		volumes = append(volumes, corev1.Volume{
			Name: name.DataVolumeName(),
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: room.Spec.ProblemSettings.DataPVCName,
					ReadOnly:  true,
				},
			},
		})
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      name.DataVolumeName(),
			MountPath: name.DataVolumeMountPath(),
			ReadOnly:  true,
		})
	}

	labels := map[string]string{
		name.ActorIDLabel(): actor.ID,
		name.RoomIDLabel():  room.Spec.ID,
		name.PodTypeLabel(): name.PodTypeActor(),
	}

	cpu, err := resource.ParseQuantity(room.Spec.ProblemSettings.ResourceCPULimit)
	if err != nil {
		return nil, err
	}

	memory, err := resource.ParseQuantity(room.Spec.ProblemSettings.ResourceMemoryLimit)
	if err != nil {
		return nil, err
	}

	ephemeral, err := resource.ParseQuantity(room.Spec.ProblemSettings.ResourceEphemeralLimit)
	if err != nil {
		return nil, err
	}

	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name.ActorPodName(actor.ID),
			Namespace: room.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Volumes:       volumes,
			RestartPolicy: corev1.RestartPolicyNever,
			Containers: []corev1.Container{
				{
					Name:            name.ActorContainerName(),
					Image:           actor.Image,
					ImagePullPolicy: corev1.PullIfNotPresent,
					VolumeMounts:    volumeMounts,
					Env: []corev1.EnvVar{
						{
							Name:  "GIMULATOR_HOST",
							Value: fmt.Sprintf("%s:%d", name.GimulatorServiceName(room.Spec.ID), name.GimulatorServicePort()),
						},
						{
							Name:  "GIMULATOR_CLIENT_ID",
							Value: actor.ID,
						},
						{
							Name:  "GIMULATOR_ROLE",
							Value: actor.Role,
						},
						{
							Name:  "GIMULATOR_TOKEN",
							Value: actor.Token,
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
		},
	}, nil
}
