package config

import (
	"errors"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var logMap = log.Log.WithName("config.map")

// GetMapKeyValue returns the key and value from the given map with length one
func GetMapKeyValue(m map[string]string) (string, string, error) {
	if len(m) == 0 {
		err := errors.New("length of the map is 0")
		logMap.Error(err, "Length of the given map should be 1")
		return "", "", err
	}
	if len(m) > 1 {
		err := errors.New("length of the map is more than 1")
		logMap.Error(err, "Length of the given map should be 1")
		return "", "", err
	}

	key := reflect.ValueOf(m).MapKeys()[0].String()
	return key, m[key], nil
}
