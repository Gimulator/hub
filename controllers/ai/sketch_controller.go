package ai

import (
	"time"

	"github.com/Gimulator/Gimulator/auth"
	"github.com/go-logr/logr"
	cache "github.com/patrickmn/go-cache"
	uuid "github.com/satori/go.uuid"
	aiv1 "gitlab.com/Syfract/Xerac/hub/apis/ai/v1"
	"gitlab.com/Syfract/Xerac/hub/utils/deployer"
	env "gitlab.com/Syfract/Xerac/hub/utils/environment"
	"gitlab.com/Syfract/Xerac/hub/utils/name"
	"gitlab.com/Syfract/Xerac/hub/utils/storage"
	"gopkg.in/yaml.v3"
	batch "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type SketchReconciler struct {
	client.Client
	log      logr.Logger
	scheme   *runtime.Scheme
	deployer *deployer.Deployer
	storage  storage.Storage
	cache    *cache.Cache
	config   auth.Config
}

func NewSketchReconciler(mgr manager.Manager, log logr.Logger) (*SketchReconciler, error) {
	storage, err := storage.NewMock()
	if err != nil {
		return nil, err
	}

	return &SketchReconciler{
		Client:   mgr.GetClient(),
		deployer: deployer.NewDeployer(mgr.GetClient()),
		scheme:   mgr.GetScheme(),
		log:      log,
		storage:  storage,
		cache:    cache.New(time.Minute*10, time.Minute*20),
	}, nil
}

func (r *SketchReconciler) Reconcile(room aiv1.Room, job *batch.Job) error {
	err := r.reconcileData(room.Spec.Sketch.Bucket, room.Spec.Sketch.Key, &r.config)
	if err != nil {
		return err
	}

	err = r.reconcileSketch(room, job)
	if err != nil {
		return err
	}

	err = r.reconcileConfigMap(room)
	if err != nil {
		return err
	}

	return nil
}

func (r *SketchReconciler) reconcileSketch(room aiv1.Room, job *batch.Job) error {
	for _, actor := range room.Spec.Actors {
		err := r.reconcileConfigForActor(actor, job)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *SketchReconciler) reconcileConfigForActor(actor aiv1.Actor, job *batch.Job) error {
	role := actor.Role
	username := name.NameDashID(actor.Name, actor.ID)
	password := uuid.NewV4().String()

	authActor := auth.Actor{
		Username: username,
		Role:     role,
		Password: password,
	}
	r.config.Actors = append(r.config.Actors, authActor)

	for i := range job.Spec.Template.Spec.Containers {
		con := &job.Spec.Template.Spec.Containers[i]
		if con.Name != username {
			continue
		}
		con.Env = append(con.Env, core.EnvVar{
			Name:  env.UsernameEnvVarKey,
			Value: username,
		})
		con.Env = append(con.Env, core.EnvVar{
			Name:  env.PasswordEnvVarKey,
			Value: password,
		})

		break
	}

	return nil
}

func (r *SketchReconciler) reconcileData(bucket, key string, config *auth.Config) error {
	name := name.BucketDashKey(bucket, key)
	var (
		b  []byte
		ok bool = false
	)

	data, exists := r.cache.Get(name)
	if exists {
		b, ok = data.([]byte)
	}

	if !ok {
		str, err := r.storage.GetConfigYamlToString(bucket, key)
		if err != nil {
			return err
		}
		r.cache.Set(name, str, cache.DefaultExpiration)
		b = []byte(str)
	}

	return yaml.Unmarshal(b, config)
}

func (r *SketchReconciler) reconcileConfigMap(room aiv1.Room) error {
	data, err := yaml.Marshal(r.config)
	if err != nil {
		return err
	}

	coreConfigMap := &core.ConfigMap{
		ObjectMeta: meta.ObjectMeta{
			Name:      name.GimulatorConfigMap(room.Spec.ID),
			Namespace: env.Namespace(),
		},
		Data: map[string]string{
			"data": string(data),
		},
	}

	return r.deployer.SyncConfigMap(coreConfigMap)
}
