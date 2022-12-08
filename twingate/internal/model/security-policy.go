package model

type SecurityPolicy struct {
	ID     string
	Name   string
	Type   string
	Groups []string
}
