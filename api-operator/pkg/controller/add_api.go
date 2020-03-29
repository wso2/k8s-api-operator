package controller

import (
	"github.com/wso2/k8s-api-operator/api-operator/pkg/controller/api"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, api.Add)
}
