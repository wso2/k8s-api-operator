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

package str

import "testing"

func TestContainsString(t *testing.T) {

	var strValue bool

	args := []string{"a", "b"}
	strValue = ContainsString(args, "a")
	if strValue == false {
		t.Error("provides invalid value for given slice and expected true value")
	}

	strValue = ContainsString(args, "c")
	if strValue == true {
		t.Error("provides invalid value for given slice and expected false value")
	}
}

func TestRemoveString(t *testing.T) {

	var resultSet1 []string
	var resultSet2 []string

	args := []string{"a", "b"}
	resultSet1 = RemoveString(args, "a")

	if len(resultSet1) == 2 {
		t.Error("value has not been removed from the array")
	}

	resultSet2 = RemoveString(args, "a")
	resultSet2 = RemoveString(args, "c")

	if len(resultSet2) != 2 {
		t.Error("value has been removed from the array")
	}
}
