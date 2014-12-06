package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/cloudfoundry-incubator/receptor/fake_receptor"
	"github.com/luan/teapot"
	. "github.com/luan/teapot/handlers"
	"github.com/pivotal-golang/lager"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("WorkstationHandler", func() {
	var (
		logger           lager.Logger
		responseRecorder *httptest.ResponseRecorder
		handler          *WorkstationHandler
		// request            *http.Request
		fakeReceptorClient *fake_receptor.FakeClient
	)

	BeforeEach(func() {
		logger = lager.NewLogger("test")
		logger.RegisterSink(lager.NewWriterSink(GinkgoWriter, lager.DEBUG))
		responseRecorder = httptest.NewRecorder()
		fakeReceptorClient = new(fake_receptor.FakeClient)
		handler = NewWorkstationHandler(fakeReceptorClient, logger)
	})

	Describe("Create", func() {
		validCreateRequest := teapot.WorkstationCreateRequest{
			Name:        "workstation-name-1",
			DockerImage: "docker:///docker",
		}

		invalidCreateRequest := teapot.WorkstationCreateRequest{
			DockerImage: "docker:///docker",
		}

		Context("when everything succeeds", func() {
			JustBeforeEach(func() {
				handler.Create(responseRecorder, newTestRequest(validCreateRequest))
			})

			It("responds with 201 CREATED", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusCreated))
			})

			It("responds with an empty body", func() {
				Expect(responseRecorder.Body.String()).To(Equal(""))
			})
		})

		Context("when the requested workstation is invalid", func() {
			BeforeEach(func() {
				handler.Create(responseRecorder, newTestRequest(invalidCreateRequest))
			})

			It("responds with 418 I'M A TEAPOT", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusTeapot))
			})

			It("responds with a relevant error message", func() {
				expectedBody, _ := json.Marshal(teapot.Error{
					Type:    teapot.InvalidWorkstation,
					Message: "Invalid field: name",
				})
				Expect(responseRecorder.Body.String()).To(Equal(string(expectedBody)))
			})
		})

		Context("when the request does not contain a WorkstationCreateRequest", func() {
			var garbageRequest = []byte(`hello`)

			BeforeEach(func() {
				handler.Create(responseRecorder, newTestRequest(garbageRequest))
			})

			It("responds with 400 BAD REQUEST", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusBadRequest))
			})

			It("responds with a relevant error message", func() {
				err := json.Unmarshal(garbageRequest, &teapot.WorkstationCreateRequest{})
				expectedBody, _ := json.Marshal(teapot.Error{
					Type:    teapot.InvalidJSON,
					Message: err.Error(),
				})
				Expect(responseRecorder.Body.String()).To(Equal(string(expectedBody)))
			})
		})
	})
})
