package client

import (
	"context"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/controller/common"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/envoy/action"
)

type GatewayClient interface {
	Update(ctx context.Context, reqInfo *common.RequestInfo, projects *action.ProjectsMap) (Response, error)
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

type ResponseType string

const (
	Failed  = ResponseType("Failed")
	Updated = ResponseType("Updated")
	Deleted = ResponseType("Deleted")
)
