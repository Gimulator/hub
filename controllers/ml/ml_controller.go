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
	"encoding/json"
	"strconv"
	"time"

	"github.com/go-logr/logr"
	batch "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	mlv1 "github.com/Gimulator/hub/apis/ml/v1"
	"github.com/Gimulator/hub/utils/deployer"
	"github.com/Gimulator/hub/utils/name"
	rabbit "github.com/Gimulator/hub/utils/rabbitMQ"
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

	job := &batch.Job{}

	//if err := m.reconcileDataPersistentVolumeClaim(src, job); err != nil {
	//	return ctrl.Result{}, err
	//}

	//if err := m.reconcileEvaluationPersistentVolumeClaim(src, job); err != nil {
	//	return ctrl.Result{}, err
	//}

	if err := m.jobManifest(src, job); err != nil {
		return ctrl.Result{}, err
	}

	if err := m.initContainerManifest(src, job); err != nil {
		return ctrl.Result{}, err
	}

	if err := m.evaluatorContainerManifest(src, job); err != nil {
		return ctrl.Result{}, err
	}

	syncedJob, err := m.deployer.SyncJob(src, job)
	if err != nil {
		return ctrl.Result{}, err
	}

	if err := m.reconcileSyncedJob(src, syncedJob); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (m *MLReconciler) reconcileSyncedJob(src *mlv1.ML, job *batch.Job) error {
	if job.Status.Active > 0 {
		return nil
	}

	if job.Status.Succeeded > 0 {
		return m.reconcileFailedML(src, job.Status.Conditions)
	}

	if job.Status.Conditions != nil && len(job.Status.Conditions) > 0 {
		con := job.Status.Conditions[0]
		if con.Type == batch.JobComplete {
			m.deployer.DeleteML(src)
		}
		if con.Type == batch.JobFailed {
			return m.reconcileFailedML(src, job.Status.Conditions)
		}
	}

	creationTime := job.CreationTimestamp
	if creationTime.IsZero() {
		return nil
	}

	diff := time.Now().Sub(creationTime.Time)
	if diff > time.Minute*20 {
		m.deployer.DeleteML(src)
	}

	return nil
}

func (m *MLReconciler) reconcileFailedML(src *mlv1.ML, conditions []batch.JobCondition) error {
	result := struct {
		RoomID  int    `json:"run_id"`
		Status  string `json:"status"`
		Message string `json:"message"`
	}{
		RoomID:  src.Spec.ID,
		Status:  "FAIL",
		Message: "could not find error",
	}

	if conditions != nil {
		bytes, err := json.Marshal(conditions)
		if err == nil {
			result.Message = string(bytes)
		}
	}

	bytes, err := json.Marshal(result)
	if err != nil {
		return err
	}

	if err := m.rabbit.Send(bytes); err != nil {
		return err
	}

	return m.deployer.DeleteML(src)
}

func (m *MLReconciler) jobManifest(src *mlv1.ML, job *batch.Job) error {
	job.Name = name.MLJobName(src.Spec.ID)
	job.Namespace = src.Namespace
	job.Spec.BackoffLimit = &src.Spec.BackoffLimit

	job.Spec.Template.Spec.Volumes = []core.Volume{
		{
			Name: "data-volume",
			VolumeSource: core.VolumeSource{
				PersistentVolumeClaim: &core.PersistentVolumeClaimVolumeSource{
					ClaimName: "data-persist-volume-claim",
					ReadOnly:  true,
				},
			},
		},
		{
			Name: "evaluation-volume",
			VolumeSource: core.VolumeSource{
				PersistentVolumeClaim: &core.PersistentVolumeClaimVolumeSource{
					ClaimName: "evaluation-persist-volume-claim",
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
					Name:  "ID",
					Value: strconv.Itoa(src.Spec.ID),
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
	return ctrl.NewControllerManagedBy(mgr).For(&mlv1.ML{}).Complete(m)
}
