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

package ml

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/go-logr/logr"
	"go.etcd.io/etcd/client"
	batch "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	mlv1 "github.com/Gimulator/hub/apis/ml/v1"
	"github.com/Gimulator/hub/utils/deployer"
	env "github.com/Gimulator/hub/utils/environment"
	"github.com/Gimulator/hub/utils/name"
	rabbit "github.com/Gimulator/hub/utils/rabbitMQ"
	"github.com/Gimulator/hub/utils/storage"
)

// MLReconciler reconciles a ML object
type MLReconciler struct {
	client.Client
	log      logr.Logger
	deployer *deployer.Deployer
	Scheme   *runtime.Scheme
	rabbit   *rabbit.Rabbit
}

func NewMLReconciler(mgr manager.Manager, log logr.Logger) (*MLReconciler, error) {
	rabbit, err := rabbit.NewRabbit()
	if err != nil {
		return nil, err
	}

	scheme := mgr.GetScheme()

	return &MLReconciler{
		log:      log,
		Scheme:   scheme,
		deployer: deployer.NewDeployer(mgr.GetClient(), scheme),
		rabbit:   rabbit,
	}, nil
}

// +kubebuilder:rbac:groups=hub.xerac.cloud,resources=mls,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=hub.xerac.cloud,resources=mls/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch,resources=jobs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods/log,verbs=get
// +kubebuilder:rbac:groups=core,resources=resourcequotas,verbs=get;list;watch;create;update;patch;delete

