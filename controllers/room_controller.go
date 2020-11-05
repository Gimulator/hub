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
	uuid "github.com/satori/go.uuid"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	hubv1 "github.com/Gimulator/hub/api/v1"
	"github.com/Gimulator/hub/pkg/client"
	"github.com/Gimulator/hub/pkg/config"
	"github.com/Gimulator/hub/pkg/reporter"
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

	Log      logr.Logger
	Scheme   *runtime.Scheme
	reporter *reporter.Reporter
}

// NewRoomReconciler returns new instance of RoomReconciler
func NewRoomReconciler(mgr manager.Manager, log logr.Logger, reporter *reporter.Reporter) (*RoomReconciler, error) {
	client, err := client.NewClient(mgr.GetClient(), mgr.GetScheme())
	if err != nil {
		return nil, err
	}

	gimulatorReconciler, err := newGimulatorReconciler(client, log)
	if err != nil {
		return nil, err
	}

	directorReconciler, err := newDirectorReconciler(client, log)
	if err != nil {
		return nil, err
	}

	actorReconciler, err := newActorReconciler(client, log)
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
		reporter:            reporter,
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
	logger.Info("starting to reconcile room")

	room, err := r.GetRoom(ctx, req.NamespacedName)
	if errors.IsNotFound(err) {
		logger.Info("room does not exist")
	} else if err != nil {
		logger.Error(err, "could not get room object")
		return ctrl.Result{}, err
	}

	// tokens should be generated one time in the life-cycle of a room,
	// so we update room with generated tokens and then for next call of reconcile,
	// we will do nothing
	logger.Info("starting to generate tokens if neeeded")
	if wasGenerated, err := r.generateTokens(room); err != nil {
		logger.Error(err, "cloud not generate tokens")
		return ctrl.Result{}, err
	} else if wasGenerated {
		logger.Info(("starting to update room after generating tokens"))
		if room, err = r.SyncRoom(ctx, room); err != nil {
			logger.Error(err, "could not update room after generating tokens")
			return ctrl.Result{}, err
		}
	}

	logger.Info("starting to fetch game configuration")
	if err := config.FetchProblemSettings(room); err != nil {
		logger.Error(err, "could not fetch game configuration", "game", room.Spec.ProblemID)
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
	for _, actor := range room.Spec.Actors {
		if err := r.reconcileActor(ctx, room, actor); err != nil {
			logger.Error(err, "could not reconcile actor", "actor", actor.ID)
			return ctrl.Result{}, err
		}
	}

	logger.Info("starting to sync room")
	if _, err := r.SyncRoom(ctx, room); err != nil {
		logger.Error(err, "could  not sync room")
		return ctrl.Result{}, err
	}

	logger.Info("starting to reconcile statuses")
	if shouldDelete, err := r.reporter.Report(ctx, room); err != nil {
		logger.Error(err, "could not reconcile statuses")
		return ctrl.Result{}, err
	} else if shouldDelete {
		return ctrl.Result{}, r.DeleteRoom(ctx, room)
	}

	return ctrl.Result{}, nil
}

func (r *RoomReconciler) generateTokens(room *hubv1.Room) (bool, error) {
	flag := false

	if room.Spec.Director.Token == "" {
		room.Spec.Director.Token = uuid.NewV4().String()
		flag = true
	}

	for _, actor := range room.Spec.Actors {
		if actor.Token == "" {
			actor.Token = uuid.NewV4().String()
			flag = true
		}
	}

	return flag, nil
}

func (r *RoomReconciler) checkPVCs(ctx context.Context, room *hubv1.Room) error {
	if room.Spec.ProblemSettings.DataPVCName != "" {
		key := types.NamespacedName{
			Name:      room.Spec.ProblemSettings.DataPVCName,
			Namespace: room.Namespace,
		}
		if _, err := r.GetPVC(ctx, key); err != nil {
			return err
		}
	}
	if room.Spec.ProblemSettings.FactPVCName != "" {
		key := types.NamespacedName{
			Name:      room.Spec.ProblemSettings.FactPVCName,
			Namespace: room.Namespace,
		}
		if _, err := r.GetPVC(ctx, key); err != nil {
			return err
		}
	}

	return nil
}

func (r *RoomReconciler) SetupWithManager(mgr ctrl.Manager) error {
	builder := ctrl.NewControllerManagedBy(mgr)
	builder = builder.For(&hubv1.Room{})
	builder = builder.Owns(&corev1.Pod{})
	return builder.Complete(r)
}
