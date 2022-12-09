package client

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

func gqlVar(val interface{}, name string) gqlVarOption {
	return func(values map[string]interface{}) map[string]interface{} {
		gqlValue := convertToGQL(val)
		if gqlValue != nil {
			values[name] = gqlValue
		}

		return values
	}
}

func convertToGQL(val interface{}) interface{} {
	var gqlValue interface{}

	switch value := val.(type) {
	case string:
		gqlValue = graphql.String(value)
	case bool:
		gqlValue = graphql.Boolean(value)
	case int:
		gqlValue = graphql.Int(value)
	case int32:
		gqlValue = graphql.Int(value)
	case int64:
		gqlValue = graphql.Int(int32(value))
	case float64:
		gqlValue = graphql.Float(value)
	case float32:
		gqlValue = graphql.Float(value)
	}

	if gqlValue != nil {
		return gqlValue
	}

	return val
}

func gqlNullable(val interface{}, name string) gqlVarOption {
	return func(values map[string]interface{}) map[string]interface{} {
		var gqlValue interface{}
		if isDefaultValue(val) {
			gqlValue = getDefaultGQLValue(val)
		} else {
			gqlValue = convertToGQL(val)
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

func getDefaultGQLValue(val interface{}) interface{} {
	if val == nil {
		return nil
	}

	var (
		defaultString *graphql.String
		defaultInt    *graphql.Int
		defaultBool   *graphql.Boolean
		defaultFloat  *graphql.Float
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
