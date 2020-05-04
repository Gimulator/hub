package ai

import (
	"time"

	"github.com/go-logr/logr"
	cache "github.com/patrickmn/go-cache"
	aiv1 "gitlab.com/Syfract/Xerac/hub/apis/ai/v1"
	"gitlab.com/Syfract/Xerac/hub/utils/deployer"
	env "gitlab.com/Syfract/Xerac/hub/utils/environment"
	"gitlab.com/Syfract/Xerac/hub/utils/name"
	"gitlab.com/Syfract/Xerac/hub/utils/storage"
	batch "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type ConfigMapReconciler struct {
	client.Client
	log      logr.Logger
	scheme   *runtime.Scheme
	deployer *deployer.Deployer
	storage  storage.Storage
	cache    *cache.Cache
}

func NewConfigMapReconciler(mgr manager.Manager, log logr.Logger) (*ConfigMapReconciler, error) {
	storage, err := storage.NewMock()
	if err != nil {
		return nil, err
	}

	return &ConfigMapReconciler{
		Client:   mgr.GetClient(),
		deployer: deployer.NewDeployer(mgr.GetClient()),
		scheme:   mgr.GetScheme(),
		log:      log,
		storage:  storage,
		cache:    cache.New(time.Minute*10, time.Minute*20),
	}, nil
}

func (r *ConfigMapReconciler) Reconcile(room aiv1.Room, job *batch.Job) error {
	err := r.reconcileConfigMaps(room)
	if err != nil {
		return err
	}

	return nil
}

func (r *ConfigMapReconciler) reconcileConfigMaps(room aiv1.Room) error {
	for _, aiConfigMap := range room.Spec.ConfigMaps {
		err := r.reconcileConfigMap(aiConfigMap)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *ConfigMapReconciler) reconcileConfigMap(aiConfigMap aiv1.ConfigMap) error {
	data, err := r.reconcileData(aiConfigMap.Bucket, aiConfigMap.Key)
	if err != nil {
		return err
	}

	coreConfigMap := &core.ConfigMap{
		ObjectMeta: meta.ObjectMeta{
			Name:      aiConfigMap.Name,
			Namespace: env.Namespace(),
		},
		Data: map[string]string{
			"data": data,
		},
	}

	return r.deployer.SyncConfigMap(coreConfigMap)
}

func (r *ConfigMapReconciler) reconcileData(bucket, key string) (string, error) {
	name := name.BucketDashKey(bucket, key)

	data, exists := r.cache.Get(name)
	if exists {
		str, ok := data.(string)
		if ok {
			return str, nil
		}
	}

	str, err := r.storage.GetConfigYamlToString(bucket, key)
	if err != nil {
		return "", err
	}
	r.cache.Set(name, str, cache.DefaultExpiration)
	return str, nil
}

// func (r *ConfigMapReconciler) reconcileOwnership(room aiv1.Room, coreConfigMap *core.ConfigMap) error {
// 	err := controllerutil.SetControllerReference(&room, coreConfigMap, r.scheme)
// 	if err != nil {
// 		return err
// 	}

// 	ctx, cancel := context.WithTimeout(context.Background(), env.APICallTimeout)
// 	defer cancel()
// 	return r.Update(ctx, coreConfigMap)
// }
