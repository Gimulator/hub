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

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(env.APICallTimeout))
	defer cancel()

	syncedConfigMap := &core.ConfigMap{}
	if err := d.Get(ctx, types.NamespacedName{Name: configMap.Name, Namespace: configMap.Namespace}, syncedConfigMap); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
		syncedConfigMap = configMap.DeepCopy()
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(env.APICallTimeout))
		defer cancel()
		return d.Create(ctx, syncedConfigMap)
	}

	if !reflect.DeepEqual(configMap.Data, syncedConfigMap.Data) {
		syncedConfigMap = configMap.DeepCopy()
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(env.APICallTimeout))
		defer cancel()
		return d.Update(ctx, syncedConfigMap)
	}
	return nil
}

func (d *Deployer) SyncJob(job *batch.Job) (*batch.Job, error) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(env.APICallTimeout))
	defer cancel()

	syncedJob := &batch.Job{}
	if err := d.Get(ctx, types.NamespacedName{Name: job.Name, Namespace: job.Namespace}, syncedJob); err != nil {
		if !errors.IsNotFound(err) {
			return syncedJob, err
		}
		syncedJob = job.DeepCopy()

		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(env.APICallTimeout))
		defer cancel()

		err = d.Create(ctx, syncedJob)
		return syncedJob, err
	}
	return syncedJob, nil
}
