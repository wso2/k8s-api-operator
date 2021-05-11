// Copyright (c) WSO2 Inc. (http://www.wso2.org) All Rights Reserved.
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

package vol

import "testing"

func TestContextIsEquals(t *testing.T) {
	tests := []struct {
		name    string
		ctx     Context
		ctxStr  string
		isEqual bool
	}{
		{
			name:    "no_input_str",
			ctx:     DefaultContext,
			ctxStr:  "",
			isEqual: true,
		},
		{
			name:    "no_input_str_kaniko",
			ctx:     KanikoContext,
			ctxStr:  "",
			isEqual: false,
		},
		{
			name:    "input_str_default",
			ctx:     DefaultContext,
			ctxStr:  "default",
			isEqual: true,
		},
		{
			name:    "input_str_ignore_case",
			ctx:     DefaultContext,
			ctxStr:  "DeFault",
			isEqual: true,
		},

		{
			name:    "wrong_context",
			ctx:     KanikoContext,
			ctxStr:  "DeFault",
			isEqual: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			isEqual := test.ctx.isEqual(test.ctxStr)
			if isEqual != test.isEqual {
				t.Errorf("got %v, want %v", isEqual, test.isEqual)
			}
		})
	}
}
