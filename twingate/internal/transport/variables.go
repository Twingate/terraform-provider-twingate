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

func gqlField(val, name string) gqlVarOption {
	return func(values map[string]interface{}) map[string]interface{} {
		values[name] = graphql.String(val)
		return values
	}
}
