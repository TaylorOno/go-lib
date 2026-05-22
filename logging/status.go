package logging

import "fmt"

// Status log status to be used as a metric
type Status struct {
	name    string
	message string
}

// NewStatus returns a new status.
func NewStatus(name string, message string) Status {
	return Status{name: name, message: message}
}

// Name of the metrics
func (status Status) Name() string {
	return status.name
}

// Message to capture
func (status Status) Message() string {
	return status.message
}

// String representation
func (status Status) String() string {
	return status.name + "/" + status.message
}

func (status Status) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("{\"%s\":\"%s\"}", status.name, status.message)), nil
}
