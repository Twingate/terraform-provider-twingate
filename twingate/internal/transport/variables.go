package transport

import "github.com/twingate/go-graphql-client"

func newVars(options ...gqlVarOption) map[string]interface{} {
	values := make(map[string]interface{})

	for _, opt := range options {
		values = opt(values)
	}

	return values
}

type gqlVarOption func(values map[string]interface{}) map[string]interface{}

func gqlID(val string, name ...string) gqlVarOption {
	key := "id"
	if len(name) > 0 {
		key = name[0]
	}

	return func(values map[string]interface{}) map[string]interface{} {
		values[key] = graphql.ID(val)
		return values
	}
}

func gqlField(val interface{}, name string) gqlVarOption {
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

	return gqlValue
}
