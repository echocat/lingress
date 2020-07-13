package rules

var _ = RegisterDefaultOptionsPart(&OptionsSecure{})

const (
	optionsSecureKey = "secure"

	annotationForceSecure        = "lingress.echocat.org/force-secure"
	annotationWhitelistedRemotes = "lingress.echocat.org/whitelisted-remotes"

	annotationNginxForceSslRedirect     = "nginx.ingress.kubernetes.io/force-ssl-redirect"
	annotationNginxWhitelistSourceRange = "nginx.ingress.kubernetes.io/whitelist-source-range"
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
	ForceSecure        Bool      `json:"forceSecure,omitempty"`
	WhitelistedRemotes []Address `json:"whitelistedRemotes,omitempty"`
}

func (instance OptionsSecure) Name() string {
	return optionsSecureKey
}

func (instance OptionsSecure) IsRelevant() bool {
	return instance.ForceSecure > 0 ||
		len(instance.WhitelistedRemotes) > 0
}

func (instance *OptionsSecure) Set(annotations Annotations) (err error) {
	if instance.ForceSecure, err = evaluateOptionForceSecure(annotations); err != nil {
		return
	}
	if instance.WhitelistedRemotes, err = evaluateOptionWhitelistedRemotes(annotations); err != nil {
		return
	}
	return
}

func evaluateOptionForceSecure(annotations map[string]string) (Bool, error) {
	if v, ok := annotations[annotationForceSecure]; ok {
		return AnnotationIsTrue(annotationForceSecure, v)
	}
	if v, ok := annotations[annotationNginxForceSslRedirect]; ok {
		return AnnotationIsTrue(annotationNginxForceSslRedirect, v)
	}
	return NotDefined, nil
}

func evaluateOptionWhitelistedRemotes(annotations map[string]string) ([]Address, error) {
	if v, ok := annotations[annotationWhitelistedRemotes]; ok {
		return AnnotationAddresses(annotationWhitelistedRemotes, v)
	}
	if v, ok := annotations[annotationNginxWhitelistSourceRange]; ok {
		return AnnotationAddresses(annotationNginxWhitelistSourceRange, v)
	}
	return nil, nil
}
