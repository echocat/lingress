package rules

import (
	value2 "github.com/echocat/lingress/rules/value"
	"github.com/echocat/lingress/value"
)

var _ = RegisterDefaultOptionsPart(&OptionsCors{})

const (
	optionsCorsKey = "cors"

	annotationCorsEnabledAlternative  = "lingress.echocat.org/cors"
	annotationCorsEnabled             = "lingress.echocat.org/cors.enabled"
	annotationCorsAllowedOriginsHosts = "lingress.echocat.org/cors.allowed-origin-hosts"
	annotationCorsAllowedMethods      = "lingress.echocat.org/cors.allowed-methods"
	annotationCorsAllowedHeaders      = "lingress.echocat.org/cors.allowed-headers"
	annotationCorsAllowedCredentials  = "lingress.echocat.org/cors.credentials"
	annotationCorsMaxAge              = "lingress.echocat.org/cors.max-age"
)

func OptionsCorsOf(rule Rule) *OptionsCors {
	if rule == nil {
		return &OptionsCors{}
	}
	if v, ok := rule.Options()[optionsCorsKey].(*OptionsCors); ok {
		return v
	}
	return &OptionsCors{}
}

type OptionsCors struct {
	Enabled            value.ForcibleBool  `json:"enabled,omitempty"`
	AllowedOriginsHost value2.HostPatterns `json:"allowedOriginsHost,omitempty"`
	AllowedMethods     value2.Methods      `json:"allowedMethods,omitempty"`
	AllowedHeaders     value2.HeaderNames  `json:"allowedHeaders,omitempty"`
	AllowedCredentials value.Bool          `json:"allowedCredentials,omitempty"`
	MaxAge             value.Duration      `json:"maxAge,omitempty"`
}

func (this OptionsCors) Name() string {
	return optionsCorsKey
}

func (this OptionsCors) IsRelevant() bool {
	return this.Enabled.IsPresent() ||
		this.AllowedOriginsHost.IsPresent() ||
		this.AllowedMethods.IsPresent() ||
		this.AllowedHeaders.IsPresent() ||
		this.AllowedCredentials.IsPresent() ||
		this.MaxAge.IsPresent()
}

func (this *OptionsCors) Set(annotations Annotations) (err error) {
	if this.Enabled, err = evaluateOptionCorsEnable(annotations); err != nil {
		return
	}
	if this.AllowedOriginsHost, err = evaluateOptionCorsAllowedOriginsHosts(annotations); err != nil {
		return
	}
	if this.AllowedMethods, err = evaluateOptionCorsAllowedMethods(annotations); err != nil {
		return
	}
	if this.AllowedHeaders, err = evaluateOptionCorsAllowedHeaders(annotations); err != nil {
		return
	}
	if this.AllowedCredentials, err = evaluateOptionCorsAllowedCredentials(annotations); err != nil {
		return
	}
	if this.MaxAge, err = evaluateOptionCorsMaxAge(annotations); err != nil {
		return
	}
	return
}

func evaluateOptionCorsEnable(annotations map[string]string) (value.ForcibleBool, error) {
	if v, ok := annotations[annotationCorsEnabled]; ok {
		return AnnotationIsForcibleBool(annotationCorsEnabled, v)
	}
	if v, ok := annotations[annotationCorsEnabledAlternative]; ok {
		return AnnotationIsForcibleBool(annotationCorsEnabledAlternative, v)
	}
	return value.NewForcibleBool(value.UndefinedBool(), false), nil
}

func evaluateOptionCorsAllowedOriginsHosts(annotations map[string]string) (value2.HostPatterns, error) {
	if v, ok := annotations[annotationCorsAllowedOriginsHosts]; ok {
		return value2.ParseHostPatterns(v)
	}
	return nil, nil
}

func evaluateOptionCorsAllowedMethods(annotations map[string]string) (value2.Methods, error) {
	if v, ok := annotations[annotationCorsAllowedMethods]; ok {
		return value2.ParseMethods(v)
	}
	return nil, nil
}

func evaluateOptionCorsAllowedHeaders(annotations map[string]string) (value2.HeaderNames, error) {
	if v, ok := annotations[annotationCorsAllowedHeaders]; ok {
		return value2.ParseHeaderNames(v)
	}
	return nil, nil
}

func evaluateOptionCorsAllowedCredentials(annotations map[string]string) (value.Bool, error) {
	if v, ok := annotations[annotationCorsAllowedCredentials]; ok {
		return AnnotationIsTrue(annotationCorsAllowedCredentials, v)
	}
	return value.UndefinedBool(), nil
}

func evaluateOptionCorsMaxAge(annotations map[string]string) (value.Duration, error) {
	if v, ok := annotations[annotationCorsMaxAge]; ok {
		return value.ParseDuration(v)
	}
	return value.Duration{}, nil
}
