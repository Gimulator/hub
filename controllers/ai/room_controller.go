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
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/Gimulator/Gimulator/auth"
	"github.com/Gimulator/hub/utils/cache"
	"github.com/Gimulator/hub/utils/convertor"
	env "github.com/Gimulator/hub/utils/environment"
	"github.com/Gimulator/hub/utils/name"
	"github.com/Gimulator/hub/utils/storage"
	"github.com/go-logr/logr"
	uuid "github.com/satori/go.uuid"
	"gopkg.in/yaml.v2"
	batch "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	aiv1 "github.com/Gimulator/hub/apis/ai/v1"
	"github.com/Gimulator/hub/utils/deployer"
)

// RoomReconciler reconciles a Room object
type RoomReconciler struct {
	client.Client
	log      logr.Logger
	Scheme   *runtime.Scheme
	deployer *deployer.Deployer
}

func NewRoomReconciler(mgr manager.Manager, log logr.Logger) (*RoomReconciler, error) {
	return &RoomReconciler{
		Client:   mgr.GetClient(),
		log:      log,
		Scheme:   mgr.GetScheme(),
		deployer: deployer.NewDeployer(mgr.GetClient()),
	}, nil
}

// +kubebuilder:rbac:groups=hub.xerac.cloud,resources=rooms,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=hub.xerac.cloud,resources=rooms/status,verbs=get;update;patch

func (r *RoomReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(env.APICallTimeout))
	defer cancel()

	src := &aiv1.Room{}
	if err := r.Get(ctx, req.NamespacedName, src); err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	dst := &aiv1.Room{}
	src.DeepCopyInto(dst)

	if err := r.reconcileActors(src, dst); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileConfigMaps(src, dst); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileSketch(src, dst); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileVolumes(src, dst); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileArgs(src, dst); err != nil {
		return ctrl.Result{}, err
	}

	for _, cm := range dst.Spec.ConfigMaps {
		configMap, err := convertor.ConvertConfigMap(cm)
		if err != nil {
			return ctrl.Result{}, err
		}

		if err := r.reconcileConfigMap(dst, configMap); err != nil {
			return ctrl.Result{}, err
		}
	}

	job, err := convertor.ConvertRoom(dst)
	if err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileJob(src, job); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// ********************************* reconcile jobs *********************************

func (r *RoomReconciler) reconcileJob(src *aiv1.Room, job *batch.Job) error {
	syncedJob, err := r.deployer.SyncJob(job)
	if err != nil {
		return err
	}
	fmt.Println(syncedJob.Status.Conditions)
	return nil
}

// ********************************* reconcile config map *********************************

func (r *RoomReconciler) reconcileConfigMap(src *aiv1.Room, configMap *core.ConfigMap) error {
	if err := r.deployer.SyncConfigMap(configMap); err != nil {
		return err
	}

	return nil
}

// ********************************* reconcile actors *********************************

func (r *RoomReconciler) reconcileActors(src, dst *aiv1.Room) error {
	if err := r.reconcileGimulatorActor(src, dst); err != nil {
		return err
	}

	if err := r.reconcileLoggerActor(src, dst); err != nil {
		return err
	}
	return nil
}

func (r *RoomReconciler) reconcileGimulatorActor(src, dst *aiv1.Room) error {
	dst.Spec.Actors = append(dst.Spec.Actors, aiv1.Actor{
		Name:    env.GimulatorName(),
		ID:      env.GimulatorID(),
		Image:   env.GimulatorImage(),
		Type:    aiv1.AIActorType(env.GimulatorType()),
		Command: env.GimulatorCmd(),
		Resources: aiv1.Resources{
			Limits: aiv1.Resource{
				CPU:       env.GimulatorResourceLimitsCPU(),
				Memory:    env.GimulatorResourceLimitsMemory(),
				Ephemeral: env.GimulatorResourceLimitsEphemeral(),
			},
			Requests: aiv1.Resource{
				CPU:       env.GimulatorResourceRequestsCPU(),
				Memory:    env.GimulatorResourceRequestsMemory(),
				Ephemeral: env.GimulatorResourceRequestsEphemeral(),
			},
		},
		EnvVars: make([]aiv1.EnvVar, 0),
		VolumeMounts: []aiv1.VolumeMount{
			{
				Name: env.GimulatorConfigVolumeName(),
				Path: env.GimulatorConfigVolumePath(),
			},
		},
	})
	return nil
}

func (r *RoomReconciler) reconcileLoggerActor(src, dst *aiv1.Room) error {
	dst.Spec.Actors = append(dst.Spec.Actors, aiv1.Actor{
		Name:    env.LoggerName(),
		ID:      env.LoggerID(),
		Image:   env.LoggerImage(),
		Type:    aiv1.AIActorType(env.LoggerType()),
		Command: env.LoggerCmd(),
		Resources: aiv1.Resources{
			Limits: aiv1.Resource{
				CPU:       env.LoggerResourceLimitsCPU(),
				Memory:    env.LoggerResourceLimitsMemory(),
				Ephemeral: env.LoggerResourceLimitsEphemeral(),
			},
			Requests: aiv1.Resource{
				CPU:       env.LoggerResourceRequestsCPU(),
				Memory:    env.LoggerResourceRequestsMemory(),
				Ephemeral: env.LoggerResourceRequestsEphemeral(),
			},
		},
		EnvVars: []aiv1.EnvVar{
			{Key: "LOGGER_ADDRESS", Value: env.GimulatorIP()},
			{Key: "LOGGER_S3_URL", Value: env.S3URL()},
			{Key: "LOGGER_S3_ACCESS_KEY", Value: env.S3AccessKey()},
			{Key: "LOGGER_S3_SECRET_KEY", Value: env.S3SecretKey()},
			{Key: "LOGGER_S3_BUCKET", Value: env.LoggerS3Bucket()},
			{Key: "LOGGER_S3_KEY", Value: name.S3LoggerKeyName(dst.Spec.ID)},
			{Key: "LOGGER_RECORD_DIR", Value: env.LoggerRecordDir()},
			{Key: "LOGGER_RECORD_END_KEY", Value: env.GimulatorEndOfGameKey()},
			{Key: "LOGGER_RABBIT_URI", Value: env.LoggerRabbitURI()},
			{Key: "LOGGER_RABBIT_QUEUE", Value: env.LoggerRabbitQueue()},
		},
		VolumeMounts: make([]aiv1.VolumeMount, 0),
	})
	return nil
}

