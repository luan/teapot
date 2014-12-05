package testrunner

import (
	"os/exec"

	"github.com/tedsuo/ifrit/ginkgomon"
)

type Args struct {
	Address         string
	ReceptorAddress string
}

func (args Args) ArgSlice() []string {
	return []string{
		"-address", args.Address,
		"-receptorAddress", args.ReceptorAddress,
	}
}

func New(binPath string, args Args) *ginkgomon.Runner {
	return ginkgomon.New(ginkgomon.Config{
		Name:       "teapot",
		Command:    exec.Command(binPath, args.ArgSlice()...),
		StartCheck: "started",
	})
}
