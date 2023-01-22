package client

import (
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/utils"
	"github.com/hasura/go-graphql-client"
)

func newVars(options ...gqlVarOption) map[string]interface{} {
	values := make(map[string]interface{})

	for _, opt := range options {
		values = opt(values)
	}

	return values
}

type gqlVarOption func(values map[string]interface{}) map[string]interface{}

func gqlID(val interface{}, name ...string) gqlVarOption {
	key := "id"
	if len(name) > 0 {
		key = name[0]
	}

	return func(values map[string]interface{}) map[string]interface{} {
		values[key] = graphql.ToID(val)

		return values
	}
}

func gqlIDs(ids []string, name string) gqlVarOption {
	gqlValues := utils.Map[string, graphql.ID](ids,
		func(val string) graphql.ID {
			return graphql.ID(val)
		})

	return func(values map[string]interface{}) map[string]interface{} {
		values[name] = gqlValues

		return values
	}
}

func gqlVar(val interface{}, name string) gqlVarOption {
	return func(values map[string]interface{}) map[string]interface{} {
		if val != nil {
			values[name] = val
		}

		return values
	}
}

func gqlNullable(val interface{}, name string) gqlVarOption {
	return func(values map[string]interface{}) map[string]interface{} {
		var gqlValue interface{}
		if isZeroValue(val) {
			gqlValue = getNullableValue(val)
		} else {
			gqlValue = val
		}

		values[name] = gqlValue

		return values
	}
}

func isZeroValue(val interface{}) bool {
	if val == nil {
		return true
	}

	var (
		defaultString  string
		defaultInt     int
		defaultInt32   int32
		defaultInt64   int64
		defaultBool    bool
		defaultFloat64 float64
		defaultFloat32 float32
	)

	switch value := val.(type) {
	case string:
		return value == defaultString
	case bool:
		return value == defaultBool
	case int:
		return value == defaultInt
	case int32:
		return value == defaultInt32
	case int64:
		return value == defaultInt64
	case float32:
		return value == defaultFloat32
	case float64:
		return value == defaultFloat64
	}

	return false
}

func getNullableValue(val interface{}) interface{} {
	if val == nil {
		return nil
	}

	var (
		defaultString *string
		defaultInt    *int
		defaultBool   *bool
		defaultFloat  *float64
	)

	switch val.(type) {
	case string:
		return defaultString
	case bool:
		return defaultBool
	case int, int32, int64:
		return defaultInt
	case float32, float64:
		return defaultFloat
	}

	return nil
}
