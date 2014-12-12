package metrics_test

import (
	"github.com/cloudfoundry/dropsonde/metric_sender/fake"
	"github.com/cloudfoundry/dropsonde/metrics"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Metrics", func() {
	var fakeMetricSender *fake.FakeMetricSender

	BeforeEach(func() {
		fakeMetricSender = fake.NewFakeMetricSender()
		metrics.Initialize(fakeMetricSender)
	})

	It("delegates SendValue", func() {
		metrics.SendValue("metric", 42.42, "answers")

		Expect(fakeMetricSender.GetValue("metric").Value).To(Equal(42.42))
		Expect(fakeMetricSender.GetValue("metric").Unit).To(Equal("answers"))
	})

	It("delegates IncrementCounter", func() {
		metrics.IncrementCounter("count")

		Expect(fakeMetricSender.GetCounter("count")).To(BeEquivalentTo(1))

		metrics.IncrementCounter("count")

		Expect(fakeMetricSender.GetCounter("count")).To(BeEquivalentTo(2))
	})

	It("delegates AddToCounter", func() {
		metrics.AddToCounter("count", 5)

		Expect(fakeMetricSender.GetCounter("count")).To(BeEquivalentTo(5))

		metrics.AddToCounter("count", 10)

		Expect(fakeMetricSender.GetCounter("count")).To(BeEquivalentTo(15))
	})

	Context("when Metric Sender is not initialized", func() {

		BeforeEach(func() {
			metrics.Initialize(nil)
		})

		It("SendValue is a no-op", func() {
			err := metrics.SendValue("metric", 42.42, "answers")

			Expect(err).ToNot(HaveOccurred())
		})

		It("IncrementCounter is a no-op", func() {
			err := metrics.IncrementCounter("count")

			Expect(err).ToNot(HaveOccurred())
		})

		It("AddToCounter is a no-op", func() {
			err := metrics.AddToCounter("count", 10)

			Expect(err).ToNot(HaveOccurred())
		})

	})
})
