package cache

import (
	"fmt"
	"time"

	"github.com/getlantern/deepcopy"
	"github.com/patrickmn/go-cache"
)

var (
	CacheExpirationTime  = time.Hour * 24
	CacheCleanupInterval = time.Hour * 24
	c                    *cache.Cache
)

func init() {
	c = cache.New(CacheExpirationTime, CacheCleanupInterval)
}

func SetStruct(key string, value interface{}) {
	c.SetDefault(key, value)
}

func GetStruct(key string, value interface{}) error {
	data, exists := c.Get(key)
	if !exists {
		return fmt.Errorf("asked key not found")
	}

	if err := deepcopy.Copy(value, data); err != nil {
		return fmt.Errorf("can not convert asked data to asked structure")
	}

	return nil
}

func SetString(key string, value string) {
	c.SetDefault(key, value)
}

func GetString(key string) (string, error) {
	data, exists := c.Get(key)
	if !exists {
		return "", fmt.Errorf("asked key not found")
	}

	str, ok := data.(string)
	if !ok {
		return "", fmt.Errorf("asked key not found")
	}
	return str, nil
}
