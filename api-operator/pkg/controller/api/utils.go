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

package api

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"math/rand"
	"regexp"
	"strings"
)

// removeVersionTag removes version number in a url provided
func removeVersionTag(url string) string {
	regExpString := `\/v[\d.-]*\/?$`
	regExp := regexp.MustCompile(regExpString)
	return regExp.ReplaceAllString(url, "")
}

// isStringArrayContains checks the given text contains in the given arr
func isStringArrayContains(arr []string, text string) bool {
	for _, s := range arr {
		if s == text {
			return true
		}
	}
	return false
}

// getRandFileName returns a file name with suffixing a random number
func getRandFileName(filename string) string {
	fileSplits := strings.SplitN(filename, ".", 2)
	return fmt.Sprintf("%v-%v.%v", fileSplits[0], rand.Intn(10000), fileSplits[1])
}

// copySecret copies secret from given to destination given
func copySecret(r *ReconcileAPI, fromName, fromNamespace, toName, toNamespace string) error {
	// Get volume
	fromScrt := &corev1.Secret{}
	fromErr := r.client.Get(context.TODO(), types.NamespacedName{Name: fromName, Namespace: fromNamespace}, fromScrt)
	if fromErr != nil && errors.IsNotFound(fromErr) {
		log.Info("Secret not found", "namespace", fromNamespace, "name", fromName)
		return fromErr
	} else if fromErr != nil {
		log.Error(fromErr, "Error getting secret", "namespace", fromNamespace, "name", fromName)
		return fromErr
	}

	toScrt := &corev1.Secret{}
	toErr := r.client.Get(context.TODO(), types.NamespacedName{Name: toName, Namespace: toNamespace}, toScrt)
	toScrt.Data = fromScrt.Data
	toScrt.StringData = fromScrt.StringData
	toScrt.Type = fromScrt.Type
	toScrt.Namespace = toNamespace
	toScrt.Name = toName

	// Place volume in to namespace
	if toErr != nil && errors.IsNotFound(toErr) {
		log.Info("Coping secret to users namespace", "from namespace", fromNamespace, "from name", fromName,
			"to namespace", toNamespace, "to name", toName)

		createErr := r.client.Create(context.TODO(), toScrt)
		if createErr != nil {
			log.Error(createErr, "Error creating secret", "namespace", toNamespace, "name", toName)
			return createErr
		}
		return nil
	} else if toErr != nil {
		log.Error(toErr, "Error getting secret", "namespace", toNamespace, "name", toName)
		return toErr
	}

	// toScrt already exists and update it
	updateErr := r.client.Update(context.TODO(), toScrt)
	if updateErr != nil {
		log.Error(updateErr, "Error updating secret", "namespace", toNamespace, "name", toName)
		return updateErr
	}

	return nil
}
