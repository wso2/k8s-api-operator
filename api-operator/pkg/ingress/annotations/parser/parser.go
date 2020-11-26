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

type Parser interface {
	Parse(*networking.Ingress)
}

func GetAnnotationWithPrefix(name string) string {
	return fmt.Sprintf("%v/%v", Prefix, name)
}

func GetStringAnnotation(ing *networking.Ingress, name string) (string, error) {
	fullName := GetAnnotationWithPrefix(name)
	val, ok := ing.Annotations[fullName]
	if ok {
		return val, nil
	}

	return "", errors.NewAnnotationNotExists(fullName)
}

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
