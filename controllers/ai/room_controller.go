/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package ai

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/Gimulator/Gimulator/auth"
	"github.com/Gimulator/hub/utils/cache"
	"github.com/Gimulator/hub/utils/convertor"
	env "github.com/Gimulator/hub/utils/environment"
	"github.com/Gimulator/hub/utils/name"
	rabbit "github.com/Gimulator/hub/utils/rabbitMQ"
	"github.com/Gimulator/hub/utils/storage"
	"github.com/go-logr/logr"
	"gopkg.in/yaml.v2"
	batch "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/source"

	aiv1 "github.com/Gimulator/hub/apis/ai/v1"
	"github.com/Gimulator/hub/utils/deployer"
)

// RoomReconciler reconciles a Room object
type RoomReconciler struct {
	log      logr.Logger
	Scheme   *runtime.Scheme
	deployer *deployer.Deployer
	rabbit   *rabbit.Rabbit
}

func NewRoomReconciler(mgr manager.Manager, log logr.Logger) (*RoomReconciler, error) {
	rabbit, err := rabbit.NewRabbit()
	if err != nil {
		return nil, err
	}

	scheme := mgr.GetScheme()

	return &RoomReconciler{
		log:      log,
		Scheme:   scheme,
		deployer: deployer.NewDeployer(mgr.GetClient(), scheme),
		rabbit:   rabbit,
	}, nil
}

// +kubebuilder:rbac:groups=hub.xerac.cloud,resources=rooms,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=hub.xerac.cloud,resources=rooms/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch,resources=jobs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=configmaps/status,verbs=get;update;patch

