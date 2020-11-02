package common

import (
	"context"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type RequestInfo struct {
	reconcile.Request
	Ctx    context.Context
	Client *client.Client
	Object runtime.Object
	Log    logr.Logger
}
