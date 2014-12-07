package managers

import (
	"github.com/cloudfoundry-incubator/receptor"
	diego_models "github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/luan/teapot/models"
	"github.com/pivotal-golang/lager"
)

type WorkstationManager interface {
	Create(models.Workstation) error
}

type workstationManager struct {
	receptorClient receptor.Client
	logger         lager.Logger
}

func NewWorkstationManager(receptorClient receptor.Client, logger lager.Logger) WorkstationManager {
	return &workstationManager{
		receptorClient: receptorClient,
		logger:         logger,
	}
}

func (m *workstationManager) Create(workstation models.Workstation) error {
	if err := workstation.Validate(); err != nil {
		return err
	}

	desiredLRP, err := m.receptorClient.GetDesiredLRP(workstation.Name)
	if err == nil && desiredLRP.ProcessGuid == workstation.Name {
		return models.ValidationError{models.ErrDuplicateField{"name"}}
	}

	err = m.receptorClient.CreateDesiredLRP(receptor.DesiredLRPCreateRequest{
		ProcessGuid: workstation.Name,
		Domain:      "teapot",
		Instances:   1,
		Stack:       "lucid64",
		RootFSPath:  workstation.DockerImage,
		DiskMB:      128,
		MemoryMB:    64,
		LogGuid:     workstation.Name,
		LogSource:   "TEAPOT-WORKSTATION",
		Action: &diego_models.RunAction{
			Path:      "/bin/sh",
			LogSource: "TEA",
		},
	})

	return err
}
