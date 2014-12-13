package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	cf_lager "github.com/cloudfoundry-incubator/cf-lager"
	"github.com/cloudfoundry-incubator/receptor"
	"github.com/luan/teapot/handlers"
	"github.com/luan/teapot/managers"
	"github.com/pivotal-golang/lager"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/grouper"
	"github.com/tedsuo/ifrit/http_server"
	"github.com/tedsuo/ifrit/sigmon"
)

var serverAddress = flag.String(
	"address",
	"",
	"The host:port that the server is bound to.",
)

var receptorAddress = flag.String(
	"receptorAddress",
	"",
	"The url for the receptor.",
)

var username = flag.String(
	"username",
	"",
	"username for basic auth, enables basic auth if set",
)

var password = flag.String(
	"password",
	"",
	"password for basic auth",
)

func PrintUsageAndExit() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(1)
}

func main() {
	flag.Parse()
	problems := []string{}
	if len(*receptorAddress) == 0 {
		problems = append(problems, "-receptorAddress")
	}
	if len(*serverAddress) == 0 {
		problems = append(problems, "-address")
	}

	if len(problems) > 0 {
		fmt.Fprintf(os.Stderr, "Missing arguments: %s\n\n", strings.Join(problems, ", "))
		PrintUsageAndExit()
	}

	logger := cf_lager.New("teapot")
	logger.Info("starting", lager.Data{"listen_address": *serverAddress, "receptor_address": *receptorAddress})
	receptorClient := receptor.NewClient(*receptorAddress)
	workstationManager := managers.NewWorkstationManager(receptorClient, logger)
	handler := handlers.New(workstationManager, logger, *username, *password)

	members := grouper.Members{
		{"server", http_server.New(*serverAddress, handler)},
	}

	group := grouper.NewOrdered(os.Interrupt, members)

	monitor := ifrit.Invoke(sigmon.New(group))

	logger.Info("started")

	err := <-monitor.Wait()
	if err != nil {
		logger.Error("exited-with-failure", err)
		os.Exit(1)
	}
	logger.Info("exited")
}
