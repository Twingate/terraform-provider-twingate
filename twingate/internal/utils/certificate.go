package utils

import (
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"
)

var ErrFailedDecodeCertificate = errors.New("failed to decode PEM certificate")

// CalculateCertificateFingerprint returns the SHA-256 fingerprint of a PEM-encoded certificate
// formatted as colon-separated uppercase hex pairs (e.g. "AB:CD:EF:...").
func CalculateCertificateFingerprint(pemCert string) (string, error) {
	block, _ := pem.Decode([]byte(pemCert))
	if block == nil {
		return "", ErrFailedDecodeCertificate
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse certificate: %w", err)
	}

	hash := sha256.Sum256(cert.Raw)
	hexStr := strings.ToUpper(hex.EncodeToString(hash[:]))

	var result strings.Builder

	for i, char := range hexStr {
		if i > 0 && i%2 == 0 {
			result.WriteRune(':')
		}

		result.WriteRune(char)
	}

	return result.String(), nil
}
