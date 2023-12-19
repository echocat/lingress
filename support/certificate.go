package support

import (
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"github.com/google/uuid"
	"math/big"
	"time"
)

func CreateDummyCertificate() (tls.Certificate, error) {
	return CreateDummyCertificateFor(uuid.New().String())
}

func CreateDummyCertificateFor(name string) (tls.Certificate, error) {
	fail := func(err error) (tls.Certificate, error) {
		return tls.Certificate{}, fmt.Errorf("cannot create dummy certificate: %w", err)
	}
	pk, err := CreatePrivateKey()
	if err != nil {
		return fail(err)
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return fail(err)
	}

	now := time.Now()

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: name,
		},
		NotBefore: now,
		NotAfter:  now.Add(time.Hour * 24 * 365 * 10),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, PublicKeyOf(pk), pk)
	if err != nil {
		return fail(err)
	}

	leaf, err := x509.ParseCertificate(derBytes)
	if err != nil {
		return fail(err)
	}

	return tls.Certificate{
		Certificate: [][]byte{
			derBytes,
		},
		PrivateKey: pk,
		Leaf:       leaf,
	}, nil
}
