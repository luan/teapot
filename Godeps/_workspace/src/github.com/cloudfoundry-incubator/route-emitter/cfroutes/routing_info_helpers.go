package cfroutes

import (
	"encoding/json"

	"github.com/cloudfoundry-incubator/receptor"
)

const CF_ROUTER = "cf-router"

type CFRoutes []CFRoute

type CFRoute struct {
	Hostnames []string `json:"hostnames"`
	Port      uint16   `json:"port"`
}

func (c CFRoutes) RoutingInfo() receptor.RoutingInfo {
	data, _ := json.Marshal(c)
	routingInfo := json.RawMessage(data)
	return receptor.RoutingInfo{
		CF_ROUTER: &routingInfo,
	}
}

func CFRoutesFromRoutingInfo(routingInfo receptor.RoutingInfo) (CFRoutes, error) {
	if routingInfo == nil {
		return nil, nil
	}

	data, found := routingInfo[CF_ROUTER]
	if !found {
		return nil, nil
	}

	if data == nil {
		return nil, nil
	}

	routes := CFRoutes{}
	err := json.Unmarshal(*data, &routes)

	return routes, err
}
