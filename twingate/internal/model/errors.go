package model

import "fmt"

func ErrInvalidPortRange(portRange string, err error) error {
	return fmt.Errorf(`failed to parse protocols port range "%s": %w`, portRange, err)
}

type PortNotInRangeError struct {
	Port int64
}

func NewPortNotInRangeError(port int64) *PortNotInRangeError {
	return &PortNotInRangeError{
		Port: port,
	}
}

func (e *PortNotInRangeError) Error() string {
	return fmt.Sprintf("port %d not in the range of %d-%d", e.Port, minPortValue, maxPortValue)
}

type PortRangeNotRisingSequenceError struct {
	Start int32
	End   int32
}

func NewPortRangeNotRisingSequenceError(start, end int32) *PortRangeNotRisingSequenceError {
	return &PortRangeNotRisingSequenceError{
		Start: start,
		End:   end,
	}
}

func (e *PortRangeNotRisingSequenceError) Error() string {
	return fmt.Sprintf("ports %d, %d needs to be in a rising sequence", e.Start, e.End)
}
