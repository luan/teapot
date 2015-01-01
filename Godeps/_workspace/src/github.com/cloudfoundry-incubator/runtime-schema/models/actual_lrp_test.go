package models_test

import (
	"encoding/json"
	"fmt"

	"github.com/cloudfoundry-incubator/runtime-schema/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ActualLRP", func() {
	Describe("ActualLRPKey", func() {
		Describe("Validate", func() {
			var actualLRPKey models.ActualLRPKey

			BeforeEach(func() {
				actualLRPKey = models.NewActualLRPKey("process-guid", 1, "domain")
			})

			Context("when valid", func() {
				It("returns nil", func() {
					Ω(actualLRPKey.Validate()).Should(BeNil())
				})
			})

			Context("when the ProcessGuid is blank", func() {
				BeforeEach(func() {
					actualLRPKey.ProcessGuid = ""
				})

				It("returns a validation error", func() {
					Ω(actualLRPKey.Validate()).Should(ConsistOf(models.ErrInvalidField{"process_guid"}))
				})
			})

			Context("when the Domain is blank", func() {
				BeforeEach(func() {
					actualLRPKey.Domain = ""
				})

				It("returns a validation error", func() {
					Ω(actualLRPKey.Validate()).Should(ConsistOf(models.ErrInvalidField{"domain"}))
				})
			})

			Context("when the Index is negative", func() {
				BeforeEach(func() {
					actualLRPKey.Index = -1
				})

				It("returns a validation error", func() {
					Ω(actualLRPKey.Validate()).Should(ConsistOf(models.ErrInvalidField{"index"}))
				})
			})
		})
	})

	Describe("ActualLRPContainerKey", func() {
		Describe("Validate", func() {
			var actualLRPContainerKey models.ActualLRPContainerKey

			Context("when both instance guid and cell id are specified", func() {
				It("returns nil", func() {
					actualLRPContainerKey = models.NewActualLRPContainerKey("instance-guid", "cell-id")
					Ω(actualLRPContainerKey.Validate()).Should(BeNil())
				})
			})

			Context("when both instance guid and cell id are empty", func() {
				It("returns a validation error", func() {
					actualLRPContainerKey = models.NewActualLRPContainerKey("", "")
					Ω(actualLRPContainerKey.Validate()).Should(ConsistOf(
						models.ErrInvalidField{"cell_id"},
						models.ErrInvalidField{"instance_guid"},
					))
				})
			})

			Context("when only the instance guid is specified", func() {
				It("returns a validation error", func() {
					actualLRPContainerKey = models.NewActualLRPContainerKey("instance-guid", "")
					Ω(actualLRPContainerKey.Validate()).Should(ConsistOf(models.ErrInvalidField{"cell_id"}))
				})
			})

			Context("when only the cell id is specified", func() {
				It("returns a validation error", func() {
					actualLRPContainerKey = models.NewActualLRPContainerKey("", "cell-id")
					Ω(actualLRPContainerKey.Validate()).Should(ConsistOf(models.ErrInvalidField{"instance_guid"}))
				})
			})
		})
	})

	Describe("ActualLRP", func() {
		var lrp models.ActualLRP
		var lrpKey models.ActualLRPKey
		var containerKey models.ActualLRPContainerKey
		var netInfo models.ActualLRPNetInfo

		BeforeEach(func() {
		})
		lrpPayload := `{
    "process_guid":"some-guid",
    "instance_guid":"some-instance-guid",
    "address": "1.2.3.4",
    "ports": [
      { "container_port": 8080 },
      { "container_port": 8081, "host_port": 1234 }
    ],
    "index": 2,
    "state": "RUNNING",
    "since": 1138,
    "cell_id":"some-cell-id",
    "domain":"some-domain"
  }`

		BeforeEach(func() {
			lrpKey = models.NewActualLRPKey("some-guid", 2, "some-domain")
			containerKey = models.NewActualLRPContainerKey("some-instance-guid", "some-cell-id")
			netInfo = models.NewActualLRPNetInfo("1.2.3.4", []models.PortMapping{
				{ContainerPort: 8080},
				{ContainerPort: 8081, HostPort: 1234},
			})

			lrp = models.ActualLRP{
				ActualLRPKey:          lrpKey,
				ActualLRPContainerKey: containerKey,
				ActualLRPNetInfo:      netInfo,
				State:                 models.ActualLRPStateRunning,
				Since:                 1138,
			}
		})

		Describe("To JSON", func() {
			It("should JSONify", func() {
				marshalled, err := json.Marshal(&lrp)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(string(marshalled)).Should(MatchJSON(lrpPayload))
			})
		})

		Describe("FromJSON", func() {
			It("returns a LRP with correct fields", func() {
				aLRP := &models.ActualLRP{}
				err := models.FromJSON([]byte(lrpPayload), aLRP)
				Ω(err).ShouldNot(HaveOccurred())

				Ω(aLRP).Should(Equal(&lrp))
			})

			Context("with an invalid payload", func() {
				It("returns the error", func() {
					aLRP := &models.ActualLRP{}
					err := models.FromJSON([]byte("something lol"), aLRP)
					Ω(err).Should(HaveOccurred())
				})
			})

			for field, payload := range map[string]string{
				"process_guid":  `{"instance_guid": "instance_guid", "cell_id": "cell_id", "domain": "domain"}`,
				"instance_guid": `{"process_guid": "process-guid", "cell_id": "cell_id", "domain": "domain","state":"CLAIMED"}`,
				"cell_id":       `{"process_guid": "process-guid", "instance_guid": "instance_guid", "domain": "domain", "state":"RUNNING"}`,
				"domain":        `{"process_guid": "process-guid", "cell_id": "cell_id", "instance_guid": "instance_guid"}`,
			} {
				missingField := field
				jsonPayload := payload

				Context("when the json is missing a "+missingField, func() {
					It("returns an error indicating so", func() {
						aLRP := &models.ActualLRP{}
						err := models.FromJSON([]byte(jsonPayload), aLRP)
						Ω(err.Error()).Should(ContainSubstring(missingField))
					})
				})
			}
		})

		Describe("AllowsTransitionTo", func() {
			var (
				before   models.ActualLRP
				afterKey models.ActualLRPKey
			)

			BeforeEach(func() {
				before = models.ActualLRP{
					ActualLRPKey: models.NewActualLRPKey("fake-process-guid", 1, "fake-domain"),
				}
				afterKey = before.ActualLRPKey
			})

			Context("when the ProcessGuid fields differ", func() {
				BeforeEach(func() {
					before.ProcessGuid = "some-process-guid"
					afterKey.ProcessGuid = "another-process-guid"
				})

				It("is not allowed", func() {
					Ω(before.AllowsTransitionTo(afterKey, before.ActualLRPContainerKey, before.State)).Should(BeFalse())
				})
			})

			Context("when the Index fields differ", func() {
				BeforeEach(func() {
					before.Index = 1138
					afterKey.Index = 3417
				})

				It("is not allowed", func() {
					Ω(before.AllowsTransitionTo(afterKey, before.ActualLRPContainerKey, before.State)).Should(BeFalse())
				})
			})

			Context("when the Domain fields differ", func() {
				BeforeEach(func() {
					before.Domain = "some-domain"
					afterKey.Domain = "another-domain"
				})

				It("is not allowed", func() {
					Ω(before.AllowsTransitionTo(afterKey, before.ActualLRPContainerKey, before.State)).Should(BeFalse())
				})
			})

			Context("when the ProcessGuid, Index, and Domain are equivalent", func() {
				var (
					emptyKey                 = models.NewActualLRPContainerKey("", "")
					claimedKey               = models.NewActualLRPContainerKey("some-instance-guid", "some-cell-id")
					differentInstanceGuidKey = models.NewActualLRPContainerKey("some-other-instance-guid", "some-cell-id")
					differentCellIDKey       = models.NewActualLRPContainerKey("some-instance-guid", "some-other-cell-id")
				)

				type stateTableEntry struct {
					BeforeState        models.ActualLRPState
					AfterState         models.ActualLRPState
					BeforeContainerKey models.ActualLRPContainerKey
					AfterContainerKey  models.ActualLRPContainerKey
					Allowed            bool
				}

				var EntryToString = func(entry stateTableEntry) string {
					return fmt.Sprintf("is %t when the before has state %s and instance guid '%s' and cell id '%s' and the after has state %s and instance guid '%s' and cell id '%s'",
						entry.Allowed,
						entry.BeforeState,
						entry.BeforeContainerKey.InstanceGuid,
						entry.BeforeContainerKey.CellID,
						entry.AfterState,
						entry.AfterContainerKey.InstanceGuid,
						entry.AfterContainerKey.CellID,
					)
				}

				stateTable := []stateTableEntry{
					{models.ActualLRPStateUnclaimed, models.ActualLRPStateUnclaimed, emptyKey, emptyKey, true},
					{models.ActualLRPStateUnclaimed, models.ActualLRPStateClaimed, emptyKey, claimedKey, true},
					{models.ActualLRPStateUnclaimed, models.ActualLRPStateRunning, emptyKey, claimedKey, true},
					{models.ActualLRPStateClaimed, models.ActualLRPStateUnclaimed, claimedKey, emptyKey, true},
					{models.ActualLRPStateClaimed, models.ActualLRPStateClaimed, claimedKey, claimedKey, true},
					{models.ActualLRPStateClaimed, models.ActualLRPStateClaimed, claimedKey, differentInstanceGuidKey, false},
					{models.ActualLRPStateClaimed, models.ActualLRPStateClaimed, claimedKey, differentCellIDKey, false},
					{models.ActualLRPStateClaimed, models.ActualLRPStateRunning, claimedKey, claimedKey, true},
					{models.ActualLRPStateClaimed, models.ActualLRPStateRunning, claimedKey, differentInstanceGuidKey, true},
					{models.ActualLRPStateClaimed, models.ActualLRPStateRunning, claimedKey, differentCellIDKey, true},
					{models.ActualLRPStateRunning, models.ActualLRPStateUnclaimed, claimedKey, emptyKey, true},
					{models.ActualLRPStateRunning, models.ActualLRPStateClaimed, claimedKey, claimedKey, true},
					{models.ActualLRPStateRunning, models.ActualLRPStateClaimed, claimedKey, differentInstanceGuidKey, false},
					{models.ActualLRPStateRunning, models.ActualLRPStateClaimed, claimedKey, differentCellIDKey, false},
					{models.ActualLRPStateRunning, models.ActualLRPStateRunning, claimedKey, claimedKey, true},
					{models.ActualLRPStateRunning, models.ActualLRPStateClaimed, claimedKey, differentInstanceGuidKey, false},
					{models.ActualLRPStateRunning, models.ActualLRPStateClaimed, claimedKey, differentCellIDKey, false},
				}

				for _, entry := range stateTable {
					entry := entry
					It(EntryToString(entry), func() {
						before.State = entry.BeforeState
						before.ActualLRPContainerKey = entry.BeforeContainerKey
						Ω(before.AllowsTransitionTo(before.ActualLRPKey, entry.AfterContainerKey, entry.AfterState)).Should(Equal(entry.Allowed))
					})
				}
			})
		})

		Describe("Validate", func() {

			Context("when state is unclaimed", func() {
				BeforeEach(func() {
					lrp = models.ActualLRP{
						ActualLRPKey: lrpKey,
						State:        models.ActualLRPStateUnclaimed,
						Since:        1138,
					}
				})

				itValidatesPresenceOfTheLRPKey(&lrp)
				itValidatesAbsenceOfTheContainerKey(&lrp)
				itValidatesAbsenceOfNetInfo(&lrp)
			})

			Context("when state is claimed", func() {
				BeforeEach(func() {
					lrp = models.ActualLRP{
						ActualLRPKey:          lrpKey,
						ActualLRPContainerKey: containerKey,
						State: models.ActualLRPStateClaimed,
						Since: 1138,
					}
				})

				itValidatesPresenceOfTheLRPKey(&lrp)
				itValidatesPresenceOfTheContainerKey(&lrp)
				itValidatesAbsenceOfNetInfo(&lrp)
			})

			Context("when state is running", func() {
				BeforeEach(func() {
					lrp = models.ActualLRP{
						ActualLRPKey:          lrpKey,
						ActualLRPContainerKey: containerKey,
						ActualLRPNetInfo:      netInfo,
						State:                 models.ActualLRPStateRunning,
						Since:                 1138,
					}
				})

				itValidatesPresenceOfTheLRPKey(&lrp)
				itValidatesPresenceOfTheContainerKey(&lrp)
				itValidatesPresenceOfNetInfo(&lrp)
			})

			Context("when state is not set", func() {
				BeforeEach(func() {
					lrp = models.ActualLRP{
						ActualLRPKey: lrpKey,
						State:        "",
						Since:        1138,
					}
				})

				It("validate returns an error", func() {
					err := lrp.Validate()
					Ω(err).Should(HaveOccurred())
					Ω(err.Error()).Should(ContainSubstring("state"))
				})
			})

			Context("when since is not set", func() {
				BeforeEach(func() {
					lrp = models.ActualLRP{
						ActualLRPKey: lrpKey,
						State:        models.ActualLRPStateUnclaimed,
						Since:        0,
					}
				})

				It("validate returns an error", func() {
					err := lrp.Validate()
					Ω(err).Should(HaveOccurred())
					Ω(err.Error()).Should(ContainSubstring("since"))
				})
			})
		})
	})
})

