package ai

import (
	"fmt"
	"path/filepath"

	"github.com/getlantern/deepcopy"
	"github.com/go-logr/logr"
	aiv1 "gitlab.com/Syfract/Xerac/hub/apis/ai/v1"
	env "gitlab.com/Syfract/Xerac/hub/utils/environment"
	"gitlab.com/Syfract/Xerac/hub/utils/name"
	batch "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	res "k8s.io/apimachinery/pkg/api/resource"
)

type ActorReconciler struct {
	Log logr.Logger
}

func NewActorReconciler(log logr.Logger) (*ActorReconciler, error) {
	return &ActorReconciler{
		Log: log,
	}, nil
}

func (r *ActorReconciler) Reconcile(room aiv1.Room, job *batch.Job) error {
	if job.Spec.Template.Spec.Containers == nil {
		job.Spec.Template.Spec.Containers = make([]core.Container, 0)
	}

	var err error = nil
	if job.Spec.Template.Spec.Containers, err = r.reconcileActors(room.Spec.Actors, job.Spec.Template.Spec.Containers); err != nil {
		return err
	}

	return nil
}

func (r *ActorReconciler) reconcileActors(actors []aiv1.Actor, containers []core.Container) ([]core.Container, error) {
	actorList := []aiv1.Actor{}
	err := deepcopy.Copy(&actorList, actors)
	if err != nil {
		r.Log.Error(err, "can not copy actors")
		return containers, err
	}

	gimActor, err := r.reconcileGimulatorActor()
	if err != nil {
		return containers, err
	}
	actorList = append(actorList, gimActor)

	resActor, err := r.reconcileResultActor()
	if err != nil {
		return containers, err
	}
	actorList = append(actorList, resActor)

	logActor, err := r.reconcileLoggerActor()
	if err != nil {
		return containers, err
	}
	actorList = append(actorList, logActor)

	r.Log.Info("start to reconcile actors")
	for _, actor := range actorList {
		container := core.Container{}
		err := r.reconcileActor(actor, &container)
		if err != nil {
			r.Log.Error(err, "cannot reconcile actor")
			return containers, err
		}
		containers = append(containers, container)
	}

	if err = r.reconcileArgs(actors, containers); err != nil {
		return containers, err
	}

	return containers, nil
}

func (r *ActorReconciler) reconcileActor(actor aiv1.Actor, container *core.Container) error {
	var err error = nil

	container.VolumeMounts = make([]core.VolumeMount, 0)
	if container.VolumeMounts, err = r.reconcileVolumeMounts(actor.VolumeMounts, container.VolumeMounts); err != nil {
		return err
	}

	container.Env = make([]core.EnvVar, 0)
	if container.Env, err = r.reconcileEnvVars(actor.EnvVars, container.Env); err != nil {
		return err
	}

	container.Resources = core.ResourceRequirements{}
	if err := r.reconcileResources(actor.Resources, &container.Resources); err != nil {
		return err
	}

	container.Command = []string{"/bin/bash", "-c"}
	container.Name = name.NameDashID(actor.Name, actor.ID)
	container.Image = actor.Image

	return nil
}

func (r *ActorReconciler) reconcileVolumeMounts(aVMs []aiv1.VolumeMount, cVMs []core.VolumeMount) ([]core.VolumeMount, error) {
	for _, avm := range aVMs {
		cvm := core.VolumeMount{}
		if err := r.reconcileVolumeMount(avm, &cvm); err != nil {
			return cVMs, err
		}
		cVMs = append(cVMs, cvm)
	}

	if cVMs, err := r.reconcileSharedVolumeMount(cVMs); err != nil {
		return cVMs, err
	}
	return cVMs, nil
}

func (r *ActorReconciler) reconcileVolumeMount(aVM aiv1.VolumeMount, cVM *core.VolumeMount) error {
	cVM = &core.VolumeMount{
		Name:      aVM.Name,
		MountPath: aVM.Path,
	}
	return nil
}

func (r *ActorReconciler) reconcileSharedVolumeMount(cVMs []core.VolumeMount) ([]core.VolumeMount, error) {
	cVMs = append(cVMs, core.VolumeMount{
		Name:      env.SharedVolumeName(),
		MountPath: env.SharedVolumePath(),
	})
	return cVMs, nil
}

func (r *ActorReconciler) reconcileEnvVars(aEvnVars []aiv1.EnvVar, cEnvVars []core.EnvVar) ([]core.EnvVar, error) {
	for _, aEnvVar := range aEvnVars {
		cEnvVar := core.EnvVar{}
		if err := r.reconcileEnvVar(aEnvVar, &cEnvVar); err != nil {
			return cEnvVars, err
		}
		cEnvVars = append(cEnvVars, cEnvVar)
	}
	return cEnvVars, nil
}

func (r *ActorReconciler) reconcileEnvVar(aEnvVar aiv1.EnvVar, cEnvVar *core.EnvVar) error {
	if aEnvVar.Key == "" {
		return fmt.Errorf("nil key for EvnVar")
	}

	*cEnvVar = core.EnvVar{
		Name:  aEnvVar.Key,
		Value: aEnvVar.Value,
	}
	return nil
}

