package rules

var _ = RegisterDefaultOptionsPart(&OptionsPrefix{})

const (
	optionsPrefixKey = "prefix"

	annotationStripRulePathPrefix = "lingress.echocat.org/strip-rule-path-prefix"
	annotationPathPrefix          = "lingress.echocat.org/path-prefix"
	annotationXForwardedPrefix    = "lingress.echocat.org/x-forwarded-prefix"

	annotationNginxRewriteTarget    = "nginx.ingress.kubernetes.io/rewrite-target"
	annotationNginxXForwardedPrefix = "nginx.ingress.kubernetes.io/x-forwarded-prefix"
)

func OptionsPrefixOf(options Options) *OptionsPrefix {
	if v, ok := options[optionsPrefixKey].(*OptionsPrefix); ok {
		return v
	}
	return &OptionsPrefix{}
}

type OptionsPrefix struct {
	StripRulePathPrefix OptionalBool `json:"stripRulePathPrefix,omitempty"`
	PathPrefix          []string     `json:"pathPrefix,omitempty"`
	XForwardedPrefix    OptionalBool `json:"xForwardedPrefix,omitempty"`
}

func (instance OptionsPrefix) Name() string {
	return optionsPrefixKey
}

func (instance OptionsPrefix) IsRelevant() bool {
	return instance.StripRulePathPrefix > 0 ||
		len(instance.PathPrefix) > 0 ||
		instance.XForwardedPrefix > 0
}

func (instance *OptionsPrefix) Set(annotations Annotations) (err error) {
	if instance.PathPrefix, err = evaluateOptionPathPrefix(annotations); err != nil {
		return
	}
	if instance.StripRulePathPrefix, err = evaluateOptionStripRulePathPrefix(annotations); err != nil {
		return
	}
	if instance.XForwardedPrefix, err = evaluateOptionXForwardedPrefix(annotations); err != nil {
		return
	}
	return
}

func evaluateOptionStripRulePathPrefix(annotations map[string]string) (OptionalBool, error) {
	if v, ok := annotations[annotationStripRulePathPrefix]; ok {
		return AnnotationIsTrue(annotationStripRulePathPrefix, v)
	}
	if _, ok := annotations[annotationNginxRewriteTarget]; ok {
		return True, nil
	}
	return NotDefined, nil
}

func evaluateOptionPathPrefix(annotations map[string]string) ([]string, error) {
	if v, ok := annotations[annotationPathPrefix]; ok {
		return ParsePath(v, false)
	}
	if v := annotations[annotationNginxRewriteTarget]; v != "" {
		return ParsePath(v, false)
	}
	return []string{}, nil
}

func evaluateOptionXForwardedPrefix(annotations map[string]string) (OptionalBool, error) {
	if v, ok := annotations[annotationXForwardedPrefix]; ok {
		return AnnotationIsTrue(annotationXForwardedPrefix, v)
	}
	if v, ok := annotations[annotationNginxXForwardedPrefix]; ok {
		return AnnotationIsTrue(annotationNginxXForwardedPrefix, v)
	}
	return NotDefined, nil
}
