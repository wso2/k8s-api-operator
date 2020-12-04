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

package parser

import (
	"fmt"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/ingress/errors"
	networking "k8s.io/api/networking/v1beta1"
	"strconv"
)

const (
	// DefaultPrefix defines the default annotation prefix used in the WSO2 microgateway ingress controller.
	DefaultPrefix = "microgateway.ingress.wso2.com"
)

var (
	// Prefix defines the annotation prefix which is mutable
	Prefix = DefaultPrefix
)

// Parser defines annotation parser
type Parser interface {
	Parse(*networking.Ingress)
}

// GetAnnotationWithPrefix returns annotation key with the annotation prefix
func GetAnnotationWithPrefix(name string) string {
	return fmt.Sprintf("%v/%v", Prefix, name)
}

// GetStringAnnotation returns string value of the annotation from given ingress
func GetStringAnnotation(ing *networking.Ingress, name string) (string, error) {
	fullName := GetAnnotationWithPrefix(name)
	val, ok := ing.Annotations[fullName]
	if ok {
		return val, nil
	}

	return "", errors.NewAnnotationNotExists(fullName)
}

// GetBoolAnnotation returns boolean value of the annotation from given ingress
func GetBoolAnnotation(ing *networking.Ingress, name string) (bool, error) {
	fullName := GetAnnotationWithPrefix(name)
	val, ok := ing.Annotations[fullName]
	if ok {
		b, err := strconv.ParseBool(val)
		if err != nil {
			return false, errors.IngressError{ErrReason: errors.InvalidContent, Message: err.Error()}
		}
		return b, nil
	}

	return false, errors.NewAnnotationNotExists(fullName)
}
