package twingate

type IDNameResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type OkErrorResponse struct {
	Ok    bool   `json:"ok"`
	Error string `json:"error"`
}

func (r *OkErrorResponse) ErrorMessage() string {
	return r.Error
}

func (r *OkErrorResponse) IsOk() bool {
	return r.Ok
}

type EdgesResponse struct {
	Node *IDNameResponse `json:"node"`
}
