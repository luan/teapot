package teapot

import "github.com/tedsuo/rata"

const (
	// Workstations
	CreateWorkstationRoute = "CreateWorkstation"
)

var Routes = rata.Routes{
	// Workstations
	{Path: "/workstations", Method: "POST", Name: CreateWorkstationRoute},
}
