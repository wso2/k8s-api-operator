package maps

import "testing"

func TestOneKey(t *testing.T) {
	var key string
	var err error

	key, err = OneKey("string")
	if err == nil {
		t.Error("string argument should return an error")
	}

	key, err = OneKey(map[string]int{"one": 1})
	if err != nil {
		t.Error("map with one key should not return an error")
	}
	if key != "one" {
		t.Errorf("for map {\"one\": 1} want: 'key' but was: %s", key)
	}

	key, err = OneKey(map[string]int{"one": 1, "two": 2})
	if err == nil {
		t.Error("map with multiple keys should return error")
	}
}
