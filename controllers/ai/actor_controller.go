package ai

import (
	"fmt"
	"path/filepath"

	"github.com/go-logr/logr"
	aiv1 "gitlab.com/Syfract/Xerac/hub/apis/ai/v1"
	env "gitlab.com/Syfract/Xerac/hub/utils/environment"
	"gitlab.com/Syfract/Xerac/hub/utils/name"
)

type ActorReconciler struct {
	log logr.Logger
}

func NewActorReconciler(log logr.Logger) (*ActorReconciler, error) {
	return &ActorReconciler{
		log: log,
	}, nil
}

func (r *ActorReconciler) Reconcile(src, dst *aiv1.Room) error {
	return r.reconcileActors(src, dst)
}

func (r *ActorReconciler) reconcileActors(src, dst *aiv1.Room) error {
	if err := r.reconcileGimulatorActor(src, dst); err != nil {
		return err
	}

	if err := r.reconcileLoggerActor(src, dst); err != nil {
		return err
	}

	if err := r.reconcileArgs(src, dst); err != nil {
		return err
	}

	return nil
}

func (r *ActorReconciler) reconcileGimulatorActor(src, dst *aiv1.Room) error {
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

func (r *ActorReconciler) reconcileLoggerActor(src, dst *aiv1.Room) error {
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
		EnvVars:      make([]aiv1.EnvVar, 0),
		VolumeMounts: make([]aiv1.VolumeMount, 0),
	})
	return nil
}

func (r *ActorReconciler) reconcileArgs(src, dst *aiv1.Room) error {
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
