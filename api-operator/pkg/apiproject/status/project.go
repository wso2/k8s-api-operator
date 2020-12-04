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
	"context"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/apiproject"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/apiproject/client"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/config"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/controller/common"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

const ingressProjectStatusCm = "ingress-project-status"
const ingressProjectStatusKey = "project-status"

// ProjectsStatus represents a list of Open API Spec projects updated in the microgateway.
// Maps ingress -> projects
//
// default/ing1:
//   example1_com: _
//   example2_com: _
// default/ing2:
//   example2_com: _
//   example3_com: _
//
// Require a go routine mutex if handled by multiple go routines.
// Since, only one go routine updates ingresses this is not required.
type ProjectsStatus map[string]map[string]string

// UpdateToConfigMap updates current project status in the ingress configmap
func (s *ProjectsStatus) UpdateToConfigMap(ctx context.Context, reqInfo *common.RequestInfo) error {
	// Marshal yaml
	bytes, err := yaml.Marshal(s)
	if err != nil {
		return err
	}

	// Check ingress-configs from configmap
	ingresCm := &v1.ConfigMap{}
	if err := reqInfo.Client.Get(ctx, types.NamespacedName{
		Namespace: config.SystemNamespace, Name: ingressProjectStatusCm,
	}, ingresCm); err != nil {
		if errors.IsNotFound(err) {
			// ConfigMap not found and create it.
			ingresCm.Namespace = config.SystemNamespace
			ingresCm.Name = ingressProjectStatusCm
			ingresCm.Data = map[string]string{ingressProjectStatusKey: string(bytes)}
			err = reqInfo.Client.Create(ctx, ingresCm)
			return err
		}
		return err
	}

	// Update ConfigMap
	if ingresCm.Data == nil {
		ingresCm.Data = map[string]string{ingressProjectStatusKey: string(bytes)}
	} else {
		ingresCm.Data[ingressProjectStatusKey] = string(bytes)
	}

	if err := reqInfo.Client.Update(ctx, ingresCm); err != nil {
		return err
	}
	return nil
}

// UpdatedProjects returns project names that needed to be updated with the new state
func (s *ProjectsStatus) UpdatedProjects(sDiff *ProjectsStatus) apiproject.ProjectSet {
	// returns all projects in both s and sDiff
	projects := apiproject.ProjectSet{}
	for ing, ps := range *sDiff {
		// projects from new state: sDiff
		for p := range ps {
			projects[p] = true
		}

		// projects from current state: s
		for p := range (*s)[ing] {
			projects[p] = true
		}
	}

	return projects
}

// RemovedProjects returns removed project names from current state with given new state
func (s *ProjectsStatus) RemovedProjects(sDiff *ProjectsStatus) []string {
	// returns what presents in s but not in sDiff
	var projects []string
	for ing, ps := range *s {
		if _, ok := (*sDiff)[ing]; ok {
			for p := range ps {
				if _, ok := (*sDiff)[ing][p]; !ok {
					projects = append(projects, p)
				}
			}
		}
	}
	return projects
}

// Update updates the gateway current state according to the new state changes and response of gateway for changes
func (s *ProjectsStatus) Update(sDiff *ProjectsStatus, gatewayResponse client.Response) {
	// allIngresses are the all updated or deleted ingresses
	allIngresses := make([]string, 0, len(*s)+len(*sDiff))
	for ing := range *s {
		allIngresses = append(allIngresses, ing)
	}
	for ing := range *sDiff {
		allIngresses = append(allIngresses, ing)
	}

	for resProject, resType := range gatewayResponse {
		switch resType {
		case client.Failed:
			// project is failed to update or delete
			// ignore it
			continue
		case client.Deleted:
			for ing := range *s {
				s.removeProject(ing, resProject)
			}
		case client.Updated:
			for _, ing := range allIngresses {
				sProjectFound := s.ContainsProject(ing, resProject)
				_, sDiffIngFound := (*sDiff)[ing]
				sDiffProjectFound := sDiff.ContainsProject(ing, resProject)

				// project should be deleted from the current state if
				// current state contains it and diff state contains ingress but do not contains project
				if sProjectFound && sDiffIngFound && !sDiffProjectFound {
					s.removeProject(ing, resProject)
				}

				// project should be added to the current state if
				// current state do not contains it and diff state contains it
				if !sProjectFound && sDiffProjectFound {
					s.addProject(ing, resProject)
				}

				// otherwise do not need to update current state
			}
		}
	}
}

// ProjectSet returns existing projects in the status configmap
func (s *ProjectsStatus) ProjectSet() apiproject.ProjectSet {
	projects := apiproject.ProjectSet{}
	for _, ps := range *s {
		for p := range ps {
			projects[p] = true
		}
	}

	return projects
}

// ContainsProject checks the given ingress is exists in the given project
func (s *ProjectsStatus) ContainsProject(ing, project string) bool {
	if _, ok := (*s)[ing]; ok {
		_, found := (*s)[ing][project]
		return found
	}
	return false
}

func (s *ProjectsStatus) removeProject(ing, project string) {
	delete((*s)[ing], project)
	// if there are no any project, delete the ingress from current state
	if len((*s)[ing]) == 0 {
		delete(*s, ing)
	}
}

func (s *ProjectsStatus) addProject(ing, project string) {
	if _, ok := (*s)[ing]; ok {
		(*s)[ing][project] = "_"
	} else {
		(*s)[ing] = map[string]string{project: "_"}
	}
}
