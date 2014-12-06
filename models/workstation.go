package models

import "regexp"

type Workstation struct {
	Name        string `json:"name"`
	DockerImage string `json:"docker_image"`
}

func NewWorkstation(arguments ...string) Workstation {
	var name, dockerImage string

	if len(arguments) > 0 {
		name = arguments[0]
	}
	if len(arguments) > 1 {
		dockerImage = arguments[1]
	}

	return Workstation{name, dockerImage}
}

func (workstation Workstation) Validate() error {
	var validationError ValidationError

	matched, err := regexp.MatchString("^[\\w-]+$", workstation.Name)
	if err != nil || !matched {
		validationError = append(validationError, ErrInvalidField{"name"})
	}

	matched, err = regexp.MatchString("^docker:///[\\w-]+#?[\\w-]*$", workstation.DockerImage)
	if err != nil || !matched {
		validationError = append(validationError, ErrInvalidField{"docker_image"})
	}

	if len(validationError) > 0 {
		return validationError
	}
	return nil
}
