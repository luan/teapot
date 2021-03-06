package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/route-emitter/cfroutes"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
)

var bucket = flag.String(
	"bucket",
	"tiego-artifacts",
	"The bucket where teapot will be downloaded from.",
)

var filename = flag.String(
	"filename",
	"",
	"the filename to download from the bucket",
)

var receptorAddr string

func DockerTeapot(client receptor.Client, routeRoot string) error {
	teapotDownloadURL := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", *bucket, *filename)
	fmt.Println(teapotDownloadURL)
	client.DeleteDesiredLRP("teapot")
	route := fmt.Sprintf("teapot.%s", routeRoot)
	username := os.Getenv("TEAPOT_USERNAME")
	password := os.Getenv("TEAPOT_PASSWORD")
	teaSecret := os.Getenv("TEAPOT_TEA_SECRET")
	devMode := os.Getenv("TEAPOT_DEVMODE")
	if devMode != "true" && (len(username) == 0 || len(password) == 0 || len(teaSecret) == 0) {
		fmt.Println("Either set TEAPOT_USERNAME and TEAPOT_PASSWORD and TEAPOT_TEA_SECRET or, to disable authentication, TEAPOT_DEVMODE=true")
		os.Exit(1)
	}
	if len(teaSecret) == 0 {
		teaSecret = "p"
	}

	bytes, _ := json.Marshal(cfroutes.CFRoutes{
		{Hostnames: []string{route}, Port: 8080},
	})

	routeRawJson := json.RawMessage(bytes)
	routingInfo := receptor.RoutingInfo{}
	routingInfo[cfroutes.CF_ROUTER] = &routeRawJson

	err := client.CreateDesiredLRP(receptor.DesiredLRPCreateRequest{
		ProcessGuid: "teapot",
		Domain:      "teapot",
		RootFSPath:  "docker:///busybox#ubuntu-14.04",
		Instances:   1,
		Stack:       "lucid64",
		Setup: &models.ParallelAction{[]models.Action{
			&models.DownloadAction{
				From: teapotDownloadURL,
				To:   "/tmp",
			},
		}, ""},
		Action: &models.RunAction{
			Path: "/tmp/teapot",
			Args: []string{
				"-address", "0.0.0.0:8080",
				"-receptorAddress", receptorAddr,
				"-username", username,
				"-password", password,
				"-appsDomain", routeRoot,
				"-teaSecret", teaSecret,
			},
			LogSource: "TEAPOT",
		},
		DiskMB:    128,
		MemoryMB:  64,
		Ports:     []uint16{8080},
		Routes:    routingInfo,
		LogGuid:   "teapot",
		LogSource: "TEAPOT",
	})
	if err != nil {
		return err
	}

	fmt.Println("Teapot is deployed.")
	fmt.Printf("To make contact:\n  http://%s/\n", route)

	return nil
}

func main() {
	receptorAddr = os.Getenv("RECEPTOR")
	if receptorAddr == "" {
		fmt.Println("No RECEPTOR set")
		os.Exit(1)
	}
	flag.Parse()

	client := receptor.NewClient(receptorAddr)
	routeRoot := strings.Split(receptorAddr, "receptor.")[1]
	DockerTeapot(client, routeRoot)
}
