package settings

import (
	"github.com/echocat/lingress/support"
)

var (
	defaultIngressClasses = []string{"lingress", ""}
)

func NewIngress() (Ingress, error) {
	return Ingress{
		Classes: []string{},
	}, nil
}

type Ingress struct {
	Classes []string `yaml:"classes,omitempty" json:"classes,omitempty"`
}

func (this *Ingress) RegisterFlags(fe support.FlagEnabled, appPrefix string) {
	fe.Flag("ingress.class", "Ingress classes which this application should respect.").
		PlaceHolder("<class[,...]>").
		Envar(support.FlagEnvName(appPrefix, "INGRESS_CLASS")).
		StringsVar(&this.Classes)
}

func (this *Ingress) GetClasses() []string {
	if v := this.Classes; len(v) != 0 {
		return v
	}
	return defaultIngressClasses
}
