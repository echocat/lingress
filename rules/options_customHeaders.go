package rules

import (
	"fmt"
	"net/textproto"
	"strings"
)

var _ = RegisterDefaultOptionsPart(&OptionsCustomHeaders{})

const (
	optionsCustomerHeadersKey = "custom-headers"

	annotationRequestHeaders  = "lingress.echocat.org/headers-request"
	annotationResponseHeaders = "lingress.echocat.org/headers-response"
)

func OptionsCustomHeadersOf(options Options) *OptionsCustomHeaders {
	if v, ok := options[optionsCustomerHeadersKey].(*OptionsCustomHeaders); ok {
		return v
	}
	return &OptionsCustomHeaders{}
}

type OptionsCustomHeaders struct {
	RequestHeaders  Headers `json:"requestHeaders,omitempty"`
	ResponseHeaders Headers `json:"responseHeaders,omitempty"`
}

func (instance OptionsCustomHeaders) Name() string {
	return optionsCustomerHeadersKey
}

func (instance OptionsCustomHeaders) IsRelevant() bool {
	return len(instance.RequestHeaders) > 0 ||
		len(instance.ResponseHeaders) > 0
}

func (instance *OptionsCustomHeaders) Set(annotations Annotations) (err error) {
	if instance.RequestHeaders, err = evaluateOptionHeaders(annotations, annotationRequestHeaders); err != nil {
		return
	}
	if instance.ResponseHeaders, err = evaluateOptionHeaders(annotations, annotationResponseHeaders); err != nil {
		return
	}
	return
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
