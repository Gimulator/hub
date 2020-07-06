package convertor

import (
	"fmt"

	aiv1 "github.com/Gimulator/hub/apis/ai/v1"
	env "github.com/Gimulator/hub/utils/environment"
	"github.com/Gimulator/hub/utils/name"

	batch "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	resource "k8s.io/apimachinery/pkg/api/resource"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ConvertVolumeMount(src aiv1.VolumeMount) (core.VolumeMount, error) {
	if src.Name == "" {
		return core.VolumeMount{}, fmt.Errorf("convertor: empty Name for VolumeMount")
	}

	if src.Path == "" {
		return core.VolumeMount{}, fmt.Errorf("convertor: empty Path for VolumeMount")
	}

	return core.VolumeMount{
		Name:      src.Name,
		MountPath: src.Path,
	}, nil
}

func ConvertEnvVar(src aiv1.EnvVar) (core.EnvVar, error) {
	if src.Key == "" {
		return core.EnvVar{}, fmt.Errorf("convertor: empty Key for EnvVar")
	}

	return core.EnvVar{
		Name:  src.Key,
		Value: src.Value,
	}, nil
}

func ConvertResources(src aiv1.Resources) (core.ResourceRequirements, error) {
	requests, err := ConvertResourceRequests(src.Requests)
	if err != nil {
		return core.ResourceRequirements{}, err
	}

	limits, err := ConvertResourceLimits(src.Limits)
	if err != nil {
		return core.ResourceRequirements{}, err
	}

	return core.ResourceRequirements{
		Requests: requests,
		Limits:   limits,
	}, nil
}

func ConvertResourceRequests(src aiv1.Resource) (core.ResourceList, error) {
	cpu := resource.MustParse(env.DefaultRequestsCPU())
	if src.CPU != "" {
		cpu = resource.MustParse(src.CPU)
	}

	memory := resource.MustParse(env.DefaultRequestsMemory())
	if src.Memory != "" {
		memory = resource.MustParse(src.Memory)
	}

	ephemeral := resource.MustParse(env.DefaultRequestsEphemeral())
	if src.Ephemeral != "" {
		memory = resource.MustParse(src.Ephemeral)
	}

	return core.ResourceList{
		core.ResourceCPU:              cpu,
		core.ResourceMemory:           memory,
		core.ResourceEphemeralStorage: ephemeral,
	}, nil
}

func ConvertResourceLimits(src aiv1.Resource) (core.ResourceList, error) {
	cpu := resource.MustParse(env.DefaultLimitsCPU())
	if src.CPU != "" {
		cpu = resource.MustParse(src.CPU)
	}

	memory := resource.MustParse(env.DefaultLimitsMemory())
	if src.Memory != "" {
		memory = resource.MustParse(src.Memory)
	}

	ephemeral := resource.MustParse(env.DefaultLimitsEphemeral())
	if src.Ephemeral != "" {
		memory = resource.MustParse(src.Ephemeral)
	}

	return core.ResourceList{
		core.ResourceCPU:              cpu,
		core.ResourceMemory:           memory,
		core.ResourceEphemeralStorage: ephemeral,
	}, nil
}

func ConvertVolume(src aiv1.Volume) (core.Volume, error) {
	if src.ConfigMapVolumes != nil && src.EmptyDirVolume != nil {
		return core.Volume{}, fmt.Errorf("convertor: EmptyDir and ConfigMap both are not nil")
	}

	if src.ConfigMapVolumes != nil {
		return ConvertConfigMapVolume(src.ConfigMapVolumes)
	}

	if src.EmptyDirVolume != nil {
		return ConvertEmptyDirVolume(src.EmptyDirVolume)
	}

	return core.Volume{}, fmt.Errorf("convertor: EmptyDir and ConfigMap both are nil")
}

