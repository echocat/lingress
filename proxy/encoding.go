package proxy

import (
	"github.com/echocat/lingress/context"
	"github.com/echocat/lingress/rules"
	"github.com/echocat/lingress/support"
	"net/textproto"
	"strings"
)

func init() {
	DefaultInterceptors.Add(NewEncoding())
}

type Encoding struct{}

func NewEncoding() *Encoding {
	return &Encoding{}
}

func (instance *Encoding) Name() string {
	return "encoding"
}

func (instance *Encoding) HandlesStages() []context.Stage {
	return []context.Stage{context.StagePrepareUpstreamRequest}
}

func (instance *Encoding) Handle(ctx *context.Context) (proceed bool, err error) {
	switch ctx.Stage {
	case context.StagePrepareUpstreamRequest:
		return instance.handleRequest(ctx)
	}
	return true, nil
}

func (instance *Encoding) RegisterFlag(support.FlagEnabled, string) error {
	return nil
}

func (instance *Encoding) handleRequest(ctx *context.Context) (proceed bool, err error) {
	req := ctx.Upstream.Request
	if req == nil {
		return true, nil
	}

	if r := ctx.Rule; r != nil {
		if v := rules.OptionsEncodingOf(r.Options()).TransportEncoding; len(v) > 0 {
			req.TransferEncoding = v
			req.Header[textproto.CanonicalMIMEHeaderKey("Transfer-Encoding")] = v
			for _, p := range v {
				if strings.ToLower(p) == "chunked" {
					req.ContentLength = -1
					req.Header.Del("Content-Length")
				}
			}
		}
	}

	return true, nil
}
