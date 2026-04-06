package model

type CertificateAuthority struct {
	ID          string
	Name        string
	Fingerprint string
}

func (c CertificateAuthority) GetID() string {
	return c.ID
}

func (c CertificateAuthority) GetName() string {
	return c.Name
}