func ConvertEmptyDirVolume(src *aiv1.EmptyDirVolume) (core.Volume, error) {
	if src.Name == "" {
		return core.Volume{}, fmt.Errorf("convertor: empty Name for EmptyDirVolume")
	}

	return core.Volume{
		Name: src.Name,
		VolumeSource: core.VolumeSource{
			EmptyDir: &core.EmptyDirVolumeSource{},
		},
	}, nil
}

func ConvertConfigMapVolume(src *aiv1.ConfigMapVolume) (core.Volume, error) {
	if src.Name == "" {
		return core.Volume{}, fmt.Errorf("convertor: empty Name for ConfigMapVolume")
	}

	if src.ConfigMapName == "" {
		return core.Volume{}, fmt.Errorf("convertor: empty ConfigMapName for ConfigMapVolume")
	}

	return core.Volume{
		Name: src.Name,
		VolumeSource: core.VolumeSource{
			ConfigMap: &core.ConfigMapVolumeSource{
				LocalObjectReference: core.LocalObjectReference{
					Name: src.ConfigMapName,
				},
				Items: []core.KeyToPath{
					{
						Key:  env.ConfigMapItemKey(),
						Path: src.Path,
					},
				},
			},
		},
	}, nil
}

func ConvertActor(src aiv1.Actor) (core.Container, error) {
	dst := core.Container{}

	volumeMounts := make([]core.VolumeMount, 0)
	for _, vm := range src.VolumeMounts {
		volumeMount, err := ConvertVolumeMount(vm)
		if err != nil {
			return dst, err
		}
		volumeMounts = append(volumeMounts, volumeMount)
	}

	envVars := make([]core.EnvVar, 0)
	for _, ev := range src.EnvVars {
		envVar, err := ConvertEnvVar(ev)
		if err != nil {
			return dst, err
		}
		envVars = append(envVars, envVar)
	}

	resources, err := ConvertResources(src.Resources)
	if err != nil {
		return dst, err
	}

	return core.Container{
		Name:            name.ContainerName(src.Name, src.ID),
		Image:           src.Image,
		VolumeMounts:    volumeMounts,
		Env:             envVars,
		Resources:       resources,
		Args:            src.Args,
		Command:         []string{"/bin/sh", "-c"},
		ImagePullPolicy: core.PullIfNotPresent,
	}, nil
}

func ConvertConfigMap(src aiv1.ConfigMap) (*core.ConfigMap, error) {
	if src.Name == "" {
		return &core.ConfigMap{}, fmt.Errorf("convertor: empty Name for ConfigMap")
	}

	return &core.ConfigMap{
		ObjectMeta: meta.ObjectMeta{
			Name:      src.Name,
			Namespace: env.RoomNamespace(),
		},
		Data: map[string]string{
			env.ConfigMapItemKey(): src.Data,
		},
	}, nil
}

func ConvertRoom(src *aiv1.Room) (*batch.Job, error) {
	dst := &batch.Job{
		ObjectMeta: meta.ObjectMeta{
			Name:      name.RoomJobName(src.Spec.ID),
			Namespace: env.RoomNamespace(),
		},
		Spec: batch.JobSpec{
			BackoffLimit:          &src.Spec.BackoffLimit,
			ActiveDeadlineSeconds: &src.Spec.ActiveDeadLineSeconds,
			Template: core.PodTemplateSpec{
				Spec: core.PodSpec{
					Volumes:       make([]core.Volume, 0),
					Containers:    make([]core.Container, 0),
					RestartPolicy: "OnFailure",
				},
			},
		},
	}

	for _, actor := range src.Spec.Actors {
		container, err := ConvertActor(actor)
		if err != nil {
			return &batch.Job{}, err
		}
		dst.Spec.Template.Spec.Containers = append(dst.Spec.Template.Spec.Containers, container)
	}

	for _, v := range src.Spec.Volumes {
		volume, err := ConvertVolume(v)
		if err != nil {
			return &batch.Job{}, err
		}
		dst.Spec.Template.Spec.Volumes = append(dst.Spec.Template.Spec.Volumes, volume)
	}

	return dst, nil
}
