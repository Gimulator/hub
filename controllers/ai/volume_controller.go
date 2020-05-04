package ai

import (
	"github.com/go-logr/logr"
	aiv1 "gitlab.com/Syfract/Xerac/hub/apis/ai/v1"
	env "gitlab.com/Syfract/Xerac/hub/utils/environment"
)

type VolumeReconciler struct {
	log logr.Logger
}

func NewVolumeReconciler(log logr.Logger) (*VolumeReconciler, error) {
	return &VolumeReconciler{
		log: log,
	}, nil
}

func (r *VolumeReconciler) Reconcile(src, dst *aiv1.Room) error {
	if dst.Spec.Volumes == nil {
		dst.Spec.Volumes = make([]aiv1.Volume, 0)
	}

	if err := r.reconcileSharedVolumes(src, dst); err != nil {
		return err
	}

	if err := r.reconcileGimulatorVolume(src, dst); err != nil {
		return err
	}

	if err := r.reconcileLoggerVolume(src, dst); err != nil {
		return err
	}

	return nil
}

func (r *VolumeReconciler) reconcileSharedVolumes(src, dst *aiv1.Room) error {
	sharedVolume := aiv1.Volume{
		EmptyDirVolume: &aiv1.EmptyDirVolume{
			Name: env.SharedVolumeName(),
		},
	}
	dst.Spec.Volumes = append(dst.Spec.Volumes, sharedVolume)

	return nil
}

func (r *VolumeReconciler) reconcileGimulatorVolume(src, dst *aiv1.Room) error {
	gimulatorVolume := aiv1.Volume{
		EmptyDirVolume: &aiv1.EmptyDirVolume{
			Name: env.GimulatorConfigVolumeName(),
		},
	}
	dst.Spec.Volumes = append(dst.Spec.Volumes, gimulatorVolume)

	return nil
}

func (r *VolumeReconciler) reconcileLoggerVolume(src, dst *aiv1.Room) error {
	loggerVolume := aiv1.Volume{
		EmptyDirVolume: &aiv1.EmptyDirVolume{
			Name: env.LoggerLogDirName(),
		},
	}
	dst.Spec.Volumes = append(dst.Spec.Volumes, loggerVolume)

	return nil
}
