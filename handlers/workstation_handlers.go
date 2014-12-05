package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/luan/teapot"
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

	err = h.receptorClient.CreateDesiredLRP(receptor.DesiredLRPCreateRequest{
		ProcessGuid: workstationRequest.Name,
		Domain:      "teapot",
		Instances:   1,
		Stack:       "lucid64",
		RootFSPath:  workstationRequest.DockerImage,
		DiskMB:      128,
		MemoryMB:    64,
		LogGuid:     workstationRequest.Name,
		LogSource:   "TEAPOT-WORKSTATION",
		Action: &models.RunAction{
			Path:      "/bin/sh",
			LogSource: "TEA",
		},
	})
	if err != nil {
		log.Error("create-desired-lrp", err)
		w.WriteHeader(http.StatusTeapot)
	} else {
		log.Info("created", lager.Data{"workstation-name": workstationRequest.Name})
		w.WriteHeader(http.StatusCreated)
	}
}
