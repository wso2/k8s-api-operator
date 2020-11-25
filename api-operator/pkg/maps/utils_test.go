// Copyright (c)  WSO2 Inc. (http://www.wso2.org) All Rights Reserved.
//
// WSO2 Inc. licenses this file to you under the Apache License,
// Version 2.0 (the "License"); you may not use this file except
// in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

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
