package config

import (
	hubv1 "github.com/Gimulator/hub/api/v1"
	"github.com/Gimulator/hub/pkg/cache"
	"github.com/Gimulator/hub/pkg/name"
	"github.com/Gimulator/hub/pkg/s3"
)

func FetchProblemSettings(room *hubv1.Room) error {
	if room.Spec.ProblemSettings != nil {
		return nil
	}

	if err := cache.GetStruct(name.CacheKeyForProblemSettings(room.Spec.ProblemID), room.Spec.ProblemSettings); err == nil {
		return nil
	}

	if err := s3.GetStruct(name.S3ProblemSettingsBucket(), name.S3ProblemSettingsObjectName(room.Spec.ProblemID), room.Spec.ProblemSettings); err != nil {
		return err
	}
	cache.SetStruct(name.CacheKeyForProblemSettings(room.Spec.ProblemID), room.Spec.ProblemSettings.DeepCopy())

	return nil
}
