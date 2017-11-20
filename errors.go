package mon

import (
	"fmt"
)

type ErrMetricNotFound struct {
	Metric string
}
func (e *ErrMetricNotFound) Error() string {
    return fmt.Sprintf("No such metric: %s", e.Metric)
}

type ErrMetricAlreadyRegistered struct {
	Metric string
}
func (e *ErrMetricAlreadyRegistered) Error() string {
    return fmt.Sprintf("Metric already registered: %s", e.Metric)
}
