package rules

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"github.com/echocat/lingress/value"
	"slices"
	"strings"
	"sync"
)

type CertificateRepository interface {
	FindCertificatesBy(CertificateQuery) (Certificates, error)
}

type Certificates []*tls.Certificate

func (this *Certificates) add(in *tls.Certificate) bool {
	for _, candidate := range *this {
		if len(in.Certificate) == len(candidate.Certificate) &&
			bytes.Equal(in.Certificate[0], candidate.Certificate[0]) {
			return false
		}
	}
	*this = append(*this, in)
	return true
}

func (this *Certificates) addAll(ins []*tls.Certificate) bool {
	anyAdded := false
	for _, in := range ins {
		if this.add(in) {
			anyAdded = true
		}
	}
	return anyAdded
}

type certificates struct {
	all                  Certificates
	sourceToCertificates map[string]*Certificates
}

func (this certificates) hasContent() bool {
	return len(this.all) > 0
}

func (this *certificates) add(source string, ins ...*tls.Certificate) bool {
	if this.sourceToCertificates == nil {
		this.sourceToCertificates = map[string]*Certificates{
			source: (*Certificates)(&ins),
		}
		this.all = ins
		return true
	}

	ofSource, ok := this.sourceToCertificates[source]
	if !ok {
		ofSource = &Certificates{}
		this.sourceToCertificates[source] = ofSource
	}
	return ofSource.addAll(ins) || this.all.addAll(ins)
}

func (this *certificates) remove(source string) bool {
	if this.sourceToCertificates == nil {
		return false
	}

	_, ok := this.sourceToCertificates[source]
	if !ok {
		return false
	}
	delete(this.sourceToCertificates, source)

	anyRemoved := false

	this.all = slices.DeleteFunc(this.all, func(candidate *tls.Certificate) bool {
		anyOtherContains := false
		for _, others := range this.sourceToCertificates {
			for _, other := range *others {
				if len(other.Certificate) == len(candidate.Certificate) &&
					bytes.Equal(other.Certificate[0], candidate.Certificate[0]) {
					anyOtherContains = true
					break
				}
			}
			if anyOtherContains {
				break
			}
		}
		if !anyOtherContains {
			anyRemoved = true
			return true
		}
		return false
	})

	return anyRemoved
}

type certificatesByHostValue struct {
	direct   certificates
	wildcard certificates
}

func (this *certificatesByHostValue) remove(source string) bool {
	return this.direct.remove(source) || this.wildcard.remove(source)
}

func (this *certificatesByHostValue) hasContent() bool {
	return this.direct.hasContent() || this.wildcard.hasContent()
}

type CertificatesByHost struct {
	values map[value.WildcardSupportingFqdn]*certificatesByHostValue
	mutex  sync.RWMutex
}

func (this *CertificatesByHost) Find(host value.WildcardSupportingFqdn) Certificates {
	this.mutex.RLock()
	defer this.mutex.RUnlock()

	values := this.values
	if values == nil {
		return nil
	}
	direct := values[host]
	if len(direct.direct.all) > 0 {
		return direct.direct.all
	}

	i := strings.IndexRune(string(host), '.')
	if i <= 0 {
		return nil
	}
	shortHost := host[i+1:]
	return values[shortHost].wildcard.all
}

func (this *CertificatesByHost) RemoveBySource(source string) ([]value.WildcardSupportingFqdn, error) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	removed := map[value.WildcardSupportingFqdn]struct{}{}
	toPurge := map[value.WildcardSupportingFqdn]struct{}{}

	for host, target := range this.values {
		if target.remove(source) {
			removed[host] = struct{}{}
		}
		if !target.hasContent() {
			toPurge[host] = struct{}{}
		}
	}

	for host := range toPurge {
		delete(this.values, host)
	}

	removedS := make([]value.WildcardSupportingFqdn, len(removed))
	var ri int
	for host := range removed {
		removedS[ri] = host
		ri++
	}

	return removedS, nil
}

func (this *CertificatesByHost) Add(source string, certificate tls.Certificate) ([]value.WildcardSupportingFqdn, error) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	if len(certificate.Certificate) <= 0 {
		return nil, errors.New("empty certificate")
	}
	if certificate.PrivateKey == nil {
		return nil, errors.New("certificate without privateKey")
	}
	if certificate.Leaf == nil {
		leaf, err := x509.ParseCertificate(certificate.Certificate[0])
		if err != nil {
			return nil, err
		}
		certificate.Leaf = leaf
	}

	added := map[value.WildcardSupportingFqdn]struct{}{}

	var cnFqdn value.WildcardSupportingFqdn
	if err := cnFqdn.Set(certificate.Leaf.Subject.CommonName); err == nil {
		if this.add(source, cnFqdn, &certificate) {
			added[cnFqdn] = struct{}{}
		}
	}
	for _, dns := range certificate.Leaf.DNSNames {
		var dnsFqdn value.WildcardSupportingFqdn
		if err := dnsFqdn.Set(dns); err == nil {
			if this.add(source, dnsFqdn, &certificate) {
				added[dnsFqdn] = struct{}{}
			}
		}
	}

	addedS := make([]value.WildcardSupportingFqdn, len(added))
	var ai int
	for host := range added {
		addedS[ai] = host
		ai++
	}

	return addedS, nil
}

func (this *CertificatesByHost) add(source string, host value.WildcardSupportingFqdn, certificate *tls.Certificate) bool {
	wildcarded := false
	if strings.HasPrefix(string(host), "*.") {
		host = host[2:]
		wildcarded = true
	}
	if this.values == nil {
		this.values = make(map[value.WildcardSupportingFqdn]*certificatesByHostValue, 1)
	}
	target, ok := this.values[host]
	if !ok {
		target = &certificatesByHostValue{}
		this.values[host] = target
	}
	if wildcarded {
		return target.wildcard.add(source, certificate)
	} else {
		return target.direct.add(source, certificate)
	}
}

func (this *CertificatesByHost) AddBytes(source string, certificate, privateKey []byte) ([]value.WildcardSupportingFqdn, error) {
	cert, err := tls.X509KeyPair(certificate, privateKey)
	if err != nil {
		return nil, err
	}
	return this.Add(source, cert)
}
