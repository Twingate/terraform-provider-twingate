package model

type Gateway struct {
	ID              string
	RemoteNetworkID string
	Address         string
	X509CAID        string
	SSHCAID         string // empty when not set
}
