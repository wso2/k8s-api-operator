package controller

import (
	"github.com/example-inc/k8s-apim-operator/apim-operator/pkg/controller/targetendpoint"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, targetendpoint.Add)
}
