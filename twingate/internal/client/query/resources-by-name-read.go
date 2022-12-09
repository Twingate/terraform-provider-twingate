package query

type ReadResourcesByName struct {
	Resources `graphql:"resources(filter: {name: {eq: $name}}, after: $resourcesEndCursor)"`
}