// ********************************* reconcile args *********************************

func (r *RoomReconciler) reconcileArgs(src, dst *aiv1.Room) error {
	condition := ""
	for _, actor := range dst.Spec.Actors {
		if actor.Type != aiv1.AIActorTypeFinisher {
			continue
		}
		path := filepath.Join(env.SharedVolumePath(), name.TerminatedFileName(actor.Name))
		condition += fmt.Sprintf("-f %s && ", path)
	}
	condition += "true"

	for i := range dst.Spec.Actors {
		actor := &dst.Spec.Actors[i]

		switch actor.Type {
		case aiv1.AIActorTypeFinisher:
			path := filepath.Join(env.SharedVolumePath(), name.TerminatedFileName(actor.Name))
			actor.Args = []string{fmt.Sprintf(env.FinisherArgs, path, actor.Command)}
		case aiv1.AIActorTypeMaster:
			actor.Args = []string{fmt.Sprintf(env.MasterArgs, actor.Command, condition, condition)}
		case aiv1.AIActorTypeSlave:
			actor.Args = []string{fmt.Sprintf(env.SlaveArgs, actor.Command, condition)}
		default:
			return fmt.Errorf("invalid actor type")
		}
	}
	return nil
}

// ********************************* reconcile config maps *********************************

func (r *RoomReconciler) reconcileConfigMaps(src, dst *aiv1.Room) error {
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

// ********************************* reconcile sektch *********************************

func (r *RoomReconciler) reconcileSketch(src, dst *aiv1.Room) error {
	sketch, err := r.reconcilePrimitiveSketch(src, dst)
	if err != nil {
		return err
	}

	for i := range dst.Spec.Actors {
		actor := &dst.Spec.Actors[i]

		if actor.Name == env.GimulatorName() {
			continue
		}

		role := actor.Role
		username := name.ContainerName(actor.Name, actor.ID)
		password := uuid.NewV4().String()

		sketch.Actors = append(sketch.Actors, auth.Actor{
			Role:     role,
			Username: username,
			Password: password,
		})

		actor.EnvVars = append(actor.EnvVars,
			aiv1.EnvVar{Key: env.UsernameEnvVarKey, Value: username},
			aiv1.EnvVar{Key: env.PasswordEnvVarKey, Value: password},
		)
	}

	return r.reconcileFinalSketch(src, dst, sketch)
}

func (r *RoomReconciler) reconcilePrimitiveSketch(src, dst *aiv1.Room) (*auth.Config, error) {
	sketch := &auth.Config{}

	for _, cm := range dst.Spec.ConfigMaps {
		if cm.Name != dst.Spec.Sketch {
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

func (r *RoomReconciler) reconcileFinalSketch(src, dst *aiv1.Room, sketch *auth.Config) error {
	for _, cm := range dst.Spec.ConfigMaps {
		if cm.Name != dst.Spec.Sketch {
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

func (r *RoomReconciler) reconcileVolumes(src, dst *aiv1.Room) error {
	if dst.Spec.Volumes == nil {
		dst.Spec.Volumes = make([]aiv1.Volume, 0)
	}

	if err := r.reconcileSharedVolumes(src, dst); err != nil {
		return err
	}

	if err := r.reconcileGimulatorVolume(src, dst); err != nil {
		return err
	}

	//if err := r.reconcileLoggerVolume(src, dst); err != nil {
	//	return err
	//}

	return nil
}

func (r *RoomReconciler) reconcileSharedVolumes(src, dst *aiv1.Room) error {
	sharedVolume := aiv1.Volume{
		EmptyDirVolume: &aiv1.EmptyDirVolume{
			Name: env.SharedVolumeName(),
		},
	}
	dst.Spec.Volumes = append(dst.Spec.Volumes, sharedVolume)

	return nil
}

func (r *RoomReconciler) reconcileGimulatorVolume(src, dst *aiv1.Room) error {
	gimulatorVolume := aiv1.Volume{
		EmptyDirVolume: &aiv1.EmptyDirVolume{
			Name: env.GimulatorConfigVolumeName(),
		},
	}
	dst.Spec.Volumes = append(dst.Spec.Volumes, gimulatorVolume)

	return nil
}

//func (r *RoomReconciler) reconcileLoggerVolume(src, dst *aiv1.Room) error {
//	loggerVolume := aiv1.Volume{
//		EmptyDirVolume: &aiv1.EmptyDirVolume{
//			Name: env.LoggerLogDirName(),
//		},
//	}
//	dst.Spec.Volumes = append(dst.Spec.Volumes, loggerVolume)
//
//	return nil
//}

func (r *RoomReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).For(&aiv1.Room{}).Complete(r)
}
