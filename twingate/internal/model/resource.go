package model

var Policies = []string{PolicyRestricted, PolicyAllowAll, PolicyDenyAll}

const (
	PolicyRestricted = "RESTRICTED"
	PolicyAllowAll   = "ALLOW_ALL"
	PolicyDenyAll    = "DENY_ALL"
)
