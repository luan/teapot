package receptor

import (
	"encoding/json"

	"github.com/cloudfoundry-incubator/runtime-schema/models"
)

type EnvironmentVariable struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type PortMapping struct {
	ContainerPort uint16 `json:"container_port"`
	HostPort      uint16 `json:"host_port,omitempty"`
}

const (
	TaskStateInvalid   = "INVALID"
	TaskStatePending   = "PENDING"
	TaskStateRunning   = "RUNNING"
	TaskStateCompleted = "COMPLETED"
	TaskStateResolving = "RESOLVING"
)

type TaskCreateRequest struct {
	Action                models.Action              `json:"-"`
	Annotation            string                     `json:"annotation,omitempty"`
	CompletionCallbackURL string                     `json:"completion_callback_url"`
	CPUWeight             uint                       `json:"cpu_weight"`
	DiskMB                int                        `json:"disk_mb"`
	Domain                string                     `json:"domain"`
	LogGuid               string                     `json:"log_guid"`
	LogSource             string                     `json:"log_source"`
	MetricsGuid           string                     `json:"metrics_guid"`
	MemoryMB              int                        `json:"memory_mb"`
	ResultFile            string                     `json:"result_file"`
	Stack                 string                     `json:"stack"`
	TaskGuid              string                     `json:"task_guid"`
	RootFSPath            string                     `json:"rootfs"`
	Privileged            bool                       `json:"privileged"`
	EnvironmentVariables  []EnvironmentVariable      `json:"env,omitempty"`
	EgressRules           []models.SecurityGroupRule `json:"egress_rules,omitempty"`
}

type InnerTaskCreateRequest TaskCreateRequest

type mTaskCreateRequest struct {
	ActionRaw json.RawMessage `json:"action"`
	*InnerTaskCreateRequest
}

func (request TaskCreateRequest) MarshalJSON() ([]byte, error) {
	actionRaw, err := models.MarshalAction(request.Action)
	if err != nil {
		return nil, err
	}

	innerRequest := InnerTaskCreateRequest(request)
	mRequest := &mTaskCreateRequest{
		ActionRaw:              actionRaw,
		InnerTaskCreateRequest: &innerRequest,
	}

	return json.Marshal(mRequest)
}

func (request *TaskCreateRequest) UnmarshalJSON(payload []byte) error {
	mRequest := &mTaskCreateRequest{InnerTaskCreateRequest: (*InnerTaskCreateRequest)(request)}
	err := json.Unmarshal(payload, mRequest)
	if err != nil {
		return err
	}

	var a models.Action
	if mRequest.ActionRaw == nil {
		a = nil
	} else {
		a, err = models.UnmarshalAction(mRequest.ActionRaw)
		if err != nil {
			return err
		}
	}
	request.Action = a

	return nil
}

type TaskResponse struct {
	Action                models.Action              `json:"-"`
	Annotation            string                     `json:"annotation,omitempty"`
	CompletionCallbackURL string                     `json:"completion_callback_url"`
	CPUWeight             uint                       `json:"cpu_weight"`
	DiskMB                int                        `json:"disk_mb"`
	Domain                string                     `json:"domain"`
	LogGuid               string                     `json:"log_guid"`
	LogSource             string                     `json:"log_source"`
	MetricsGuid           string                     `json:"metrics_guid"`
	MemoryMB              int                        `json:"memory_mb"`
	ResultFile            string                     `json:"result_file"`
	Stack                 string                     `json:"stack"`
	TaskGuid              string                     `json:"task_guid"`
	RootFSPath            string                     `json:"rootfs"`
	Privileged            bool                       `json:"privileged"`
	EnvironmentVariables  []EnvironmentVariable      `json:"env,omitempty"`
	CellID                string                     `json:"cell_id"`
	CreatedAt             int64                      `json:"created_at"`
	Failed                bool                       `json:"failed"`
	FailureReason         string                     `json:"failure_reason"`
	Result                string                     `json:"result"`
	State                 string                     `json:"state"`
	EgressRules           []models.SecurityGroupRule `json:"egress_rules,omitempty"`
}

type InnerTaskResponse TaskResponse

type mTaskResponse struct {
	ActionRaw json.RawMessage `json:"action"`
	*InnerTaskResponse
}

