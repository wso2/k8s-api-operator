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

import (
	"errors"
	"fmt"
	"reflect"
)

func OneKey(m interface{}) (string, error) {
	if reflect.TypeOf(m).Kind().String() != "map" {
		err := errors.New("type of the argument is not a map")
		return "", err
	}
	keys := reflect.ValueOf(m).MapKeys()

	if len(keys) != 1 {
		err := fmt.Errorf("length of the map should be 1 but was %v", len(keys))
		return "", err
	}

	return keys[0].String(), nil
}
