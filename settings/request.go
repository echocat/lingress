package settings

import (
	"github.com/echocat/lingress/support"
	"github.com/echocat/lingress/value"
)

func NewRequest() (Request, error) {
	return Request{
		Headers: value.Headers{},
	}, nil
}

type Request struct {
	Headers value.Headers `json:"headers,omitempty" yaml:"headers,omitempty" `
}

func (this *Request) RegisterFlags(fe support.FlagEnabled, _ string) {
	fe.Flag("request.headers", "Could be defined multiple times and will set, add(+) or remove(-) headers going to upstream. If ! is prefixed it will override these headers regardless what the ingress config or client says.").
		PlaceHolder("[!][-|+]<name>[:<value>]").
		SetValue(&this.Headers)
}
