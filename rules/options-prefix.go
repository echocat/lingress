package rules

import "github.com/echocat/lingress/value"

var _ = RegisterDefaultOptionsPart(&OptionsPrefix{})

const (
	optionsPrefixKey = "prefix"

	annotationStripRulePathPrefix = "lingress.echocat.org/strip-rule-path-prefix"
	annotationPathPrefix          = "lingress.echocat.org/path-prefix"
	annotationXForwardedPrefix    = "lingress.echocat.org/x-forwarded-prefix"
)

func OptionsPrefixOf(rule Rule) *OptionsPrefix {
	if rule == nil {
		return &OptionsPrefix{}
	}
	if v, ok := rule.Options()[optionsPrefixKey].(*OptionsPrefix); ok {
		return v
	}
	return &OptionsPrefix{}
}

type OptionsPrefix struct {
	StripRulePathPrefix value.Bool `json:"stripRulePathPrefix,omitempty"`
	PathPrefix          []string   `json:"pathPrefix,omitempty"`
	XForwardedPrefix    value.Bool `json:"xForwardedPrefix,omitempty"`
}

func (this OptionsPrefix) Name() string {
	return optionsPrefixKey
}

func (this OptionsPrefix) IsRelevant() bool {
	return this.StripRulePathPrefix.IsPresent() ||
		len(this.PathPrefix) > 0 ||
		this.XForwardedPrefix.IsPresent()
}

func (this *OptionsPrefix) Set(annotations Annotations) (err error) {
	if this.PathPrefix, err = evaluateOptionPathPrefix(annotations); err != nil {
		return
	}
	if this.StripRulePathPrefix, err = evaluateOptionStripRulePathPrefix(annotations); err != nil {
		return
	}
	if this.XForwardedPrefix, err = evaluateOptionXForwardedPrefix(annotations); err != nil {
		return
	}
	return
}

func evaluateOptionStripRulePathPrefix(annotations map[string]string) (value.Bool, error) {
	if v, ok := annotations[annotationStripRulePathPrefix]; ok {
		return AnnotationIsBool(annotationStripRulePathPrefix, v)
	}
	return value.UndefinedBool(), nil
}

func evaluateOptionPathPrefix(annotations map[string]string) ([]string, error) {
	if v, ok := annotations[annotationPathPrefix]; ok {
		return ParsePath(v, false)
	}
	return []string{}, nil
}

func evaluateOptionXForwardedPrefix(annotations map[string]string) (value.Bool, error) {
	if v, ok := annotations[annotationXForwardedPrefix]; ok {
		return AnnotationIsBool(annotationXForwardedPrefix, v)
	}
	return value.UndefinedBool(), nil
}
