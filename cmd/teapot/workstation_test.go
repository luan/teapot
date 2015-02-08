package main_test

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/route-emitter/cfroutes"
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

			bytes, _ := json.Marshal(cfroutes.CFRoutes{
				{Hostnames: []string{"tiego-my-workstation.tiego.com"}, Port: 3000},
			})
			routeJson := json.RawMessage(bytes)
			routingInfo := receptor.RoutingInfo{}
			routingInfo[cfroutes.CF_ROUTER] = &routeJson

			openRule := models.SecurityGroupRule{
				Protocol:     models.AllProtocol,
				Destinations: []string{"0.0.0.0/0"},
			}
			openRules := []models.SecurityGroupRule{openRule}

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
						CPUWeight:  2,
						DiskMB:     1024,
						MemoryMB:   512,
						LogGuid:    "my-workstation",
						LogSource:  "TEAPOT-WORKSTATION",
						Ports:      []uint16{8080, 3000},
						Routes:     routingInfo,
						Action: &models.RunAction{
							Path:       "/tmp/tea",
							LogSource:  "TEA",
							Privileged: false,
						},
						EgressRules: openRules,
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
			desiredLRPsRoute, _ := receptor.Routes.FindRouteByName(receptor.DesiredLRPsRoute)
			actualLRPsRoute, _ := receptor.Routes.FindRouteByName(receptor.ActualLRPsRoute)
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

			_, listErr = client.ListWorkstations()
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
							Address: teaHost,
							Ports:   []receptor.PortMapping{{HostPort: uint16(teaPort)}},
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
		CPUWeight:   2,
		DiskMB:      1024,
		MemoryMB:    512,
	}
}
