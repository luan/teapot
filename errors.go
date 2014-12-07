package teapot

type Error struct {
	Type    string `json:"name"`
	Message string `json:"message"`
}

func (err Error) Error() string {
	return err.Message
}

const (
	InvalidWorkstation   = "InvalidWorkstation"
	DuplicateWorkstation = "DuplicateWorkstation"

	InvalidJSON = "InvalidJSON"

	UnknownError = "UnknownError"
)
