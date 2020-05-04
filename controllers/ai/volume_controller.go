package ai

import (
	"fmt"

	"github.com/go-logr/logr"
	aiv1 "gitlab.com/Syfract/Xerac/hub/apis/ai/v1"
	env "gitlab.com/Syfract/Xerac/hub/utils/environment"
	batch "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
)

type VolumeReconciler struct {
	Log logr.Logger
}

func NewVolumeReconciler(log logr.Logger) (*VolumeReconciler, error) {
	return &VolumeReconciler{
		Log: log,
	}, nil
}

func (r *VolumeReconciler) Reconcile(room aiv1.Room, job *batch.Job) error {
	if job.Spec.Template.Spec.Volumes == nil {
		job.Spec.Template.Spec.Volumes = make([]core.Volume, 0)
	}

	return nil
}

func (r *VolumeReconciler) ReconcileSharedVolumes(room aiv1.Room, job *batch.Job) error {
	emptyDirVolume := &aiv1.EmptyDirVolume{
		Name: env.SharedVolumeName(),
	}

	err := r.ReconcileEmptyDirAskedVolume(emptyDirVolume, job)
	if err != nil {
		return err
	}

	return nil
}

func (r *VolumeReconciler) ReconcileAskedVolumes(room aiv1.Room, job *batch.Job) error {
	for _, volume := range room.Spec.Volumes {
		err := r.ReconcileAskedVolume(volume, job)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *VolumeReconciler) ReconcileAskedVolume(volume aiv1.Volume, job *batch.Job) error {
	if volume.ConfigMapVolumes != nil && volume.EmptyDirVolume != nil {
		return fmt.Errorf("invalid volume")
	}

	switch {
	case volume.EmptyDirVolume != nil:
		err := r.ReconcileEmptyDirAskedVolume(volume.EmptyDirVolume, job)
		if err != nil {
			return err
		}
	case volume.ConfigMapVolumes != nil:
		err := r.ReconcileConfigMapAskedVolume(volume.ConfigMapVolumes, job)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("there is no specific volume type")
	}

	return nil
}

func (r *VolumeReconciler) ReconcileEmptyDirAskedVolume(emptyDir *aiv1.EmptyDirVolume, job *batch.Job) error {
	job.Spec.Template.Spec.Volumes = append(job.Spec.Template.Spec.Volumes, core.Volume{
		Name: emptyDir.Name,
		VolumeSource: core.VolumeSource{
			EmptyDir: &core.EmptyDirVolumeSource{},
		},
	})
	return nil
}

func (r *VolumeReconciler) ReconcileConfigMapAskedVolume(configMap *aiv1.ConfigMapVolume, job *batch.Job) error {
	job.Spec.Template.Spec.Volumes = append(job.Spec.Template.Spec.Volumes, core.Volume{
		Name: configMap.Name,
		VolumeSource: core.VolumeSource{
			ConfigMap: &core.ConfigMapVolumeSource{
				LocalObjectReference: core.LocalObjectReference{
					Name: configMap.ConfigMapName,
				},
				Items: []core.KeyToPath{
					{
						Key: "data",
					},
				},
			},
		},
	})
	return nil
}
