package deployer

import (
	"context"
	"reflect"
	"time"

	aiv1 "github.com/Gimulator/hub/apis/ai/v1"
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

func (d *Deployer) SyncConfigMap(configMap *core.ConfigMap) (*core.ConfigMap, error) {
	syncedConfigMap := &core.ConfigMap{}
	getCtx, getCancel := context.WithTimeout(context.Background(), time.Duration(env.APICallTimeout))
	defer getCancel()

	err := d.Get(getCtx, types.NamespacedName{Name: configMap.Name, Namespace: configMap.Namespace}, syncedConfigMap)
	switch {
	case err != nil && !errors.IsNotFound(err):
		return syncedConfigMap, err
	case errors.IsNotFound(err):
		syncedConfigMap = configMap.DeepCopy()
		createCtx, createCancel := context.WithTimeout(context.Background(), time.Duration(env.APICallTimeout))
		defer createCancel()

		err = d.Create(createCtx, syncedConfigMap)
		if !errors.IsAlreadyExists(err) {
			return syncedConfigMap, err
		}
		return syncedConfigMap, nil
	case !reflect.DeepEqual(configMap.Data, syncedConfigMap.Data):
		syncedConfigMap = configMap.DeepCopy()
		updateCtx, updateCancel := context.WithTimeout(context.Background(), time.Duration(env.APICallTimeout))
		defer updateCancel()

		err = d.Update(updateCtx, syncedConfigMap)
		return syncedConfigMap, err
	default:
		return syncedConfigMap, nil
	}
}

func (d *Deployer) SyncJob(job *batch.Job) (*batch.Job, error) {
	syncedJob := &batch.Job{}
	getCtx, getCancel := context.WithTimeout(context.Background(), time.Duration(env.APICallTimeout))
	defer getCancel()

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

func (d *Deployer) SyncRoom(room *aiv1.Room) (*aiv1.Room, error) {
	syncedRoom := &aiv1.Room{}
	getCtx, getCancel := context.WithTimeout(context.Background(), time.Duration(env.APICallTimeout))
	defer getCancel()

	err := d.Get(getCtx, types.NamespacedName{Name: room.Name, Namespace: room.Namespace}, syncedRoom)
	switch {
	case err != nil:
		return syncedRoom, err
	case !reflect.DeepEqual(room.Status, syncedRoom.Status):
		syncedRoom.Status = *room.Status.DeepCopy()
		updateCtx, updateCancel := context.WithTimeout(context.Background(), time.Duration(env.APICallTimeout))
		defer updateCancel()

		err = d.Update(updateCtx, syncedRoom)
		return syncedRoom, err
	default:
		return syncedRoom, nil
	}
}
