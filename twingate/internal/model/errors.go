package model

import (
	"errors"
	"fmt"
)

var ErrInvalidPortRangeLen = errors.New("port range expects 2 values")

func ErrInvalidPortRange(portRange string, err error) error {
	return fmt.Errorf(`failed to parse protocols port range "%s": %w`, portRange, err)
}

type PortNotInRangeError struct {
	Port int
}

func NewPortNotInRangeError(port int) *PortNotInRangeError {
	return &PortNotInRangeError{
		Port: port,
	}
}

func (e *PortNotInRangeError) Error() string {
	return fmt.Sprintf("port %d not in the range of %d-%d", e.Port, minPortValue, maxPortValue)
}

type PortRangeNotRisingSequenceError struct {
	Start int
	End   int
}

func NewPortRangeNotRisingSequenceError(start, end int) *PortRangeNotRisingSequenceError {
	return &PortRangeNotRisingSequenceError{
		Start: start,
		End:   end,
	}
}

func (e *PortRangeNotRisingSequenceError) Error() string {
	return fmt.Sprintf("ports %d, %d needs to be in a rising sequence", e.Start, e.End)
}
