package transport

import (
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/utils"
	"github.com/twingate/go-graphql-client"
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
		values[key] = graphql.ID(val)
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

func gqlField(val interface{}, name string) gqlVarOption {
	return func(values map[string]interface{}) map[string]interface{} {
		gqlValue := tryConvertToGQL(val)
		if gqlValue != nil {
			values[name] = gqlValue
		}

		return values
	}
}

func tryConvertToGQL(val interface{}) interface{} {
	var gqlValue interface{}

	switch v := val.(type) {
	case string:
		gqlValue = graphql.String(v)
	case bool:
		gqlValue = graphql.Boolean(v)
	case int:
		gqlValue = graphql.Int(v)
	case int32:
		gqlValue = graphql.Int(v)
	case int64:
		// TODO: handle int32 overflow
		gqlValue = graphql.Int(v)
	case float64:
		gqlValue = graphql.Float(v)
	case float32:
		gqlValue = graphql.Float(v)
	}

	if gqlValue != nil {
		return gqlValue
	}

	return val
}

func gqlNullableField(val interface{}, name string) gqlVarOption {
	return func(values map[string]interface{}) map[string]interface{} {
		var gqlValue interface{}
		if isDefaultValue(val) {
			gqlValue = nil
		} else {
			gqlValue = tryConvertToGQL(val)
		}

		values[name] = gqlValue

		return values
	}
}

func isDefaultValue(val interface{}) bool {
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

	switch v := val.(type) {
	case string:
		return v == defaultString
	case bool:
		return v == defaultBool
	case int:
		return v == defaultInt
	case int32:
		return v == defaultInt32
	case int64:
		return v == defaultInt64
	case float32:
		return v == defaultFloat32
	case float64:
		return v == defaultFloat64
	}

	return false
}
