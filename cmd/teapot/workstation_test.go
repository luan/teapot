package main_test

import (
	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/luan/teapot"
	"github.com/tedsuo/ifrit/ginkgomon"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("Workstation API", func() {
	BeforeEach(func() {
		teapotProcess = ginkgomon.Invoke(teapotRunner)
	})

	AfterEach(func() {
		ginkgomon.Kill(teapotProcess)
	})

	Describe("POST /workstatations/", func() {
		var workstationToCreate teapot.WorkstationCreateRequest
		var createErr error

		BeforeEach(func() {
			createDesiredLRPRoute, _ := receptor.Routes.FindRouteByName(receptor.CreateDesiredLRPRoute)
			receptorServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(createDesiredLRPRoute.Method, createDesiredLRPRoute.Path),
					ghttp.VerifyJSONRepresenting(receptor.DesiredLRPCreateRequest{
						ProcessGuid: "my-workstation",
						Domain:      "teapot",
						Instances:   1,
						Stack:       "lucid64",
						RootFSPath:  "docker:///debian#wheezy",
						DiskMB:      128,
						MemoryMB:    64,
						LogGuid:     "my-workstation",
						LogSource:   "TEAPOT-WORKSTATION",
						Action: &models.RunAction{
							Path:      "/bin/sh",
							LogSource: "TEA",
						},
					}),
				),
			)

			workstationToCreate = newValidWorkstationCreateRequest()
			createErr = client.CreateWorkstation(workstationToCreate)
		})

		It("responds without an error", func() {
			Expect(createErr).NotTo(HaveOccurred())
		})

		It("requests an LRP from the receptor", func() {
			Expect(receptorServer.ReceivedRequests()).To(HaveLen(1))
		})
	})
})

func newValidWorkstationCreateRequest() teapot.WorkstationCreateRequest {
	return teapot.WorkstationCreateRequest{
		Name:        "my-workstation",
		DockerImage: "docker:///debian#wheezy",
	}
}