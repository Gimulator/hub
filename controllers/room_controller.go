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
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/source"

	hubv1 "github.com/Gimulator/hub/api/v1"
	"github.com/Gimulator/hub/pkg/client"
	"github.com/Gimulator/hub/pkg/config"
	"github.com/Gimulator/hub/pkg/reporter"
	"github.com/Gimulator/hub/pkg/timer"
)

var (
	ReconcilationTimeout = time.Second * 20
)

// RoomReconciler reconciles a Room object
type RoomReconciler struct {
	*client.Client
	*actorReconciler
	*gimulatorReconciler
	*directorReconciler

	Log       logr.Logger
	Scheme    *runtime.Scheme
	clientset *kubernetes.Clientset
	reporter  *reporter.Reporter
	timer     *timer.Timer
}

// NewRoomReconciler returns new instance of RoomReconciler
func NewRoomReconciler(mgr manager.Manager, log logr.Logger, reporter *reporter.Reporter, client *client.Client, clientset *kubernetes.Clientset) (*RoomReconciler, error) {
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

	roomTimer, err := timer.NewTimer(clientset, ctrl.Log.WithName("timer"), reporter, client)
	if err != nil {
		return nil, err
	}

	return &RoomReconciler{
		Log:                 log,
		Scheme:              mgr.GetScheme(),
		Client:              client,
		clientset:           clientset,
		actorReconciler:     actorReconciler,
		gimulatorReconciler: gimulatorReconciler,
		directorReconciler:  directorReconciler,
		reporter:            reporter,
		timer:               roomTimer,
	}, nil
}

// +kubebuilder:rbac:groups=hub.roboepics.com,resources=rooms,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=hub.roboepics.com,resources=rooms/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=configmaps/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete

// Reconcile reconciles a request for a Room object
func (r *RoomReconciler) Reconcile(cx context.Context, req ctrl.Request) (ctrl.Result, error) {
	ctx, cancle := context.WithTimeout(context.TODO(), ReconcilationTimeout)
	defer cancle()

	logger := r.Log.WithValues("reconciler", "Room", "room", req.NamespacedName)
	logger.Info("starting to reconcile room")

	room, err := r.GetRoom(ctx, req.NamespacedName)
	if errors.IsNotFound(err) {
		logger.Info("room does not exist")
		return ctrl.Result{}, nil
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

		room.Status.DirectorStatus = corev1.PodUnknown
		room.Status.GimulatorStatus = corev1.PodUnknown
		room.Status.ActorStatuses = make(map[string]corev1.PodPhase)
		for _, actor := range room.Spec.Actors {
			room.Status.ActorStatuses[actor.Name] = corev1.PodUnknown
		}

		if room, err = r.SyncRoom(ctx, room); err != nil {
			logger.Error(err, "could not update room after generating tokens")
			return ctrl.Result{}, err
		}
	}

	logger.Info("starting to fetch setting")
	if err := config.FetchSetting(ctx, room); err != nil {
		logger.Error(err, "could not fetch setting", "problem", room.Spec.ProblemID)
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
		return ctrl.Result{}, err
	}

	logger.Info("starting to reconcile director")
	if err := r.reconcileDirector(ctx, room); err != nil {
		logger.Error(err, "could not reconcile director")
		return ctrl.Result{}, err
	}

	logger.Info("starting to reconcile actors")
	for _, actor := range room.Spec.Actors {
		if err := r.reconcileActor(ctx, room, actor); err != nil {
			logger.Error(err, "could not reconcile actor", "actor", actor.Name)
			return ctrl.Result{}, err
		}
	}

	logger.Info("starting to sync timers")
	r.timer.SyncTimers(room)

	logger.Info("starting to sync room")
	if _, err := r.SyncRoom(ctx, room); err != nil {
		logger.Error(err, "could  not sync room")
		return ctrl.Result{}, err
	}

	logger.Info("starting to reconcile status and report it")
	if shouldDelete, err := r.reporter.Report(ctx, room); err != nil {
		logger.Error(err, "could not report")
		return ctrl.Result{}, err
	} else if shouldDelete {
		if err := r.DeleteRoom(ctx, room); err != nil {
			logger.Error(err, "could not reconcile statuses")
			return ctrl.Result{}, err
		}
	}

	logger.Info("end of reconciling")
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
	if room.Spec.Setting.DataPVCNames == nil {
		return nil
	}
	if room.Spec.Setting.DataPVCNames.Public != nil {
		for _, pvcName := range room.Spec.Setting.DataPVCNames.Public {
			key := types.NamespacedName{
				Name:      pvcName,
				Namespace: room.Namespace,
			}
			if _, err := r.GetPVC(ctx, key); err != nil {
				return err
			}
		}
	}
	if room.Spec.Setting.DataPVCNames.Private != nil {
		for _, pvcName := range room.Spec.Setting.DataPVCNames.Private {
			key := types.NamespacedName{
				Name:      pvcName,
				Namespace: room.Namespace,
			}
			if _, err := r.GetPVC(ctx, key); err != nil {
				return err
			}
		}
	}

	// if room.Spec.ProblemSettings.FactPVCName != "" {
	// 	key := types.NamespacedName{
	// 		Name:      room.Spec.ProblemSettings.FactPVCName,
	// 		Namespace: room.Namespace,
	// 	}
	// 	if _, err := r.GetPVC(ctx, key); err != nil {
	// 		return err
	// 	}
	// }

	return nil
}

func (r *RoomReconciler) SetupWithManager(mgr ctrl.Manager) error {
	builder := ctrl.NewControllerManagedBy(mgr)
	builder = builder.For(&hubv1.Room{})
	builder = builder.Watches(
		&source.Kind{Type: &corev1.Pod{}},
		&handler.EnqueueRequestForOwner{
			OwnerType: &hubv1.Room{},
		},
	)

	return builder.Complete(r)
}
