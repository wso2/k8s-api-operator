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
