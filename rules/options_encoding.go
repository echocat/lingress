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

func (instance OptionsEncoding) Name() string {
	return optionsEncodingKey
}

func (instance OptionsEncoding) IsRelevant() bool {
	return len(instance.TransportEncoding) > 0
}

func (instance *OptionsEncoding) Set(annotations Annotations) (err error) {
	if instance.TransportEncoding, err = evaluateOptionTransportEncoding(annotations); err != nil {
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
