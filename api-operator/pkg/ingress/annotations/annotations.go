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

package annotations

import (
	"github.com/wso2/k8s-api-operator/api-operator/pkg/ingress/annotations/apim"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/ingress/annotations/tls"
	"k8s.io/api/networking/v1beta1"
)

type Ingress struct {
	Tls  tls.Config
	APIM apim.Config
}

func ParseIngress(ing *v1beta1.Ingress) Ingress {
	return Ingress{Tls: tls.Parse(ing), APIM: apim.Parse(ing)}
}
