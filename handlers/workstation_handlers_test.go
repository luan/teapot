package handlers_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/fake_receptor"
	"github.com/luan/teapot"
	. "github.com/luan/teapot/handlers"
	"github.com/luan/teapot/managers"
	"github.com/pivotal-golang/lager"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("WorkstationHandler", func() {
	var (
		logger             lager.Logger
		responseRecorder   *httptest.ResponseRecorder
		handler            *WorkstationHandler
		fakeReceptorClient *fake_receptor.FakeClient
		manager            managers.WorkstationManager
	)

	BeforeEach(func() {
		logger = lager.NewLogger("test")
		logger.RegisterSink(lager.NewWriterSink(GinkgoWriter, lager.DEBUG))
		responseRecorder = httptest.NewRecorder()
		fakeReceptorClient = new(fake_receptor.FakeClient)
		manager = managers.NewWorkstationManager(fakeReceptorClient, logger)
		handler = NewWorkstationHandler(manager, logger)
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

		Context("when the workstation already exists", func() {
			JustBeforeEach(func() {
				fakeReceptorClient.GetDesiredLRPReturns(receptor.DesiredLRPResponse{ProcessGuid: validCreateRequest.Name}, nil)
				handler.Create(responseRecorder, newTestRequest(validCreateRequest))
			})

			It("fails with a 400 BAD REQUEST", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusBadRequest))
			})

			It("responds with an an error including the validation details", func() {
				expectedBody, _ := json.Marshal(receptor.Error{
					Type:    teapot.InvalidWorkstation,
					Message: "Unique constraint failed for: name",
				})
				Expect(responseRecorder.Body.String()).To(Equal(string(expectedBody)))
			})
		})

		Context("when the requested workstation is invalid", func() {
			BeforeEach(func() {
				handler.Create(responseRecorder, newTestRequest(invalidCreateRequest))
			})

			It("responds with 400 BAD REQUEST", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusBadRequest))
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

	Describe("Delete", func() {
		var req *http.Request

		BeforeEach(func() {
			req = newTestRequest("")
			req.Form = url.Values{":process_guid": []string{"workstation-name"}}
		})

		Context("when everything succeeds", func() {
			BeforeEach(func() {
				handler.Delete(responseRecorder, req)
			})

			It("responds with 204 NO CONTENT", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusNoContent))
			})

			It("responds with an empty body", func() {
				Expect(responseRecorder.Body.String()).To(Equal(""))
			})
		})

		Context("when the workstation doesn't exists", func() {
			BeforeEach(func() {
				fakeReceptorClient.DeleteDesiredLRPReturns(errors.New("receptor error"))
				handler.Delete(responseRecorder, req)
			})

			It("fails with a 404 NOT FOUND", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusNotFound))
			})
		})
	})

	Describe("Attach", func() {
		var req *http.Request

		BeforeEach(func() {
			req = newTestRequest("")
			req.Form = url.Values{":process_guid": []string{"workstation-name"}}
		})

		Context("when the workstation doesn't exists", func() {
			BeforeEach(func() {
				handler.Attach(responseRecorder, req)
			})

			It("fails with a 404 NOT FOUND", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusNotFound))
			})
		})

		Context("when the receptor returns an error", func() {
			BeforeEach(func() {
				fakeReceptorClient.ActualLRPsByProcessGuidReturns(nil, errors.New("receptor error"))
				handler.Attach(responseRecorder, req)
			})

			It("fails with a 404 NOT FOUND", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusNotFound))
			})
		})
	})
})
