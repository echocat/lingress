package settings

import (
	"fmt"
	value2 "github.com/echocat/lingress/rules/value"
	"github.com/echocat/lingress/support"
	"github.com/echocat/lingress/value"
	"time"
)

var (
	defaultCorsMaxAge = value.NewDuration(24 * time.Hour)
)

func NewCors() (Cors, error) {
	return Cors{
		AllowedOriginsHost: value2.NewForcibleHostPatterns(value2.HostPatterns{}, false),
		AllowedMethods:     value2.NewForcibleMethods(value2.Methods{}, false),
		AllowedHeaders:     value2.NewForcibleHeaders(value2.HeaderNames{}, false),
		AllowedCredentials: value.NewForcibleBool(value.True(), false),
		MaxAge:             value.NewForcibleDuration(defaultCorsMaxAge, false),
		Enabled:            value.NewForcibleBool(value.False(), false),
	}, nil
}

type Cors struct {
	AllowedOriginsHost value2.ForcibleHostPatterns `json:"allowedOriginsHost,omitempty" yaml:"allowedOriginsHost,omitempty"`
	AllowedMethods     value2.ForcibleMethods      `json:"allowedMethods,omitempty" yaml:"allowedMethods,omitempty"`
	AllowedHeaders     value2.ForcibleHeaderNames  `json:"allowedHeaders,omitempty" yaml:"allowedHeaders,omitempty"`
	AllowedCredentials value.ForcibleBool          `json:"allowedCredentials,omitempty" yaml:"allowedCredentials,omitempty"`
	MaxAge             value.ForcibleDuration      `json:"maxAge,omitempty" yaml:"maxAge,omitempty"`
	Enabled            value.ForcibleBool          `json:"enabled,omitempty" yaml:"enabled,omitempty"`
}

func (this *Cors) RegisterFlags(fe support.FlagEnabled, appPrefix string) {
	fe.Flag("cors.enabled", "If set if will be used if annotation lingress.echocat.org/cors is absent. If this value is prefix with ! it overrides everything regardless what was set in the annotation.").
		PlaceHolder(fmt.Sprint(this.Enabled)).
		Envar(support.FlagEnvName(appPrefix, "CORS_ENABLED")).
		SetValue(&this.Enabled)
	fe.Flag("cors.allowedOriginHosts", "Host patterns which are allowed to be referenced by Origin headers. '*' means all.").
		PlaceHolder(this.AllowedOriginsHost.String()).
		Envar(support.FlagEnvName(appPrefix, "CORS_ALLOWED_ORIGIN_HOSTS")).
		SetValue(&this.AllowedOriginsHost)
	fe.Flag("cors.allowedMethods", "Methods which are allowed to be referenced.").
		PlaceHolder(fmt.Sprint(this.AllowedMethods)).
		Envar(support.FlagEnvName(appPrefix, "CORS_ALLOWED_METHODS")).
		SetValue(&this.AllowedMethods)
	fe.Flag("cors.allowedHeaders", "Headers which are allowed to be referenced. '*' means all.").
		PlaceHolder(fmt.Sprint(this.AllowedHeaders)).
		Envar(support.FlagEnvName(appPrefix, "CORS_ALLOWED_HEADERS")).
		SetValue(&this.AllowedHeaders)
	fe.Flag("cors.allowedCredentials", "Credentials are allowed.").
		PlaceHolder(fmt.Sprint(this.AllowedCredentials)).
		Envar(support.FlagEnvName(appPrefix, "CORS_ALLOWED_CREDENTIALS")).
		SetValue(&this.AllowedCredentials)
	fe.Flag("cors.maxAge", "How long the response to the preflight request can be cached for without sending another preflight request").
		PlaceHolder(fmt.Sprint(this.MaxAge)).
		Envar(support.FlagEnvName(appPrefix, "CORS_MAX_AGE")).
		SetValue(&this.MaxAge)
}
