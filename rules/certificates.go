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

func (this CertificatesByHost) Find(host string) Certificates {
	values := this.values
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

func (this *CertificatesByHost) Add(certificate tls.Certificate) error {
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
		this.add(certificate.Leaf.Subject.CommonName, &certificate)
	}
	for _, dns := range certificate.Leaf.DNSNames {
		this.add(dns, &certificate)
	}

	return nil
}

func (this *CertificatesByHost) add(host string, certificate *tls.Certificate) {
	wildcarded := false
	if strings.HasPrefix(host, "*.") {
		host = host[2:]
		wildcarded = true
	}
	if this.values == nil {
		this.values = map[string]certificatesByHostValue{}
	}
	existing := this.values[host]
	if wildcarded {
		existing.wildcard = append(existing.wildcard, certificate)
	} else {
		existing.direct = append(existing.direct, certificate)
	}
	this.values[host] = existing
}

func (this *CertificatesByHost) AddBytes(certificate, privateKey []byte) error {
	cert, err := tls.X509KeyPair(certificate, privateKey)
	if err != nil {
		return err
	}
	return this.Add(cert)
}
