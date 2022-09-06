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

package main

import (
	"flag"
	"os"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	hubv1 "github.com/Gimulator/hub/api/v1"
	"github.com/Gimulator/hub/controllers"
	"github.com/Gimulator/hub/pkg/client"
	"github.com/Gimulator/hub/pkg/mq"
	"github.com/Gimulator/hub/pkg/reporter"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = hubv1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func main() {
	namespace := os.Getenv("HUB_NAMESPACE")
	if namespace == "" {
		namespace = "hub-system"
	}

	rabbitHost := os.Getenv("HUB_RABBIT_HOST")
	rabbitUsername := os.Getenv("HUB_RABBIT_USERNAME")
	rabbitPassword := os.Getenv("HUB_RABBIT_PASSWORD")
	rabbitQueue := os.Getenv("HUB_RABBIT_RESULT_QUEUE")
	token := os.Getenv("HUB_GIMULATOR_TOKEN")

	var metricsAddr string
	var enableLeaderElection bool
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	resyncPeriod := time.Second * 60
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		Port:               9443,
		LeaderElection:     enableLeaderElection,
		LeaderElectionID:   "fa842ee4.roboepics.com",
		Namespace:          namespace,
		SyncPeriod:         &resyncPeriod,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// Setting up RabbitMq
	rabbit, err := mq.NewRabbit(rabbitHost, rabbitUsername, rabbitPassword, rabbitQueue)
	if err != nil {
		setupLog.Error(err, "unable to create rabbit instance")
		os.Exit(1)
	}

	controllerClient, err := client.NewClient(mgr.GetClient(), mgr.GetScheme())
	if err != nil {
		setupLog.Error(err, "unable to create client instance")
		os.Exit(1)
	}

	clientSetConfig, err := rest.InClusterConfig()
	if err != nil {
		setupLog.Error(err, "unable to get client set config")
		os.Exit(1)
	}

	clientSet, err := kubernetes.NewForConfig(clientSetConfig)
	if err != nil {
		setupLog.Error(err, "unable to initialize k8s client set")
		os.Exit(1)
	}

	reporterObj, err := reporter.NewReporter(token, rabbit, controllerClient, clientSet)
	if err != nil {
		setupLog.Error(err, "unable to create reporter instance")
		os.Exit(1)
	}

	// Setting up room controller
	roomReconciler, err := controllers.NewRoomReconciler(mgr, ctrl.Log.WithName("room-controller"), reporterObj, controllerClient, clientSet)
	if err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "room-controller")
		os.Exit(1)
	}

	if err := roomReconciler.SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to setup controller", "controller", "room-controller")
		os.Exit(1)
	}

	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
