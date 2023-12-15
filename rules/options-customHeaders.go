package rules

import (
	"fmt"
	"github.com/echocat/lingress/value"
	"strings"
)

var _ = RegisterDefaultOptionsPart(&OptionsCustomHeaders{})

const (
	optionsCustomerHeadersKey = "custom-headers"

	annotationRequestHeaders  = "lingress.echocat.org/headers-request"
	annotationResponseHeaders = "lingress.echocat.org/headers-response"
)

func OptionsCustomHeadersOf(rule Rule) *OptionsCustomHeaders {
	if rule == nil {
		return &OptionsCustomHeaders{}
	}
	if v, ok := rule.Options()[optionsCustomerHeadersKey].(*OptionsCustomHeaders); ok {
		return v
	}
	return &OptionsCustomHeaders{}
}

type OptionsCustomHeaders struct {
	RequestHeaders  value.Headers `json:"requestHeaders,omitempty"`
	ResponseHeaders value.Headers `json:"responseHeaders,omitempty"`
}

func (this OptionsCustomHeaders) Name() string {
	return optionsCustomerHeadersKey
}

func (this OptionsCustomHeaders) IsRelevant() bool {
	return len(this.RequestHeaders) > 0 ||
		len(this.ResponseHeaders) > 0
}

func (this *OptionsCustomHeaders) Set(annotations Annotations) (err error) {
	if this.RequestHeaders, err = evaluateOptionHeaders(annotations, annotationRequestHeaders); err != nil {
		return
	}
	if this.ResponseHeaders, err = evaluateOptionHeaders(annotations, annotationResponseHeaders); err != nil {
		return
	}
	return
}

func evaluateOptionHeaders(annotations map[string]string, name string) (result value.Headers, err error) {
	result = value.Headers{}
	if pvs, ok := annotations[name]; ok {
		vs := strings.Split(strings.ReplaceAll(pvs, "\r", ""), "\n")
		for _, v := range vs {
			v = strings.TrimSpace(v)
			if err := result.Set(v); err != nil {
				return value.Headers{}, fmt.Errorf("illegal header value for annotation %s: %s", name, v)
			}
		}
	}
	return
}
