package client

import (
	"github.com/wso2/k8s-api-operator/api-operator/pkg/envoy/action"
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

func (c *Fake) Update(projects *action.ProjectsMap) (Response, error) {
	c.ProjectMap = projects
	c.Response = c.responseMethod(projects)
	return c.Response, nil
}
