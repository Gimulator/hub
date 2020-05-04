package convertor

import (
	"fmt"

	aiv1 "gitlab.com/Syfract/Xerac/hub/apis/ai/v1"
	env "gitlab.com/Syfract/Xerac/hub/utils/environment"

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
	cpu := resource.Format(env.ResourceRequestsCPU())
	if src.CPU != "" {
		cpu = resource.Format(src.CPU)
	}

	memory := resource.Format(env.ResourceRequestsMemory())
	if src.Memory != "" {
		memory = resource.Format(src.Memory)
	}

	ephemeral := resource.Format(env.ResourceRequestsEphemeral())
	if src.Ephemeral != "" {
		memory = resource.Format(src.Ephemeral)
	}

	return core.ResourceList{
		core.ResourceCPU:              resource.Quantity{Format: cpu},
		core.ResourceMemory:           resource.Quantity{Format: memory},
		core.ResourceEphemeralStorage: resource.Quantity{Format: ephemeral},
	}, nil
}

func ConvertResourceLimits(src aiv1.Resource) (core.ResourceList, error) {
	cpu := resource.Format(env.ResourceLimitsCPU())
	if src.CPU != "" {
		cpu = resource.Format(src.CPU)
	}

	memory := resource.Format(env.ResourceLimitsMemory())
	if src.Memory != "" {
		memory = resource.Format(src.Memory)
	}

	ephemeral := resource.Format(env.ResourceLimitsEphemeral())
	if src.Ephemeral != "" {
		memory = resource.Format(src.Ephemeral)
	}

	return core.ResourceList{
		core.ResourceCPU:              resource.Quantity{Format: cpu},
		core.ResourceMemory:           resource.Quantity{Format: memory},
		core.ResourceEphemeralStorage: resource.Quantity{Format: ephemeral},
	}, nil
}

func ConvertEmptyDirVolume(src aiv1.EmptyDirVolume) (*core.Volume, error) {
	if src.Name == "" {
		return nil, fmt.Errorf("convertor: empty Name for EmptyDirVolume")
	}

	return &core.Volume{
		Name: src.Name,
		VolumeSource: core.VolumeSource{
			EmptyDir: &core.EmptyDirVolumeSource{},
		},
	}, nil
}

func ConvertConfigMapVolume(src aiv1.ConfigMapVolume) (*core.Volume, error) {
	if src.Name == "" {
		return nil, fmt.Errorf("convertor: empty Name for ConfigMapVolume")
	}

	if src.ConfigMapName == "" {
		return nil, fmt.Errorf("convertor: empty ConfigMapName for ConfigMapVolume")
	}

	return &core.Volume{
		Name: src.Name,
		VolumeSource: core.VolumeSource{
			ConfigMap: &core.ConfigMapVolumeSource{
				LocalObjectReference: core.LocalObjectReference{
					Name: src.ConfigMapName,
				},
				Items: []core.KeyToPath{
					{Key: "data"}, //TODO: Fix hardcode
				},
			},
		},
	}, nil
}

func ConvertConfigMap(src aiv1.ConfigMap, data string) (core.ConfigMap, error) {
	if src.Name == "" {
		return core.ConfigMap{}, fmt.Errorf("convertor: empty Name for ConfigMap")
	}

	return core.ConfigMap{
		ObjectMeta: meta.ObjectMeta{
			Name:      src.Name,
			Namespace: env.Namespace(),
		},
		Data: map[string]string{
			"data": data,
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
		Name:         src.Name,
		Image:        src.Image,
		VolumeMounts: volumeMounts,
		Env:          envVars,
		Resources:    resources,
		Args:         src.Args,
		Command:      []string{"/bin/bash", "-c"},
	}, nil
}
