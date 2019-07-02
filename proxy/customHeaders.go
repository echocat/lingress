package proxy

import (
	"github.com/echocat/lingress/context"
	"github.com/echocat/lingress/rules"
	"github.com/echocat/lingress/support"
)

func init() {
	DefaultInterceptors.Add(NewCustomHeaders())
}

type CustomHeaders struct {
	Request  rules.Headers
	Response rules.Headers
}

func NewCustomHeaders() *CustomHeaders {
	return &CustomHeaders{
		Request:  rules.Headers{},
		Response: rules.Headers{},
	}
}

func (instance *CustomHeaders) Name() string {
	return "custom-headers"
}

func (instance *CustomHeaders) HandlesStages() []context.Stage {
	return []context.Stage{context.StagePrepareUpstreamRequest, context.StagePrepareClientResponse}
}

func (instance *CustomHeaders) Handle(ctx *context.Context) (proceed bool, err error) {
	switch ctx.Stage {
	case context.StagePrepareUpstreamRequest:
		return instance.handleRequest(ctx)
	case context.StagePrepareClientResponse:
		return instance.handleResponse(ctx)
	}
	return true, nil
}

func (instance *CustomHeaders) RegisterFlag(fe support.FlagEnabled, appPrefix string) error {
	fe.Flag("header.request", "Could be defined multiple times and will set, add(+) or remove(-) headers going to upstream. If ! is prefixed it will override these headers regardless what the ingress config or client says.").
		PlaceHolder("[!][-|+]<name>[:<value>]").
		SetValue(&instance.Request)
	fe.Flag("header.response", "Could be defined multiple times and will set, add(+) or remove(-) headers going to client. If ! is prefixed it will override these headers regardless what the ingress config or upstream says.").
		PlaceHolder("[!][-|+]<name>[:<value>]").
		SetValue(&instance.Response)
	return nil
}

func (instance *CustomHeaders) handleRequest(ctx *context.Context) (proceed bool, err error) {
	req := ctx.Upstream.Request
	if req == nil {
		return true, nil
	}

	apply := func(rs rules.Headers) {
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
		apply(r.Options().RequestHeaders)
	}

	apply(instance.Request)

	return true, nil
}

func (instance *CustomHeaders) handleResponse(ctx *context.Context) (proceed bool, err error) {
	resp := ctx.Client.Response
	if resp == nil {
		return true, nil
	}

	ih := resp.Header()
	apply := func(rs rules.Headers) {
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
		apply(r.Options().ResponseHeaders)
	}

	apply(instance.Response)

	return true, nil
}
