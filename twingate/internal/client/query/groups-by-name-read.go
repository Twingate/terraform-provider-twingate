package query

type ReadGroupsByName struct {
	Groups `graphql:"groups(filter: {name: {eq: $name}}, after: $groupsEndCursor)"`
}
