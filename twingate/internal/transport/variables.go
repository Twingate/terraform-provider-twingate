package transport

import "github.com/twingate/go-graphql-client"

type gqlVars struct {
	values map[string]interface{}
}

func newVariables() *gqlVars {
	return &gqlVars{
		values: make(map[string]interface{}),
	}
}

func (v *gqlVars) withID(val string, name ...string) *gqlVars {
	key := "id"
	if len(name) > 0 {
		key = name[0]
	}

	v.values[key] = graphql.ID(val)
	return v
}

func (v *gqlVars) withField(val, key string) *gqlVars {
	v.values[key] = graphql.String(val)
	return v
}

func (v *gqlVars) value() map[string]interface{} {
	return v.values
}
