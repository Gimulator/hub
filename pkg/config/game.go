package config

import (
	hubv1 "github.com/Gimulator/hub/api/v1"
	"github.com/Gimulator/hub/pkg/cache"
	"github.com/Gimulator/hub/pkg/name"
	"github.com/Gimulator/hub/pkg/s3"
)

func FetchGameConfig(room *hubv1.Room) error {
	if room.Spec.GameConfig != nil {
		return nil
	}

	if err := cache.GetStruct(name.CacheKeyForGame(room.Spec.Game), room.Spec.GameConfig); err == nil {
		return nil
	}

	if err := s3.GetStruct(name.S3GameConfigBucket(), room.Spec.Game, room.Spec.GameConfig); err != nil {
		return err
	}

	// is it OK to ignore error of cache system?
	_ = cache.SetStruct(name.CacheKeyForGame(room.Spec.Game), room.Spec.GameConfig.DeepCopy())

	return nil
}
