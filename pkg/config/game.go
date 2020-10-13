package config

import (
	hubv1 "github.com/Gimulator/hub/api/v1"
	"github.com/Gimulator/hub/pkg/cache"
	"github.com/Gimulator/hub/pkg/name"
	"github.com/Gimulator/hub/pkg/s3"
)

type GameConfig struct {
	DataPVCName      string `json:"dataPVCName"`
	FactPVCName      string `json:"factPVCName"`
	GimulatorImage   string `json:"gimulatorImage"`
	OutputVolumeSize string `json:"outputVolumeSize"`

	Namespace string
	RoomID    string
}

func FetchGameConfig(room *hubv1.Room) (GameConfig, error) {
	gameConfig := GameConfig{}
	if err := cache.GetStruct(name.CacheKeyForGame(room.Spec.Game), &gameConfig); err == nil {
		return gameConfig, nil
	}

	if err := s3.GetStruct(name.S3GameConfigBucket(), room.Spec.Game, gameConfig); err != nil {
		return GameConfig{}, err
	}

	// is it OK to ignore error of cache system?
	_ = cache.SetStruct(name.CacheKeyForGame(room.Spec.Game), gameConfig)

	gameConfig.Namespace = room.Namespace
	gameConfig.RoomID = room.Spec.ID

	return gameConfig, nil
}
