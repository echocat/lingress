package rules

import "github.com/echocat/lingress/value"

var _ = RegisterDefaultOptionsPart(&OptionsSecure{})

const (
	optionsSecureKey = "secure"

	annotationForceSecure        = "lingress.echocat.org/force-secure"
	annotationWhitelistedRemotes = "lingress.echocat.org/whitelisted-remotes"
)

func OptionsSecureOf(rule Rule) *OptionsSecure {
	if rule == nil {
		return &OptionsSecure{}
	}
	if v, ok := rule.Options()[optionsSecureKey].(*OptionsSecure); ok {
		return v
	}
	return &OptionsSecure{}
}

type OptionsSecure struct {
	ForceSecure        value.Bool `json:"forceSecure,omitempty"`
	WhitelistedRemotes []Address  `json:"whitelistedRemotes,omitempty"`
}

func (this OptionsSecure) Name() string {
	return optionsSecureKey
}

func (this OptionsSecure) IsRelevant() bool {
	return this.ForceSecure.IsPresent() ||
		len(this.WhitelistedRemotes) > 0
}

func (this *OptionsSecure) Set(annotations Annotations) (err error) {
	if this.ForceSecure, err = evaluateOptionForceSecure(annotations); err != nil {
		return
	}
	if this.WhitelistedRemotes, err = evaluateOptionWhitelistedRemotes(annotations); err != nil {
		return
	}
	return
}

func evaluateOptionForceSecure(annotations map[string]string) (value.Bool, error) {
	if v, ok := annotations[annotationForceSecure]; ok {
		return AnnotationIsBool(annotationForceSecure, v)
	}
	return value.UndefinedBool(), nil
}

func evaluateOptionWhitelistedRemotes(annotations map[string]string) ([]Address, error) {
	if v, ok := annotations[annotationWhitelistedRemotes]; ok {
		return AnnotationAddresses(annotationWhitelistedRemotes, v)
	}
	return nil, nil
}
