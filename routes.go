package teapot

import "github.com/tedsuo/rata"

const (
	// Workstations
	CreateWorkstationRoute = "CreateWorkstation"
	DeleteWorkstationRoute = "DeleteWorkstation"
)

var Routes = rata.Routes{
	// Workstations
	{Path: "/workstations", Method: "POST", Name: CreateWorkstationRoute},
	{Path: "/workstations/:name", Method: "DELETE", Name: DeleteWorkstationRoute},
}
