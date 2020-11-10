package common

import (
	"context"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type key int

const requestInfoKey key = 0

type RequestInfo struct {
	reconcile.Request
	Client      client.Client
	Object      runtime.Object
	EvnRecorder record.EventRecorder
	Log         logr.Logger
}

func (r RequestInfo) NewContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, requestInfoKey, r)
}

func FromContext(ctx context.Context) (RequestInfo, bool) {
	value, ok := ctx.Value(requestInfoKey).(RequestInfo)
	return value, ok
}