func (response TaskResponse) MarshalJSON() ([]byte, error) {
	actionRaw, err := models.MarshalAction(response.Action)
	if err != nil {
		return nil, err
	}

	innerResponse := InnerTaskResponse(response)
	mResponse := &mTaskResponse{
		ActionRaw:         actionRaw,
		InnerTaskResponse: &innerResponse,
	}

	return json.Marshal(mResponse)
}

func (response *TaskResponse) UnmarshalJSON(payload []byte) error {
	mResponse := &mTaskResponse{InnerTaskResponse: (*InnerTaskResponse)(response)}
	err := json.Unmarshal(payload, mResponse)
	if err != nil {
		return err
	}

	var a models.Action
	if mResponse.ActionRaw == nil {
		a = nil
	} else {
		a, err = models.UnmarshalAction(mResponse.ActionRaw)
		if err != nil {
			return err
		}
	}
	response.Action = a

	return nil
}

type RoutingInfo map[string]*json.RawMessage

type DesiredLRPCreateRequest struct {
	ProcessGuid          string                     `json:"process_guid"`
	Domain               string                     `json:"domain"`
	RootFSPath           string                     `json:"rootfs"`
	Instances            int                        `json:"instances"`
	Stack                string                     `json:"stack"`
	EnvironmentVariables []EnvironmentVariable      `json:"env,omitempty"`
	Setup                models.Action              `json:"-"`
	Action               models.Action              `json:"-"`
	Monitor              models.Action              `json:"-"`
	StartTimeout         uint                       `json:"start_timeout"`
	DiskMB               int                        `json:"disk_mb"`
	MemoryMB             int                        `json:"memory_mb"`
	CPUWeight            uint                       `json:"cpu_weight"`
	Privileged           bool                       `json:"privileged"`
	Ports                []uint16                   `json:"ports"`
	Routes               RoutingInfo                `json:"routes,omitempty"`
	LogGuid              string                     `json:"log_guid"`
	LogSource            string                     `json:"log_source"`
	MetricsGuid          string                     `json:"metrics_guid"`
	Annotation           string                     `json:"annotation,omitempty"`
	EgressRules          []models.SecurityGroupRule `json:"egress_rules,omitempty"`
}

type InnerDesiredLRPCreateRequest DesiredLRPCreateRequest

type mDesiredLRPCreateRequest struct {
	SetupRaw   *json.RawMessage `json:"setup,omitempty"`
	ActionRaw  *json.RawMessage `json:"action"`
	MonitorRaw *json.RawMessage `json:"monitor,omitempty"`
	*InnerDesiredLRPCreateRequest
}

func (request DesiredLRPCreateRequest) MarshalJSON() ([]byte, error) {
	var setupRaw, actionRaw, monitorRaw *json.RawMessage

	if request.Action != nil {
		raw, err := models.MarshalAction(request.Action)
		if err != nil {
			return nil, err
		}
		rm := json.RawMessage(raw)
		actionRaw = &rm
	}

	if request.Setup != nil {
		raw, err := models.MarshalAction(request.Setup)
		if err != nil {
			return nil, err
		}
		rm := json.RawMessage(raw)
		setupRaw = &rm
	}

	if request.Monitor != nil {
		raw, err := models.MarshalAction(request.Monitor)
		if err != nil {
			return nil, err
		}
		rm := json.RawMessage(raw)
		monitorRaw = &rm
	}

	innerRequest := InnerDesiredLRPCreateRequest(request)
	mRequest := &mDesiredLRPCreateRequest{
		SetupRaw:                     setupRaw,
		ActionRaw:                    actionRaw,
		MonitorRaw:                   monitorRaw,
		InnerDesiredLRPCreateRequest: &innerRequest,
	}

	return json.Marshal(mRequest)
}

func (request *DesiredLRPCreateRequest) UnmarshalJSON(payload []byte) error {
	mRequest := &mDesiredLRPCreateRequest{InnerDesiredLRPCreateRequest: (*InnerDesiredLRPCreateRequest)(request)}
	err := json.Unmarshal(payload, mRequest)
	if err != nil {
		return err
	}

	var a models.Action

	if mRequest.ActionRaw == nil {
		a = nil
	} else {
		a, err = models.UnmarshalAction(*mRequest.ActionRaw)
		if err != nil {
			return err
		}
	}
	request.Action = a

	if mRequest.SetupRaw == nil {
		a = nil
	} else {
		a, err = models.UnmarshalAction(*mRequest.SetupRaw)
		if err != nil {
			return err
		}
	}
	request.Setup = a

	if mRequest.MonitorRaw == nil {
		a = nil
	} else {
		a, err = models.UnmarshalAction(*mRequest.MonitorRaw)
		if err != nil {
			return err
		}
	}
	request.Monitor = a

	return nil
}

