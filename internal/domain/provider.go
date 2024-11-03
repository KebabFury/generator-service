package domain

type Provider struct {
	Name          string `json:"name,omitempty"`
	ActionCode    string `json:"actionCode,omitempty"`
	Documentation string `json:"documentation,omitempty"`
}
