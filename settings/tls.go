package settings

import (
	"fmt"
	"github.com/echocat/lingress/support"
	"github.com/echocat/lingress/value"
	"strings"
)

func NewTls() (Tls, error) {
	return Tls{
		SecretNames:         []string{"certificates"},
		Forced:              value.NewForcibleBool(value.False(), false),
		FallbackCertificate: false,
	}, nil
}

type Tls struct {
	SecretNames         []string           `yaml:"secretNames,omitempty" json:"secretNames,omitempty"`
	Forced              value.ForcibleBool `yaml:"forced,omitempty" json:"forced,omitempty"`
	FallbackCertificate bool               `yaml:"fallbackCertificate,omitempty" json:"fallbackCertificate,omitempty"`
}

func (this *Tls) RegisterFlags(fe support.FlagEnabled, appPrefix string) {
	fe.Flag("tls.secretNames", "Name of the secret that contains the tls keys and certificates.").
		PlaceHolder(strings.Join(this.SecretNames, ",")).
		Envar(support.FlagEnvName(appPrefix, "TLS_SECRET_NAMES")).
		StringsVar(&this.SecretNames)
	fe.Flag("tls.forced", "If set if will be used if annotation lingress.echocat.org/force-secure is absent. If this value is prefix with ! it overrides everything regardless what was set in the annotation.").
		PlaceHolder(fmt.Sprint(this.Forced)).
		Envar(support.FlagEnvName(appPrefix, "TLS_FORCED")).
		SetValue(&this.Forced)
	fe.Flag("tls.fallbackCertificate", "If set lingress will use a fallback certificate for every request for which no other certificate can be determined.").
		PlaceHolder(fmt.Sprint(this.FallbackCertificate)).
		Envar(support.FlagEnvName(appPrefix, "TLS_FALLBACK_CERTIFICATE")).
		BoolVar(&this.FallbackCertificate)
}
