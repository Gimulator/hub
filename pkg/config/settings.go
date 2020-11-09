package config

import (
	hubv1 "github.com/Gimulator/hub/api/v1"
	"github.com/Gimulator/hub/pkg/cache"
	"github.com/Gimulator/hub/pkg/name"
	"github.com/Gimulator/hub/pkg/s3"
)

func FetchSetting(room *hubv1.Room) error {
	if room.Spec.Setting != nil {
		return nil
	}

	if err := cache.GetStruct(name.CacheKeyForSetting(room.Spec.ProblemID), room.Spec.Setting); err == nil {
		return nil
	}

	setting := &hubv1.Setting{}
	if err := s3.GetStruct(name.S3SettingBucket(), name.S3SettingObjectName(room.Spec.ProblemID), setting); err != nil {
		return err
	}
	room.Spec.Setting = setting

	cache.SetStruct(name.CacheKeyForSetting(room.Spec.ProblemID), room.Spec.Setting.DeepCopy())

	return nil
}
