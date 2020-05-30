package deployer

import (
	"context"
	"reflect"
	"sync"
	"time"

	aiv1 "github.com/Gimulator/hub/apis/ai/v1"
	env "github.com/Gimulator/hub/utils/environment"
	batch "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Deployer implements helper methods to sync resources
type Deployer struct {
	client.Client

	configMapMutex sync.Mutex
	jobMutex       sync.Mutex
	roomMutex      sync.Mutex
}

// NewDeployer creates a new Deployer
func NewDeployer(c client.Client) *Deployer {
	return &Deployer{
		Client: c,

		configMapMutex: sync.Mutex{},
		jobMutex:       sync.Mutex{},
		roomMutex:      sync.Mutex{},
	}
}

// ********************************************************** sync ConfigMap

func (d *Deployer) SyncConfigMap(configMap *core.ConfigMap) (*core.ConfigMap, error) {
	d.configMapMutex.Lock()
	defer d.configMapMutex.Unlock()

	_, err := d.getConfigMap(configMap)

	if errors.IsNotFound(err) {
		return d.createConfigMap(configMap)
	}

	if err == nil {
		return d.updateConfigMap(configMap)
	}

	return nil, err
}

func (d *Deployer) getConfigMap(configMap *core.ConfigMap) (*core.ConfigMap, error) {
	oldConfigMap := &core.ConfigMap{}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(env.APICallTimeout))
	defer cancel()

	err := d.Get(ctx, types.NamespacedName{Name: configMap.Name, Namespace: configMap.Namespace}, oldConfigMap)
	return oldConfigMap, err
}

func (d *Deployer) createConfigMap(configMap *core.ConfigMap) (*core.ConfigMap, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(env.APICallTimeout))
	defer cancel()

	err := d.Create(ctx, configMap)
	return configMap, err
}

func (d *Deployer) updateConfigMap(configMap *core.ConfigMap) (*core.ConfigMap, error) {
	syncedConfigMap := configMap.DeepCopy()

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		var err error
		syncedConfigMap, err = d.getConfigMap(configMap)
		if err != nil {
			return err
		}

		if reflect.DeepEqual(syncedConfigMap.Data, configMap.Data) {
			return nil
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(env.APICallTimeout))
		defer cancel()

		syncedConfigMap.Data = configMap.Data
		err = d.Update(ctx, syncedConfigMap)
		return err
	})

	return syncedConfigMap, retryErr
}

// ********************************************************** sync Job

func (d *Deployer) SyncJob(job *batch.Job) (*batch.Job, error) {
	d.jobMutex.Lock()
	defer d.jobMutex.Unlock()

	oldJob, err := d.getJob(job)

	if errors.IsNotFound(err) {
		return d.createJob(job)
	}

	return oldJob, err
}

func (d *Deployer) getJob(job *batch.Job) (*batch.Job, error) {
	oldJob := &batch.Job{}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(env.APICallTimeout))
	defer cancel()

	err := d.Get(ctx, types.NamespacedName{Name: job.Name, Namespace: job.Namespace}, oldJob)
	return oldJob, err
}

func (d *Deployer) createJob(job *batch.Job) (*batch.Job, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(env.APICallTimeout))
	defer cancel()

	err := d.Create(ctx, job)
	return job, err
}

// ********************************************************** sync Room

func (d *Deployer) SyncRoom(room *aiv1.Room) (*aiv1.Room, error) {
	d.roomMutex.Lock()
	defer d.roomMutex.Unlock()

	_, err := d.getRoom(room)

	if errors.IsNotFound(err) {
		return d.createRoom(room)
	}

	if err == nil {
		return d.updateRoom(room)
	}

	return nil, err
}

func (d *Deployer) getRoom(room *aiv1.Room) (*aiv1.Room, error) {
	oldRoom := &aiv1.Room{}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(env.APICallTimeout))
	defer cancel()

	err := d.Get(ctx, types.NamespacedName{Name: room.Name, Namespace: room.Namespace}, oldRoom)
	return oldRoom, err
}

func (d *Deployer) createRoom(room *aiv1.Room) (*aiv1.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(env.APICallTimeout))
	defer cancel()

	err := d.Create(ctx, room)
	return room, err
}

func (d *Deployer) updateRoom(room *aiv1.Room) (*aiv1.Room, error) {
	syncedRoom := room.DeepCopy()

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		var err error
		syncedRoom, err = d.getRoom(room)
		if err != nil {
			return err
		}

		if reflect.DeepEqual(syncedRoom.Status, room.Status) {
			return nil
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(env.APICallTimeout))
		defer cancel()

		syncedRoom.Status = room.Status
		err = d.Update(ctx, syncedRoom)
		return err
	})

	return syncedRoom, retryErr
}
