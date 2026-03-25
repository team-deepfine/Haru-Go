package appstore

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/base64"
	"fmt"
)

// Apple Root CA - G3 (Root certificate for App Store Server API)
// Subject: CN=Apple Root CA - G3, OU=Apple Certification Authority, O=Apple Inc., C=US
// This is the root of trust for all App Store Server API JWS signatures.
// Download: https://www.apple.com/certificateauthority/AppleRootCA-G3.cer
//
//nolint:lll
const appleRootCAPEM = `-----BEGIN CERTIFICATE-----
MIICQzCCAcmgAwIBAgIILcX8iNLFS5UwCgYIKoZIzj0EAwMwZzEbMBkGA1UEAwwS
QXBwbGUgUm9vdCBDQSAtIEczMSYwJAYDVQQLDB1BcHBsZSBDZXJ0aWZpY2F0aW9u
IEF1dGhvcml0eTETMBEGA1UECgwKQXBwbGUgSW5jLjELMAkGA1UEBhMCVVMwHhcN
MTQwNDMwMTgxOTA2WhcNMzkwNDMwMTgxOTA2WjBnMRswGQYDVQQDDBJBcHBsZSBS
b290IENBIC0gRzMxJjAkBgNVBAsMHUFwcGxlIENlcnRpZmljYXRpb24gQXV0aG9y
aXR5MRMwEQYDVQQKDApBcHBsZSBJbmMuMQswCQYDVQQGEwJVUzB2MBAGByqGSM49
AgEGBSuBBAAiA2IABJjpLz1AcqTtkyJygRMc3RCV8cWjTnHcFBbZDuWmBSp3ZHtf
TjjTuxxEtX/1H7YyYl3J6YRbTzBPEVoA/VhYDKX1DyxNB0cTddqXl5dvMVztK517
IDvYuVTZXpmkOlEKMaNCMEAwHQYDVR0OBBYEFLuw3qFYM4iapIqZ3r6966/ayySr
MA8GA1UdEwEB/wQFMAMBAf8wDgYDVR0PAQH/BAQDAgEGMAoGCCqGSM49BAMDA2gA
MGUCMQCD6cHEFl4aXTQY2e3v9GwOAEZLuN+yRhHFD/3meoyhpmvOwgPUnPWTxnS4
at+qIxUCMG1mihDK1A3UT82NQz60imOlM27jbdoXt2QfyFMm+YhidDkLF1vLUagM
6BgD56KyKA==
-----END CERTIFICATE-----`

var appleRootCAPool *x509.CertPool

func init() {
	appleRootCAPool = x509.NewCertPool()
	if !appleRootCAPool.AppendCertsFromPEM([]byte(appleRootCAPEM)) {
		panic("failed to parse Apple Root CA certificate")
	}
}

// verifyX5cChain verifies the x5c certificate chain against the Apple Root CA
// and returns the leaf certificate's ECDSA public key.
func verifyX5cChain(x5c []interface{}) (*ecdsa.PublicKey, error) {
	if len(x5c) == 0 {
		return nil, fmt.Errorf("empty x5c chain")
	}

	// Parse all certificates in the chain
	certs := make([]*x509.Certificate, 0, len(x5c))
	for i, raw := range x5c {
		certStr, ok := raw.(string)
		if !ok {
			return nil, fmt.Errorf("x5c[%d] is not a string", i)
		}
		certDER, err := base64.StdEncoding.DecodeString(certStr)
		if err != nil {
			return nil, fmt.Errorf("decode x5c[%d]: %w", i, err)
		}
		cert, err := x509.ParseCertificate(certDER)
		if err != nil {
			return nil, fmt.Errorf("parse x5c[%d]: %w", i, err)
		}
		certs = append(certs, cert)
	}

	// Build intermediate pool from non-leaf certificates
	intermediates := x509.NewCertPool()
	for _, cert := range certs[1:] {
		intermediates.AddCert(cert)
	}

	// Verify the leaf certificate chains to Apple Root CA
	leaf := certs[0]
	opts := x509.VerifyOptions{
		Roots:         appleRootCAPool,
		Intermediates: intermediates,
	}
	if _, err := leaf.Verify(opts); err != nil {
		return nil, fmt.Errorf("x5c chain does not chain to Apple Root CA: %w", err)
	}

	ecKey, ok := leaf.PublicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("leaf certificate does not contain ECDSA public key")
	}

	return ecKey, nil
}
