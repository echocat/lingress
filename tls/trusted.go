package tls

import (
	"crypto/tls"
	"crypto/x509"
	"embed"
	"fmt"
	"net/http"
	"path"
)

var (
	//go:embed ca-certificates-mozilla
	caCertificates embed.FS

	Pool *x509.CertPool
)

func init() {
	fail := func(err error) {
		panic(fmt.Errorf("cannot load ca-certificates: %w", err))
	}
	failf := func(msg string, args ...any) {
		fail(fmt.Errorf(msg, args...))
	}

	Pool = x509.NewCertPool()

	entries, err := caCertificates.ReadDir(".")
	if err != nil {
		fail(err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if ext := path.Ext(entry.Name()); ext != ".pem" {
			continue
		}
		if content, err := caCertificates.ReadFile(entry.Name()); err != nil {
			failf("ca-certificate %q cannot be loaded: %w", entry, err)
		} else if !Pool.AppendCertsFromPEM(content) {
			failf("ca-certificate %q cannot be loaded: illegal format", entry)
		}
	}
	http.DefaultTransport = &http.Transport{
		TLSClientConfig: &tls.Config{RootCAs: Pool},
	}
}
