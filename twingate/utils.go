package twingate

func toList[E, O any](edges []E, getObj func(edge E) O) []O {
	out := make([]O, 0, len(edges))
	for _, elem := range edges {
		out = append(out, getObj(elem))
	}

	return out
}
