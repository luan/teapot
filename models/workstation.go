package models

import "regexp"

type Workstation struct {
	Name        string `json:"name"`
	DockerImage string `json:"docker_image"`
	State       string `json:"state"`
}

func NewWorkstation(arguments ...string) Workstation {
	var name, dockerImage, state string

	if len(arguments) > 0 {
		name = arguments[0]
	}
	if len(arguments) > 1 && len(arguments[1]) > 0 {
		dockerImage = arguments[1]
	} else {
		dockerImage = "docker:///ubuntu#trusty"
	}

	if len(arguments) > 2 && len(arguments[2]) > 0 {
		state = arguments[2]
	} else {
		state = "STOPPED"
	}

	return Workstation{name, dockerImage, state}
}

func (workstation Workstation) Validate() error {
	var validationError ValidationError

	matched, err := regexp.MatchString("^[\\w-.]+$", workstation.Name)
	if err != nil || !matched {
		validationError = append(validationError, ErrInvalidField{"name"})
	}

	matched, err = regexp.MatchString("^docker:///[\\w-.]+#?[\\w-.]*$", workstation.DockerImage)
	if err != nil || !matched {
		validationError = append(validationError, ErrInvalidField{"docker_image"})
	}

	if len(validationError) > 0 {
		return validationError
	}
	return nil
}
