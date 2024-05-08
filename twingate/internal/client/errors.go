package client

import (
	"errors"
	"fmt"

	"github.com/hasura/go-graphql-client"
)

var (
	ErrGraphqlIDIsEmpty          = errors.New("id is empty")
	ErrGraphqlNameIsEmpty        = errors.New("name is empty")
	ErrGraphqlEmptyBothNameAndID = errors.New("both name and id should not be empty")
	ErrGraphqlResultIsEmpty      = errors.New("query result is empty")
	ErrGraphqlConnectorIDIsEmpty = errors.New("connector id is empty")
	ErrGraphqlNetworkIDIsEmpty   = errors.New("network id is empty")
	ErrGraphqlNetworkNameIsEmpty = errors.New("network name is empty")
	ErrGraphqlEmailIsEmpty       = errors.New("email is empty")
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

func NewAPIErrorWithID(wrappedError error, operation, resource, id string) *APIError {
	return &APIError{
		WrappedError: wrappedError,
		Operation:    operation,
		Resource:     resource,
		ID:           graphql.ID(id),
	}
}

func NewAPIErrorWithName(wrappedError error, operation, resource, name string) *APIError {
	return &APIError{
		WrappedError: wrappedError,
		Operation:    operation,
		Resource:     resource,
		Name:         name,
	}
}

func NewAPIError(wrappedError error, operation, resource string) *APIError {
	return &APIError{
		WrappedError: wrappedError,
		Operation:    operation,
		Resource:     resource,
	}
}

func (e *APIError) Error() string {
	args := []interface{}{e.Operation, e.Resource}

	var format = "failed to %s %s"

	if e.ID != "" {
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
	Message string
}

func NewMutationError(message string) *MutationError {
	return &MutationError{
		Message: message,
	}
}

func (e *MutationError) Error() string {
	return e.Message
}
