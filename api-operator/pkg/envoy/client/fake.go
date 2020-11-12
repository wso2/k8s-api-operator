package client

import (
	"context"
	"fmt"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/controller/common"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/envoy/action"
	"github.com/wso2/k8s-api-operator/api-operator/pkg/swagger"
	"math/rand"
	"time"
)

type Fake struct {
	ProjectMap     *action.ProjectsMap
	Response       Response
	responseMethod func(projects *action.ProjectsMap) Response
}

func NewFake(response Response) *Fake {
	return &Fake{
		responseMethod: func(projects *action.ProjectsMap) Response {
			return response
		},
	}
}

func NewFakeAllSucceeded() *Fake {
	return &Fake{
		responseMethod: func(projects *action.ProjectsMap) Response {
			r := Response{}

			for name, project := range *projects {
				switch project.Type {
				case action.ForceUpdate:
					r[name] = Updated
				case action.Delete:
					r[name] = Deleted
				}
			}
			return r
		},
	}
}

func NewFakeAllFailed() *Fake {
	return &Fake{
		responseMethod: func(projects *action.ProjectsMap) Response {
			r := Response{}

			for name := range *projects {
				r[name] = Failed
			}
			return r
		},
	}
}

func NewFakeWithRandomResponse() *Fake {
	return &Fake{
		responseMethod: func(projects *action.ProjectsMap) Response {
			r := Response{}
			rand.Seed(time.Now().UnixNano())

			for name, project := range *projects {
				if rand.Intn(2) == 0 {
					r[name] = Failed
				} else {
					switch project.Type {
					case action.ForceUpdate:
						r[name] = Updated
					case action.Delete:
						r[name] = Deleted
					}
				}
			}
			return r
		},
	}
}

func (c *Fake) Update(ctx context.Context, reqInfo *common.RequestInfo, projects *action.ProjectsMap) (Response, error) {
	for s, project := range *projects {
		fmt.Println("")
		fmt.Println("******* PRINT PROJECT ******")
		fmt.Printf("Project name: %s\n", s)
		fmt.Printf("Action: %s\n", project.Type)

		if project.Type != action.ForceUpdate {
			continue
		}
		err := project.OAS.Validate(ctx)
		fmt.Printf("Swagger validation: %v\n", err == nil)
		fmt.Printf("Tls certs: %v\n", project.TlsCertificate)
		fmt.Println(swagger.PrettyString(project.OAS))
	}

	c.ProjectMap = projects
	c.Response = c.responseMethod(projects)
	return c.Response, nil
}
