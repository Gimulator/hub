package deployer

import (
	"context"
	"reflect"
	"time"

	env "github.com/Gimulator/hub/utils/environment"
	batch "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Deployer implements helper methods to sync resources
type Deployer struct {
	client.Client
}

// NewDeployer creates a new Deployer
func NewDeployer(c client.Client) *Deployer {
	return &Deployer{c}
}

func (d *Deployer) SyncConfigMap(configMap *core.ConfigMap) error {
	getCtx, getCancel := context.WithTimeout(context.Background(), time.Duration(env.APICallTimeout))
	defer getCancel()

	syncedConfigMap := &core.ConfigMap{}
	err := d.Get(getCtx, types.NamespacedName{Name: configMap.Name, Namespace: configMap.Namespace}, syncedConfigMap)
	switch {
	case err != nil && !errors.IsNotFound(err):
		return err
	case err != nil:
		syncedConfigMap = configMap.DeepCopy()
		createCtx, createCancel := context.WithTimeout(context.Background(), time.Duration(env.APICallTimeout))
		defer createCancel()

		err = d.Create(createCtx, syncedConfigMap)
		if !errors.IsAlreadyExists(err) {
			return err
		}
		return nil
	case !reflect.DeepEqual(configMap.Data, syncedConfigMap.Data):
		syncedConfigMap = configMap.DeepCopy()
		updateCtx, updateCancel := context.WithTimeout(context.Background(), time.Duration(env.APICallTimeout))
		defer updateCancel()

		return d.Update(updateCtx, syncedConfigMap)
	default:
		return nil
	}
}

func (d *Deployer) SyncJob(job *batch.Job) (*batch.Job, error) {

	getCtx, getCancel := context.WithTimeout(context.Background(), time.Duration(env.APICallTimeout))
	defer getCancel()

	syncedJob := &batch.Job{}
	err := d.Get(getCtx, types.NamespacedName{Name: job.Name, Namespace: job.Namespace}, syncedJob)
	if errors.IsNotFound(err) {
		syncedJob = job.DeepCopy()
		createCtx, createCancel := context.WithTimeout(context.Background(), time.Duration(env.APICallTimeout))
		defer createCancel()

		err = d.Create(createCtx, syncedJob)
		if errors.IsAlreadyExists(err) {
			return syncedJob, nil
		}
		return syncedJob, err
	}
	return syncedJob, err
}
