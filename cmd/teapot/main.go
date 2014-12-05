package main

import (
	"flag"
	"os"

	cf_lager "github.com/cloudfoundry-incubator/cf-lager"
	"github.com/cloudfoundry-incubator/receptor"
	"github.com/luan/teapot/handlers"
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

func main() {
	logger := cf_lager.New("teapot")
	logger.Info("starting")
	receptorClient := receptor.NewClient(*receptorAddress)
	handler := handlers.New(receptorClient, logger)

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
