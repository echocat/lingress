package rules

import "github.com/echocat/lingress/value"

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
	Enabled            value.ForcibleBool `json:"enabled,omitempty"`
	AllowedOriginsHost HostPatterns       `json:"allowedOriginsHost,omitempty"`
	AllowedMethods     Methods            `json:"allowedMethods,omitempty"`
	AllowedHeaders     HeaderNames        `json:"allowedHeaders,omitempty"`
	AllowedCredentials value.Bool         `json:"allowedCredentials,omitempty"`
	MaxAge             value.Duration     `json:"maxAge,omitempty"`
}

func (instance OptionsCors) Name() string {
	return optionsCorsKey
}

func (instance OptionsCors) IsRelevant() bool {
	return instance.Enabled.IsPresent() ||
		instance.AllowedOriginsHost.IsPresent() ||
		instance.AllowedMethods.IsPresent() ||
		instance.AllowedHeaders.IsPresent() ||
		instance.AllowedCredentials.IsPresent() ||
		instance.MaxAge.IsPresent()
}

func (instance *OptionsCors) Set(annotations Annotations) (err error) {
	if instance.Enabled, err = evaluateOptionCorsEnable(annotations); err != nil {
		return
	}
	if instance.AllowedOriginsHost, err = evaluateOptionCorsAllowedOriginsHosts(annotations); err != nil {
		return
	}
	if instance.AllowedMethods, err = evaluateOptionCorsAllowedMethods(annotations); err != nil {
		return
	}
	if instance.AllowedHeaders, err = evaluateOptionCorsAllowedHeaders(annotations); err != nil {
		return
	}
	if instance.AllowedCredentials, err = evaluateOptionCorsAllowedCredentials(annotations); err != nil {
		return
	}
	if instance.MaxAge, err = evaluateOptionCorsMaxAge(annotations); err != nil {
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
	return value.NewForcibleBool(value.UndefinedBool, false), nil
}

func evaluateOptionCorsAllowedOriginsHosts(annotations map[string]string) (HostPatterns, error) {
	if v, ok := annotations[annotationCorsAllowedOriginsHosts]; ok {
		return ParseHostPatterns(v)
	}
	return nil, nil
}

func evaluateOptionCorsAllowedMethods(annotations map[string]string) (Methods, error) {
	if v, ok := annotations[annotationCorsAllowedMethods]; ok {
		return ParseMethods(v)
	}
	return nil, nil
}

func evaluateOptionCorsAllowedHeaders(annotations map[string]string) (HeaderNames, error) {
	if v, ok := annotations[annotationCorsAllowedHeaders]; ok {
		return ParseHeaderNames(v)
	}
	return nil, nil
}

func evaluateOptionCorsAllowedCredentials(annotations map[string]string) (value.Bool, error) {
	if v, ok := annotations[annotationCorsAllowedCredentials]; ok {
		return AnnotationIsTrue(annotationCorsAllowedCredentials, v)
	}
	return value.UndefinedBool, nil
}

func evaluateOptionCorsMaxAge(annotations map[string]string) (value.Duration, error) {
	if v, ok := annotations[annotationCorsMaxAge]; ok {
		return value.ParseDuration(v)
	}
	return 0, nil
}
