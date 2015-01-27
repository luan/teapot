package models

import (
	"regexp"

	"github.com/luan/teapot"
)

const (
	StoppedState = "STOPPED"
	RunningState = "RUNNING"
	ClaimedState = "CLAIMED"
)

type Workstation struct {
	Name        string `json:"name"`
	DockerImage string `json:"docker_image"`
	State       string `json:"state"`
	CPUWeight   uint   `json:"cpu_weight"`
	DiskMB      int    `json:"disk_mb"`
	MemoryMB    int    `json:"memory_mb"`
}

func NewWorkstation(request teapot.WorkstationCreateRequest) Workstation {
	if len(request.DockerImage) == 0 {
		request.DockerImage = "docker:///ubuntu#trusty"
	}

	return Workstation{
		Name:        request.Name,
		DockerImage: request.DockerImage,
		CPUWeight:   request.CPUWeight,
		DiskMB:      request.DiskMB,
		MemoryMB:    request.MemoryMB,
		State:       StoppedState,
	}
}

func (workstation Workstation) Validate() error {
	var validationError ValidationError

	matched, err := regexp.MatchString("^[\\w-.]+$", workstation.Name)
	if err != nil || !matched {
		validationError = append(validationError, ErrInvalidField{"name"})
	}

	matched, err = regexp.MatchString("^docker:///[\\w-.]+(/[\\w-.]+)?#?[\\w-.]*$", workstation.DockerImage)
	if err != nil || !matched {
		validationError = append(validationError, ErrInvalidField{"docker_image"})
	}

	if len(validationError) > 0 {
		return validationError
	}
	return nil
}
