package controllers

import (
	"context"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	hubv1 "github.com/Gimulator/hub/api/v1"
	"github.com/Gimulator/hub/pkg/client"
	"github.com/Gimulator/hub/pkg/name"
)

// directorReconciler reconciles an director of a Room object
type directorReconciler struct {
	*client.Client
	Log logr.Logger
}

// newDirectorReconciler returns new instance of DirectorReconciler
func newDirectorReconciler(client *client.Client, log logr.Logger) (*directorReconciler, error) {
	return &directorReconciler{
		Log:    log,
		Client: client,
	}, nil
}

func (a *directorReconciler) reconcileDirector(ctx context.Context, room *hubv1.Room) error {
	logger := a.Log.WithValues("reconciler", "Director", "director", room.Spec.Director.ID, "room", room.Spec.ID)

	logger.Info("starting to reconcile director's output PVC")
	if err := a.reconcileOutputPVC(ctx, room); err != nil {
		logger.Error(err, "could not reconcile director's output PVC")
		return err
	}

	logger.Info("starting to create director's manifest")
	dirPod := a.directorPodManifest(room)

	logger.Info("starting to sync director's pod")
	syncedDirPod, err := a.SyncPod(ctx, dirPod, room)
	if err != nil {
		logger.Error(err, "could not sync director's pod")
		return err
	}

	logger.Info("starting to update status of director")
	a.updateDirectorStatus(room, syncedDirPod)

	return nil
}

func (a *directorReconciler) reconcileOutputPVC(ctx context.Context, room *hubv1.Room) error {
	quantity, err := resource.ParseQuantity(room.Spec.GameConfig.OutputVolumeSize)
	if err != nil {
		return err
	}

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name.OutputPVCName(room.Spec.Director.ID),
			Namespace: room.Namespace,
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

	_, err = a.SyncPVC(ctx, pvc, room)
	return err
}

func (a *directorReconciler) updateDirectorStatus(room *hubv1.Room, pod *corev1.Pod) {
	room.Status.DirectorStatus = pod.Status.DeepCopy()
}

func (a *directorReconciler) directorPodManifest(room *hubv1.Room) *corev1.Pod {
	volumes := make([]corev1.Volume, 0)
	volumeMounts := make([]corev1.VolumeMount, 0)

	volumes = append(volumes, corev1.Volume{
		Name: name.OutputVolumeName(),
		VolumeSource: corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: name.OutputPVCName(room.Spec.Director.ID),
			},
		},
	})

	volumeMounts = append(volumeMounts, corev1.VolumeMount{
		Name:      name.OutputVolumeName(),
		MountPath: name.OutputVolumeMountDir(),
	})

	if room.Spec.GameConfig.DataPVCName != "" {
		volumes = append(volumes, corev1.Volume{
			Name: name.DataVolumeName(),
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: room.Spec.GameConfig.DataPVCName,
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

	if room.Spec.GameConfig.FactPVCName != "" {
		volumes = append(volumes, corev1.Volume{
			Name: name.FactVolumeName(),
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: room.Spec.GameConfig.FactPVCName,
					ReadOnly:  true,
				},
			},
		})
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      name.FactVolumeName(),
			MountPath: name.FactVolumeMountDir(),
			ReadOnly:  true,
		})
	}

	labels := map[string]string{
		name.DirectorIDLabel(): room.Spec.Director.ID,
		name.RoomIDLabel():     room.Spec.ID,
		name.PodTypeLabel():    name.PodTypeDirector(),
	}

	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name.DirectorPodName(room.Spec.Director.ID),
			Namespace: room.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Volumes:       volumes,
			RestartPolicy: corev1.RestartPolicyNever,
			Containers: []corev1.Container{
				{
					Name:            name.DirectorContainerName(),
					Image:           room.Spec.Director.Image,
					ImagePullPolicy: corev1.PullIfNotPresent,
					VolumeMounts:    volumeMounts,
					Resources:       corev1.ResourceRequirements{},
					Env:             []corev1.EnvVar{},
				},
			},
		},
	}
}
