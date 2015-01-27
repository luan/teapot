package teapot

type WorkstationCreateRequest struct {
	Name        string `json:"name"`
	DockerImage string `json:"docker_image"`
	CPUWeight   uint   `json:"cpu_weight"`
	DiskMB      int    `json:"disk_mb"`
	MemoryMB    int    `json:"memory_mb"`
}

type WorkstationResponse struct {
	Name        string `json:"name"`
	DockerImage string `json:"docker_image"`
	State       string `json:"state"`
}
