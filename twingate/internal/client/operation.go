package client

import (
	"fmt"

	"github.com/iancoleman/strcase"
)

type resource string

const (
	resourceConnector      resource = "connector"
	resourceGroup          resource = "group"
	resourceRemoteNetwork  resource = "remote network"
	resourceResource       resource = "resource"
	resourceResourceAccess resource = "resource access"
	resourceSecurityPolicy resource = "security policy"
	resourceServiceAccount resource = "service account"
	resourceServiceKey     resource = "service account key"
	resourceUser           resource = "user"
)

const (
	operationCreate = "create"
	operationRead   = "read"
	operationUpdate = "update"
	operationDelete = "delete"
	operationRevoke = "revoke"
)

type operation struct {
	customName string
	resource   string
	name       string
}

func (r resource) create() operation {
	return operation{
		resource: string(r),
		name:     operationCreate,
	}
}

func (r resource) update() operation {
	return operation{
		resource: string(r),
		name:     operationUpdate,
	}
}

func (r resource) delete() operation {
	return operation{
		resource: string(r),
		name:     operationDelete,
	}
}

func (r resource) read() operation {
	return operation{
		resource: string(r),
		name:     operationRead,
	}
}

func (r resource) revoke() operation {
	return operation{
		resource: string(r),
		name:     operationRevoke,
	}
}

type attr struct {
	id   string
	name string
}

func (o operation) apiError(err error, attrs ...attr) *APIError {
	// prevent double wrapping
	if e, ok := err.(*APIError); ok { //nolint
		return e
	}

	if len(attrs) == 0 {
		return NewAPIError(err, o.name, o.resource)
	}

	atr := attrs[0]
	if atr.name != "" {
		return NewAPIErrorWithName(err, o.name, o.resource, atr.name)
	}

	if atr.id != "" {
		return NewAPIErrorWithID(err, o.name, o.resource, atr.id)
	}

	return NewAPIError(err, o.name, o.resource)
}

func (o operation) String() string {
	if o.customName != "" {
		return o.customName
	}

	return strcase.ToLowerCamel(fmt.Sprintf("%s %s", o.name, o.resource))
}

func (o operation) withCustomName(name string) operation {
	o.customName = name

	return o
}
