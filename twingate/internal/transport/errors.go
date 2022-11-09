package transport

import (
	"errors"
	"fmt"

	"github.com/twingate/go-graphql-client"
)

var (
	ErrTooManyGroupsError = fmt.Errorf("provider does not support more than %d groups per resource", readResourceQueryGroupsSize)

	ErrGraphqlIDIsEmpty        = errors.New("id is empty")
	ErrGraphqlNameIsEmpty      = errors.New("name is empty")
	ErrGraphqlResourceNotFound = errors.New("not found")

	ErrGraphqlResultIsEmpty      = errors.New("query result is empty")
	ErrGraphqlConnectorIDIsEmpty = errors.New("connector id is empty")
	ErrGraphqlNetworkIDIsEmpty   = errors.New("network id is empty")
	ErrGraphqlNetworkNameIsEmpty = errors.New("network name is empty")
	ErrGraphqlGroupNameIsEmpty   = errors.New("group name is empty")
)

type HTTPError struct {
	RequestURI string
	StatusCode int
	Body       []byte
}

func NewHTTPError(requestURI string, statusCode int, body []byte) *HTTPError {
	return &HTTPError{
		RequestURI: requestURI,
		StatusCode: statusCode,
		Body:       body,
	}
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("request %s failed, status %d, body %s", e.RequestURI, e.StatusCode, e.Body)
}

type APIError struct {
	WrappedError error
	Operation    string
	Resource     string
	ID           graphql.ID
	Name         string
}

func NewAPIErrorWithID(wrappedError error, operation string, resource string, id graphql.ID) *APIError {
	return &APIError{
		WrappedError: wrappedError,
		Operation:    operation,
		Resource:     resource,
		ID:           id,
	}
}

func NewAPIErrorWithName(wrappedError error, operation string, resource string, name string) *APIError {
	return &APIError{
		WrappedError: wrappedError,
		Operation:    operation,
		Resource:     resource,
		Name:         name,
	}
}

func NewAPIError(wrappedError error, operation string, resource string) *APIError {
	return &APIError{
		WrappedError: wrappedError,
		Operation:    operation,
		Resource:     resource,
	}
}

func (e *APIError) Error() string {
	var args = make([]interface{}, 0, 2) //nolint:gomnd
	args = append(args, e.Operation, e.Resource)

	var format = "failed to %s %s"

	if e.ID != nil && e.ID.(string) != "" {
		format += " with id %s"

		args = append(args, e.ID)
	}

	if e.Name != "" {
		format += " with name %s"

		args = append(args, e.Name)
	}

	if e.WrappedError != nil {
		format += ": %s"

		args = append(args, e.WrappedError)
	}

	return fmt.Sprintf(format, args...)
}

func (e *APIError) Unwrap() error {
	return e.WrappedError
}

type MutationError struct {
	Message graphql.String
}

func NewMutationError(message graphql.String) *MutationError {
	return &MutationError{
		Message: message,
	}
}

func (e *MutationError) Error() string {
	return string(e.Message)
}
