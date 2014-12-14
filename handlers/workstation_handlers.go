package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/luan/teapot"
	"github.com/luan/teapot/managers"
	"github.com/luan/teapot/models"
	"github.com/pivotal-golang/lager"
)

type WorkstationHandler struct {
	manager managers.WorkstationManager
	logger  lager.Logger
}

func NewWorkstationHandler(manager managers.WorkstationManager, logger lager.Logger) *WorkstationHandler {
	return &WorkstationHandler{
		manager: manager,
		logger:  logger,
	}
}

func (h *WorkstationHandler) Create(w http.ResponseWriter, r *http.Request) {
	log := h.logger.Session("create")
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

	err = h.manager.Create(workstation)

	if err != nil {
		switch t := err.(type) {
		default:
			log.Error("unknown-error", err, lager.Data{"type": t})
			log.Info("did you set RECEPTOR correctly?")
			writeJSONResponse(w, http.StatusInternalServerError, teapot.Error{
				Type:    teapot.UnknownError,
				Message: err.Error(),
			})
		case models.ValidationError:
			log.Error("invalid-workstation", err)
			writeJSONResponse(w, http.StatusBadRequest, teapot.Error{
				Type:    teapot.InvalidWorkstation,
				Message: err.Error(),
			})
		}
		return
	}

	log.Info("created", lager.Data{"workstation-name": workstation.Name})

	w.WriteHeader(http.StatusCreated)
}

func (h *WorkstationHandler) Delete(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue(":name")
	log := h.logger.Session("delete", lager.Data{
		"Name": name,
	})

	err := h.manager.Delete(name)
	if err != nil {
		log.Info("delete-failed", lager.Data{"workstation-name": name})
		w.WriteHeader(http.StatusNotFound)
		return
	}

	log.Info("deleted", lager.Data{"workstation-name": name})

	w.WriteHeader(http.StatusNoContent)
}