func (m *MLReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	log := m.log.WithValues("name", req.Name, "namespace", req.Namespace)
	log.Info("starting to reconcile")

	src, err := m.deployer.GetML(req.NamespacedName)
	if errors.IsNotFound(err) {
		log.Info("ml does not exist")
		return ctrl.Result{}, nil
	}
	if err != nil {
		log.Error(err, "could not get ml")
		return ctrl.Result{}, err
	}

	log.Info("starting to check quota")
	if err := m.checkQuota(src); err != nil {
		log.WithValues("error", err.Error()).Info("not enough quota")
		src.Status.StatusType = mlv1.MLStatusTypePending
		if _, err := m.deployer.SyncML(src); err != nil {
			log.Error(err, "could not update ml status")
		}
		return ctrl.Result{}, nil
	}

	if src.Status.StatusType == "" {
		log.Info("starting to set status to unknown")
		src.Status.StatusType = mlv1.MLStatusTypeUnknown
		if _, err := m.deployer.SyncML(src); err != nil {
			log.Error(err, "could not update ml status")
			return ctrl.Result{}, err
		}
	}

	job := &batch.Job{}

	log.Info("starting to create job manifest")
	if err := m.jobManifest(src, job); err != nil {
		log.Error(err, "could not create job manifest")
		return ctrl.Result{}, err
	}

	log.Info("starting to create init container manifest")
	if err := m.initContainerManifest(src, job); err != nil {
		log.Error(err, "could not create init container manifest")
		return ctrl.Result{}, err
	}

	log.Info("starting to create evaluator manifest")
	if err := m.evaluatorContainerManifest(src, job); err != nil {
		log.Error(err, "could not create evaluator manifest")
		return ctrl.Result{}, err
	}

	log.Info("starting to sync job")
	syncedJob, err := m.deployer.SyncJob(src, job)
	if err != nil {
		log.Error(err, "could not sync job")
		return ctrl.Result{}, err
	}

	log.Info("starting to set status to running")
	src.Status.StatusType = mlv1.MLStatusTypeRunning
	if _, err := m.deployer.SyncML(src); err != nil {
		return ctrl.Result{}, err
	}

	log.Info("starting to reconcile syncedJob")
	if err := m.reconcileSyncedJob(src, syncedJob); err != nil {
		log.Error(err, "could not reconcile syncedJob manifest")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (m *MLReconciler) checkQuota(src *mlv1.ML) error {
	log := m.log.WithValues("name", src.Name, "namespace", src.Namespace)

	if src.Status.StatusType == mlv1.MLStatusTypeRunning || src.Status.StatusType == mlv1.MLStatusTypeUnknown {
		log.Info("resource is already in progress")
		return nil
	}

	quota, err := m.deployer.GetResourceQuota(types.NamespacedName{Name: "resource-quota", Namespace: "hub-system"})
	if err != nil {
		return err
	}
	log.Info("-----------------------------------------")
	log.Info(quota.Status.String())

	cpuLimit := quota.Status.Hard["limits.cpu"].DeepCopy()
	cpuUsed := quota.Status.Used["limits.cpu"].DeepCopy()
	log.WithValues("used cpu", cpuUsed.String()).WithValues("hard cpu", cpuLimit.String()).Info("more information")
	cpuLimit.Sub(cpuUsed.DeepCopy())
	log.WithValues("available cpu", cpuLimit.String()).Info("more information")

	memoryLimit := quota.Status.Hard["limits.memory"].DeepCopy()
	memoryUsed := quota.Status.Used["limits.memory"].DeepCopy()
	log.WithValues("used memory", memoryUsed.String()).WithValues("hard memory", memoryLimit.String()).Info("more information")
	memoryLimit.Sub(memoryUsed.DeepCopy())
	log.WithValues("available memory", memoryLimit.String()).Info("more information")

	resourceCPU := resource.MustParse(src.Spec.CPUResourceLimit)
	resourceMemory := resource.MustParse(src.Spec.MemoryResourceLimit)

	log.WithValues("resource cpu", resourceCPU.String()).WithValues("resource memory", resourceMemory.String()).Info("more information")

	if cpuLimit.Cmp(resourceCPU) < 0 {
		return fmt.Errorf("not enough cpu quota")
	}

	if memoryLimit.Cmp(resourceMemory) < 0 {
		return fmt.Errorf("not enough memory quota")
	}

	return nil
}

func (m *MLReconciler) reconcileSyncedJob(src *mlv1.ML, job *batch.Job) error {
	log := m.log.WithValues("name", src.Name, "namespace", src.Namespace)
	if job.Status.Active > 0 {
		log.Info("job has still active pods")

		creationTime := job.CreationTimestamp
		if creationTime.IsZero() {
			log.Info("job's time is zero")
			return nil
		}

		diff := time.Now().Sub(creationTime.Time).Seconds()
		if diff > 7200 {
			log.Info("job's deadline has expired")
			return m.reconcileTimeLimitExceededJob(src, job)
		}
		return nil
	}

	if job.Status.Succeeded > 0 {
		log.Info("job has been successful")
		return m.deleteML(src, job)
	}

	if job.Status.Conditions != nil && len(job.Status.Conditions) > 0 {
		con := job.Status.Conditions[0]
		if con.Type == batch.JobComplete {
			log.Info("job has been completed")
			return m.deleteML(src, job)
		}
		if con.Type == batch.JobFailed {
			log.Info("job has been failed")
			return m.reconcileFailedML(src, job)
		}
	}

	return nil
}

func (m *MLReconciler) reconcileTimeLimitExceededJob(src *mlv1.ML, job *batch.Job) error {
	log := m.log.WithValues("name", src.Name, "namespace", src.Namespace)
	log.Info("starting to delete time limit exceeded job")

	result := struct {
		RoomID  int    `json:"run_id"`
		Status  string `json:"status"`
		Message string `json:"message"`
	}{
		RoomID:  src.Spec.RunID,
		Status:  "FAIL",
		Message: "Time limit exceeded",
	}

	bytes, err := json.Marshal(result)
	if err != nil {
		return err
	}

	if err := m.rabbit.Send(bytes); err != nil {
		return err
	}

	return m.deleteML(src, job)
}

func (m *MLReconciler) reconcileFailedML(src *mlv1.ML, job *batch.Job) error {
	log := m.log.WithValues("name", src.Name, "namespace", src.Namespace)
	log.Info("starting to delete failed job")

	result := struct {
		RoomID  int    `json:"run_id"`
		Status  string `json:"status"`
		Message string `json:"message"`
	}{
		RoomID:  src.Spec.RunID,
		Status:  "FAIL",
		Message: "The cause of the failure is unknown",
	}

	conditions := job.Status.Conditions
	if conditions != nil {
		bytes, err := json.Marshal(conditions)
		if err == nil {
			result.Message = string(bytes)
		} else {
			log.Error(err, "could not marshal conditions")
		}
	}

	bytes, err := json.Marshal(result)
	if err != nil {
		return err
	}

	if err := m.rabbit.Send(bytes); err != nil {
		return err
	}

	return m.deleteML(src, job)
}

func (m *MLReconciler) deleteML(src *mlv1.ML, job *batch.Job) error {
	log := m.log.WithValues("name", src.Name, "namespace", src.Namespace)
	log.Info("starting to delete ML by deployer")

	output := ""

	podLog, err := m.getPodLogs(job)
	if err != nil {
		podLog = ""
		//return err
	}
	output += podLog

	conditions := job.Status.Conditions
	if conditions != nil {
		bytes, err := json.Marshal(conditions)
		if err != nil {
			return err
		}
		output += "\n\nXerac Conditions:\n" + string(bytes)
	}

	log.Info("starting to put string to s3")
	if err := storage.PutString(output, "pod-logs", strconv.Itoa(src.Spec.SubmissionID)); err != nil {
		return err
	}

	log.Info("starting to delete ML with deployer")
	return m.deployer.DeleteML(src)
}

func (m *MLReconciler) getPodLogs(job *batch.Job) (string, error) {
	podList, err := m.deployer.GetPodListWithJob(job)
	if err != nil {
		return "", err
	}
	log := ""

	for i, pod := range podList.Items {
		m.log.WithValues("podName", pod.Name, "podNamespace", pod.Namespace).Info("starting to handle result of ml")

		config, err := rest.InClusterConfig()
		if err != nil {
			return "", err
		}

		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			return "", err
		}

		podLogOpts := core.PodLogOptions{
			Container: "submission",
		}
		req := clientset.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &podLogOpts)

		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(env.APICallTimeout))
		defer cancel()

		podLogs, err := req.Stream(ctx)
		if err != nil {
			return "", err
		}
		defer podLogs.Close()

		buf := new(bytes.Buffer)
		_, err = io.Copy(buf, podLogs)
		if err != nil {
			return "", err
		}

		str := buf.String()
		if len(str) > 10000 {
			str = str[:10000] + "\n..."
		}

		log += fmt.Sprintf("\n\nlog of %d-th time of run: \n%s", i, str)
	}

	return log, nil
}

