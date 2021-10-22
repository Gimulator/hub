package config

import (
	"context"

	hubv1 "github.com/Gimulator/hub/api/v1"
	"github.com/Gimulator/hub/pkg/cache"
	"github.com/Gimulator/hub/pkg/name"
	"github.com/Gimulator/hub/pkg/s3"
)

func FetchRules(ctx context.Context, room *hubv1.Room) (string, error) {
	str, err := cache.GetString(name.CacheKeyForRules(room.Spec.ProblemID))
	if err == nil {
		return str, nil
	}

	str, err = s3.GetString(ctx, name.S3RulesBucket(), name.S3RulesObjectName(room.Spec.ProblemID))
	if err != nil {
		return "", err
	}

	// is it OK to ignore error of cache system?
	cache.SetString(name.CacheKeyForRules(room.Spec.ProblemID), str)

	return str, nil
}
