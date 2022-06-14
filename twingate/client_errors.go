package twingate

import (
	"errors"
	"fmt"
)

var (
	ErrTooManyGroupsError = fmt.Errorf("provider does not support more than %d groups per resource", readResourceQueryGroupsSize)

	ErrGraphqlIDIsEmpty        = errors.New("id is empty")
	ErrGraphqlNameIsEmpty      = errors.New("name is empty")
	ErrGraphqlResourceNotFound = errors.New("not found")

	ErrGraphqlConnectorIDIsEmpty = errors.New("connector id is empty")
	ErrGraphqlNetworkIDIsEmpty   = errors.New("network id is empty")
	ErrGraphqlNetworkNameIsEmpty = errors.New("network name is empty")
	ErrGraphqlGroupNameIsEmpty   = errors.New("group name is empty")
)

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
	return fmt.Sprintf("port %d not in the range of 0-65535", e.Port)
}

type PortRangeNotRisingSequenceError struct {
	Start int64
	End   int64
}

func NewPortRangeNotRisingSequenceError(start int64, end int64) *PortRangeNotRisingSequenceError {
	return &PortRangeNotRisingSequenceError{
		Start: start,
		End:   end,
	}
}

func (e *PortRangeNotRisingSequenceError) Error() string {
	return fmt.Sprintf("ports %d, %d needs to be in a rising sequence", e.Start, e.End)
}
