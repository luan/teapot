package managers

import (
	"github.com/cloudfoundry-incubator/receptor"
	diego_models "github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/luan/teapot/models"
	"github.com/pivotal-golang/lager"
)

type WorkstationManager interface {
	Create(workstation models.Workstation) error
	Delete(name string) error
	Fetch(name string) ([]receptor.ActualLRPResponse, error)
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
	log := m.logger.Session("workstation-manager-create", lager.Data{"workstation": workstation})

	if err := workstation.Validate(); err != nil {
		return err
	}

	desiredLRP, err := m.receptorClient.GetDesiredLRP(workstation.Name)
	if err == nil && desiredLRP.ProcessGuid == workstation.Name {
		return models.ValidationError{models.ErrDuplicateField{"name"}}
	}

	lrpRequest := receptor.DesiredLRPCreateRequest{
		ProcessGuid: workstation.Name,
		Setup: &diego_models.SerialAction{
			Actions: []diego_models.Action{
				&diego_models.DownloadAction{
					From:     "https://dl.dropboxusercontent.com/u/33868236/tea.tar.gz",
					To:       "/tmp",
					CacheKey: "tea",
				},
			},
		},
		Domain:     "tiego",
		Instances:  1,
		Stack:      "lucid64",
		RootFSPath: workstation.DockerImage,
		DiskMB:     128,
		MemoryMB:   64,
		LogGuid:    workstation.Name,
		LogSource:  "TEAPOT-WORKSTATION",
		Ports:      []uint32{8080},
		Action: &diego_models.RunAction{
			Path:       "/tmp/tea",
			LogSource:  "TEA",
			Privileged: true,
		},
	}

	log.Debug("requesting-lrp", lager.Data{"lrp_request": lrpRequest})
	err = m.receptorClient.CreateDesiredLRP(lrpRequest)
	if err != nil {
		log.Debug("request-failed", lager.Data{"error": err})
	} else {
		log.Debug("request-suceeded")
	}

	return err
}

func (m *workstationManager) Delete(name string) error {
	return m.receptorClient.DeleteDesiredLRP(name)
}

func (m *workstationManager) Fetch(name string) ([]receptor.ActualLRPResponse, error) {
	return m.receptorClient.ActualLRPsByProcessGuid(name)
}
