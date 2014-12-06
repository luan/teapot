package models_test

import (
	"github.com/luan/teapot/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

type ValidatorErrorCase struct {
	Message string
	models.Validator
}

func testValidatorErrorCase(testCase ValidatorErrorCase) {
	message := testCase.Message
	invalid := testCase.Validator

	Context("when invalid", func() {
		It("returns an error indicating '"+message+"'", func() {
			err := invalid.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(message))
		})
	})
}

func TestModels(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Models Suite")
}
