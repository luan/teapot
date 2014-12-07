package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/cloudfoundry-incubator/receptor"
	diego_models "github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/luan/teapot"
	"github.com/luan/teapot/models"
	"github.com/pivotal-golang/lager"
)

type WorkstationHandler struct {
	receptorClient receptor.Client
	logger         lager.Logger
}

func NewWorkstationHandler(receptorClient receptor.Client, logger lager.Logger) *WorkstationHandler {
	return &WorkstationHandler{
		receptorClient: receptorClient,
		logger:         logger,
	}
}

func (h *WorkstationHandler) Create(w http.ResponseWriter, r *http.Request) {
	log := h.logger.Session("create-workstation-handler")
	workstationRequest := teapot.WorkstationCreateRequest{}

	err := json.NewDecoder(r.Body).Decode(&workstationRequest)
	if err != nil {
		log.Error("invalid-json", err)
		writeJSONResponse(w, http.StatusBadRequest, teapot.Error{
			Type:    teapot.InvalidJSON,
			Message: err.Error(),
		})
		return
	}

	workstation := models.NewWorkstation(workstationRequest.Name, workstationRequest.DockerImage)

	if err = workstation.Validate(); err != nil {
		log.Error("invalid-workstation", err)
		writeJSONResponse(w, http.StatusTeapot, teapot.Error{
			Type:    teapot.InvalidWorkstation,
			Message: err.Error(),
		})
		return
	}

	desiredLRP, err := h.receptorClient.GetDesiredLRP(workstation.Name)
	if err == nil && desiredLRP.ProcessGuid == workstation.Name {
		log.Info("create-desired-lrp", lager.Data{"message": "trying to create duplicate workstation"})
		writeJSONResponse(w, http.StatusBadRequest, teapot.Error{
			Type:    teapot.DuplicateWorkstation,
			Message: "workstation '" + workstation.Name + "' already exists",
		})
		return
	}

	err = h.receptorClient.CreateDesiredLRP(receptor.DesiredLRPCreateRequest{
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
	if err != nil {
		log.Error("create-desired-lrp", err)
		w.WriteHeader(http.StatusBadGateway)
	} else {
		log.Info("created", lager.Data{"workstation-name": workstationRequest.Name})
		w.WriteHeader(http.StatusCreated)
	}
}
