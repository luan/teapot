package main

import (
	"os"
	"strings"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
)

var domain = "tiego"
var routeRoot string
var client receptor.Client

func contains(s []receptor.ActualLRPResponse, e string) int {
	for i, a := range s {
		if a.ProcessGuid == e {
			return i
		}
	}
	return -1
}

func namify(guid string) string {
	return strings.Replace(guid, "tiego-", "", 1)
}

func listWorkstations() []string {
	// desiredLRPs, _ := client.DesiredLRPsByDomain(domain)
	actualLRPs, _ := client.ActualLRPsByDomain(domain)
	lrps := []receptor.ActualLRPResponse{}
	list := []string{}

	for _, lrp := range actualLRPs {
		lrps = append(lrps, lrp)
		name := namify(lrp.ProcessGuid)
		list = append(list, name)
	}

	// for _, lrp := range desiredLRPs {
	// 	// state := ""
	// 	// if i := contains(lrps, lrp.ProcessGuid); i >= 0 {
	// 	// 	state = lrps[i].State
	// 	// }
	// 	// name := namify(lrp.ProcessGuid)
	// 	// list = append(list, name)
	// }

	return list
}

func main() {
	receptorAddr := os.Getenv("RECEPTOR")
	if receptorAddr == "" {
		panic("No RECEPTOR set")
	}

	client = receptor.NewClient(receptorAddr)
	routeRoot = strings.Split(receptorAddr, "receptor.")[1]

	m := martini.Classic()
	// render html templates from templates directory
	m.Use(render.Renderer(render.Options{
		Layout: "layout",
	}))

	m.Get("/", func(params martini.Params, r render.Render) {
		r.HTML(200, "index", listWorkstations())
	})

	m.Get("/shell/:name", func(params martini.Params, r render.Render) {
		r.HTML(200, "terminal", params["name"])
	})

	m.Use(martini.Static("public"))

	m.Run()
}
