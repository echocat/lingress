package settings

import (
	"fmt"
	"github.com/echocat/lingress/support"
	"github.com/echocat/lingress/value"
	"regexp"
)

func NewTls() (Tls, error) {
	return Tls{
		SecretNames:         []string{},
		SecretNamePattern:   nil,
		SecretLabelSelector: []string{},
		SecretFieldSelector: []string{},
		Forced:              value.NewForcibleBool(value.False(), false),
		FallbackCertificate: value.False(),
	}, nil
}

type Tls struct {
	SecretNames         []string           `yaml:"secretNames,omitempty" json:"secretNames,omitempty"`
	SecretNamePattern   *regexp.Regexp     `yaml:"secretNamePattern,omitempty" json:"secretNamePattern,omitempty"`
	SecretLabelSelector []string           `yaml:"secretLabelSelector,omitempty" json:"secretLabelSelector,omitempty"`
	SecretFieldSelector []string           `yaml:"secretFieldSelector,omitempty" json:"secretFieldSelector,omitempty"`
	Forced              value.ForcibleBool `yaml:"forced,omitempty" json:"forced,omitempty"`
	FallbackCertificate value.Bool         `yaml:"fallbackCertificate,omitempty" json:"fallbackCertificate,omitempty"`
}

func (this *Tls) RegisterFlags(fe support.FlagEnabled, appPrefix string) {
	fe.Flag("tls.secretNames", "Name of the secrets that contains the tls keys and certificates.").
		PlaceHolder("<secret name[,...<secret name>]>").
		Envar(support.FlagEnvName(appPrefix, "TLS_SECRET_NAMES")).
		StringsVar(&this.SecretNames)
	fe.Flag("tls.secretNamePatterns", "Patterns for name of the secrets that contains the tls keys and certificates.").
		PlaceHolder("<regex_pattern>").
		Envar(support.FlagEnvName(appPrefix, "TLS_SECRET_NAME_PATTERNS")).
		RegexpVar(&this.SecretNamePattern)
	fe.Flag("tls.secretLabelSelector", "Label selector to filter the secrets that contains the tls keys and certificates by.").
		PlaceHolder("<label>=<value>[,..]").
		Envar(support.FlagEnvName(appPrefix, "TLS_SECRET_LABEL_SELECTOR")).
		StringsVar(&this.SecretLabelSelector)
	fe.Flag("tls.secretFieldSelector", "Field selector to filter the secrets that contains the tls keys and certificates by.").
		PlaceHolder("<field>=<value>[,..]").
		Envar(support.FlagEnvName(appPrefix, "TLS_SECRET_FIELD_SELECTOR")).
		StringsVar(&this.SecretFieldSelector)
	fe.Flag("tls.forced", "If set if will be used if annotation lingress.echocat.org/force-secure is absent. If this value is prefix with ! it overrides everything regardless what was set in the annotation.").
		PlaceHolder(fmt.Sprint(this.Forced)).
		Envar(support.FlagEnvName(appPrefix, "TLS_FORCED")).
		SetValue(&this.Forced)
	fe.Flag("tls.fallbackCertificate", "If set lingress will use a fallback certificate for every request for which no other certificate can be determined.").
		PlaceHolder(this.FallbackCertificate.String()).
		Envar(support.FlagEnvName(appPrefix, "TLS_FALLBACK_CERTIFICATE")).
		SetValue(&this.FallbackCertificate)
}
