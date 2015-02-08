package handlers

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/gorilla/websocket"
	"github.com/luan/teapot"
	"github.com/luan/teapot/managers"
	"github.com/luan/teapot/models"
	"github.com/pivotal-golang/lager"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

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

	workstation := models.NewWorkstation(workstationRequest)

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

	log.Info("created", lager.Data{"workstation_name": workstation.Name})

	w.WriteHeader(http.StatusCreated)
}

func (h *WorkstationHandler) List(w http.ResponseWriter, r *http.Request) {
	workstations, _ := h.manager.List()

	js, _ := json.Marshal(&workstations)

	w.WriteHeader(http.StatusOK)
	w.Write(js)
}

func (h *WorkstationHandler) Delete(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue(":name")
	log := h.logger.Session("delete", lager.Data{
		"Name": name,
	})

	err := h.manager.Delete(name)
	if err != nil {
		log.Info("delete-failed", lager.Data{"workstation_name": name})
		writeWorkstationNotFoundResponse(w, name)
		return
	}

	log.Info("deleted", lager.Data{"workstation_name": name})

	w.WriteHeader(http.StatusNoContent)
}

func (h *WorkstationHandler) Attach(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue(":name")
	log := h.logger.Session("attach", lager.Data{
		"Name": name,
	})

	actualLRPs, err := h.manager.Fetch(name)
	if err != nil || len(actualLRPs) == 0 {
		log.Info("attach-failed", lager.Data{"workstation_name": name, "actual_lrps": actualLRPs, "error": err})
		writeWorkstationNotFoundResponse(w, name)
		return
	}

	workstation := actualLRPs[0]
	if workstation.State != receptor.ActualLRPStateRunning {
		log.Info("attach-failed", lager.Data{"workstation_name": name, "actual_lrps": actualLRPs, "error": err})
		writeInvalidWorkstationResponse(w, workstation)
		return
	}

	attachURL := fmt.Sprintf("ws://%s:%d/shell", workstation.Address, workstation.Ports[0].HostPort)
	log.Debug("attaching-to", lager.Data{"attach_url": attachURL, "actual_lrps": actualLRPs})

	u, _ := url.Parse(attachURL)

	log.Debug("opening-tcp", lager.Data{"address": u.Host})
	conn, err := net.Dial("tcp", u.Host)
	if err != nil {
		log.Error("attach-failed", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Debug("tcp-connection-open", lager.Data{"conn": conn.RemoteAddr()})

	wsClient, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error("attach-failed", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer wsClient.Close()
	log.Debug("websocket-open", lager.Data{"ws": wsClient.RemoteAddr()})

	wsServer, _, err := websocket.NewClient(conn, u, http.Header{"Origin": {attachURL}}, 1024, 1024)
	if err != nil {
		log.Error("attach-failed", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer wsServer.Close()

	w.WriteHeader(http.StatusOK)
	log.Info("attached", lager.Data{"workstation_name": name})

	done := make(chan bool)
	go h.proxyWebsocket(wsServer, wsClient, done)
	go h.proxyWebsocket(wsClient, wsServer, done)
	<-done

	log.Info("unattached", lager.Data{"workstation_name": name})
}

func (h *WorkstationHandler) proxyWebsocket(s *websocket.Conn, d *websocket.Conn, done chan bool) {
	log := h.logger.Session("proxy-websocket")

	for {
		mType, m, err := s.ReadMessage()
		if err != nil {
			log.Error("read-error", err)
			done <- true
			return
		}

		d.WriteMessage(mType, m)
	}
}

func writeWorkstationNotFoundResponse(w http.ResponseWriter, name string) {
	writeJSONResponse(w, http.StatusNotFound, receptor.Error{
		Type:    teapot.WorkstationNotFound,
		Message: fmt.Sprintf("Workstation with name '%s' not found", name),
	})
}

func writeInvalidWorkstationResponse(w http.ResponseWriter, workstation receptor.ActualLRPResponse) {
	writeJSONResponse(w, http.StatusBadRequest, receptor.Error{
		Type:    teapot.InvalidWorkstation,
		Message: fmt.Sprintf("Workstation %s is not RUNNING.", workstation.ProcessGuid),
	})
}
