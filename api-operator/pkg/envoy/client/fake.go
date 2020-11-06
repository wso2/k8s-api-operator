package client

import (
	"github.com/wso2/k8s-api-operator/api-operator/pkg/envoy/action"
	"math/rand"
	"time"
)

func NewFakeWithRandomResponse() *Fake {
	return &Fake{
		randomResponse: true,
	}
}

func NewFake(response Response) *Fake {
	return &Fake{
		Response:       response,
		randomResponse: false,
	}
}

type Fake struct {
	ProjectMap     *action.ProjectsMap
	Response       Response
	randomResponse bool
}

func (c *Fake) Update(projects *action.ProjectsMap) (Response, error) {
	c.ProjectMap = projects

	if !c.randomResponse {
		return c.Response, nil
	}

	// Random response
	r := Response{}
	rand.Seed(time.Now().UnixNano())

	for name, project := range *projects {
		if rand.Intn(2) == 0 {
			r[name] = Failed
		} else {
			switch project.Type {
			case action.Update:
				r[name] = Updated
			case action.Delete:
				r[name] = Deleted
			}
		}
	}
	c.Response = r
	return c.Response, nil
}
