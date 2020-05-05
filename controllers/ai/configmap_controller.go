package ai

import (
	"github.com/go-logr/logr"
	aiv1 "gitlab.com/Syfract/Xerac/hub/apis/ai/v1"
	"gitlab.com/Syfract/Xerac/hub/utils/cache"
	"gitlab.com/Syfract/Xerac/hub/utils/name"
	"gitlab.com/Syfract/Xerac/hub/utils/storage"
)

type ConfigMapReconciler struct {
	log logr.Logger
}

func NewConfigMapReconciler(log logr.Logger) (*ConfigMapReconciler, error) {
	return &ConfigMapReconciler{
		log: log,
	}, nil
}

func (r *ConfigMapReconciler) Reconcile(src, dst *aiv1.Room) error {
	if dst.Spec.ConfigMaps == nil {
		dst.Spec.ConfigMaps = make([]aiv1.ConfigMap, 0)
	}

	err := r.reconcileConfigMaps(src, dst)
	if err != nil {
		return err
	}

	return nil
}

func (r *ConfigMapReconciler) reconcileConfigMaps(src, dst *aiv1.Room) error {
	for _, cm := range src.Spec.ConfigMaps {

		if cm.Data != "" {
			continue
		}

		name := name.ConfigMapName(cm.Bucket, cm.Name)
		data, err := cache.GetYamlString(name)
		if err != nil {
			data, err = storage.Get(cm.Bucket, cm.Key)
			if err != nil {
				return err
			}
			cache.SetYamlString(name, data)
		}

		dst.Spec.ConfigMaps = append(dst.Spec.ConfigMaps, aiv1.ConfigMap{
			Name:   cm.Name,
			Bucket: cm.Bucket,
			Key:    cm.Key,
			Data:   data,
		})
	}
	return nil
}
