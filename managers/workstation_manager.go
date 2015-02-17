package managers

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/route-emitter/cfroutes"
	diego_models "github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/luan/teapot/models"
	"github.com/pivotal-golang/lager"
)

var setupAction = &diego_models.SerialAction{
	Actions: []diego_models.Action{
		&diego_models.DownloadAction{
			From:     "https://tiego-artifacts.s3.amazonaws.com/dropbear/dropbear.tar.gz",
			To:       "/tmp",
			CacheKey: "dropbear",
		},
		&diego_models.DownloadAction{
			From:     "https://tiego-artifacts.s3.amazonaws.com/tea-builds/tea-latest.tgz",
			To:       "/tmp",
			CacheKey: "tea",
		},
		&diego_models.RunAction{
			Path:      "/tmp/dropbearkey",
			LogSource: "KEYGEN",
			Args:      []string{"-t", "rsa", "-f", "/tmp/dropbear_rsa_host_key"},
		},
		&diego_models.RunAction{
			Path:      "/tmp/dropbearkey",
			LogSource: "KEYGEN",
			Args:      []string{"-t", "dss", "-f", "/tmp/dropbear_dss_host_key"},
		},
		&diego_models.RunAction{
			Path:      "/tmp/dropbearkey",
			LogSource: "KEYGEN",
			Args:      []string{"-t", "ecdsa", "-f", "/tmp/dropbear_ecdsa_host_key"},
		},
	},
}

type WorkstationManager interface {
	Create(workstation models.Workstation) error
	Delete(name string) error
	Fetch(name string) ([]receptor.ActualLRPResponse, error)
	List() ([]models.Workstation, error)
	AddKey(name string, key []byte) error
}

type workstationManager struct {
	receptorClient receptor.Client
	logger         lager.Logger
	teaSecret      string
	routeProvider  models.RouteProvider
}

func NewWorkstationManager(receptorClient receptor.Client, routeProvider models.RouteProvider, teaSecret string, logger lager.Logger) WorkstationManager {
	return &workstationManager{
		receptorClient: receptorClient,
		logger:         logger,
		teaSecret:      teaSecret,
		routeProvider:  routeProvider,
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

	tiegoRoute := m.routeProvider.TiegoRoute(workstation.Name)
	sshRoute := m.routeProvider.SSHRoute(workstation.Name)
	routingInfo := cfroutes.CFRoutes{
		{Hostnames: []string{tiegoRoute}, Port: 3000},
		{Hostnames: []string{sshRoute}, Port: 8080},
	}.RoutingInfo()

	openRules := []diego_models.SecurityGroupRule{
		{
			Protocol:     diego_models.AllProtocol,
			Destinations: []string{"0.0.0.0/0"},
		},
	}

	if err != nil {
		log.Debug("marshalling-route-json-failed", lager.Data{"error": err})
	}

	mainAction := &diego_models.ParallelAction{
		Actions: []diego_models.Action{
			&diego_models.RunAction{
				Path:      "/bin/bash",
				LogSource: "SSHD",
				Args: []string{
					"-c",
					`set -e && /tmp/dropbear -p 127.0.0.1:22000 -r /tmp/dropbear_rsa_host_key -r /tmp/dropbear_dss_host_key -r /tmp/dropbear_ecdsa_host_key`,
				},
				Privileged: false,
			},
			&diego_models.RunAction{
				Path: "/tmp/tea",
				Args: []string{
					"-secret", m.teaSecret,
				},
				LogSource:  "TEA",
				Privileged: false,
			},
		},
	}

	lrpRequest := receptor.DesiredLRPCreateRequest{
		ProcessGuid: workstation.Name,
		Setup:       setupAction,
		Domain:      "tiego",
		Instances:   1,
		Stack:       "lucid64",
		RootFSPath:  workstation.DockerImage,
		CPUWeight:   workstation.CPUWeight,
		DiskMB:      workstation.DiskMB,
		MemoryMB:    workstation.MemoryMB,
		LogGuid:     workstation.Name,
		LogSource:   "TEAPOT-WORKSTATION",
		Ports:       []uint16{8080, 3000},
		Routes:      routingInfo,
		Privileged:  true,
		Action:      mainAction,
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

func (m *workstationManager) AddKey(name string, key []byte) error {
	route := m.routeProvider.SSHRoute(name)
	if !strings.HasPrefix(route, "http://") && !strings.HasPrefix(route, "https://") {
		route = "http://" + route
	}
	u, _ := url.Parse(route)
	u.Path = path.Join("add-key", m.teaSecret)

	body := bytes.NewReader(key)
	_, err := http.Post(u.String(), "text/plain", body)
	if err != nil {
		return err
	}

	return nil
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
