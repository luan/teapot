package handlers

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"

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

func (h *WorkstationHandler) Attach(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue(":name")
	log := h.logger.Session("attach", lager.Data{
		"Name": name,
	})

	actualLRPs, err := h.manager.Fetch(name)
	if err != nil || len(actualLRPs) == 0 {
		log.Info("attach-failed", lager.Data{"workstation-name": name, "actual-lrps": actualLRPs, "error": err})
		w.WriteHeader(http.StatusNotFound)
		return
	}
	attachURL := fmt.Sprintf("ws://%s:%d/shell", actualLRPs[0].Host, actualLRPs[0].Ports[0].HostPort)

	u, _ := url.Parse(attachURL)

	conn, err := net.Dial("tcp", u.Host)
	if err != nil {
		log.Error("attach-failed", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	wsClient, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error("attach-failed", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer wsClient.Close()

	wsServer, _, err := websocket.NewClient(conn, u, http.Header{"Origin": {attachURL}}, 1024, 1024)
	if err != nil {
		log.Error("attach-failed", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer wsServer.Close()

	w.WriteHeader(http.StatusOK)
	log.Info("attached", lager.Data{"workstation-name": name})

	done := make(chan bool)
	go h.proxyWebsocket(wsServer, wsClient, done)
	go h.proxyWebsocket(wsClient, wsServer, done)
	<-done

	log.Info("unattached", lager.Data{"workstation-name": name})
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
