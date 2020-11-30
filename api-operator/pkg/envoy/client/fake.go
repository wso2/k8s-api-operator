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

package client

import (
	"context"
	"fmt"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/controller/common"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/envoy/action"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/swagger"
	"math/rand"
	"time"
)

type Fake struct {
	ProjectMap     *action.ProjectsMap
	Response       Response
	responseMethod func(projects *action.ProjectsMap) Response
}

func NewFake(response Response) *Fake {
	return &Fake{
		responseMethod: func(projects *action.ProjectsMap) Response {
			return response
		},
	}
}

func NewFakeAllSucceeded() *Fake {
	return &Fake{
		responseMethod: func(projects *action.ProjectsMap) Response {
			r := Response{}

			for name, project := range *projects {
				switch project.Action {
				case action.ForceUpdate:
					r[name] = Updated
				case action.Delete:
					r[name] = Deleted
				}
			}
			return r
		},
	}
}

func NewFakeAllFailed() *Fake {
	return &Fake{
		responseMethod: func(projects *action.ProjectsMap) Response {
			r := Response{}

			for name := range *projects {
				r[name] = Failed
			}
			return r
		},
	}
}

func NewFakeWithRandomResponse() *Fake {
	return &Fake{
		responseMethod: func(projects *action.ProjectsMap) Response {
			r := Response{}
			rand.Seed(time.Now().UnixNano())

			for name, project := range *projects {
				if rand.Intn(2) == 0 {
					r[name] = Failed
				} else {
					switch project.Action {
					case action.ForceUpdate:
						r[name] = Updated
					case action.Delete:
						r[name] = Deleted
					}
				}
			}
			return r
		},
	}
}

func (c *Fake) Update(ctx context.Context, reqInfo *common.RequestInfo, projects *action.ProjectsMap) (Response, error) {
	for s, project := range *projects {
		fmt.Println("")
		fmt.Println("******* PRINT PROJECT ******")
		fmt.Printf("Project name: %s\n", s)
		fmt.Printf("Action: %s\n", project.Action)

		if project.Action != action.ForceUpdate {
			continue
		}
		err := project.OAS.Validate(ctx)
		fmt.Printf("Swagger validation: %v\n", err == nil)
		fmt.Printf("Tls certs: %v\n", project.TlsCertificate)
		fmt.Println(swagger.PrettyString(project.OAS))
	}

	c.ProjectMap = projects
	c.Response = c.responseMethod(projects)
	return c.Response, nil
}
