package rules

import (
	"fmt"
	"k8s.io/api/extensions/v1beta1"
	"net"
	"net/textproto"
	"strings"
)

const (
	annotationCors                = "lingress.echocat.org/cors"
	annotationStripRulePathPrefix = "lingress.echocat.org/strip-rule-path-prefix"
	annotationPrefix              = "lingress.echocat.org/path-prefix"
	annotationXForwardedPrefix    = "lingress.echocat.org/x-forwarded-prefix"
	annotationForceSecure         = "lingress.echocat.org/force-secure"
	annotationWhitelistedRemotes  = "lingress.echocat.org/whitelisted-remotes"
	annotationRequestHeaders      = "lingress.echocat.org/headers-request"
	annotationResponseHeaders     = "lingress.echocat.org/headers-response"

	annotationNginxEnableCors           = "nginx.ingress.kubernetes.io/enable-cors"
	annotationNginxRewriteTarget        = "nginx.ingress.kubernetes.io/rewrite-target"
	annotationNginxXForwardedPrefix     = "nginx.ingress.kubernetes.io/x-forwarded-prefix"
	annotationNginxForceSslRedirect     = "nginx.ingress.kubernetes.io/force-ssl-redirect"
	annotationNginxWhitelistSourceRange = "nginx.ingress.kubernetes.io/whitelist-source-range"
)

type Options struct {
	Cors OptionalBool `json:"cors,omitempty"`

	ForceSecure OptionalBool `json:"forceSecure,omitempty"`

	StripRulePathPrefix OptionalBool `json:"stripRulePathPrefix,omitempty"`
	PathPrefix          []string     `json:"pathPrefix,omitempty"`
	XForwardedPrefix    OptionalBool `json:"xForwardedPrefix,omitempty"`

	WhitelistedRemotes []Address `json:"whitelistedRemotes,omitempty"`

	RequestHeaders  Headers `json:"requestHeaders,omitempty"`
	ResponseHeaders Headers `json:"responseHeaders,omitempty"`
}

func (instance Options) IsRelevant() bool {
	return instance.Cors > 0 ||
		instance.ForceSecure > 0 ||
		instance.StripRulePathPrefix > 0 ||
		len(instance.PathPrefix) > 0 ||
		instance.XForwardedPrefix > 0 ||
		len(instance.WhitelistedRemotes) > 0
}

func optionsForIngress(ingress *v1beta1.Ingress) (result Options, err error) {
	annotations := ingress.GetAnnotations()

	if result.Cors, err = evaluateOptionEnableCors(annotations); err != nil {
		return
	}
	if result.ForceSecure, err = evaluateOptionForceSecure(annotations); err != nil {
		return
	}
	if result.PathPrefix, err = evaluateOptionPathPrefix(annotations); err != nil {
		return
	}
	if result.StripRulePathPrefix, err = evaluateOptionStripRulePathPrefix(annotations); err != nil {
		return
	}
	if result.XForwardedPrefix, err = evaluateOptionXForwardedPrefix(annotations); err != nil {
		return
	}
	if result.WhitelistedRemotes, err = evaluateOptionWhitelistedRemotes(annotations); err != nil {
		return
	}
	if result.RequestHeaders, err = evaluateOptionHeaders(annotations, annotationRequestHeaders); err != nil {
		return
	}
	if result.ResponseHeaders, err = evaluateOptionHeaders(annotations, annotationResponseHeaders); err != nil {
		return
	}

	return
}

func evaluateOptionEnableCors(annotations map[string]string) (OptionalBool, error) {
	if v, ok := annotations[annotationForceSecure]; ok {
		return annotationIsTrue(annotationForceSecure, v)
	}
	if v, ok := annotations[annotationNginxEnableCors]; ok {
		return annotationIsTrue(annotationNginxEnableCors, v)
	}
	return NotDefined, nil
}

func evaluateOptionForceSecure(annotations map[string]string) (OptionalBool, error) {
	if v, ok := annotations[annotationCors]; ok {
		return annotationIsTrue(annotationCors, v)
	}
	if v, ok := annotations[annotationNginxForceSslRedirect]; ok {
		return annotationIsTrue(annotationNginxForceSslRedirect, v)
	}
	return NotDefined, nil
}

func evaluateOptionStripRulePathPrefix(annotations map[string]string) (OptionalBool, error) {
	if v, ok := annotations[annotationStripRulePathPrefix]; ok {
		return annotationIsTrue(annotationStripRulePathPrefix, v)
	}
	if _, ok := annotations[annotationNginxRewriteTarget]; ok {
		return True, nil
	}
	return NotDefined, nil
}

func evaluateOptionPathPrefix(annotations map[string]string) ([]string, error) {
	if v, ok := annotations[annotationPrefix]; ok {
		return ParsePath(v, false)
	}
	if v := annotations[annotationNginxRewriteTarget]; v != "" {
		return ParsePath(v, false)
	}
	return []string{}, nil
}

func evaluateOptionXForwardedPrefix(annotations map[string]string) (OptionalBool, error) {
	if v, ok := annotations[annotationXForwardedPrefix]; ok {
		return annotationIsTrue(annotationXForwardedPrefix, v)
	}
	if v, ok := annotations[annotationNginxXForwardedPrefix]; ok {
		return annotationIsTrue(annotationNginxXForwardedPrefix, v)
	}
	return NotDefined, nil
}

