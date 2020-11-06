package client

import "github.com/wso2/k8s-api-operator/api-operator/pkg/envoy/action"

type GatewayClient interface {
	Update(projects *action.ProjectsMap) (Response, error)
}

// Response represents the response code list after updating the microgateway
// maps [project -> response code]
//
// a_com
//   Updated
// b_com
//   Failed
// c_com
//   Deleted
//
type Response map[string]ResponseType

type ResponseType int

const (
	Failed  = ResponseType(0)
	Updated = ResponseType(1)
	Deleted = ResponseType(2)
)
