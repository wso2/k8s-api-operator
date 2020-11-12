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

type ResponseType int

func (t ResponseType) String() string {
	switch t {
	case Failed:
		return "Failed"
	case Updated:
		return "Updated"
	case Deleted:
		return "Deleted"
	}
	return "Unsupported Response Type"
}

const (
	Failed  = ResponseType(0)
	Updated = ResponseType(1)
	Deleted = ResponseType(2)
)