func (r *ActorReconciler) reconcileResources(aRs aiv1.Resources, cRs *core.ResourceRequirements) error {
	if err := r.reconcileResource(aRs.Limits, &cRs.Limits); err != nil {
		return err
	}

	if err := r.reconcileResource(aRs.Requests, &cRs.Requests); err != nil {
		return err
	}

	return nil
}

func (r *ActorReconciler) reconcileResource(aR aiv1.Resource, cR *core.ResourceList) error {
	tmp := core.ResourceList{}

	if aR.CPU != "" {
		tmp[core.ResourceCPU] = res.Quantity{
			Format: res.Format(aR.CPU),
		}
	}

	if aR.Memory != "" {
		tmp[core.ResourceMemory] = res.Quantity{
			Format: res.Format(aR.Memory),
		}
	}

	if aR.Ephemeral != "" {
		tmp[core.ResourceEphemeralStorage] = res.Quantity{
			Format: res.Format(aR.Ephemeral),
		}
	}

	*cR = tmp
	return nil
}

func (r *ActorReconciler) reconcileGimulatorActor() (aiv1.Actor, error) {
	return aiv1.Actor{
		Name:  env.GimulatorName(),
		Image: env.GimulatorImage(),
		Type:  aiv1.AIActorType(env.GimulatorType()),
		Role:  "",
		Resources: aiv1.Resources{
			Limits: aiv1.Resource{
				CPU:    "1000m",
				Memory: "1G",
			},
			Requests: aiv1.Resource{
				CPU:    "500m",
				Memory: "500M",
			},
		},
		EnvVars:      make([]aiv1.EnvVar, 0),
		VolumeMounts: make([]aiv1.VolumeMount, 0),
	}, nil
}

func (r *ActorReconciler) reconcileResultActor() (aiv1.Actor, error) {
	return aiv1.Actor{
		Name:  env.ResultName(),
		Image: env.ResultImage(),
		Type:  aiv1.AIActorType(env.ResultType()),
		Role:  "",
		Resources: aiv1.Resources{
			Limits: aiv1.Resource{
				CPU:    "1000m",
				Memory: "1G",
			},
			Requests: aiv1.Resource{
				CPU:    "500m",
				Memory: "500M",
			},
		},
		EnvVars:      make([]aiv1.EnvVar, 0),
		VolumeMounts: make([]aiv1.VolumeMount, 0),
	}, nil
}

func (r *ActorReconciler) reconcileLoggerActor() (aiv1.Actor, error) {
	return aiv1.Actor{
		Name:  env.LoggerName(),
		Image: env.LoggerImage(),
		Type:  aiv1.AIActorType(env.LoggerType()),
		Role:  "",
		Resources: aiv1.Resources{
			Limits: aiv1.Resource{
				CPU:    "1000m",
				Memory: "1G",
			},
			Requests: aiv1.Resource{
				CPU:    "500m",
				Memory: "500M",
			},
		},
		EnvVars:      make([]aiv1.EnvVar, 0),
		VolumeMounts: make([]aiv1.VolumeMount, 0),
	}, nil
}

func (r *ActorReconciler) reconcileArgs(actors []aiv1.Actor, containers []core.Container) error {
	dir := env.SharedVolumePath()
	fins := make([]string, 0)
	for _, actor := range actors {
		if actor.Type == aiv1.AIActorTypeFinisher {
			fins = append(fins, name.TerminatedFile(actor.Name))
		}
	}

	for _, actor := range actors {
		var args = ""
		switch actor.Type {
		case aiv1.AIActorTypeFinisher:
			args = newFinisherArgs(actor.Name, dir, "cmd")
		case aiv1.AIActorTypeMaster:
			args = newMasterArgs(dir, "cmd", fins)
		case aiv1.AIActorTypeSlave:
			args = newSlaveArgs(dir, "cmd", fins)
		default:
			return fmt.Errorf("invalid actor type")
		}

		for i := range containers {
			if containers[i].Name != name.NameDashID(actor.Name, actor.ID) {
				continue
			}

			containers[i].Args = []string{args}
			break
		}
	}
	return nil
}

func newFinisherArgs(n, dir, cmd string) string {
	path := filepath.Join(dir, name.TerminatedFile(n))
	return fmt.Sprintf(env.FinisherArgs, path, cmd)
}

func newMasterArgs(dir, cmd string, fins []string) string {
	condition := ""
	for i, f := range fins {
		path := filepath.Join(dir, f)
		condition += "-f " + path
		if i < len(fins)-1 {
			condition += " && "
		}
	}
	return fmt.Sprintf(env.MasterArgs, cmd, condition, condition)
}

func newSlaveArgs(dir, cmd string, fins []string) string {
	condition := ""
	for i, f := range fins {
		path := filepath.Join(dir, f)
		condition += "-f " + path
		if i < len(fins)-1 {
			condition += " && "
		}
	}
	return fmt.Sprintf(env.SlaveArgs, cmd, condition)
}