func itValidatesPresenceOfTheLRPKey(lrp *models.ActualLRP) {
	Context("when the lrp key is set", func() {
		BeforeEach(func() {
			lrp.ActualLRPKey = models.NewActualLRPKey("some-guid", 1, "domain")
		})

		It("validate does not return an error", func() {
			Ω(lrp.Validate()).ShouldNot(HaveOccurred())
		})
	})

	Context("when the lrp key is not set", func() {
		BeforeEach(func() {
			lrp.ActualLRPKey = models.ActualLRPKey{}
		})

		It("validate returns an error", func() {
			err := lrp.Validate()
			Ω(err).Should(HaveOccurred())
			Ω(err.Error()).Should(ContainSubstring("process_guid"))
		})
	})
}

func itValidatesPresenceOfTheContainerKey(lrp *models.ActualLRP) {
	Context("when the container key is set", func() {
		BeforeEach(func() {
			lrp.ActualLRPContainerKey = models.NewActualLRPContainerKey("some-instance", "some-cell")
		})

		It("validate does not return an error", func() {
			Ω(lrp.Validate()).ShouldNot(HaveOccurred())
		})
	})

	Context("when the container key is not set", func() {
		BeforeEach(func() {
			lrp.ActualLRPContainerKey = models.ActualLRPContainerKey{}
		})

		It("validate returns an error", func() {
			err := lrp.Validate()
			Ω(err).Should(HaveOccurred())
			Ω(err.Error()).Should(ContainSubstring("instance_guid"))
		})
	})
}

