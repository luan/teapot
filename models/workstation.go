package models

import "regexp"

type Workstation struct {
	Name        string `json:"name"`
	DockerImage string `json:"docker_image"`
}

func (workstation Workstation) Validate() error {
	var validationError ValidationError

	matched, err := regexp.MatchString("^[a-zA-Z0-9.-]+$", workstation.Name)
	if err != nil || !matched {
		validationError = append(validationError, ErrInvalidField{"name"})
	}

	if len(validationError) > 0 {
		return validationError
	}
	return nil
}
