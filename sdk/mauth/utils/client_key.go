package utils

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"time"

	"golang.org/x/crypto/pkcs12"
)

func ExtractClientKey(password, privateKeyData string) (privateKeyPem string, ExpiresAt time.Time, err error) {

	base64DecodeTmp, err := base64.StdEncoding.DecodeString(privateKeyData)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("base64 decode error: %w", err)
	}

	blocks, err := pkcs12.ToPEM(base64DecodeTmp, password)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("pkcs12 to pem error: %w", err)
	}
	if len(blocks) < 2 {
		return "", time.Time{}, fmt.Errorf("pkcs12 to pem error: blocks length less than 2")
	}

	certBlock, _ := pem.Decode(pem.EncodeToMemory(blocks[0]))
	if certBlock == nil {
		return "", time.Time{}, fmt.Errorf("pkcs12 to pem error: cert block is nil")
	}

	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("parse certificate error: %w", err)
	}

	ExpiresAt = cert.NotAfter

	return string(pem.EncodeToMemory(blocks[1])), ExpiresAt, nil
}
