package rules

import (
	"github.com/echocat/lingress/value"
)

var _ = RegisterDefaultOptionsPart(&OptionsCompress{})

const (
	optionsCompressKey = "compress"

	annotationCompressEnabledAlternative = "lingress.echocat.org/compress"
	annotationCompressEnabled            = "lingress.echocat.org/compress.enabled"
)

func OptionsCompressOf(rule Rule) *OptionsCompress {
	if rule == nil {
		return &OptionsCompress{}
	}
	if v, ok := rule.Options()[optionsCompressKey].(*OptionsCompress); ok {
		return v
	}
	return &OptionsCompress{}
}

type OptionsCompress struct {
	Enabled value.Bool `json:"enabled,omitempty"`
}

func (this OptionsCompress) Name() string {
	return optionsCompressKey
}

func (this OptionsCompress) IsRelevant() bool {
	return this.Enabled.IsPresent()
}

func (this *OptionsCompress) Set(annotations Annotations) (err error) {
	if this.Enabled, err = evaluateOptionCompressEnable(annotations); err != nil {
		return
	}
	return
}

func evaluateOptionCompressEnable(annotations map[string]string) (value.Bool, error) {
	if v, ok := annotations[annotationCompressEnabled]; ok {
		return AnnotationIsBool(annotationCompressEnabled, v)
	}
	if v, ok := annotations[annotationCompressEnabledAlternative]; ok {
		return AnnotationIsBool(annotationCompressEnabledAlternative, v)
	}
	return value.UndefinedBool(), nil
}
