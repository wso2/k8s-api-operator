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
	"fmt"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/ingress"
	"strings"
)

// DefaultBackendProject represents the project of default backend of ingress
// project for a host in an ingress rule can not have three "_" consecutively.
// So this name is not conflict with a project for a host in an ingress rule.
const DefaultBackendProject = "ingress-___default"

func HostToProject(host string) string {
	p := strings.ReplaceAll(host, "*.", "__")
	return fmt.Sprintf("ingress-%v", strings.ReplaceAll(p, ".", "_"))
}

func ProjectToHost(pj string) string {
	p := strings.TrimPrefix(strings.ReplaceAll(pj, "__", "*."), "ingress-")
	return strings.ReplaceAll(p, "_", ".")
}

func IngressToName(ing *ingress.Ingress) string {
	return fmt.Sprintf("%v/%v", ing.Namespace, ing.Name)
}
