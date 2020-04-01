package rules

var _ = RegisterDefaultOptionsPart(&OptionsCors{})

const (
	optionsCorsKey = "cors"

	annotationCors            = "lingress.echocat.org/cors"
	annotationNginxEnableCors = "nginx.ingress.kubernetes.io/enable-cors"
)

func OptionsCorsOf(options Options) *OptionsCors {
	if v, ok := options[optionsCorsKey].(*OptionsCors); ok {
		return v
	}
	return &OptionsCors{}
}

type OptionsCors struct {
	Cors OptionalBool `json:"cors,omitempty"`
}

func (instance OptionsCors) Name() string {
	return optionsCorsKey
}

func (instance OptionsCors) IsRelevant() bool {
	return instance.Cors > 0
}

func (instance *OptionsCors) Set(annotations Annotations) (err error) {
	if instance.Cors, err = evaluateOptionEnableCors(annotations); err != nil {
		return
	}
	return
}

func evaluateOptionEnableCors(annotations map[string]string) (OptionalBool, error) {
	if v, ok := annotations[annotationCors]; ok {
		return AnnotationIsTrue(annotationCors, v)
	}
	if v, ok := annotations[annotationNginxEnableCors]; ok {
		return AnnotationIsTrue(annotationNginxEnableCors, v)
	}
	return NotDefined, nil
}