type DesiredLRPUpdateRequest struct {
	Instances  *int        `json:"instances,omitempty"`
	Routes     RoutingInfo `json:"routes,omitempty"`
	Annotation *string     `json:"annotation,omitempty"`
}

type DesiredLRPResponse struct {
	ProcessGuid          string                     `json:"process_guid"`
	Domain               string                     `json:"domain"`
	RootFSPath           string                     `json:"rootfs"`
	Instances            int                        `json:"instances"`
	Stack                string                     `json:"stack"`
	EnvironmentVariables []EnvironmentVariable      `json:"env,omitempty"`
	Setup                models.Action              `json:"setup"`
	Action               models.Action              `json:"action"`
	Monitor              models.Action              `json:"monitor"`
	StartTimeout         uint                       `json:"start_timeout"`
	DiskMB               int                        `json:"disk_mb"`
	MemoryMB             int                        `json:"memory_mb"`
	CPUWeight            uint                       `json:"cpu_weight"`
	Privileged           bool                       `json:"privileged"`
	Ports                []uint16                   `json:"ports"`
	Routes               RoutingInfo                `json:"routes,omitempty"`
	LogGuid              string                     `json:"log_guid"`
	LogSource            string                     `json:"log_source"`
	MetricsGuid          string                     `json:"metrics_guid"`
	Annotation           string                     `json:"annotation,omitempty"`
	EgressRules          []models.SecurityGroupRule `json:"egress_rules,omitempty"`
}

type InnerDesiredLRPResponse DesiredLRPResponse

type mDesiredLRPResponse struct {
	SetupRaw   *json.RawMessage `json:"setup,omitempty"`
	ActionRaw  *json.RawMessage `json:"action"`
	MonitorRaw *json.RawMessage `json:"monitor,omitempty"`
	*InnerDesiredLRPResponse
}

func (response DesiredLRPResponse) MarshalJSON() ([]byte, error) {
	var setupRaw, actionRaw, monitorRaw *json.RawMessage

	if response.Action != nil {
		raw, err := models.MarshalAction(response.Action)
		if err != nil {
			return nil, err
		}
		rm := json.RawMessage(raw)
		actionRaw = &rm
	}

	if response.Setup != nil {
		raw, err := models.MarshalAction(response.Setup)
		if err != nil {
			return nil, err
		}
		rm := json.RawMessage(raw)
		setupRaw = &rm
	}
	if response.Monitor != nil {
		raw, err := models.MarshalAction(response.Monitor)
		if err != nil {
			return nil, err
		}
		rm := json.RawMessage(raw)
		monitorRaw = &rm
	}

	innerResponse := InnerDesiredLRPResponse(response)
	mResponse := &mDesiredLRPResponse{
		SetupRaw:                setupRaw,
		ActionRaw:               actionRaw,
		MonitorRaw:              monitorRaw,
		InnerDesiredLRPResponse: &innerResponse,
	}

	return json.Marshal(mResponse)
}

func (response *DesiredLRPResponse) UnmarshalJSON(payload []byte) error {
	mResponse := &mDesiredLRPResponse{InnerDesiredLRPResponse: (*InnerDesiredLRPResponse)(response)}
	err := json.Unmarshal(payload, mResponse)
	if err != nil {
		return err
	}

	var a models.Action

	if mResponse.ActionRaw == nil {
		a = nil
	} else {
		a, err = models.UnmarshalAction(*mResponse.ActionRaw)
		if err != nil {
			return err
		}
	}
	response.Action = a

	if mResponse.SetupRaw == nil {
		a = nil
	} else {
		a, err = models.UnmarshalAction(*mResponse.SetupRaw)
		if err != nil {
			return err
		}
	}
	response.Setup = a

	if mResponse.MonitorRaw == nil {
		a = nil
	} else {
		a, err = models.UnmarshalAction(*mResponse.MonitorRaw)
		if err != nil {
			return err
		}
	}
	response.Monitor = a

	return nil
}

type ActualLRPState string

const (
	ActualLRPStateInvalid   ActualLRPState = "INVALID"
	ActualLRPStateUnclaimed ActualLRPState = "UNCLAIMED"
	ActualLRPStateClaimed   ActualLRPState = "CLAIMED"
	ActualLRPStateRunning   ActualLRPState = "RUNNING"
	ActualLRPStateCrashed   ActualLRPState = "CRASHED"
)

