package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
)

const (
	spyDownloadURL    string = "http://file_server.service.dc1.consul:8080/v1/static/docker-circus/docker-circus.tgz"
	teapotDownloadURL string = "https://tiego-artifacts.s3.amazonaws.com/teapot.tar.gz"
)

var receptorAddr string

func DockerTeapot(client receptor.Client, routeRoot string) error {
	client.DeleteDesiredLRP("teapot")
	route := fmt.Sprintf("teapot.%s", routeRoot)
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
			&models.DownloadAction{
				From: spyDownloadURL,
				To:   "/tmp",
			},
		}, ""},
		Action: &models.RunAction{
			Path: "/tmp/teapot",
			Args: []string{
				"-address", "0.0.0.0:8080",
				"-receptorAddress", receptorAddr,
			},
			LogSource: "TEAPOT",
		},
		Monitor: &models.RunAction{
			Path:      "/tmp/spy",
			Args:      []string{"-addr", fmt.Sprintf(":%d", 8080)},
			LogSource: "SPY",
		},
		DiskMB:    128,
		MemoryMB:  64,
		Ports:     []uint32{8080},
		Routes:    []string{route},
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

	client := receptor.NewClient(receptorAddr)
	routeRoot := strings.Split(receptorAddr, "receptor.")[1]
	DockerTeapot(client, routeRoot)
}
