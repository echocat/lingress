package support

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"github.com/pkg/errors"
)

func CreatePrivateKey() (interface{}, error) {
	pk, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot create RSA 4096 private key")
	}
	return pk, err
}

func PublicKeyOf(candidate interface{}) interface{} {
	switch k := candidate.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey
	case *ecdsa.PrivateKey:
		return &k.PublicKey
	case ed25519.PrivateKey:
		return k.Public().(ed25519.PublicKey)
	default:
		return nil
	}
}
