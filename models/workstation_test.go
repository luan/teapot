package models_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/luan/teapot/models"
)

var _ = Describe("Workstation", func() {
	var workstation Workstation

	Describe("NewWorkstation", func() {
		It("defaults dockerImage to something valid", func() {
			workstation = NewWorkstation()
			Expect(workstation.DockerImage).To(Equal("docker:///ubuntu#trusty"))
		})

		It("defaults dockerImage to something valid", func() {
			workstation = NewWorkstation("a-name", "")
			Expect(workstation.DockerImage).To(Equal("docker:///ubuntu#trusty"))
		})

		It("defaults State to STOPPED", func() {
			workstation = NewWorkstation()
			Expect(workstation.State).To(Equal("STOPPED"))
		})
	})

	Describe("Validate", func() {
		Context("when the workstation has a valid name and docker_image", func() {
			It("is valid", func() {
				workstation = NewWorkstation("w_-1.1", "docker:///a#b-1.1")

				err := workstation.Validate()
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when the workstation name is present but invalid", func() {
			It("returns an error indicating so", func() {
				workstation = NewWorkstation("invalid/guid")

				err := workstation.Validate()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("name"))
			})
		})

		for _, testCase := range []ValidatorErrorCase{
			{"name",
				NewWorkstation(),
			},
			{"name",
				NewWorkstation("a b"),
			},
			{"docker_image",
				NewWorkstation("a", "blah"),
			},
			{"docker_image",
				NewWorkstation("a", "http://example.com"),
			},
			{"docker_image",
				NewWorkstation("a", "docker://ubuntu#trusty"),
			},
			{"docker_image",
				NewWorkstation("a", "docker:///ubuntu:trusty"),
			},
		} {
			testValidatorErrorCase(testCase)
		}
	})
})
