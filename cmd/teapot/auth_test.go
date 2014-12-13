package main_test

import (
	"net/http"

	"github.com/luan/teapot/cmd/teapot/testrunner"
	"github.com/tedsuo/ifrit/ginkgomon"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Basic Auth", func() {
	JustBeforeEach(func() {
		teapotProcess = ginkgomon.Invoke(teapotRunner)
	})

	AfterEach(func() {
		ginkgomon.Kill(teapotProcess)
	})

	Context("when a request without auth is made", func() {
		var res *http.Response
		JustBeforeEach(func() {
			var err error
			httpClient := new(http.Client)
			res, err = httpClient.Get("http://" + teapotAddress)
			Expect(err).NotTo(HaveOccurred())
			res.Body.Close()
		})

		Context("when the username and password have been set", func() {
			It("returns 401 for all requests", func() {
				Expect(res.StatusCode).To(Equal(http.StatusUnauthorized))
			})
		})

		Context("and the username and password have not been set", func() {
			BeforeEach(func() {
				teapotArgs.Username = ""
				teapotArgs.Password = ""
				teapotRunner = testrunner.New(teapotBinPath, teapotArgs)
			})

			It("does not return 401", func() {
				Expect(res.StatusCode).To(Equal(http.StatusNotFound))
			})
		})
	})
})
