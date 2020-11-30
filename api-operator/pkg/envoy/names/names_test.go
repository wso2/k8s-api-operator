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

package names

import (
	"testing"
)

func TestHostToProject(t *testing.T) {
	var tests = []struct {
		host string
		want string
	}{
		{
			host: "foo.example.com",
			want: "ingress-foo_example_com",
		},
		{
			host: "*.example.com",
			want: "ingress-__example_com",
		},
		{
			host: "*.org",
			want: "ingress-__org",
		},
		// host = "*" is not valid since IngressRules.Host can not be "*"
		// host = "*.*.com", host = "foo.*.com" are invalid
		// regex used for validation: '\*\.[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*'
	}

	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			p := HostToProject(tt.host)
			if p != tt.want {
				t.Errorf("HostToProject(%v) = %v; want %v", tt.host, p, tt.want)
			}
		})
	}
}

func TestProjectToHost(t *testing.T) {
	var tests = []struct {
		project string
		want    string
	}{
		{
			project: "ingress-foo_example_com",
			want:    "foo.example.com",
		},
		{
			project: "ingress-__example_com",
			want:    "*.example.com",
		},
		{
			project: "ingress-__org",
			want:    "*.org",
		},
	}

	for _, tt := range tests {
		t.Run(tt.project, func(t *testing.T) {
			p := ProjectToHost(tt.project)
			if p != tt.want {
				t.Errorf("ProjectToHost(%v) = %v; want %v", tt.project, p, tt.want)
			}
		})
	}
}
