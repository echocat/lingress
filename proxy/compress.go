package proxy

import (
	"github.com/CAFxX/httpcompression"
	"github.com/echocat/lingress/context"
	"github.com/echocat/lingress/rules"
	"github.com/echocat/lingress/support"
	"net/http"
)

const (
	compressFinalizerCtxKey = "compressFinalizer"
)

func init() {
	DefaultInterceptors.Add(NewCompressInterceptor())
}

type CompressInterceptor struct {
	responseWriterFactory *httpcompression.ResponseWriterFactoryFactory
}

func NewCompressInterceptor() *CompressInterceptor {
	f, err := httpcompression.NewDefaultResponseWriterFactory(
		httpcompression.ErrorHandler(func(w http.ResponseWriter, r *http.Request, err error) {
			context.OfRequest(r).
				Log().
				WithError(err).
				Error("Problems while handling compression of request.")
		}),
	)
	support.Must(err)

	return &CompressInterceptor{
		responseWriterFactory: f,
	}
}

func (this *CompressInterceptor) Name() string {
	return "compress"
}

func (this *CompressInterceptor) HandlesStages() []context.Stage {
	return []context.Stage{context.StagePrepareClientResponse, context.StageDone}
}

func (this *CompressInterceptor) Handle(ctx *context.Context) (proceed bool, err error) {
	switch ctx.Stage {
	case context.StagePrepareClientResponse:
		return this.prepareHeaders(ctx)
	case context.StageDone:
		return this.finalize(ctx)
	}
	return true, nil
}

func (this *CompressInterceptor) prepareHeaders(ctx *context.Context) (proceed bool, err error) {
	optionsCompress := rules.OptionsCompressOf(ctx.Rule)
	if !ctx.Settings.Response.Compress.Evaluate(optionsCompress.Enabled).GetOr(false) {
		return true, nil
	}
	wrappedRw, finalizer, err := this.responseWriterFactory.Create(ctx.Client.Response, ctx.Client.Request)
	if err != nil {
		return false, err
	}
	ctx.Properties[compressFinalizerCtxKey] = finalizer
	if wrappedRw != ctx.Client.Response {
		ctx.Client.Response = wrappedRw
	}

	return true, nil
}

func (this *CompressInterceptor) finalize(ctx *context.Context) (proceed bool, err error) {
	finalizer := ctx.Properties[compressFinalizerCtxKey].(httpcompression.Finalizer)
	if finalizer == nil {
		return true, nil
	}
	defer func() {
		if fErr := finalizer(); fErr != nil && err == nil {
			err = fErr
		}
	}()

	return true, nil
}
