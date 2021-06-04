package twingate

type IdNameResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type OkErrorResponse struct {
	Ok    bool   `json:"ok"`
	Error string `json:"error"`
}

type EdgesResponse struct {
	Node *IdNameResponse `json:"node"`
}
