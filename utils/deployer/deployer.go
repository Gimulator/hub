package deployer

import (
	"context"
	"sync"
	"time"

	aiv1 "github.com/Gimulator/hub/apis/ai/v1"
	env "github.com/Gimulator/hub/utils/environment"
	batch "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// Deployer implements helper methods to sync resources
type Deployer struct {
	client.Client

	Scheme         *runtime.Scheme
	configMapMutex sync.Mutex
	jobMutex       sync.Mutex
	roomMutex      sync.Mutex
}

// NewDeployer creates a new Deployer
func NewDeployer(c client.Client, scheme *runtime.Scheme) *Deployer {
	return &Deployer{
		Client: c,

		Scheme:         scheme,
		configMapMutex: sync.Mutex{},
		jobMutex:       sync.Mutex{},
		roomMutex:      sync.Mutex{},
	}
}

// ********************************************************** sync ConfigMap

func (d *Deployer) SyncConfigMap(room *aiv1.Room, configMap *core.ConfigMap) (*core.ConfigMap, error) {
	d.configMapMutex.Lock()
	defer d.configMapMutex.Unlock()

	syncedConfigMap := &core.ConfigMap{
		ObjectMeta: meta.ObjectMeta{
			Name:      configMap.Name,
			Namespace: configMap.Namespace,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(env.APICallTimeout))
	defer cancel()

	if _, err := controllerutil.CreateOrUpdate(ctx, d.Client, syncedConfigMap, func() error {
		syncedConfigMap.Data = configMap.Data
		controllerutil.SetOwnerReference(room, syncedConfigMap, d.Scheme)
		return nil
	}); err != nil {
		return nil, err
	}

	return syncedConfigMap, nil
}

// ********************************************************** sync Job

func (d *Deployer) SyncJob(room *aiv1.Room, job *batch.Job) (*batch.Job, error) {
	d.jobMutex.Lock()
	defer d.jobMutex.Unlock()

	syncedJob, err := d.GetJob(types.NamespacedName{Name: job.Name, Namespace: job.Namespace})
	if err == nil {
		return syncedJob, nil
	}
	if !errors.IsNotFound(err) {
		return nil, err
	}

	if err := controllerutil.SetOwnerReference(room, job, d.Scheme); err != nil {
		return nil, err
	}

	return job, d.CreateJob(job)
}

func (d *Deployer) GetJob(nn types.NamespacedName) (*batch.Job, error) {
	job := &batch.Job{}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(env.APICallTimeout))
	defer cancel()

	err := d.Get(ctx, nn, job)
	return job, err
}

func (d *Deployer) CreateJob(job *batch.Job) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(env.APICallTimeout))
	defer cancel()

	return d.Create(ctx, job)
}

// ********************************************************** sync Room

func (d *Deployer) SyncRoom(room *aiv1.Room) (*aiv1.Room, error) {
	d.roomMutex.Lock()
	defer d.roomMutex.Unlock()

	syncedRoom := &aiv1.Room{
		ObjectMeta: meta.ObjectMeta{
			Name:      room.Name,
			Namespace: room.Namespace,
		},
	}

	retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(env.APICallTimeout))
		defer cancel()

		if _, err := controllerutil.CreateOrUpdate(ctx, d.Client, syncedRoom, func() error {
			syncedRoom.Status = *room.Status.DeepCopy()
			return nil
		}); err != nil {
			return err
		}
		return nil
	})
	return syncedRoom, nil
}

func (d *Deployer) GetRoom(nn types.NamespacedName) (*aiv1.Room, error) {
	room := &aiv1.Room{}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(env.APICallTimeout))
	defer cancel()

	err := d.Get(ctx, nn, room)
	return room, err
}

func (d *Deployer) DeleteRoom(room *aiv1.Room) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(env.APICallTimeout))
	defer cancel()

	if err := d.Delete(ctx, room); !errors.IsNotFound(err) {
		return err
	}
	return nil
}
