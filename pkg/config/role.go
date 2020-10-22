package config

import (
	hubv1 "github.com/Gimulator/hub/api/v1"
	"github.com/Gimulator/hub/pkg/cache"
	"github.com/Gimulator/hub/pkg/name"
	"github.com/Gimulator/hub/pkg/s3"
)

func FetchRoles(room *hubv1.Room) (string, error) {
	str, err := cache.GetString(name.CacheKeyForProblemSettings(room.Spec.ProblemID))
	if err == nil {
		return str, nil
	}

	str, err = s3.GetString(name.S3ProblemSettingsBucket(), name.S3ProblemSettingsObjectName(room.Spec.ProblemID))
	if err != nil {
		return "", err
	}

	// is it OK to ignore error of cache system?
	cache.SetString(name.CacheKeyForProblemSettings(room.Spec.ProblemID), str)

	return str, nil
}
