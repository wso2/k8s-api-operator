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

package config

import (
	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	"os"
	"strconv"
)

const DefaultSystemNamespace = "wso2-system"
const SystemNamespaceEnv = "SYSTEM_NAMESPACE"
const WatchClusterLevel = "WATCH_CLUSTER_LEVEL"

var (
	SystemNamespace   = DefaultSystemNamespace
	OperatorNamespace = DefaultSystemNamespace
)

func SetSystemNamespaceFromEnv() (found bool) {
	ns, found := os.LookupEnv(SystemNamespaceEnv)
	if !found {
		ns = DefaultSystemNamespace
	}
	SystemNamespace = ns
	return
}

func SetOperatorNamespace() {
	if ns, err := k8sutil.GetOperatorNamespace(); err == nil {
		OperatorNamespace = ns
	}
}

// Checks whether operator is used cluster wide or namespace wide.
func getWatchClusterLevel() (watchClusterLevel bool) {
	watchClusterLevelValue, _ := os.LookupEnv(WatchClusterLevel)
	watchClusterLevel, _ = strconv.ParseBool(watchClusterLevelValue)
	return watchClusterLevel
}

// If cluster wide is enabled, return empty list or comma separated namespaces. If cluster wide is disabled,
// used the operator deployed namespace.
func GetWatchNamespaces() (watchNamespaces string) {

	watchClusterLevel := getWatchClusterLevel()

	if watchClusterLevel {
		ns, _ := k8sutil.GetWatchNamespace()
		if ns != "" {
			watchNamespaces = ns
		} else {
			watchNamespaces = ""
		}

	} else {
		watchNamespaces = OperatorNamespace
	}

	return watchNamespaces
}
