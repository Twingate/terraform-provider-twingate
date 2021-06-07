package twingate

type IDNameResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type OkErrorResponse struct {
	Ok    bool   `json:"ok"`
	Error string `json:"error"`
}

type EdgesResponse struct {
	Node *IDNameResponse `json:"node"`
}

type queryResponseErrors struct {
	Message   string                         `json:"message"`
	Locations []*queryResponseErrorsLocation `json:"locations"`
	Path      []string                       `json:"path"`
}

type queryResponseErrorsLocation struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

type responseErrors interface {
	checkErrors() []*queryResponseErrors
}

func parseErrors(responseErrors []*queryResponseErrors) []string {
	messages := []string{}

	for _, e := range responseErrors {
		messages = append(messages, e.Message)
	}

	return messages
}