func itValidatesAbsenceOfTheContainerKey(lrp *models.ActualLRP) {
	Context("when the container key is set", func() {
		BeforeEach(func() {
			lrp.ActualLRPContainerKey = models.NewActualLRPContainerKey("some-instance", "some-cell")
		})

		It("validate returns an error", func() {
			err := lrp.Validate()
			Ω(err).Should(HaveOccurred())
			Ω(err.Error()).Should(ContainSubstring("container"))
		})
	})

	Context("when the container key is not set", func() {
		BeforeEach(func() {
			lrp.ActualLRPContainerKey = models.ActualLRPContainerKey{}
		})

		It("validate does not return an error", func() {
			Ω(lrp.Validate()).ShouldNot(HaveOccurred())
		})
	})
}

func itValidatesPresenceOfNetInfo(lrp *models.ActualLRP) {
	Context("when net info is set", func() {
		BeforeEach(func() {
			lrp.ActualLRPNetInfo = models.NewActualLRPNetInfo("1.2.3.4", []models.PortMapping{})
		})

		It("validate does not return an error", func() {
			Ω(lrp.Validate()).ShouldNot(HaveOccurred())
		})
	})

	Context("when net info is not set", func() {
		BeforeEach(func() {
			lrp.ActualLRPNetInfo = models.ActualLRPNetInfo{}
		})

		It("validate returns an error", func() {
			err := lrp.Validate()
			Ω(err).Should(HaveOccurred())
			Ω(err.Error()).Should(ContainSubstring("address"))
		})
	})
}

func itValidatesAbsenceOfNetInfo(lrp *models.ActualLRP) {
	Context("when net info is set", func() {
		BeforeEach(func() {
			lrp.ActualLRPNetInfo = models.NewActualLRPNetInfo("1.2.3.4", []models.PortMapping{})
		})

		It("validate returns an error", func() {
			err := lrp.Validate()
			Ω(err).Should(HaveOccurred())
			Ω(err.Error()).Should(ContainSubstring("net info"))
		})
	})

	Context("when net info is not set", func() {
		BeforeEach(func() {
			lrp.ActualLRPNetInfo = models.ActualLRPNetInfo{}
		})

		It("validate does not return an error", func() {
			Ω(lrp.Validate()).ShouldNot(HaveOccurred())
		})
	})
}
