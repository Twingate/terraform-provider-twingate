package model

import (
	"fmt"
	"strconv"
)

const (
	minPortValue = 0
	maxPortValue = 65535
)

func validatePort(str string) (int32, error) {
	port, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("port is not a valid integer: %w", err)
	}

	if port < minPortValue || port > maxPortValue {
		return 0, NewPortNotInRangeError(port)
	}

	return int32(port), nil
}