type ActualLRPResponse struct {
	ProcessGuid    string         `json:"process_guid"`
	InstanceGuid   string         `json:"instance_guid"`
	CellID         string         `json:"cell_id"`
	Domain         string         `json:"domain"`
	Index          int            `json:"index"`
	Address        string         `json:"address"`
	Ports          []PortMapping  `json:"ports"`
	State          ActualLRPState `json:"state"`
	CrashCount     int            `json:"crash_count"`
	PlacementError string         `json:"placement_error,omitempty"`
	Since          int64          `json:"since"`
}

type CellResponse struct {
	CellID   string       `json:"cell_id"`
	Stack    string       `json:"stack"`
	Zone     string       `json:"zone"`
	Capacity CellCapacity `json:"capacity"`
}

type CellCapacity struct {
	MemoryMB   int `json:"memory_mb"`
	DiskMB     int `json:"disk_mb"`
	Containers int `json:"containers"`
}

type Event interface {
	EventType() EventType
}

type EventType string

const (
	EventTypeInvalid EventType = ""

	EventTypeDesiredLRPCreated EventType = "desired_lrp_created"
	EventTypeDesiredLRPChanged EventType = "desired_lrp_changed"
	EventTypeDesiredLRPRemoved EventType = "desired_lrp_removed"
	EventTypeActualLRPCreated  EventType = "actual_lrp_created"
	EventTypeActualLRPChanged  EventType = "actual_lrp_changed"
	EventTypeActualLRPRemoved  EventType = "actual_lrp_removed"
)

type DesiredLRPCreatedEvent struct {
	DesiredLRPResponse DesiredLRPResponse `json:"desired_lrp"`
}

func NewDesiredLRPCreatedEvent(desiredLRP DesiredLRPResponse) DesiredLRPCreatedEvent {
	return DesiredLRPCreatedEvent{
		DesiredLRPResponse: desiredLRP,
	}
}

func (DesiredLRPCreatedEvent) EventType() EventType { return EventTypeDesiredLRPCreated }

type DesiredLRPChangedEvent struct {
	Before DesiredLRPResponse `json:"desired_lrp_before"`
	After  DesiredLRPResponse `json:"desired_lrp_after"`
}

func NewDesiredLRPChangedEvent(before, after DesiredLRPResponse) DesiredLRPChangedEvent {
	return DesiredLRPChangedEvent{
		Before: before,
		After:  after,
	}
}

func (DesiredLRPChangedEvent) EventType() EventType { return EventTypeDesiredLRPChanged }

type DesiredLRPRemovedEvent struct {
	DesiredLRPResponse DesiredLRPResponse `json:"desired_lrp"`
}

func NewDesiredLRPRemovedEvent(desiredLRP DesiredLRPResponse) DesiredLRPRemovedEvent {
	return DesiredLRPRemovedEvent{
		DesiredLRPResponse: desiredLRP,
	}
}

func (DesiredLRPRemovedEvent) EventType() EventType { return EventTypeDesiredLRPRemoved }

type ActualLRPCreatedEvent struct {
	ActualLRPResponse ActualLRPResponse `json:"actual_lrp"`
}

func NewActualLRPCreatedEvent(actualLRP ActualLRPResponse) ActualLRPCreatedEvent {
	return ActualLRPCreatedEvent{
		ActualLRPResponse: actualLRP,
	}
}

func (ActualLRPCreatedEvent) EventType() EventType { return EventTypeActualLRPCreated }

type ActualLRPChangedEvent struct {
	Before ActualLRPResponse `json:"actual_lrp_before"`
	After  ActualLRPResponse `json:"actual_lrp_after"`
}

func NewActualLRPChangedEvent(before, after ActualLRPResponse) ActualLRPChangedEvent {
	return ActualLRPChangedEvent{
		Before: before,
		After:  after,
	}
}

func (ActualLRPChangedEvent) EventType() EventType { return EventTypeActualLRPChanged }

type ActualLRPRemovedEvent struct {
	ActualLRPResponse ActualLRPResponse `json:"actual_lrp"`
}

func NewActualLRPRemovedEvent(actualLRP ActualLRPResponse) ActualLRPRemovedEvent {
	return ActualLRPRemovedEvent{
		ActualLRPResponse: actualLRP,
	}
}

func (ActualLRPRemovedEvent) EventType() EventType { return EventTypeActualLRPRemoved }
