package controller

import (
	"github.com/wso2/k8s-api-operator/apim-operator/pkg/controller/ratelimiting"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, ratelimiting.Add)
}
