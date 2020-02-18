package utils

import (
	"fmt"
	"github.com/go-logr/logr"
	registryclient "github.com/heroku/docker-registry-client/registry"
	"strings"
)

type RegAuth struct {
	RegistryUrl string
	Username    string
	Password    string
}

func IsImageExists(auth RegAuth, image string, tag string, logger logr.Logger) (bool, error) {
	hub, err := registryclient.New(auth.RegistryUrl, auth.Username, auth.Password)
	if err != nil {
		logger.Error(err, "Error connecting to the docker registry", "registry-url", auth.RegistryUrl)
		return false, err
	}

	// remove registry name if exists in the image name
	imageWithoutReg := image
	splits := strings.Split(image, "/")
	if len(splits) == 3 {
		imageWithoutReg = fmt.Sprintf("%s/%s", splits[1], splits[2])
	}

	tags, err := hub.Tags(imageWithoutReg)
	if err != nil {
		logger.Error(err, "Error getting tags from the image in the docker registry", "registry-url", auth.RegistryUrl, "image", image)
		return false, err
	}
	for _, foundTag := range tags {
		if foundTag == tag {
			logger.Info("Found the image tag from the registry", "image", imageWithoutReg, "tag", foundTag)
			return true, nil
		}
	}
	return false, nil
}
