package models

type ActualLRPState string

const (
	ActualLRPStateInvalid   ActualLRPState = "INVALID"
	ActualLRPStateUnclaimed ActualLRPState = "UNCLAIMED"
	ActualLRPStateClaimed   ActualLRPState = "CLAIMED"
	ActualLRPStateRunning   ActualLRPState = "RUNNING"
)

type ActualLRPChange struct {
	Before *ActualLRP
	After  *ActualLRP
}

type ActualLRP struct {
	ProcessGuid  string `json:"process_guid"`
	InstanceGuid string `json:"instance_guid"`
	CellID       string `json:"cell_id"`
	Domain       string `json:"domain"`

	Index int `json:"index"`

	Host  string        `json:"host"`
	Ports []PortMapping `json:"ports"`

	State ActualLRPState `json:"state"`
	Since int64          `json:"since"`
}

func NewActualLRP(
	processGuid string,
	instanceGuid string,
	cellID string,
	domain string,
	index int,
	state ActualLRPState,
) ActualLRP {

	lrp := ActualLRP{
		ProcessGuid:  processGuid,
		InstanceGuid: instanceGuid,
		CellID:       cellID,
		Domain:       domain,

		Index: index,
		State: state,
	}

	return lrp
}

func (actual ActualLRP) Validate() error {
	var validationError ValidationError

	if actual.ProcessGuid == "" {
		validationError = append(validationError, ErrInvalidField{"process_guid"})
	}

	if actual.Domain == "" {
		validationError = append(validationError, ErrInvalidField{"domain"})
	}

	if actual.State == ActualLRPStateUnclaimed {
		if actual.InstanceGuid != "" {
			validationError = append(validationError, ErrInvalidField{"instance_guid"})
		}

		if actual.CellID != "" {
			validationError = append(validationError, ErrInvalidField{"cell_id"})
		}
	} else {
		if actual.InstanceGuid == "" {
			validationError = append(validationError, ErrInvalidField{"instance_guid"})
		}

		if actual.CellID == "" {
			validationError = append(validationError, ErrInvalidField{"cell_id"})
		}
	}

	if len(validationError) > 0 {
		return validationError
	}

	return nil
}
