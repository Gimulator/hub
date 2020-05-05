package ai

import (
	"fmt"

	"github.com/Gimulator/Gimulator/auth"
	"github.com/go-logr/logr"
	uuid "github.com/satori/go.uuid"
	aiv1 "gitlab.com/Syfract/Xerac/hub/apis/ai/v1"
	env "gitlab.com/Syfract/Xerac/hub/utils/environment"
	"gitlab.com/Syfract/Xerac/hub/utils/name"
	"gopkg.in/yaml.v3"
)

type SketchReconciler struct {
	log logr.Logger
}

func NewSketchReconciler(log logr.Logger) (*SketchReconciler, error) {
	return &SketchReconciler{
		log: log,
	}, nil
}

func (r *SketchReconciler) Reconcile(src, dst *aiv1.Room) error {
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

func (r *SketchReconciler) reconcilePrimitiveSketch(src, dst *aiv1.Room) (*auth.Config, error) {
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

func (r *SketchReconciler) reconcileFinalSketch(src, dst *aiv1.Room, sketch *auth.Config) error {
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
