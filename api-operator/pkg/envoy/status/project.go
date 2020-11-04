package status

import (
	"github.com/wso2/k8s-api-operator/api-operator/pkg/controller/common"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/envoy/controller"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

// TODO: (renuka) operatorNamespace represents the namespace of the operator
const operatorNamespace = "wso2-system"
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

func (s *ProjectsStatus) UpdateToConfigMap(reqInfo *common.RequestInfo) error {
	// Marshal yaml
	bytes, err := yaml.Marshal(s)
	if err != nil {
		return err
	}

	// Check ingress-configs from configmap
	ingresCm := &v1.ConfigMap{}
	if err := (*reqInfo.Client).Get(reqInfo.Ctx, types.NamespacedName{
		Namespace: operatorNamespace, Name: ingressProjectStatusCm,
	}, ingresCm); err != nil {
		if !errors.IsNotFound(err) {
			// ConfigMap not found and create it.
			ingresCm.BinaryData = map[string][]byte{ingressProjectStatusKey: bytes}
			if err := (*reqInfo.Client).Create(reqInfo.Ctx, ingresCm); err != nil {
				return err
			}
		}
		return err
	}

	// Update ConfigMap
	// TODO set to Data
	d := ingresCm.BinaryData
	if d == nil {
		d = map[string][]byte{ingressProjectStatusKey: bytes}
	} else {
		d[ingressProjectStatusKey] = bytes
	}
	ingresCm.BinaryData = d
	if err := (*reqInfo.Client).Update(reqInfo.Ctx, ingresCm); err != nil {
		return err
	}
	return nil
}

// UpdatedProjects returns project names that needed to be updated with the new state
func (s *ProjectsStatus) UpdatedProjects(newS *ProjectsStatus) map[string]bool {
	// returns all projects in both s and newS
	projects := make(map[string]bool)
	for ing, ps := range *newS {
		// projects from new state: newS
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
func (s *ProjectsStatus) RemovedProjects(newS *ProjectsStatus) []string {
	// returns what presents in s but not in newS
	var projects []string
	for ing, ps := range *s {
		if _, ok := (*newS)[ing]; ok {
			for p := range ps {
				if _, ok := (*newS)[ing][p]; !ok {
					projects = append(projects, p)
				}
			}
		}
	}
	return projects
}

// Update updates the gateway current state according to the new state changes and response of gateway for changes
func (s *ProjectsStatus) Update(newS *ProjectsStatus, gatewayResponse controller.Response) {
	// allIngresses are the all updated or deleted ingresses
	allIngresses := make([]string, 0, len(*s)+len(*newS))
	for ing := range *s {
		allIngresses = append(allIngresses, ing)
	}
	for ing := range *newS {
		allIngresses = append(allIngresses, ing)
	}

	for resProject, resType := range gatewayResponse {
		switch resType {
		case controller.Failed:
			// project is failed to update or delete
			// ignore it
			continue
		case controller.Deleted:
			for ing := range *s {
				delete((*s)[ing], resProject)
				// if there are no any project, delete the ingress from current state
				if len((*s)[ing]) == 0 {
					delete(*s, ing)
				}
			}
		case controller.Updated:
			for _, ing := range allIngresses {
				// check project already exists in current state
				if _, ok := (*s)[ing]; ok {
					if _, ok := (*s)[ing][resProject]; ok {
						// do not need to update, already exists
						continue
					}
				}
				// if not already exists in current state, check the new state
				if _, ok := (*newS)[ing]; ok {
					if _, ok := (*newS)[ing][resProject]; ok {
						// new state contains the project
						// update the current state
						if _, ok := (*s)[ing]; ok {
							// existing ingress
							(*s)[ing][resProject] = "_"
						} else {
							// new ingress
							(*s)[ing] = map[string]string{resProject: "_"}
						}
					}
				}
				// not exists in both current or new state, ignore the update
			}
		}
	}
}