func (m *MLReconciler) jobManifest(src *mlv1.ML, job *batch.Job) error {
	job.Name = name.MLJobName(src.Spec.RunID)
	job.Namespace = src.Namespace
	job.Spec.BackoffLimit = &src.Spec.BackoffLimit

	job.Spec.Template.Spec.Volumes = []core.Volume{
		{
			Name: "data-volume",
			VolumeSource: core.VolumeSource{
				PersistentVolumeClaim: &core.PersistentVolumeClaimVolumeSource{
					ClaimName: src.Spec.DataPersistentVolumeClaimName,
					ReadOnly:  true,
				},
			},
		},
		{
			Name: "evaluation-volume",
			VolumeSource: core.VolumeSource{
				PersistentVolumeClaim: &core.PersistentVolumeClaimVolumeSource{
					ClaimName: src.Spec.EvaluationPersistentVolumeClaimName,
					ReadOnly:  true,
				},
			},
		},
		{
			Name: "result-volume",
			VolumeSource: core.VolumeSource{
				EmptyDir: &core.EmptyDirVolumeSource{},
			},
		},
	}
	job.Spec.Template.Spec.RestartPolicy = core.RestartPolicyNever

	return nil
}

func (m *MLReconciler) initContainerManifest(src *mlv1.ML, job *batch.Job) error {
	job.Spec.Template.Spec.InitContainers = []core.Container{
		{
			Name:  "submission",
			Image: src.Spec.SubmissionImage,
			Resources: core.ResourceRequirements{
				Limits: core.ResourceList{
					core.ResourceCPU:              resource.MustParse(src.Spec.CPUResourceLimit),
					core.ResourceMemory:           resource.MustParse(src.Spec.MemoryResourceLimit),
					core.ResourceEphemeralStorage: resource.MustParse(src.Spec.EphemeralResourceLimit),
				},
				Requests: core.ResourceList{
					core.ResourceCPU:              resource.MustParse(src.Spec.CPUResourceRequest),
					core.ResourceMemory:           resource.MustParse(src.Spec.MemoryResourceRequest),
					core.ResourceEphemeralStorage: resource.MustParse(src.Spec.EphemeralResourceRequest),
				},
			},
			VolumeMounts: []core.VolumeMount{
				{
					Name:      "data-volume",
					ReadOnly:  true,
					MountPath: "/data",
				},
				{
					Name:      "result-volume",
					MountPath: "/result",
				},
			},
		},
	}

	return nil
}

func (m *MLReconciler) evaluatorContainerManifest(src *mlv1.ML, job *batch.Job) error {
	job.Spec.Template.Spec.Containers = []core.Container{
		{
			Name:  "evaluator",
			Image: src.Spec.EvaluatorImage,
			Resources: core.ResourceRequirements{
				Limits: core.ResourceList{
					core.ResourceCPU:              resource.MustParse("100m"),
					core.ResourceMemory:           resource.MustParse("1G"),
					core.ResourceEphemeralStorage: resource.MustParse("10M"),
				},
				Requests: core.ResourceList{
					core.ResourceCPU:              resource.MustParse("50m"),
					core.ResourceMemory:           resource.MustParse("500M"),
					core.ResourceEphemeralStorage: resource.MustParse("5M"),
				},
			},
			VolumeMounts: []core.VolumeMount{
				{
					Name:      "evaluation-volume",
					ReadOnly:  true,
					MountPath: "/evaluation",
				},
				{
					Name:      "result-volume",
					MountPath: "/result",
				},
			},
			Env: []core.EnvVar{
				{
					Name:  "RUN_ID",
					Value: strconv.Itoa(src.Spec.RunID),
				},
				{
					Name:  "SUBMISSION_ID",
					Value: strconv.Itoa(src.Spec.SubmissionID),
				},
				{
					Name:  "QUEUE_SERVER_URL",
					Value: env.RabbitURL(),
				},
				{
					Name:  "QUEUE_NAME",
					Value: "result",
				},
			},
		},
	}
	return nil
}

