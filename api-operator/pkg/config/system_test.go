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
	"testing"
)

func TestSetSystemNamespaceFromEnv(t *testing.T) {

	var found bool

	found = SetSystemNamespaceFromEnv()
	if found == true {
		t.Error("expected false as the system namespace has not been set.")
	}

	os.Setenv(SystemNamespaceEnv, "wso2")
	found = SetSystemNamespaceFromEnv()
	if found == false {
		t.Error("expected true as the system namespace has been set.")
	}
}

func TestGetWatchNamespaces(t *testing.T) {

	var watchNamespaces string

	os.Setenv(SystemNamespaceEnv, DefaultSystemNamespace)
	SetSystemNamespaceFromEnv()
	watchNamespaces = GetWatchNamespaces()
	if watchNamespaces != OperatorNamespace {
		t.Error("expected operator namespace as the watch cluster level has not been set.")
	}

	os.Setenv(WatchClusterLevel, "true")
	watchNamespaces = GetWatchNamespaces()
	if watchNamespaces != "" {
		t.Error("expected empty string as the namespace as the watch cluster level has been set to true")
	}

	os.Setenv(WatchClusterLevel, "true")
	os.Setenv(k8sutil.WatchNamespaceEnvVar, "test123")
	watchNamespaces = GetWatchNamespaces()
	if watchNamespaces != "test123" {
		t.Error("expected default namespace as the watch cluster level has been set to true and watch namespace" +
			"is set to test123")
	}

	os.Setenv(WatchClusterLevel, "false")
	os.Setenv(SystemNamespaceEnv, "test-ns")
	SetSystemNamespaceFromEnv()
	defer func() { recover() }()
	GetWatchNamespaces()
	t.Errorf("expected panic as system namespace is different from operator namespace")

}
