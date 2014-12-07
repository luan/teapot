package handlers

import (
	"net/http"

	"github.com/luan/teapot"
	"github.com/luan/teapot/managers"
	"github.com/pivotal-golang/lager"
	"github.com/tedsuo/rata"
)

func New(workstationManager managers.WorkstationManager, logger lager.Logger) http.Handler {
	workstationHandler := NewWorkstationHandler(workstationManager, logger)

	actions := rata.Handlers{
		// Workstations
		teapot.CreateWorkstationRoute: route(workstationHandler.Create),
	}

	handler, err := rata.NewRouter(teapot.Routes, actions)
	if err != nil {
		panic("unable to create router: " + err.Error())
	}

	handler = LogWrap(handler, logger)

	return handler
}

func route(f func(w http.ResponseWriter, r *http.Request)) http.Handler {
	return http.HandlerFunc(f)
}
