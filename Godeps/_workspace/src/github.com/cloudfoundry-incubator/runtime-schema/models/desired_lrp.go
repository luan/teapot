package models

import (
	"encoding/json"
	"reflect"
	"regexp"
)

type DesiredLRP struct {
	ProcessGuid          string                `json:"process_guid"`
	Domain               string                `json:"domain"`
	RootFSPath           string                `json:"root_fs"`
	Instances            int                   `json:"instances"`
	Stack                string                `json:"stack"`
	EnvironmentVariables []EnvironmentVariable `json:"env,omitempty"`
	Setup                Action                `json:"-"`
	Action               Action                `json:"-"`
	StartTimeout         uint                  `json:"start_timeout"`
	Monitor              Action                `json:"-"`
	DiskMB               int                   `json:"disk_mb"`
	MemoryMB             int                   `json:"memory_mb"`
	CPUWeight            uint                  `json:"cpu_weight"`
	Ports                []uint32              `json:"ports"`
	Routes               []string              `json:"routes"`
	LogSource            string                `json:"log_source"`
	LogGuid              string                `json:"log_guid"`
	Annotation           string                `json:"annotation,omitempty"`
}

type InnerDesiredLRP DesiredLRP

type mDesiredLRP struct {
	SetupRaw   *json.RawMessage `json:"setup,omitempty"`
	ActionRaw  *json.RawMessage `json:"action,omitempty"`
	MonitorRaw *json.RawMessage `json:"monitor,omitempty"`
	*InnerDesiredLRP
}

type DesiredLRPChange struct {
	Before *DesiredLRP
	After  *DesiredLRP
}

type DesiredLRPUpdate struct {
	Instances  *int
	Routes     []string
	Annotation *string
}

func (desired DesiredLRP) ApplyUpdate(update DesiredLRPUpdate) DesiredLRP {
	if update.Instances != nil {
		desired.Instances = *update.Instances
	}
	if update.Routes != nil {
		desired.Routes = update.Routes
	}
	if update.Annotation != nil {
		desired.Annotation = *update.Annotation
	}
	return desired
}

var processGuidPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

func (desired DesiredLRP) Validate() error {
	var validationError ValidationError

	if desired.Domain == "" {
		validationError = append(validationError, ErrInvalidField{"domain"})
	}

	if !processGuidPattern.MatchString(desired.ProcessGuid) {
		validationError = append(validationError, ErrInvalidField{"process_guid"})
	}

	if desired.Stack == "" {
		validationError = append(validationError, ErrInvalidField{"stack"})
	}

	if desired.Setup != nil {
		err := desired.Setup.Validate()
		if err != nil {
			validationError = append(validationError, err)
		}
	}

	if desired.Action == nil {
		validationError = append(validationError, ErrInvalidActionType)
	} else {
		err := desired.Action.Validate()
		if err != nil {
			validationError = append(validationError, err)
		}
	}

	if desired.Monitor != nil {
		err := desired.Monitor.Validate()
		if err != nil {
			validationError = append(validationError, err)
		}
	}

	if desired.Instances < 0 {
		validationError = append(validationError, ErrInvalidField{"instances"})
	}

	if desired.CPUWeight > 100 {
		validationError = append(validationError, ErrInvalidField{"cpu_weight"})
	}

	if len(desired.Annotation) > maximumAnnotationLength {
		validationError = append(validationError, ErrInvalidField{"annotation"})
	}

	if len(validationError) > 0 {
		return validationError
	}

	return nil
}

func (desired DesiredLRP) ValidateModifications(updatedModel DesiredLRP) error {
	var validationError ValidationError

	if desired.ProcessGuid != updatedModel.ProcessGuid {
		validationError = append(validationError, ErrInvalidModification{"process_guid"})
	}

	if desired.Domain != updatedModel.Domain {
		validationError = append(validationError, ErrInvalidModification{"domain"})
	}

	if desired.RootFSPath != updatedModel.RootFSPath {
		validationError = append(validationError, ErrInvalidModification{"root_fs"})
	}

	if desired.Stack != updatedModel.Stack {
		validationError = append(validationError, ErrInvalidModification{"stack"})
	}

	if !reflect.DeepEqual(desired.EnvironmentVariables, updatedModel.EnvironmentVariables) {
		validationError = append(validationError, ErrInvalidModification{"env"})
	}

	if !reflect.DeepEqual(desired.Action, updatedModel.Action) {
		validationError = append(validationError, ErrInvalidModification{"action"})
	}

	if desired.DiskMB != updatedModel.DiskMB {
		validationError = append(validationError, ErrInvalidModification{"disk_mb"})
	}

	if desired.MemoryMB != updatedModel.MemoryMB {
		validationError = append(validationError, ErrInvalidModification{"memory_mb"})
	}

	if desired.CPUWeight != updatedModel.CPUWeight {
		validationError = append(validationError, ErrInvalidModification{"cpu_weight"})
	}

	if !reflect.DeepEqual(desired.Ports, updatedModel.Ports) {
		validationError = append(validationError, ErrInvalidModification{"ports"})
	}

	if desired.LogSource != updatedModel.LogSource {
		validationError = append(validationError, ErrInvalidModification{"log_source"})
	}

	if desired.LogGuid != updatedModel.LogGuid {
		validationError = append(validationError, ErrInvalidModification{"log_guid"})
	}

	if len(validationError) > 0 {
		return validationError
	}

	return nil
}

func (desired *DesiredLRP) UnmarshalJSON(payload []byte) error {
	mLRP := &mDesiredLRP{InnerDesiredLRP: (*InnerDesiredLRP)(desired)}
	err := json.Unmarshal(payload, mLRP)
	if err != nil {
		return err
	}

	var a Action
	if mLRP.ActionRaw == nil {
		a = nil
	} else {
		a, err = UnmarshalAction(*mLRP.ActionRaw)
		if err != nil {
			return err
		}
	}
	desired.Action = a

	if mLRP.SetupRaw == nil {
		a = nil
	} else {
		a, err = UnmarshalAction(*mLRP.SetupRaw)
		if err != nil {
			return err
		}
		desired.Setup = a
	}

	if mLRP.MonitorRaw == nil {
		a = nil
	} else {
		a, err = UnmarshalAction(*mLRP.MonitorRaw)
		if err != nil {
			return err
		}
		desired.Monitor = a
	}

	return nil
}

func (desired DesiredLRP) MarshalJSON() ([]byte, error) {
	var setupRaw, actionRaw, monitorRaw *json.RawMessage

	if desired.Action != nil {
		raw, err := MarshalAction(desired.Action)
		if err != nil {
			return nil, err
		}
		rm := json.RawMessage(raw)
		actionRaw = &rm
	}

	if desired.Setup != nil {
		raw, err := MarshalAction(desired.Setup)
		if err != nil {
			return nil, err
		}
		rm := json.RawMessage(raw)
		setupRaw = &rm
	}
	if desired.Monitor != nil {
		raw, err := MarshalAction(desired.Monitor)
		if err != nil {
			return nil, err
		}
		rm := json.RawMessage(raw)
		monitorRaw = &rm
	}

	innerDesiredLRP := InnerDesiredLRP(desired)

	mLRP := &mDesiredLRP{
		SetupRaw:        setupRaw,
		ActionRaw:       actionRaw,
		MonitorRaw:      monitorRaw,
		InnerDesiredLRP: &innerDesiredLRP,
	}

	return json.Marshal(mLRP)
}
