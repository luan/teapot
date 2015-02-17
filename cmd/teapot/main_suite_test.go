package main_test

import (
	"fmt"
	"net/url"

	"github.com/luan/teapot"
	"github.com/luan/teapot/cmd/teapot/testrunner"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal-golang/lager"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/ginkgomon"

	"testing"
	"time"
)

const (
	username   = "username"
	password   = "password"
	appsDomain = "tiego.com"
	teaSecret  = "s3cret"
)

var logger lager.Logger

var teapotBinPath string

var teapotAddress string
var teapotArgs testrunner.Args
var teapotRunner *ginkgomon.Runner
var teapotProcess ifrit.Process
var client teapot.Client
var receptorServer *ghttp.Server

func TestTeapot(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Teapot Suite")
}

var _ = SynchronizedBeforeSuite(
	func() []byte {
		teapotConfig, err := gexec.Build("github.com/luan/teapot/cmd/teapot", "-race")
		Expect(err).NotTo(HaveOccurred())
		return []byte(teapotConfig)
	},
	func(teapotConfig []byte) {
		teapotBinPath = string(teapotConfig)
		SetDefaultEventuallyTimeout(15 * time.Second)
	},
)

var _ = SynchronizedAfterSuite(func() {
}, func() {
	gexec.CleanupBuildArtifacts()
})

var _ = BeforeEach(func() {
	receptorServer = ghttp.NewServer()
	logger = lager.NewLogger("test")

	teapotAddress = fmt.Sprintf("127.0.0.1:%d", 6700+GinkgoParallelNode())

	teapotURL := &url.URL{
		Scheme: "http",
		Host:   teapotAddress,
		User:   url.UserPassword(username, password),
	}

	client = teapot.NewClient(teapotURL.String())

	teapotArgs = testrunner.Args{
		Address:         teapotAddress,
		ReceptorAddress: receptorServer.URL(),
		Username:        username,
		Password:        password,
		AppsDomain:      appsDomain,
		TEASecret:       teaSecret,
	}
	teapotRunner = testrunner.New(teapotBinPath, teapotArgs)
})

var _ = AfterEach(func() {
	receptorServer.Close()
})
