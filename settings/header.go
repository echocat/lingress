package settings

import (
	"github.com/echocat/lingress/support"
	"github.com/echocat/lingress/value"
)

func NewHeader() (Header, error) {
	return Header{
		Request:  value.Headers{},
		Response: value.Headers{},
	}, nil
}

type Header struct {
	Request  value.Headers `json:"request,omitempty" yaml:"request,omitempty" `
	Response value.Headers `json:"response,omitempty" yaml:"response,omitempty"`
}

func (this *Header) RegisterFlags(fe support.FlagEnabled, _ string) {
	fe.Flag("header.request", "Could be defined multiple times and will set, add(+) or remove(-) headers going to upstream. If ! is prefixed it will override these headers regardless what the ingress config or client says.").
		PlaceHolder("[!][-|+]<name>[:<value>]").
		SetValue(&this.Request)
	fe.Flag("header.response", "Could be defined multiple times and will set, add(+) or remove(-) headers going to client. If ! is prefixed it will override these headers regardless what the ingress config or upstream says.").
		PlaceHolder("[!][-|+]<name>[:<value>]").
		SetValue(&this.Response)
}
