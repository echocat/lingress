package support

import (
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"math/big"
	"time"
)

func CreateDummyCertificate() (tls.Certificate, error) {
	fail := func(err error) (tls.Certificate, error) {
		return tls.Certificate{}, errors.Wrap(err, "cannot create dummy certificate")
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
	u := uuid.New()

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: u.String(),
		},
		NotBefore: now,
		NotAfter:  now.Add(time.Hour * 24 * 365),

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
