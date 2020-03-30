package rules

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"strings"
)

type CertificateRepository interface {
	FindCertificatesBy(CertificateQuery) (Certificates, error)
}

type Certificates []*tls.Certificate

type certificatesByHostValue struct {
	direct   Certificates
	wildcard Certificates
}

type CertificatesByHost struct {
	values map[string]certificatesByHostValue
}

func (instance CertificatesByHost) Find(host string) Certificates {
	values := instance.values
	if values == nil {
		return nil
	}
	direct := values[host]
	if len(direct.direct) > 0 {
		return direct.direct
	}

	i := strings.IndexRune(host, '.')
	if i <= 0 {
		return nil
	}
	shortHost := host[i+1:]
	return values[shortHost].wildcard
}

func (instance *CertificatesByHost) Add(certificate tls.Certificate) error {
	if len(certificate.Certificate) <= 0 {
		return errors.New("empty certificate")
	}
	if certificate.PrivateKey == nil {
		return errors.New("certificate without privateKey")
	}
	if certificate.Leaf == nil {
		leaf, err := x509.ParseCertificate(certificate.Certificate[0])
		if err != nil {
			return err
		}
		certificate.Leaf = leaf
	}

	if len(certificate.Leaf.Subject.CommonName) > 0 {
		instance.add(certificate.Leaf.Subject.CommonName, &certificate)
	}
	for _, dns := range certificate.Leaf.DNSNames {
		instance.add(dns, &certificate)
	}

	return nil
}

func (instance *CertificatesByHost) add(host string, certificate *tls.Certificate) {
	wildcarded := false
	if strings.HasPrefix(host, "*.") {
		host = host[2:]
		wildcarded = true
	}
	if instance.values == nil {
		instance.values = map[string]certificatesByHostValue{}
	}
	existing := instance.values[host]
	if wildcarded {
		existing.wildcard = append(existing.wildcard, certificate)
	} else {
		existing.direct = append(existing.direct, certificate)
	}
	instance.values[host] = existing
}

func (instance *CertificatesByHost) AddBytes(certificate, privateKey []byte) error {
	cert, err := tls.X509KeyPair(certificate, privateKey)
	if err != nil {
		return err
	}
	return instance.Add(cert)
}
