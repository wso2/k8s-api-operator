package maps

import (
	"errors"
	"reflect"
)

func OneKey(m interface{}) (string, error) {
	if reflect.TypeOf(m).Kind().String() != "map" {
		err := errors.New("type of the argument is not a map")
		return "", err
	}
	keys := reflect.ValueOf(m).MapKeys()

	if len(keys) != 1 {
		err := errors.New("length of the map should be 1 but was " + string(len(keys)))
		return "", err
	}

	return keys[0].String(), nil
}
