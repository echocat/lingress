package settings

import (
	"github.com/echocat/lingress/support"
	"github.com/echocat/lingress/value"
)

func NewResponse() (Response, error) {
	return Response{
		Headers:  value.Headers{},
		Compress: value.NewForcibleBool(value.True(), false),
	}, nil
}

type Response struct {
	Headers  value.Headers      `json:"headers,omitempty" yaml:"headers,omitempty"`
	Compress value.ForcibleBool `json:"compress,omitempty" yaml:"compress,omitempty"`
}

func (this *Response) RegisterFlags(fe support.FlagEnabled, _ string) {
	fe.Flag("response.headers", "Could be defined multiple times and will set, add(+) or remove(-) responses going to client. If ! is prefixed it will override these responses regardless what the ingress config or client says.").
		PlaceHolder("[!][-|+]<name>[:<value>]").
		SetValue(&this.Headers)
	fe.Flag("response.compress", "If set to true the response is compress before send to the client. If ! is prefixed it will override whatever the update and/or the ingress setting will define.").
		PlaceHolder("[!]<true|false>").
		SetValue(&this.Compress)
}
