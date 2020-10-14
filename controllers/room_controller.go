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

package controllers

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	hubv1 "github.com/Gimulator/hub/api/v1"
	"github.com/Gimulator/hub/pkg/client"
	"github.com/Gimulator/hub/pkg/config"
)

var (
	ReconcilationTimeout = time.Second * 10
)

// RoomReconciler reconciles a Room object
type RoomReconciler struct {
	*client.Client
	*actorReconciler
	*gimulatorReconciler
	*directorReconciler

	Log    logr.Logger
	Scheme *runtime.Scheme
}

// NewRoomReconciler returns new instance of RoomReconciler
func NewRoomReconciler(mgr manager.Manager, log logr.Logger) (*RoomReconciler, error) {
	client, err := client.NewClient(mgr.GetClient(), mgr.GetScheme())
	if err != nil {
		return nil, err
	}

	actorReconciler, err := newActorReconciler(client, log)
	if err != nil {
		return nil, err
	}

	directorReconciler, err := newDirectorReconciler(client, log)
	if err != nil {
		return nil, err
	}

	gimulatorReconciler, err := newGimulatorReconciler(client, log)
	if err != nil {
		return nil, err
	}

	return &RoomReconciler{
		Log:                 log,
		Scheme:              mgr.GetScheme(),
		Client:              client,
		actorReconciler:     actorReconciler,
		gimulatorReconciler: gimulatorReconciler,
		directorReconciler:  directorReconciler,
	}, nil
}

// +kubebuilder:rbac:groups=hub.roboepics.com,resources=rooms,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=hub.roboepics.com,resources=rooms/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=pod,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pod/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=configmaps/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=persistVolumeClaim,verbs=get;list;watch;create;update;patch;delete

// Reconcile reconciles a request for a Room object
func (r *RoomReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx, cancle := context.WithTimeout(context.TODO(), ReconcilationTimeout)
	defer cancle()

	logger := r.Log.WithValues("reconciler", "Room", "room", req.NamespacedName)

	room, err := r.GetRoom(ctx, req.NamespacedName)
	if errors.IsNotFound(err) {
		logger.Info("room does not exist")
	} else if err != nil {
		logger.Error(err, "could not get room object")
		return ctrl.Result{}, err
	}

	logger.Info("starting to fetch game configuration")
	if err := config.FetchGameConfig(room); err != nil {
		logger.Error(err, "could not fetch game configuration", "game", room.Spec.Game)
		return ctrl.Result{}, err
	}

	logger.Info("starting to checkup needed PVCs")
	if err := r.checkPVCs(ctx, room); err != nil {
		logger.Error(err, "could not checkup  needed PVCs")
		return ctrl.Result{}, err
	}

	logger.Info("starting to reconcile Gimulator")
	if err := r.reconcileGimulator(ctx, room); err != nil {
		logger.Error(err, "could not reconcile gimulator")
	}

	logger.Info("starting to reconcile director")
	if err := r.reconcileDirector(ctx, room); err != nil {
		logger.Error(err, "could not reconcile director")
	}

	logger.Info("starting to reconcile actors")
	for i := range room.Spec.Actors {
		actor := &room.Spec.Actors[i]
		if err := r.reconcileActor(ctx, room, actor); err != nil {
			logger.Error(err, "could not reconcile actor", "actor", actor.ID)
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *RoomReconciler) checkPVCs(ctx context.Context, room *hubv1.Room) error {
	if room.Spec.GameConfig.DataPVCName != "" {
		key := types.NamespacedName{
			Name:      room.Spec.GameConfig.DataPVCName,
			Namespace: room.Namespace,
		}
		if _, err := r.GetPVC(ctx, key); err != nil {
			return err
		}
	}
	if room.Spec.GameConfig.FactPVCName != "" {
		key := types.NamespacedName{
			Name:      room.Spec.GameConfig.FactPVCName,
			Namespace: room.Namespace,
		}
		if _, err := r.GetPVC(ctx, key); err != nil {
			return err
		}
	}

	return nil
}

func (r *RoomReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&hubv1.Room{}).
		Complete(r)
}
