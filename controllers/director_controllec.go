package controllers

import (
	"context"
	"strings"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
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
	logger := a.Log.WithValues("reconciler", "Director", "director", room.Spec.Director.Name, "room", room.Spec.ID)

	// logger.Info("starting to reconcile director's output PVC")
	// if err := a.reconcileOutputPVC(ctx, room); err != nil {
	// 	logger.Error(err, "could not reconcile director's output PVC")
	// 	return err
	// }

	logger.Info("starting to create director's manifest")
	dirPod, err := a.directorPodManifest(room)
	if err != nil {
		logger.Error(err, "could not create director's manifest")
		return err
	}

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

// func (a *directorReconciler) reconcileOutputPVC(ctx context.Context, room *hubv1.Room) error {
// 	quantity, err := resource.ParseQuantity(room.Spec.ProblemSettings.OutputVolumeSize)
// 	if err != nil {
// 		return err
// 	}

// 	pvc := &corev1.PersistentVolumeClaim{
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name:      name.OutputPVCName(room.Spec.Director.ID),
// 			Namespace: room.Namespace,
// 		},
// 		Spec: corev1.PersistentVolumeClaimSpec{
// 			Resources: corev1.ResourceRequirements{
// 				Requests: map[corev1.ResourceName]resource.Quantity{
// 					corev1.ResourceStorage: quantity,
// 				},
// 			},
// 			AccessModes: []corev1.PersistentVolumeAccessMode{
// 				corev1.ReadOnlyMany,
// 			},
// 		},
// 	}

// 	_, err = a.SyncPVC(ctx, pvc, room)
// 	return err
// }

func (a *directorReconciler) directorPodManifest(room *hubv1.Room) (*corev1.Pod, error) {
	volumes := make([]corev1.Volume, 0)
	volumeMounts := make([]corev1.VolumeMount, 0)

	// volumes = append(volumes, corev1.Volume{
	// 	Name: name.OutputVolumeName(room.Spec.Director.ID),
	// 	VolumeSource: corev1.VolumeSource{
	// 		PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
	// 			ClaimName: name.OutputPVCName(room.Spec.Director.ID),
	// 		},
	// 	},
	// })

	// volumeMounts = append(volumeMounts, corev1.VolumeMount{
	// 	Name:      name.OutputVolumeName(room.Spec.Director.ID),
	// 	MountPath: name.OutputVolumeMountPath(),
	// })

	if room.Spec.Setting.DataPVCNames != nil {
		// Mounting Private PVCs
		if room.Spec.Setting.DataPVCNames.Private != nil {
			for _, pvcName := range room.Spec.Setting.DataPVCNames.Private {
				fullName := strings.Join([]string{"private", pvcName}, "-")

				volumes = append(volumes, corev1.Volume{
					Name: name.DataVolumeName(fullName),
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: fullName,
							ReadOnly:  true,
						},
					},
				})
				volumeMounts = append(volumeMounts, corev1.VolumeMount{
					Name:      name.DataVolumeName(fullName),
					MountPath: name.DataVolumeMountPath(),
					ReadOnly:  true,
				})
			}
		}

		// Mounting Public PVCs
		// Comment/remove this part if you believe this functionality is unnecessary
		if room.Spec.Setting.DataPVCNames.Public != nil {
			for _, pvcName := range room.Spec.Setting.DataPVCNames.Public {
				fullName := strings.Join([]string{"public", pvcName}, "-")

				volumes = append(volumes, corev1.Volume{
					Name: name.DataVolumeName(fullName),
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: fullName,
							ReadOnly:  true,
						},
					},
				})
				volumeMounts = append(volumeMounts, corev1.VolumeMount{
					Name:      name.DataVolumeName(fullName),
					MountPath: name.DataVolumeMountPath(),
					ReadOnly:  true,
				})
			}
		}
	}

	// if room.Spec.ProblemSettings.FactPVCName != "" {
	// 	volumes = append(volumes, corev1.Volume{
	// 		Name: name.FactVolumeName(),
	// 		VolumeSource: corev1.VolumeSource{
	// 			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
	// 				ClaimName: room.Spec.ProblemSettings.FactPVCName,
	// 				ReadOnly:  true,
	// 			},
	// 		},
	// 	})
	// 	volumeMounts = append(volumeMounts, corev1.VolumeMount{
	// 		Name:      name.FactVolumeName(),
	// 		MountPath: name.FactVolumeMountPath(),
	// 		ReadOnly:  true,
	// 	})
	// }

	for _, actor := range room.Spec.Actors {
		volumes = append(volumes, corev1.Volume{
			Name: name.OutputVolumeName(actor.Name),
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: name.OutputPVCName(actor.Name),
					ReadOnly:  true,
				},
			},
		})
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      name.OutputVolumeName(actor.Name),
			MountPath: name.ActorOutputVolumeMountPathForDirector(actor.Name),
			ReadOnly:  true,
		})
	}

	labels := map[string]string{
		name.CharacterLabel(): name.CharacterDirector(),
		name.RoleLabel():      name.CharacterDirector(),
		name.RoomLabel():      room.Spec.ID,
		name.ProblemLabel():   room.Spec.ProblemID,
		name.IDLabel():        room.Spec.Director.Name,
	}

	envs := room.Spec.Director.Envs
	if envs == nil {
		envs = make([]corev1.EnvVar, 0)
	}
	envs = append(envs, corev1.EnvVar{
		Name:  "GIMULATOR_HOST",
		Value: name.GimulatorHost(room.Spec.ID),
	})
	envs = append(envs, corev1.EnvVar{
		Name:  "GIMULATOR_CHARACTER",
		Value: name.CharacterDirector(),
	})
	envs = append(envs, corev1.EnvVar{
		Name:  "GIMULATOR_ROLE",
		Value: name.CharacterDirector(),
	})
	envs = append(envs, corev1.EnvVar{
		Name:  "GIMULATOR_TOKEN",
		Value: room.Spec.Director.Token,
	})
	envs = append(envs, corev1.EnvVar{
		Name:  "GIMULATOR_NAME",
		Value: room.Spec.Director.Name,
	})
	envs = append(envs, corev1.EnvVar{
		Name:  "GIMULATOR_ROOM_ID",
		Value: room.Spec.ID,
	})

	// Priorities for resource allocations:
	// 1. room.Spec.Director.Resources
	// 2. room.Spec.Setting.DefaultResources

	resources := room.Spec.Setting.DefaultResources
	if room.Spec.Director.Resources != nil {
		resources = *room.Spec.Director.Resources
	}

	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name.DirectorPodName(room.Spec.Director.Name),
			Namespace: room.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Volumes:       volumes,
			RestartPolicy: corev1.RestartPolicyNever,
			ImagePullSecrets: []corev1.LocalObjectReference{
				{
					Name: "registry-credentials",
				},
			},
			Containers: []corev1.Container{
				{
					Name:            name.DirectorContainerName(),
					Image:           room.Spec.Director.Image,
					ImagePullPolicy: corev1.PullIfNotPresent,
					VolumeMounts:    volumeMounts,
					Env:             envs,
					Resources:       resources,
				},
			},
		},
	}, nil
}

func (a *directorReconciler) updateDirectorStatus(room *hubv1.Room, pod *corev1.Pod) {
	phase := pod.Status.DeepCopy().Phase

	room.Status.DirectorStatus = phase
}
