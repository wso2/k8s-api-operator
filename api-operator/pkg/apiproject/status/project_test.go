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

package status

import (
	"github.com/wso2/k8s-api-operator/api-operator/pkg/apiproject"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/apiproject/client"
	"reflect"
	"testing"
)

func TestUpdatedProjects(t *testing.T) {
	var tests = []struct {
		name      string
		st, newSt *ProjectsStatus
		want      apiproject.ProjectSet
	}{
		// Same ingresses
		{
			name:  "No hosts added or deleted",
			st:    &ProjectsStatus{"foo/ing1": map[string]string{"a_com": "_", "b_com": "_"}},
			newSt: &ProjectsStatus{"foo/ing1": map[string]string{"a_com": "_", "b_com": "_"}},
			want:  apiproject.ProjectSet{"a_com": true, "b_com": true},
		},
		{
			name:  "Add and delete hosts",
			st:    &ProjectsStatus{"foo/ing1": map[string]string{"a_com": "_", "b_com": "_"}},
			newSt: &ProjectsStatus{"foo/ing1": map[string]string{"a_com": "_", "c_com": "_"}},
			want:  apiproject.ProjectSet{"a_com": true, "b_com": true, "c_com": true},
		},
		// Different ingresses
		{
			name:  "Add new ingress",
			st:    &ProjectsStatus{"foo/ing1": map[string]string{"a_com": "_", "b_com": "_"}},
			newSt: &ProjectsStatus{"foo/ing2": map[string]string{"b_com": "_", "c_com": "_"}},
			want:  apiproject.ProjectSet{"b_com": true, "c_com": true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := tt.st.UpdatedProjects(tt.newSt)
			if !reflect.DeepEqual(p, tt.want) {
				t.Errorf("%v.UpdatedProjects(%v) = %v; want %v", tt.st, tt.newSt, p, tt.want)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	var tests = []struct {
		name       string
		st, newSt  *ProjectsStatus
		gwResponse client.Response
		want       *ProjectsStatus
	}{
		{
			name:       "Successful deletion & update",
			st:         &ProjectsStatus{"foo/ing1": map[string]string{"a_com": "_", "b_com": "_"}},
			newSt:      &ProjectsStatus{"foo/ing1": map[string]string{"a_com": "_"}},
			gwResponse: client.Response{"a_com": client.Updated, "b_com": client.Deleted},
			want:       &ProjectsStatus{"foo/ing1": map[string]string{"a_com": "_"}},
		},
		{
			name:       "Failed deletion & update",
			st:         &ProjectsStatus{"foo/ing1": map[string]string{"a_com": "_", "b_com": "_"}},
			newSt:      &ProjectsStatus{"foo/ing1": map[string]string{"a_com": "_"}},
			gwResponse: client.Response{"a_com": client.Failed, "b_com": client.Failed},
			want:       &ProjectsStatus{"foo/ing1": map[string]string{"a_com": "_", "b_com": "_"}},
		},
		{
			name:       "Failed update",
			st:         &ProjectsStatus{"foo/ing1": map[string]string{"a_com": "_"}},
			newSt:      &ProjectsStatus{"foo/ing1": map[string]string{"b_com": "_"}},
			gwResponse: client.Response{"a_com": client.Deleted, "b_com": client.Failed},
			want:       &ProjectsStatus{},
		},
		{
			name: "Mixed operations",
			st: &ProjectsStatus{
				"foo/ing1": map[string]string{"a_com": "_", "b_com": "_"},
				"foo/ing2": map[string]string{"a_com": "_", "c_com": "_", "d_com": "_", "e_com": "_"},
				"foo/ing3": map[string]string{"e_com": "_"},
			},
			newSt: &ProjectsStatus{
				"foo/ing1": map[string]string{"a_com": "_", "c_com": "_"},
				"foo/ing2": map[string]string{"a_com": "_", "c_com": "_", "f_com": "_"},
				"foo/ing4": map[string]string{"f_com": "_"},
				"foo/ing5": map[string]string{"a_com": "_"},
			},
			gwResponse: client.Response{
				"a_com": client.Updated,
				"b_com": client.Deleted,
				"c_com": client.Failed,
				"d_com": client.Failed,
				"e_com": client.Deleted,
				"f_com": client.Failed,
			},
			want: &ProjectsStatus{
				"foo/ing1": map[string]string{"a_com": "_"},
				"foo/ing2": map[string]string{"a_com": "_", "c_com": "_", "d_com": "_"},
				"foo/ing5": map[string]string{"a_com": "_"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.st.Update(tt.newSt, tt.gwResponse)
			if !reflect.DeepEqual(tt.st, tt.want) {
				t.Errorf("updated state: %v; want %v", tt.st, tt.want)
			}
		})
	}
}
