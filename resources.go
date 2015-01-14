package teapot

type WorkstationCreateRequest struct {
	Name        string `json:"name"`
	DockerImage string `json:"docker_image"`
}

type WorkstationResponse struct {
	Name        string `json:"name"`
	DockerImage string `json:"docker_image"`
	State       string `json:"state"`
}