func (m *MLReconciler) reconcileEvaluationPersistentVolumeClaim(src *mlv1.ML, job *batch.Job) error {
	scn := "manual"
	pvc := &core.PersistentVolumeClaim{
		ObjectMeta: meta.ObjectMeta{
			Name:      src.Name + "-evaluation-pvc",
			Namespace: src.Namespace,
		},
		Spec: core.PersistentVolumeClaimSpec{
			StorageClassName: &scn,
			AccessModes: []core.PersistentVolumeAccessMode{
				core.ReadOnlyMany,
			},
			Selector: &meta.LabelSelector{
				MatchLabels: map[string]string{
					"pv-tag": "evaluation-pv",
				},
			},
			Resources: core.ResourceRequirements{
				Requests: core.ResourceList{
					core.ResourceStorage: resource.MustParse("1G"),
				},
			},
		},
	}

	if _, err := m.deployer.SyncPVC(src, pvc); err != nil {
		return err
	}
	return nil
}

func (m *MLReconciler) reconcileDataPersistentVolumeClaim(src *mlv1.ML, job *batch.Job) error {
	scn := "manual"
	pvc := &core.PersistentVolumeClaim{
		ObjectMeta: meta.ObjectMeta{
			Name:      src.Name + "-data-pvc",
			Namespace: src.Namespace,
		},
		Spec: core.PersistentVolumeClaimSpec{
			StorageClassName: &scn,
			AccessModes: []core.PersistentVolumeAccessMode{
				core.ReadOnlyMany,
			},
			Selector: &meta.LabelSelector{
				MatchLabels: map[string]string{
					"pv-tag": "data-pv",
				},
			},
			Resources: core.ResourceRequirements{
				Requests: core.ResourceList{
					core.ResourceStorage: resource.MustParse("1G"),
				},
			},
		},
	}

	if _, err := m.deployer.SyncPVC(src, pvc); err != nil {
		return err
	}
	return nil
}

func (m *MLReconciler) SetupWithManager(mgr ctrl.Manager) error {
	c, err := controller.New("resource-ml", mgr, controller.Options{Reconciler: m})
	if err != nil {
		return err
	}

	if err = c.Watch(
		&source.Kind{Type: &batch.Job{}},
		&handler.EnqueueRequestForOwner{
			OwnerType: &mlv1.ML{},
		},
	); err != nil {
		return err
	}

	if err = c.Watch(&source.Kind{Type: &mlv1.ML{}}, &handler.EnqueueRequestsFromMapFunc{
		ToRequests: handler.ToRequestsFunc(func(mapObject handler.MapObject) []reconcile.Request {
			log := m.log.WithValues("from", "resource-controller")

			log.Info("starting to list mls")
			mls, err := m.deployer.ListMLs()
			if err != nil {
				log.Error(err, "could not list mls")
				return []reconcile.Request{}
			}

			var candidateML *mlv1.ML = nil
			for _, ml := range mls.Items {
				if ml.Status.StatusType != mlv1.MLStatusTypePending || ml.CreationTimestamp.IsZero() {
					continue
				}

				if candidateML == nil {
					candidateML = &ml
					continue
				}

				if candidateML.CreationTimestamp.After(ml.CreationTimestamp.Time) {
					candidateML = &ml
				}
			}

			if candidateML != nil {
				log.WithValues("candidate", candidateML.Name).Info("create reconcile.request")
				return []reconcile.Request{
					{NamespacedName: types.NamespacedName{Name: candidateML.Name, Namespace: candidateML.Namespace}},
				}
			}
			log.Info("nil candidate ml")
			return []reconcile.Request{}
		}),
	}, predicate.Funcs{
		DeleteFunc: func(e event.DeleteEvent) bool {
			log.Log.Info("delete func predicate")
			return true
		},
		CreateFunc: func(e event.CreateEvent) bool {
			log.Log.Info("create func predicate")
			return false
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			log.Log.Info("update func predicate")
			return false
		},
		GenericFunc: func(e event.GenericEvent) bool {
			log.Log.Info("generate func predicate")
			return false
		},
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).For(&mlv1.ML{}).Complete(m)
}
