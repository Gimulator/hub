package cache

import (
	"fmt"

	"github.com/getlantern/deepcopy"
	"github.com/patrickmn/go-cache"
	env "gitlab.com/Syfract/Xerac/hub/utils/environment"
)

var c *cache.Cache

func init() {
	c = cache.New(env.CacheExpirationTime, env.CacheCleanupInterval)
}

func GetYamlString(key string) (string, error) {
	data, exists := c.Get(key)
	if !exists {
		return "", fmt.Errorf("cache: asked string not found")
	}

	str, ok := data.(string)
	if !ok {
		return "", fmt.Errorf("cache: can not convert data to string")
	}

	return str, nil
}

func GetStruct(key string, value interface{}) error {
	data, exists := c.Get(key)
	if !exists {
		return fmt.Errorf("cache: asked structure not found")
	}

	if err := deepcopy.Copy(value, data); err != nil {
		return fmt.Errorf("cache: can not convert data to asked structure")
	}

	return nil
}

func SetYamlString(key string, value string) {
	c.Set(key, value, env.CacheExpirationTime)
}

func SetStruct(key string, value interface{}) {
	c.Set(key, value, env.CacheExpirationTime)
}
