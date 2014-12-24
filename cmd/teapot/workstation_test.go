package main_test

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/gorilla/websocket"
	"github.com/luan/teapot"
	"github.com/tedsuo/ifrit/ginkgomon"
	"github.com/tedsuo/rata"

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
			workstationToCreate = newValidWorkstationCreateRequest()

			createDesiredLRPRoute, _ := receptor.Routes.FindRouteByName(receptor.CreateDesiredLRPRoute)
			getDesiredLRPRoute, _ := receptor.Routes.FindRouteByName(receptor.GetDesiredLRPRoute)
			getDesiredLRPPath, _ := getDesiredLRPRoute.CreatePath(rata.Params{"process_guid": workstationToCreate.Name})
			receptorServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(getDesiredLRPRoute.Method, getDesiredLRPPath),
					ghttp.RespondWith(http.StatusNotFound, ""),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(createDesiredLRPRoute.Method, createDesiredLRPRoute.Path),
					ghttp.VerifyJSONRepresenting(receptor.DesiredLRPCreateRequest{
						ProcessGuid: "my-workstation",
						Setup: &models.SerialAction{
							Actions: []models.Action{
								&models.DownloadAction{
									From:     "https://tiego-artifacts.s3.amazonaws.com/tea-builds/tea-latest.tgz",
									To:       "/tmp",
									CacheKey: "tea",
								},
							},
						},
						Domain:     "tiego",
						Instances:  1,
						Stack:      "lucid64",
						RootFSPath: "docker:///debian#wheezy",
						DiskMB:     128,
						MemoryMB:   64,
						LogGuid:    "my-workstation",
						LogSource:  "TEAPOT-WORKSTATION",
						Ports:      []uint32{8080},
						Action: &models.RunAction{
							Path:       "/tmp/tea",
							LogSource:  "TEA",
							Privileged: false,
						},
					}),
				),
			)

			createErr = client.CreateWorkstation(workstationToCreate)
		})

		It("responds without an error", func() {
			Expect(createErr).NotTo(HaveOccurred())
		})

		It("requests an LRP from the receptor", func() {
			Expect(receptorServer.ReceivedRequests()).To(HaveLen(2))
		})
	})

	Describe("GET /workstations/", func() {
		var listErr error

		BeforeEach(func() {
			desiredLRPsRoute, _ := receptor.Routes.FindRouteByName(receptor.DesiredLRPsByDomainRoute)
			actualLRPsRoute, _ := receptor.Routes.FindRouteByName(receptor.ActualLRPsByDomainRoute)
			desiredLRPsPath, _ := desiredLRPsRoute.CreatePath(rata.Params{"domain": "tiego"})
			actualLRPsPath, _ := actualLRPsRoute.CreatePath(rata.Params{"domain": "tiego"})
			receptorServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(desiredLRPsRoute.Method, desiredLRPsPath),
					ghttp.RespondWith(http.StatusOK, ""),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(actualLRPsRoute.Method, actualLRPsPath),
					ghttp.RespondWith(http.StatusOK, ""),
				),
			)

			listErr = client.ListWorkstations()
		})

		It("responds without an error", func() {
			Expect(listErr).NotTo(HaveOccurred())
		})

		It("requests the LRPs from the receptor", func() {
			Expect(receptorServer.ReceivedRequests()).To(HaveLen(2))
		})
	})

	Describe("DELETE /workstatations/:name", func() {
		var deleteErr error

		BeforeEach(func() {
			workstationToDelete := "w1"
			deleteDesiredLRPRoute, _ := receptor.Routes.FindRouteByName(receptor.DeleteDesiredLRPRoute)
			deleteDesiredLRPPath, _ := deleteDesiredLRPRoute.CreatePath(rata.Params{"process_guid": workstationToDelete})
			receptorServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(deleteDesiredLRPRoute.Method, deleteDesiredLRPPath),
					ghttp.RespondWith(http.StatusNoContent, ""),
				),
			)

			deleteErr = client.DeleteWorkstation(workstationToDelete)
		})

		It("responds without an error", func() {
			Expect(deleteErr).NotTo(HaveOccurred())
		})

		It("requests an LRP from the receptor", func() {
			Expect(receptorServer.ReceivedRequests()).To(HaveLen(1))
		})
	})

	Describe("GET /workstatations/:name/attach", func() {
		var (
			attachErr error
			teaServer *ghttp.Server
			ws        *websocket.Conn
		)

		BeforeEach(func() {
			teaServer = ghttp.NewServer()
			teaURL, _ := url.Parse(teaServer.URL())
			teaHostPort := strings.Split(teaURL.Host, ":")
			teaHost := teaHostPort[0]
			teaPort, _ := strconv.Atoi(teaHostPort[1])

			workstationToAttach := "w1"
			actualLRPsByProcessGuidRoute, _ := receptor.Routes.FindRouteByName(receptor.ActualLRPsByProcessGuidRoute)
			actualLRPsByProcessGuidPath, _ := actualLRPsByProcessGuidRoute.CreatePath(rata.Params{"process_guid": workstationToAttach})
			receptorServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(actualLRPsByProcessGuidRoute.Method, actualLRPsByProcessGuidPath),
					ghttp.RespondWithJSONEncoded(http.StatusOK, []receptor.ActualLRPResponse{
						{
							Host:  teaHost,
							Ports: []receptor.PortMapping{{HostPort: uint32(teaPort)}},
						},
					}),
				),
			)
			teaServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/shell"),
					func(w http.ResponseWriter, r *http.Request) {
						upgrader := websocket.Upgrader{
							CheckOrigin: func(r *http.Request) bool { return true },
						}
						ws, err := upgrader.Upgrade(w, r, nil)
						if err != nil {
							panic(err)
						}
						defer ws.Close()
						_, m, err := ws.ReadMessage()
						Expect(err).NotTo(HaveOccurred())
						Expect(string(m)).To(Equal("hello"))
						ws.WriteMessage(websocket.TextMessage, []byte("world"))
					},
				),
			)

			ws, attachErr = client.AttachWorkstation(workstationToAttach)
		})

		AfterEach(func() {
			teaServer.Close()
		})

		It("responds without an error", func() {
			Expect(attachErr).NotTo(HaveOccurred())
		})

		It("proxies to the TEA API", func() {
			Expect(ws).NotTo(BeNil())
			ws.WriteMessage(websocket.TextMessage, []byte("hello"))
			_, m, err := ws.ReadMessage()
			Expect(err).NotTo(HaveOccurred())
			Expect(string(m)).To(Equal("world"))
		})

		It("requests an actual LRP from the receptor", func() {
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
