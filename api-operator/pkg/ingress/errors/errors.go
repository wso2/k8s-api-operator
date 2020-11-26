package errors

import "fmt"

type Reason string

const (
	AnnotationNotExists Reason = "AnnotationNotExists"
	InvalidContent      Reason = "InvalidContent"
)

type Ingress interface {
	Reason() Reason
}

type IngressError struct {
	ErrReason Reason
	Message   string
}

func (e IngressError) Error() string {
	return e.Message
}

func (e IngressError) Reason() Reason {
	return e.ErrReason
}

func NewAnnotationNotExists(name string) IngressError {
	return IngressError{
		ErrReason: AnnotationNotExists,
		Message:   fmt.Sprintf("Annotation '%s' is not provided", name),
	}
}
