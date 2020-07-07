package rules

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

	annotationNginxCorsEnable             = "nginx.ingress.kubernetes.io/enable-cors"
	annotationNginxCorsAllowedMethods     = "nginx.ingress.kubernetes.io/cors-allow-methods"
	annotationNginxCorsAllowedHeaders     = "nginx.ingress.kubernetes.io/cors-allow-headers"
	annotationNginxCorsAllowedCredentials = "nginx.ingress.kubernetes.io/cors-allow-credentials"
	annotationNginxCorsMaxAge             = "nginx.ingress.kubernetes.io/cors-max-age"
)

func OptionsCorsOf(options Options) *OptionsCors {
	if v, ok := options[optionsCorsKey].(*OptionsCors); ok {
		return v
	}
	return &OptionsCors{}
}

type OptionsCors struct {
	Enabled            ForceableBool `json:"enabled,omitempty"`
	AllowedOriginsHost HostPatterns  `json:"allowedOriginsHost,omitempty"`
	AllowedMethods     Methods       `json:"allowedMethods,omitempty"`
	AllowedHeaders     HeaderNames   `json:"allowedHeaders,omitempty"`
	AllowedCredentials Bool          `json:"allowedCredentials,omitempty"`
	MaxAge             Duration      `json:"maxAge,omitempty"`
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

func evaluateOptionCorsEnable(annotations map[string]string) (ForceableBool, error) {
	if v, ok := annotations[annotationCorsEnabled]; ok {
		return AnnotationIsForceableBool(annotationCorsEnabled, v)
	}
	if v, ok := annotations[annotationCorsEnabledAlternative]; ok {
		return AnnotationIsForceableBool(annotationCorsEnabledAlternative, v)
	}
	if v, ok := annotations[annotationNginxCorsEnable]; ok {
		return AnnotationIsForceableBool(annotationNginxCorsEnable, v)
	}
	return NewForceableBool(NotDefined, false), nil
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
	if v, ok := annotations[annotationNginxCorsAllowedMethods]; ok {
		return ParseMethods(v)
	}
	return nil, nil
}

func evaluateOptionCorsAllowedHeaders(annotations map[string]string) (HeaderNames, error) {
	if v, ok := annotations[annotationCorsAllowedHeaders]; ok {
		return ParseHeaderNames(v)
	}
	if v, ok := annotations[annotationNginxCorsAllowedHeaders]; ok {
		return ParseHeaderNames(v)
	}
	return nil, nil
}

func evaluateOptionCorsAllowedCredentials(annotations map[string]string) (Bool, error) {
	if v, ok := annotations[annotationCorsAllowedCredentials]; ok {
		return AnnotationIsTrue(annotationCorsAllowedCredentials, v)
	}
	if v, ok := annotations[annotationNginxCorsAllowedCredentials]; ok {
		return AnnotationIsTrue(annotationNginxCorsAllowedCredentials, v)
	}
	return NotDefined, nil
}

func evaluateOptionCorsMaxAge(annotations map[string]string) (Duration, error) {
	if v, ok := annotations[annotationCorsMaxAge]; ok {
		return ParseDuration(v)
	}
	if v, ok := annotations[annotationNginxCorsMaxAge]; ok {
		return ParseDuration(v)
	}
	return 0, nil
}
