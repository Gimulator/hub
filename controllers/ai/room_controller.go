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
	"time"

	"github.com/go-logr/logr"
	env "gitlab.com/Syfract/Xerac/hub/utils/environment"
	batch "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	aiv1 "gitlab.com/Syfract/Xerac/hub/apis/ai/v1"
	hubaiv1 "gitlab.com/Syfract/Xerac/hub/apis/ai/v1"
	"gitlab.com/Syfract/Xerac/hub/utils/deployer"
)

// RoomReconciler reconciles a Room object
type RoomReconciler struct {
	client.Client
	log                 logr.Logger
	Scheme              *runtime.Scheme
	deployer            *deployer.Deployer
	actorReconciler     *ActorReconciler
	configMapReconciler *ConfigMapReconciler
	sketchReconciler    *SketchReconciler
	volumeReconciler    *VolumeReconciler
}

func NewRoomReconciler(mgr manager.Manager, log logr.Logger) (*RoomReconciler, error) {
	ar, err := NewActorReconciler(log.WithName("actor"))
	if err != nil {
		return nil, err
	}

	vr, err := NewVolumeReconciler(log.WithName("volume"))
	if err != nil {
		return nil, err
	}

	sr, err := NewSketchReconciler(mgr, log.WithName("sketch"))
	if err != nil {
		return nil, err
	}

	cr, err := NewConfigMapReconciler(mgr, log.WithName("configmap"))
	if err != nil {
		return nil, err
	}

	return &RoomReconciler{
		Client:              mgr.GetClient(),
		log:                 log,
		Scheme:              mgr.GetScheme(),
		deployer:            deployer.NewDeployer(mgr.GetClient()),
		actorReconciler:     ar,
		configMapReconciler: cr,
		sketchReconciler:    sr,
		volumeReconciler:    vr,
	}, nil
}

// +kubebuilder:rbac:groups=hub.xerac.cloud,resources=rooms,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=hub.xerac.cloud,resources=rooms/status,verbs=get;update;patch

func (r *RoomReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(env.APICallTimeout))
	defer cancel()

	r.log.Info("Get aiv1.Room")
	instance := aiv1.Room{}
	if err := r.Get(ctx, req.NamespacedName, &instance); err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	job := &batch.Job{}

	r.log.Info("reconcile actor")
	if err := r.actorReconciler.Reconcile(instance, job); err != nil {
		return ctrl.Result{}, err
	}

	r.log.Info("reconcile volume")
	if err := r.volumeReconciler.Reconcile(instance, job); err != nil {
		return ctrl.Result{}, err
	}

	r.log.Info("reconcile config map")
	if err := r.configMapReconciler.Reconcile(instance, job); err != nil {
		return ctrl.Result{}, err
	}

	r.log.Info("reconcile sketch")
	if err := r.sketchReconciler.Reconcile(instance, job); err != nil {
		return ctrl.Result{}, err
	}

	r.log.Info("deploy")
	_, err := r.deployer.SyncJob(job)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *RoomReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).For(&hubaiv1.Room{}).Complete(r)
}
