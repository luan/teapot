package task_handler_test

import (
	"errors"
	"net/http"
	"net/url"
	"os"

	"github.com/cloudfoundry-incubator/receptor/task_handler"
	"github.com/cloudfoundry-incubator/runtime-schema/bbs/fake_bbs"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/pivotal-golang/lager"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/ginkgomon"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("TaskWorker", func() {
	var (
		fakeBBS *fake_bbs.FakeReceptorBBS
		enqueue chan<- models.Task

		process ifrit.Process

		fakeServer *ghttp.Server
	)

	BeforeEach(func() {
		fakeServer = ghttp.NewServer()

		logger := lager.NewLogger("task-watcher-test")
		logger.RegisterSink(lager.NewWriterSink(GinkgoWriter, lager.INFO))

		fakeBBS = new(fake_bbs.FakeReceptorBBS)

		taskWorker, enqueueTasks := task_handler.NewTaskWorkerPool(fakeBBS, logger)

		enqueue = enqueueTasks

		process = ifrit.Invoke(taskWorker)
	})

	AfterEach(func() {
		fakeServer.Close()
		ginkgomon.Kill(process)
	})

	Describe("shutting down", func() {
		Context("when sent Interrupt", func() {
			BeforeEach(func() {
				process.Signal(os.Interrupt)
			})

			It("exits", func() {
				Eventually(process.Wait()).Should(Receive(BeNil()))
			})
		})

		Context("when sent Kill", func() {
			BeforeEach(func() {
				process.Signal(os.Kill)
			})

			It("exits", func() {
				Eventually(process.Wait()).Should(Receive())
			})
		})
	})

	Describe("when a task is enqueued", func() {
		var (
			callbackURL *url.URL
			statusCodes chan int
			reqCount    chan struct{}
		)

		BeforeEach(func() {
			statusCodes = make(chan int)
			reqCount = make(chan struct{}, task_handler.POOL_SIZE)
			fakeServer.RouteToHandler("POST", "/the-callback/url", func(w http.ResponseWriter, req *http.Request) {
				reqCount <- struct{}{}
				w.WriteHeader(<-statusCodes)
			})

			var err error
			callbackURL, err = url.Parse(fakeServer.URL() + "/the-callback/url")
			Ω(err).ShouldNot(HaveOccurred())
		})

		AfterEach(func() {
			close(statusCodes)
		})

		simulateTaskCompleting := func() {
			enqueue <- models.Task{
				TaskGuid:              "the-task-guid",
				CompletionCallbackURL: callbackURL,
				Action: &models.RunAction{
					Path: "lol",
				},
			}
		}

		Context("when the task has a completion callback URL", func() {
			It("marks the task as resolving", func() {
				Ω(fakeBBS.ResolvingTaskCallCount()).Should(Equal(0))

				simulateTaskCompleting()
				statusCodes <- 200

				Eventually(fakeBBS.ResolvingTaskCallCount).Should(Equal(1))
				Ω(fakeBBS.ResolvingTaskArgsForCall(0)).Should(Equal("the-task-guid"))
			})

			It("processes tasks in parallel", func() {
				for i := 0; i < task_handler.POOL_SIZE; i++ {
					simulateTaskCompleting()
				}
				Eventually(reqCount).Should(HaveLen(task_handler.POOL_SIZE))
			})

			Context("when marking the task as resolving fails", func() {
				BeforeEach(func() {
					fakeBBS.ResolvingTaskReturns(errors.New("failed to resolve task"))
				})

				It("does not make a request to the task's callback URL", func() {
					simulateTaskCompleting()

					Consistently(fakeServer.ReceivedRequests, 0.25).Should(BeEmpty())
				})
			})

			Context("when marking the task as resolving succeeds", func() {
				It("POSTs to the task's callback URL", func() {
					simulateTaskCompleting()

					statusCodes <- 200

					Eventually(fakeServer.ReceivedRequests).Should(HaveLen(1))
				})

				Context("when the request succeeds", func() {
					It("resolves the task", func() {
						simulateTaskCompleting()

						statusCodes <- 200

						Eventually(fakeBBS.ResolveTaskCallCount).Should(Equal(1))
						Ω(fakeBBS.ResolveTaskArgsForCall(0)).Should(Equal("the-task-guid"))
					})
				})

				Context("when the request fails with a 4xx response code", func() {
					It("resolves the task", func() {
						simulateTaskCompleting()

						statusCodes <- 403

						Eventually(fakeBBS.ResolveTaskCallCount).Should(Equal(1))
						Ω(fakeBBS.ResolveTaskArgsForCall(0)).Should(Equal("the-task-guid"))
					})
				})

				Context("when the request fails with a 500 response code", func() {
					It("resolves the task", func() {
						simulateTaskCompleting()

						statusCodes <- 500

						Eventually(fakeBBS.ResolveTaskCallCount).Should(Equal(1))
						Ω(fakeBBS.ResolveTaskArgsForCall(0)).Should(Equal("the-task-guid"))
					})
				})

				Context("when the request fails with a 503 or 504 response code", func() {
					It("retries the request 2 more times", func() {
						simulateTaskCompleting()
						Eventually(fakeServer.ReceivedRequests).Should(HaveLen(1))

						statusCodes <- 503

						Consistently(fakeBBS.ResolveTaskCallCount, 0.25).Should(Equal(0))
						Eventually(fakeServer.ReceivedRequests).Should(HaveLen(2))

						statusCodes <- 504

						Consistently(fakeBBS.ResolveTaskCallCount, 0.25).Should(Equal(0))
						Eventually(fakeServer.ReceivedRequests).Should(HaveLen(3))

						statusCodes <- 200

						Eventually(fakeBBS.ResolveTaskCallCount, 0.25).Should(Equal(1))
						Ω(fakeBBS.ResolveTaskArgsForCall(0)).Should(Equal("the-task-guid"))
					})

					Context("when the request fails every time", func() {
						It("does not resolve the task", func() {
							simulateTaskCompleting()
							Eventually(fakeServer.ReceivedRequests).Should(HaveLen(1))

							statusCodes <- 503

							Consistently(fakeBBS.ResolveTaskCallCount, 0.25).Should(Equal(0))
							Eventually(fakeServer.ReceivedRequests).Should(HaveLen(2))

							statusCodes <- 504

							Consistently(fakeBBS.ResolveTaskCallCount, 0.25).Should(Equal(0))
							Eventually(fakeServer.ReceivedRequests).Should(HaveLen(3))

							statusCodes <- 503

							Consistently(fakeBBS.ResolveTaskCallCount, 0.25).Should(Equal(0))
							Consistently(fakeServer.ReceivedRequests, 0.25).Should(HaveLen(3))
						})
					})
				})
			})
		})

		Context("when the task doesn't have a completion callback URL", func() {
			BeforeEach(func() {
				callbackURL = nil
				simulateTaskCompleting()
			})

			It("does not mark the task as resolving", func() {
				Consistently(fakeBBS.ResolvingTaskCallCount).Should(Equal(0))
			})
		})
	})
})
