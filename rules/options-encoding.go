package rules

import "strings"

var _ = RegisterDefaultOptionsPart(&OptionsEncoding{})

const (
	optionsEncodingKey = "encoding"

	annotationTransportEncoding = "lingress.echocat.org/transport-encoding"
)

func OptionsEncodingOf(rule Rule) *OptionsEncoding {
	if rule == nil {
		return &OptionsEncoding{}
	}
	if v, ok := rule.Options()[optionsEncodingKey].(*OptionsEncoding); ok {
		return v
	}
	return &OptionsEncoding{}
}

type OptionsEncoding struct {
	TransportEncoding []string `json:"transportEncoding,omitempty"`
}

func (this OptionsEncoding) Name() string {
	return optionsEncodingKey
}

func (this OptionsEncoding) IsRelevant() bool {
	return len(this.TransportEncoding) > 0
}

func (this *OptionsEncoding) Set(annotations Annotations) (err error) {
	if this.TransportEncoding, err = evaluateOptionTransportEncoding(annotations); err != nil {
		return
	}
	return
}

func evaluateOptionTransportEncoding(annotations map[string]string) (result []string, err error) {
	if v, ok := annotations[annotationTransportEncoding]; ok {
		return strings.Split(v, ","), nil
	}
	return
}
