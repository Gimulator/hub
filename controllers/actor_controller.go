package controllers

import (
	"context"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	hubv1 "github.com/Gimulator/hub/api/v1"
	"github.com/Gimulator/hub/pkg/client"
	"github.com/Gimulator/hub/pkg/config"
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

func (a *actorReconciler) reconcileActor(ctx context.Context, room *hubv1.Room, actor *hubv1.Actor, gameConfig config.GameConfig) error {
	logger := a.Log.WithValues("reconciler", "Actor", "actor", actor.ID, "room", room.Spec.ID)

	key := types.NamespacedName{
		Name:      actor.ID,
		Namespace: room.Namespace,
	}

	logger.Info("starting to get actor's Pod")
	pod, err := a.GetPod(ctx, key)
	if err != nil && !errors.IsNotFound(err) {
		logger.Error(err, "could not get actor's Pod")
		return err
	} else if err == nil {
		logger.Info("starting to update status of actor because pod is present")
		a.updateActorStatus(room, actor, pod)
		return nil
	} else {
		logger.Info("actor's pod is not found")
	}

	logger.Info("starting to reconcile actor's output PVC")
	if err := a.reconcileOutputPVC(ctx, actor, gameConfig); err != nil {
		logger.Error(err, "could not reconcile actor's output PVC")
		return err
	}

	logger.Info("starting to create actor's manifest")
	pod = a.actorPodManifest(actor, gameConfig)

	logger.Info("starting to sync actor's pod")
	syncedPod, err := a.SyncPod(ctx, pod)
	if err != nil {
		logger.Error(err, "could not sync actor's pod")
		return err
	}

	logger.Info("starting to update status of actor")
	a.updateActorStatus(room, actor, syncedPod)

	return nil
}

func (a *actorReconciler) reconcileOutputPVC(ctx context.Context, actor *hubv1.Actor, gameConfig config.GameConfig) error {
	quantity, err := resource.ParseQuantity(gameConfig.OutputVolumeSize)
	if err != nil {
		return err
	}

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name.OutputPVCName(actor.ID),
			Namespace: gameConfig.Namespace,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			Resources: corev1.ResourceRequirements{
				Requests: map[corev1.ResourceName]resource.Quantity{
					corev1.ResourceStorage: quantity,
				},
			},
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadOnlyMany,
			},
		},
	}

	_, err = a.SyncPVC(ctx, pvc)
	return err
}

func (a *actorReconciler) updateActorStatus(room *hubv1.Room, actor *hubv1.Actor, pod *corev1.Pod) {
	room.Status.ActorStatuses[actor.ID] = pod.Status.DeepCopy()
}

func (a *actorReconciler) actorPodManifest(actor *hubv1.Actor, gameConfig config.GameConfig) *corev1.Pod {
	volumes := make([]corev1.Volume, 0)
	volumeMounts := make([]corev1.VolumeMount, 0)

	volumes = append(volumes, corev1.Volume{
		Name: name.OutputVolumeName(),
		VolumeSource: corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: name.OutputPVCName(actor.ID),
			},
		},
	})

	volumeMounts = append(volumeMounts, corev1.VolumeMount{
		Name:      name.OutputVolumeName(),
		MountPath: name.OutputVolumeMountDir(),
	})

	if gameConfig.DataPVCName != "" {
		volumes = append(volumes, corev1.Volume{
			Name: name.DataVolumeName(),
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: gameConfig.DataPVCName,
					ReadOnly:  true,
				},
			},
		})
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      name.DataVolumeName(),
			MountPath: name.DataVolumeMountDir(),
			ReadOnly:  true,
		})
	}

	labels := map[string]string{
		name.ActorIDLabel(): actor.ID,
		name.RoomIDLabel():  gameConfig.RoomID,
		name.PodTypeLabel(): name.PodTypeActor(),
	}

	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name.ActorPodName(actor.ID),
			Namespace: gameConfig.Namespace,
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
					Resources:       corev1.ResourceRequirements{},
					Env:             []corev1.EnvVar{},
				},
			},
		},
	}
}
