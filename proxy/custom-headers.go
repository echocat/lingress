package proxy

import (
	"github.com/echocat/lingress/context"
	"github.com/echocat/lingress/rules"
	"github.com/echocat/lingress/value"
)

func init() {
	DefaultInterceptors.Add(&CustomHeaders{})
}

type CustomHeaders struct{}

func (this *CustomHeaders) Name() string {
	return "custom-headers"
}

func (this *CustomHeaders) HandlesStages() []context.Stage {
	return []context.Stage{context.StagePrepareUpstreamRequest, context.StagePrepareClientResponse}
}

func (this *CustomHeaders) Handle(ctx *context.Context) (proceed bool, err error) {
	switch ctx.Stage {
	case context.StagePrepareUpstreamRequest:
		return this.handleRequest(ctx)
	case context.StagePrepareClientResponse:
		return this.handleResponse(ctx)
	}
	return true, nil
}

func (this *CustomHeaders) handleRequest(ctx *context.Context) (proceed bool, err error) {
	req := ctx.Upstream.Request
	if req == nil {
		return true, nil
	}

	apply := func(rs value.Headers) {
		for _, h := range rs {
			if h.Del {
				req.Header.Del(h.Key)
			} else if h.Add {
				req.Header.Add(h.Key, h.Value)
			} else if _, contains := req.Header[h.Key]; !contains || h.Forced {
				req.Header.Set(h.Key, h.Value)
			}
		}
	}

	if r := ctx.Rule; r != nil {
		apply(rules.OptionsCustomHeadersOf(r).RequestHeaders)
	}

	apply(ctx.Settings.Response.Headers)

	return true, nil
}

func (this *CustomHeaders) handleResponse(ctx *context.Context) (proceed bool, err error) {
	resp := ctx.Client.Response
	if resp == nil {
		return true, nil
	}

	ih := resp.Header()
	apply := func(rs value.Headers) {
		for _, h := range rs {
			if h.Del {
				ih.Del(h.Key)
			} else if h.Add {
				ih.Add(h.Key, h.Value)
			} else if _, contains := ih[h.Key]; !contains || h.Forced {
				ih.Set(h.Key, h.Value)
			}
		}
	}

	if r := ctx.Rule; r != nil {
		apply(rules.OptionsCustomHeadersOf(r).ResponseHeaders)
	}

	apply(ctx.Settings.Response.Headers)

	return true, nil
}
