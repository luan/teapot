package managers

import (
	"fmt"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/route-emitter/cfroutes"
	diego_models "github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/luan/teapot/models"
	"github.com/pivotal-golang/lager"
)

type WorkstationManager interface {
	Create(workstation models.Workstation) error
	Delete(name string) error
	Fetch(name string) ([]receptor.ActualLRPResponse, error)
	List() ([]models.Workstation, error)
}

type workstationManager struct {
	receptorClient receptor.Client
	logger         lager.Logger
	appsDomain     string
}

func NewWorkstationManager(receptorClient receptor.Client, appsDomain string, logger lager.Logger) WorkstationManager {
	return &workstationManager{
		receptorClient: receptorClient,
		logger:         logger,
		appsDomain:     appsDomain,
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

	route := "tiego-" + workstation.Name + "." + m.appsDomain
	routingInfo := cfroutes.CFRoutes{
		{Hostnames: []string{route}, Port: 3000},
	}.RoutingInfo()

	openRule := diego_models.SecurityGroupRule{
		Protocol:     diego_models.AllProtocol,
		Destinations: []string{"0.0.0.0/0"},
	}
	openRules := []diego_models.SecurityGroupRule{openRule}

	if err != nil {
		log.Debug("marshalling-route-json-failed", lager.Data{"error": err})
	}

	lrpRequest := receptor.DesiredLRPCreateRequest{
		ProcessGuid: workstation.Name,
		Setup: &diego_models.SerialAction{
			Actions: []diego_models.Action{
				&diego_models.DownloadAction{
					From:     "https://tiego-artifacts.s3.amazonaws.com/tea-builds/tea-latest.tgz",
					To:       "/tmp",
					CacheKey: "tea",
				},
			},
		},
		Domain:     "tiego",
		Instances:  1,
		Stack:      "lucid64",
		RootFSPath: workstation.DockerImage,
		CPUWeight:  workstation.CPUWeight,
		DiskMB:     workstation.DiskMB,
		MemoryMB:   workstation.MemoryMB,
		LogGuid:    workstation.Name,
		LogSource:  "TEAPOT-WORKSTATION",
		Ports:      []uint16{8080, 3000},
		Routes:     routingInfo,
		Action: &diego_models.RunAction{
			Path:       "/tmp/tea",
			LogSource:  "TEA",
			Privileged: false,
		},
		EgressRules: openRules,
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

func (m *workstationManager) List() ([]models.Workstation, error) {
	workstations := []models.Workstation{}

	desiredLRPs, _ := m.receptorClient.DesiredLRPsByDomain("tiego")
	actualLRPs, _ := m.receptorClient.ActualLRPsByDomain("tiego")

	for _, desiredLRP := range desiredLRPs {
		state := models.StoppedState
		if i := contains(actualLRPs, desiredLRP.ProcessGuid); i >= 0 {
			state = fmt.Sprintf("%v", actualLRPs[i].State)
		}
		workstation := models.Workstation{Name: desiredLRP.ProcessGuid, DockerImage: desiredLRP.RootFSPath, State: state}
		workstations = append(workstations, workstation)
	}

	return workstations, nil
}

func contains(s []receptor.ActualLRPResponse, e string) int {
	for i, a := range s {
		if a.ProcessGuid == e {
			return i
		}
	}
	return -1
}
