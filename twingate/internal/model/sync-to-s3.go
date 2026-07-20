package model

const (
	SyncToS3TypeOIDC = "oidc"
	SyncToS3TypeIAM  = "iam"
)

//nolint:gochecknoglobals
var SyncToS3Types = []string{SyncToS3TypeOIDC, SyncToS3TypeIAM}