func evaluateOptionWhitelistedRemotes(annotations map[string]string) ([]Address, error) {
	if v, ok := annotations[annotationWhitelistedRemotes]; ok {
		return annotationAddresses(annotationWhitelistedRemotes, v)
	}
	if v, ok := annotations[annotationNginxWhitelistSourceRange]; ok {
		return annotationAddresses(annotationNginxWhitelistSourceRange, v)
	}
	return nil, nil
}

func evaluateOptionHeaders(annotations map[string]string, name string) (result Headers, err error) {
	result = Headers{}
	if pvs, ok := annotations[name]; ok {
		vs := strings.Split(strings.ReplaceAll(pvs, "\r", ""), "\n")
		for _, v := range vs {
			v = strings.TrimSpace(v)
			if err := result.Set(v); err != nil {
				return Headers{}, fmt.Errorf("illegal header value for annotation %s: %s", name, v)
			}
		}
	}
	return
}

func annotationIsTrue(name, value string) (OptionalBool, error) {
	if value == "true" {
		return True, nil
	}
	if value == "false" {
		return False, nil
	}
	return 0, fmt.Errorf("illegal boolean value for annotation %s: %s", name, value)
}

func annotationAddresses(name, value string) (result []Address, err error) {
	for _, candidate := range strings.Split(value, ",") {
		if strings.Contains(candidate, "/") {
			if _, n, pErr := net.ParseCIDR(candidate); pErr != nil {
				return nil, fmt.Errorf("'%s' is an illegal CIDR for annotation '%s': %v", candidate, name, pErr)
			} else {
				result = append(result, &networkAddress{n})
			}
		} else {
			if ips, pErr := net.LookupIP(candidate); pErr != nil {
				return nil, fmt.Errorf("'%s' is an illegal address for annotation '%s': %v", candidate, name, pErr)
			} else {
				for _, ip := range ips {
					result = append(result, ipAddress(ip))
				}
			}
		}
	}
	return
}

type OptionalBool uint8

const (
	NotDefined = OptionalBool(0)
	False      = OptionalBool(1)
	True       = OptionalBool(2)
)

func (instance OptionalBool) IsEnabled(def bool) bool {
	switch instance {
	case True:
		return true
	case False:
		return false
	}
	return def
}

func (instance OptionalBool) IsEnabledOrForced(def ForceableBool) bool {
	if def.Forced {
		return def.Value
	}
	switch instance {
	case True:
		return true
	case False:
		return false
	}
	return def.Value
}

func (instance OptionalBool) String() string {
	switch instance {
	case True:
		return "true"
	case False:
		return "false"
	}
	return ""
}

func (instance *OptionalBool) Set(plain string) error {
	switch plain {
	case "true":
		*instance = True
		return nil
	case "false":
		*instance = False
		return nil
	case "":
		*instance = NotDefined
		return nil
	}
	return fmt.Errorf("illegal value: %s", plain)
}

type ForceableBool struct {
	Value  bool
	Forced bool
}

func (instance ForceableBool) String() string {
	if instance.Forced {
		if instance.Value {
			return "!true"
		} else {
			return "!false"
		}
	} else {
		if instance.Value {
			return "true"
		} else {
			return "false"
		}
	}
}

func (instance *ForceableBool) Set(plain string) error {
	switch plain {
	case "true":
		*instance = ForceableBool{
			Value:  true,
			Forced: false,
		}
		return nil
	case "false":
		*instance = ForceableBool{
			Value:  true,
			Forced: false,
		}
		return nil
	case "!true":
		*instance = ForceableBool{
			Value:  true,
			Forced: true,
		}
		return nil
	case "!false":
		*instance = ForceableBool{
			Value:  true,
			Forced: true,
		}
		return nil
	default:
		return fmt.Errorf("illegal value: %s", plain)
	}
}

type Header struct {
	Key    string
	Value  string
	Forced bool
	Add    bool
	Del    bool
}

func (instance *Header) Set(plain string) error {
	in := plain
	r := Header{}

	if len(in) > 0 && in[0] == '!' {
		r.Forced = true
		in = in[1:]
	}

	if len(in) > 0 && in[0] == '-' {
		r.Del = true
		in = in[1:]
	} else if len(in) > 0 && in[0] == '+' {
		r.Add = true
		in = in[1:]
	}

	ci := strings.IndexRune(in, ':')
	if ci >= 0 && r.Del {
		return fmt.Errorf("illegal header rule '%s': delete actions should not have a value", plain)
	}
	if ci < 0 && !r.Del {
		return fmt.Errorf("illegal header rule '%s': add or set actions should have a value", plain)
	}

	if ci >= 0 {
		r.Key = textproto.CanonicalMIMEHeaderKey(in[:ci])
		r.Value = strings.TrimSpace(in[ci+1:])
	} else {
		r.Key = textproto.CanonicalMIMEHeaderKey(in)
	}
	if r.Key == "" {
		return fmt.Errorf("illegal header rule '%s': key is empty", plain)
	}

	*instance = r
	return nil
}

func (instance Header) String() string {
	result := ""
	if instance.Forced {
		result += "!"
	}
	if instance.Del {
		result += "-"
	} else if instance.Add {
		result += "+"
	}

	result += instance.Key

	if !instance.Del {
		result += ":" + instance.Value
	}

	return result
}

type Headers []Header

func (instance Headers) String() string {
	result := ""
	for i, val := range instance {
		if i > 0 {
			result += "\n"
		}
		result += val.String()
	}
	return result
}

func (instance *Headers) IsCumulative() bool {
	return true
}

func (instance *Headers) Set(plain string) error {
	var v Header
	if err := v.Set(plain); err != nil {
		return err
	}
	*instance = append(*instance, v)
	return nil
}
