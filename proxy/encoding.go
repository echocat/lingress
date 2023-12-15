package proxy

import (
	"github.com/echocat/lingress/context"
	"github.com/echocat/lingress/rules"
	"net/textproto"
	"strings"
)

func init() {
	DefaultInterceptors.Add(&Encoding{})
}

type Encoding struct{}

func (this *Encoding) Name() string {
	return "encoding"
}

func (this *Encoding) HandlesStages() []context.Stage {
	return []context.Stage{context.StagePrepareUpstreamRequest}
}

func (this *Encoding) Handle(ctx *context.Context) (proceed bool, err error) {
	switch ctx.Stage {
	case context.StagePrepareUpstreamRequest:
		return this.handleRequest(ctx)
	}
	return true, nil
}

func (this *Encoding) handleRequest(ctx *context.Context) (proceed bool, err error) {
	req := ctx.Upstream.Request
	if req == nil {
		return true, nil
	}

	if r := ctx.Rule; r != nil {
		if v := rules.OptionsEncodingOf(r).TransportEncoding; len(v) > 0 {
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