func (r *RoomReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	log := r.log.WithValues("name", req.Name, "namespace", req.Namespace)
	log.Info("start to reconcile")

	src, err := r.deployer.GetRoom(req.NamespacedName)
	if errors.IsNotFound(err) {
		log.Info("room does not exist")
		return ctrl.Result{}, nil
	}
	if err != nil {
		log.Error(err, "unable to fetch Room")
		return ctrl.Result{}, err
	}

	instance := &aiv1.Room{}
	src.DeepCopyInto(instance)

	log.Info("start to reconcile actors")
	if err := r.reconcileActors(instance); err != nil {
		log.Error(err, "failed to reconcile actors")
		return ctrl.Result{}, err
	}

	log.Info("start to reconcile room's config maps")
	if err := r.reconcileConfigMaps(instance); err != nil {
		log.Error(err, "failed to reconcile room's config maps")
		return ctrl.Result{}, err
	}

	log.Info("start to reconcile sketch")
	if err := r.reconcileSketch(instance); err != nil {
		log.Error(err, "failed to reconcile sketch")
		return ctrl.Result{}, err
	}

	log.Info("start to reconcile volumes")
	if err := r.reconcileVolumes(instance); err != nil {
		log.Error(err, "failed to reconcile volumes")
		return ctrl.Result{}, err
	}

	log.Info("start to reconcile args")
	if err := r.reconcileArgs(instance); err != nil {
		log.Error(err, "failed to reconcile args")
		return ctrl.Result{}, err
	}

	log.Info("start to reconcile envvars")
	if err := r.reconcileEvnVars(instance); err != nil {
		log.Error(err, "failed to reconcile envvars")
		return ctrl.Result{}, err
	}

	log.Info("start to reconcile kube's config maps")
	for _, cm := range instance.Spec.ConfigMaps {
		configMap, err := convertor.ConvertConfigMap(cm)
		if err != nil {
			log.Error(err, "failed to reconcile kube's config maps")
			return ctrl.Result{}, err
		}

		if err := r.deployConfigMap(instance, configMap); err != nil {
			log.Error(err, "failed to reconcile kube's config maps")
			return ctrl.Result{}, err
		}
	}

	log.Info("start to reconcile job")
	job, err := convertor.ConvertRoom(instance)
	if err != nil {
		log.Error(err, "failed to reconcile job")
		return ctrl.Result{}, err
	}

	log.Info("start to deploy job")
	syncedJob, err := r.deployJob(instance, job)
	if err != nil {
		log.Error(err, "failed to deploy job")
		return ctrl.Result{}, err
	}

	log.Info("start to update room status")
	if err := r.updateRoomStatus(instance, syncedJob); err != nil {
		log.Error(err, "failed to update room status")
		return ctrl.Result{}, err
	}

	log.Info("start to reconcile room")
	if err := r.reconcileRoom(instance); err != nil {
		log.Error(err, "failed to reconcile room")
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

// ********************************* reconcile room *********************************

func (r *RoomReconciler) reconcileRoom(instance *aiv1.Room) error {
	syncedRoom, err := r.deployer.SyncRoom(instance)
	if err != nil {
		return err
	}

	switch syncedRoom.Status.RoomStatusType {
	case aiv1.RoomStatusTypeSuccess:
		return r.reconcileSuccessfulRoom(instance)
	case aiv1.RoomStatusTypeFailed:
		return r.reconcileFailedRoom(instance)
	case aiv1.RoomStatusTypeRunning:
	case aiv1.RoomStatusTypeUnknown:
		// TODO: What should I do?
	}

	return nil
}

func (r *RoomReconciler) reconcileSuccessfulRoom(instance *aiv1.Room) error {
	return r.deployer.DeleteRoom(instance)
}

func (r *RoomReconciler) reconcileFailedRoom(instance *aiv1.Room) error {
	result := struct {
		RoomID  int    `json:"run_id"`
		Status  string `json:"status"`
		Message string `json:"message"`
	}{
		RoomID:  instance.Spec.ID,
		Status:  "FAIL",
		Message: "could not find error",
	}

	if instance.Status.Job != nil {
		job, err := r.deployer.GetJob(types.NamespacedName{Name: instance.Status.Job.Name, Namespace: instance.Status.Job.Namespace})
		if err == nil && job.Status.Conditions != nil {
			bytes, err := json.Marshal(job.Status.Conditions)
			if err == nil {
				result.Message = string(bytes)
			}
		}
	}

	bytes, err := json.Marshal(result)
	if err != nil {
		return err
	}

	if err := r.rabbit.Send(bytes); err != nil {
		return err
	}

	return r.deployer.DeleteRoom(instance)
}

// ********************************* update room status *********************************

func (r *RoomReconciler) updateRoomStatus(instance *aiv1.Room, job *batch.Job) error {
	instance.Status.SetJob(&aiv1.NamespacedName{
		Name:      job.Name,
		Namespace: job.Namespace,
	})

	//fmt.Println("---------------------------------------------------------------")
	//fmt.Println("job name", job.Name)
	//fmt.Println("job completion time", job.Status.CompletionTime)
	//fmt.Println("job condition", job.Status.Conditions)
	//fmt.Println("job active", job.Status.Active)
	//fmt.Println("job succeeded", job.Status.Succeeded)
	//fmt.Println("job failed", job.Status.Failed)
	//fmt.Println("job backofflimit", *job.Spec.BackoffLimit)
	//fmt.Println("---------------------------------------------------------------")

	instance.Status.RoomStatusType = aiv1.RoomStatusTypeUnknown

	if instance.Status.Job == nil {
		return fmt.Errorf("could not identify status nil job in Room.Status")
	}

	if job.Status.Active > 0 {
		instance.Status.RoomStatusType = aiv1.RoomStatusTypeRunning
		return nil
	}

	if job.Status.Succeeded > 0 {
		instance.Status.RoomStatusType = aiv1.RoomStatusTypeSuccess
		return nil
	}

	if job.Status.Failed == *job.Spec.BackoffLimit+1 && job.Status.Failed > 0 {
		instance.Status.RoomStatusType = aiv1.RoomStatusTypeFailed
		return nil
	}
	return nil
}

// ********************************* deploy jobs *********************************

func (r *RoomReconciler) deployJob(instance *aiv1.Room, job *batch.Job) (*batch.Job, error) {
	syncedJob, err := r.deployer.SyncJob(instance, job)
	if err != nil {
		return nil, err
	}

	return syncedJob, nil
}

// ********************************* deploy config map *********************************

func (r *RoomReconciler) deployConfigMap(instance *aiv1.Room, configMap *core.ConfigMap) error {
	syncedConfigMap, err := r.deployer.SyncConfigMap(instance, configMap)
	if err != nil {
		return err
	}

	instance.Status.AddConfigMap(aiv1.NamespacedName{
		Name:      syncedConfigMap.Name,
		Namespace: syncedConfigMap.Namespace,
	})
	return nil
}

// ********************************* reconcile actors *********************************

func (r *RoomReconciler) reconcileActors(instance *aiv1.Room) error {
	if err := r.reconcileGimulatorActor(instance); err != nil {
		return err
	}

	if err := r.reconcileLoggerActor(instance); err != nil {
		return err
	}
	return nil
}

func (r *RoomReconciler) reconcileGimulatorActor(instance *aiv1.Room) error {
	instance.Spec.Actors = append(instance.Spec.Actors, aiv1.Actor{
		Name:    env.GimulatorName(),
		ID:      env.GimulatorID(),
		Image:   env.GimulatorImage(),
		Type:    aiv1.AIActorType(env.GimulatorType()),
		Command: env.GimulatorCmd(),
		Resources: aiv1.Resources{
			Limits: aiv1.Resource{
				CPU:       env.GimulatorLimitsCPU(),
				Memory:    env.GimulatorLimitsMemory(),
				Ephemeral: env.GimulatorLimitsEphemeral(),
			},
			Requests: aiv1.Resource{
				CPU:       env.GimulatorRequestsCPU(),
				Memory:    env.GimulatorRequestsMemory(),
				Ephemeral: env.GimulatorRequestsEphemeral(),
			},
		},
		EnvVars: []aiv1.EnvVar{
			{
				Key:   env.GimulatorRoleFilePathEnvKey(),
				Value: filepath.Join(env.GimulatorConfigVolumePath(), env.GimulatorRoleFileName()),
			},
		},
		VolumeMounts: []aiv1.VolumeMount{
			{
				Name: env.GimulatorConfigVolumeName(),
				Path: env.GimulatorConfigVolumePath(),
			},
		},
	})
	return nil
}

func (r *RoomReconciler) reconcileLoggerActor(instance *aiv1.Room) error {
	instance.Spec.Actors = append(instance.Spec.Actors, aiv1.Actor{
		Name:    env.LoggerName(),
		ID:      env.LoggerID(),
		Image:   env.LoggerImage(),
		Type:    aiv1.AIActorType(env.LoggerType()),
		Command: env.LoggerCmd(),
		Role:    env.LoggerRole(),
		Resources: aiv1.Resources{
			Limits: aiv1.Resource{
				CPU:       env.LoggerLimitsCPU(),
				Memory:    env.LoggerLimitsMemory(),
				Ephemeral: env.LoggerLimitsEphemeral(),
			},
			Requests: aiv1.Resource{
				CPU:       env.LoggerRequestsCPU(),
				Memory:    env.LoggerRequestsMemory(),
				Ephemeral: env.LoggerRequestsEphemeral(),
			},
		},
		EnvVars: []aiv1.EnvVar{
			{Key: env.LoggerS3URLEnvKey(), Value: env.S3URL()},
			{Key: env.LoggerS3AccessKeyEnvKey(), Value: env.S3AccessKey()},
			{Key: env.LoggerS3SecretKeyEnvKey(), Value: env.S3SecretKey()},
			{Key: env.LoggerS3BucketEnvKey(), Value: env.LoggerS3Bucket()},
			{Key: env.LoggerRecorderDirEnvKey(), Value: env.LoggerRecorderDir()},
			{Key: env.LoggerRabbitURIEnvKey(), Value: env.RabbitURI()},
			{Key: env.LoggerRabbitQueueEnvKey(), Value: env.RabbitQueue()},
		},
		VolumeMounts: []aiv1.VolumeMount{
			{
				Name: env.LoggerLogVolumeName(),
				Path: env.LoggerLogVolumePath(),
			},
		},
	})
	return nil
}

// ********************************* reconcile environment variables *********************************

func (r *RoomReconciler) reconcileEvnVars(instance *aiv1.Room) error {
	for i := range instance.Spec.Actors {
		actor := &instance.Spec.Actors[i]

		actor.EnvVars = append(actor.EnvVars,
			aiv1.EnvVar{Key: env.GimulatorHostEnvKey(), Value: env.GimulatorHost()},
			aiv1.EnvVar{Key: env.RoomIDEnvKey(), Value: strconv.Itoa(instance.Spec.ID)},
			aiv1.EnvVar{Key: env.RoomEndOfGameKeyEnvKey(), Value: env.RoomEndOfGameKey()},
			aiv1.EnvVar{Key: env.ClientIDEnvKey(), Value: strconv.Itoa(actor.ID)},
		)
	}

	return nil
}

// ********************************* reconcile args *********************************

func (r *RoomReconciler) reconcileArgs(instance *aiv1.Room) error {
	condition := ""
	for _, actor := range instance.Spec.Actors {
		if actor.Type != aiv1.AIActorTypeFinisher {
			continue
		}
		path := filepath.Join(env.SharedVolumePath(), name.TerminatedFileName(actor.Name))
		condition += fmt.Sprintf("-f %s && ", path)
	}
	condition += "true"

	for i := range instance.Spec.Actors {
		actor := &instance.Spec.Actors[i]

		if actor.Name == env.GimulatorName() {
			actor.Args = []string{fmt.Sprintf(env.GimulatorArgs, actor.Command, env.SharedVolumePath(), condition, condition)}
			continue
		}

		switch actor.Type {
		case aiv1.AIActorTypeFinisher:
			path := filepath.Join(env.SharedVolumePath(), name.TerminatedFileName(actor.Name))
			actor.Args = []string{fmt.Sprintf(env.FinisherArgs, env.SharedVolumePath(), path, actor.Command)}
		case aiv1.AIActorTypeMaster:
			actor.Args = []string{fmt.Sprintf(env.MasterArgs, env.SharedVolumePath(), actor.Command, condition, condition)}
		case aiv1.AIActorTypeSlave:
			actor.Args = []string{fmt.Sprintf(env.SlaveArgs, env.SharedVolumePath(), actor.Command, condition)}
		default:
			return fmt.Errorf("invalid actor type")
		}
	}
	return nil
}

// ********************************* reconcile config maps *********************************

func (r *RoomReconciler) reconcileConfigMaps(instance *aiv1.Room) error {
	for i := range instance.Spec.ConfigMaps {
		cm := &instance.Spec.ConfigMaps[i]

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

		cm.Data = data
	}
	return nil
}

// ********************************* reconcile sektch *********************************

func (r *RoomReconciler) reconcileSketch(instance *aiv1.Room) error {
	sketch, err := r.reconcilePrimitiveSketch(instance)
	if err != nil {
		return err
	}

	for i := range instance.Spec.Actors {
		actor := &instance.Spec.Actors[i]
		if actor.Name == env.GimulatorName() {
			continue
		}

		role := actor.Role
		id := actor.ID

		sketch.Actors = append(sketch.Actors, auth.Actor{
			Role: role,
			ID:   strconv.Itoa(id),
		})
	}

	sketch.Roles = append(sketch.Roles, auth.Role{
		Role: env.LoggerRole(),
		Rules: []auth.Rule{
			{
				Type:      "",
				Name:      "",
				Namespace: "",
				Methods:   []auth.Method{auth.Watch},
			},
		},
	})

	return r.reconcileFinalSketch(instance, sketch)
}

func (r *RoomReconciler) reconcilePrimitiveSketch(instance *aiv1.Room) (*auth.Config, error) {
	sketch := &auth.Config{}

	for _, cm := range instance.Spec.ConfigMaps {
		if cm.Name != instance.Spec.Sketch {
			continue
		}

		data := cm.Data
		err := yaml.Unmarshal([]byte(data), sketch)
		if err != nil {
			return nil, err
		}
		return sketch, nil
	}
	return nil, fmt.Errorf("can not find sketch config map")
}

func (r *RoomReconciler) reconcileFinalSketch(instance *aiv1.Room, sketch *auth.Config) error {
	for i := range instance.Spec.ConfigMaps {
		cm := &instance.Spec.ConfigMaps[i]

		if cm.Name != instance.Spec.Sketch {
			continue
		}

		b, err := yaml.Marshal(sketch)
		if err != nil {
			return err
		}
		cm.Data = string(b)

		return nil
	}
	return fmt.Errorf("can not find sketch config map")
}

// ********************************* reconcile volumes *********************************

func (r *RoomReconciler) reconcileVolumes(instance *aiv1.Room) error {
	if instance.Spec.Volumes == nil {
		instance.Spec.Volumes = make([]aiv1.Volume, 0)
	}

	if err := r.reconcileSharedVolumes(instance); err != nil {
		return err
	}

	if err := r.reconcileGimulatorVolumes(instance); err != nil {
		return err
	}

	if err := r.reconcileLoggerVolumes(instance); err != nil {
		return err
	}

	return nil
}

func (r *RoomReconciler) reconcileSharedVolumes(instance *aiv1.Room) error {
	sharedVolume := aiv1.Volume{
		EmptyDirVolume: &aiv1.EmptyDirVolume{
			Name: env.SharedVolumeName(),
		},
	}
	instance.Spec.Volumes = append(instance.Spec.Volumes, sharedVolume)

	for i := range instance.Spec.Actors {
		actor := &instance.Spec.Actors[i]
		actor.VolumeMounts = append(actor.VolumeMounts, aiv1.VolumeMount{
			Name: env.SharedVolumeName(),
			Path: env.SharedVolumePath(),
		})
	}

	return nil
}

func (r *RoomReconciler) reconcileGimulatorVolumes(instance *aiv1.Room) error {
	gimulatorVolume := aiv1.Volume{
		ConfigMapVolumes: &aiv1.ConfigMapVolume{
			Name:          env.GimulatorConfigVolumeName(),
			ConfigMapName: instance.Spec.Sketch,
			Path:          env.GimulatorRoleFileName(),
		},
	}
	instance.Spec.Volumes = append(instance.Spec.Volumes, gimulatorVolume)

	return nil
}

func (r *RoomReconciler) reconcileLoggerVolumes(instance *aiv1.Room) error {
	instance.Spec.Volumes = append(instance.Spec.Volumes, aiv1.Volume{
		EmptyDirVolume: &aiv1.EmptyDirVolume{
			Name: env.LoggerLogVolumeName(),
		},
	})

	return nil
}

// ********************************* reconcile owner reference *********************************

// func (r *RoomReconciler) reconcileOwnerReference(owner *aiv1.Room, instance meta.Object) error {
// 	return controllerutil.SetOwnerReference(owner, instance, r.Scheme)
// }

func (r *RoomReconciler) SetupWithManager(mgr ctrl.Manager) error {
	c, err := controller.New("resource", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// if err = c.Watch(
	// 	&source.Kind{Type: &aiv1.Room{}},
	// 	&handler.EnqueueRequestForObject{},
	// ); err != nil {
	// 	return err
	// }

	if err = c.Watch(
		&source.Kind{Type: &batch.Job{}},
		&handler.EnqueueRequestForOwner{
			OwnerType: &aiv1.Room{},
		},
	); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).For(&aiv1.Room{}).Complete(r)
}
