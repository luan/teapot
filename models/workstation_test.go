package models_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/luan/teapot"
	. "github.com/luan/teapot/models"
)

var _ = Describe("Workstation", func() {
	var workstation Workstation

	Describe("NewWorkstation", func() {
		It("defaults dockerImage to something valid if not set", func() {
			workstation = NewWorkstation(teapot.WorkstationCreateRequest{})
			Expect(workstation.DockerImage).To(Equal("docker:///ubuntu#trusty"))
		})

		It("defaults State to STOPPED", func() {
			workstation = NewWorkstation(teapot.WorkstationCreateRequest{})
			Expect(workstation.State).To(Equal(StoppedState))
		})
	})

	Describe("Validate", func() {
		Context("when the workstation has a valid name and docker_image", func() {
			It("is valid", func() {
				workstation = Workstation{
					Name:        "w_-1.1",
					DockerImage: "docker:///a/b#c-1.1",
				}

				err := workstation.Validate()
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when the workstation name is present but invalid", func() {
			It("returns an error indicating so", func() {
				workstation = Workstation{Name: "invalid/guid"}

				err := workstation.Validate()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("name"))
			})
		})

		for _, testCase := range []ValidatorErrorCase{
			{"name",
				Workstation{},
			},
			{"name",
				Workstation{Name: "a b"},
			},
			{"docker_image",
				Workstation{Name: "a", DockerImage: "blah"},
			},
			{"docker_image",
				Workstation{Name: "a", DockerImage: "http://example.com"},
			},
			{"docker_image",
				Workstation{Name: "a", DockerImage: "docker://ubuntu#trusty"},
			},
			{"docker_image",
				Workstation{Name: "a", DockerImage: "docker:///ubuntu:trusty"},
			},
		} {
			testValidatorErrorCase(testCase)
		}
	})
})
