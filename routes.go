package teapot

import "github.com/tedsuo/rata"

const (
	// Workstations
	CreateWorkstationRoute = "CreateWorkstation"
	DeleteWorkstationRoute = "DeleteWorkstation"
	AttachWorkstationRoute = "AttachWorkstation"
	ListWorkstationsRoute  = "ListWorkstations"
)

var Routes = rata.Routes{
	// Workstations
	{Path: "/workstations", Method: "POST", Name: CreateWorkstationRoute},
	{Path: "/workstations/:name", Method: "DELETE", Name: DeleteWorkstationRoute},
	{Path: "/workstations/:name/attach", Method: "GET", Name: AttachWorkstationRoute},
	{Path: "/workstations", Method: "GET", Name: ListWorkstationsRoute},
}
