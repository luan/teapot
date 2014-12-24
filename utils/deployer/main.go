package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/cloudfoundry-incubator/receptor"
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

const (
	spyDownloadURL string = "http://file_server.service.dc1.consul:8080/v1/static/docker-circus/docker-circus.tgz"
)

var receptorAddr string

func DockerTeapot(client receptor.Client, routeRoot string) error {
	teapotDownloadURL := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", *bucket, *filename)
	fmt.Println(teapotDownloadURL)
	client.DeleteDesiredLRP("teapot")
	route := fmt.Sprintf("teapot.%s", routeRoot)
	username := os.Getenv("TEAPOT_USERNAME")
	password := os.Getenv("TEAPOT_PASSWORD")
	devMode := os.Getenv("TEAPOT_DEVMODE")
	if devMode != "true" && (len(username) == 0 || len(password) == 0) {
		fmt.Println("Either set TEAPOT_USERNAME and TEAPOT_PASSWORD or, to disable authentication, TEAPOT_DEVMODE=true")
		os.Exit(1)
	}
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
				"-username", username,
				"-password", password,
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
	flag.Parse()

	client := receptor.NewClient(receptorAddr)
	routeRoot := strings.Split(receptorAddr, "receptor.")[1]
	DockerTeapot(client, routeRoot)
}
