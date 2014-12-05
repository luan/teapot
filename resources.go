package teapot

type WorkstationCreateRequest struct {
	Name        string `json:"name"`
	DockerImage string `json:"docker_image"`
}
