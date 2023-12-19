package proxy

import (
	"context"
	"crypto/tls"
	"fmt"
	lctx "github.com/echocat/lingress/context"
	"github.com/echocat/lingress/rules"
	"github.com/echocat/lingress/server"
	"github.com/echocat/lingress/settings"
	"github.com/echocat/lingress/support"
	ltls "github.com/echocat/lingress/tls"
	"github.com/echocat/lingress/value"
	"github.com/echocat/slf4g"
	"io"
	"net"
	"net/http"
	"net/url"
	"runtime/debug"
	"strings"
	"sync"
	"time"
)

type Proxy struct {
	settings *settings.Settings

	Dialer          net.Dialer
	Transport       http.Transport
	RulesRepository rules.Repository
	Logger          log.Logger

	ResultHandler    lctx.ResultHandler
	AccessLogger     AccessLogger
	Interceptors     Interceptors
	MetricsCollector lctx.MetricsCollector

	bufferPool sync.Pool
}

type AccessLogger func(*lctx.Context)

func New(s *settings.Settings, rules rules.Repository, logger log.Logger) (*Proxy, error) {
	result := &Proxy{
		settings: s,
		Dialer:   net.Dialer{},
		Transport: http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: ltls.Pool,
			},
		},
		RulesRepository: rules,
		Interceptors:    DefaultInterceptors.Clone(),
		Logger:          logger,
	}
	result.Transport.DialContext = result.Dialer.DialContext
	result.bufferPool.New = result.createBuffer
	return result, nil
}

func (this *Proxy) Init(support.Channel) error {
	if err := this.settings.Upstream.ApplyToNetDialer(&this.Dialer); err != nil {
		return err
	}
	if err := this.settings.Upstream.ApplyToHttpTransport(&this.Transport); err != nil {
		return err
	}
	return nil
}

func (this *Proxy) ServeHTTP(connector server.Connector, resp http.ResponseWriter, req *http.Request) {
	ctx, _, err := lctx.AcquireContext(this.settings, connector.GetId(), this.settings.Server.BehindReverseProxy.Get(), resp, req, this.Logger)
	if err != nil {
		this.Logger.
			WithError(err).
			Error("Cannot acquire context.")
		return
	}
	al := this.AccessLogger
	defer func() {
		if al == nil {
			if err := ctx.Release(); err != nil {
				this.Logger.
					WithError(err).
					Error("Problem while releasing context.")
			}
		}
	}()
	ctx.Client.Started = time.Now()
	defer func() {
		if r := recover(); r != nil {
			var err error
			if v, ok := r.(error); ok {
				err = v
			} else {
				err = fmt.Errorf("panic: %v", r)
			}

			stack := string(debug.Stack())
			this.Logger.With("stack", stack).
				WithError(err).
				Error("Unhandled error inside of the finalization stack.")
		}
	}()
	defer func() {
		if rh := this.ResultHandler; rh != nil {
			rh(ctx)
		}
		ctx.Client.Duration = time.Now().Sub(ctx.Client.Started).Truncate(time.Millisecond)
		if r := ctx.Rule; r != nil {
			r.Statistics().MarkUsed(ctx.Client.Duration)
		}
		if mc := this.MetricsCollector; mc != nil {
			mc.CollectContext(ctx)
		}
		_, _ = this.switchStageAndCallInterceptors(lctx.StageDone, ctx)
		if al != nil {
			al(ctx)
		}
	}()

	ctx.Stage = lctx.StageEvaluateClientRequest
	var host value.Fqdn
	if err := host.Set(ctx.Client.Host()); err != nil {
		this.markDone(lctx.ResultFailedWithIllegalHost, ctx, err)
		return
	}

	query := rules.Query{
		Host: host,
		Path: ctx.Client.Request.RequestURI,
	}
	if u := ctx.Client.Request.URL; u != nil {
		query.Path = u.Path
	}

	rs, err := this.RulesRepository.FindBy(query)
	if err != nil {
		this.markDone(lctx.ResultFailedWithUnexpectedError, ctx, err)
		return
	} else if rs == nil || rs.Len() == 0 {
		this.markDone(lctx.ResultFailedWithRuleNotFound, ctx, err)
		return
	}

	r, err := this.selectRule(rs)
	if err != nil {
		this.markDone(lctx.ResultFailedWithUnexpectedError, ctx, err)
		return
	}

	ctx.Rule = r

	if proceed, err := this.callInterceptors(ctx); err != nil {
		this.markDone(lctx.ResultFailedWithUnexpectedError, ctx, err)
		return
	} else if !proceed {
		return
	}

	ctx.Upstream.Address = r.Backend()

	if proceed, err := this.createBackendRequestFor(ctx, r); err != nil {
		this.markDone(lctx.ResultFailedWithUnexpectedError, ctx, err)
		return
	} else if !proceed {
		return
	}
	if cancel := ctx.Upstream.Cancel; cancel != nil {
		defer cancel()
	}

	if err := this.execute(ctx); isDialError(err) {
		this.markDone(lctx.ResultFailedWithUpstreamUnavailable, ctx, err)
		return
	} else if isClientGoneError(err) {
		this.markDone(lctx.ResultFailedWithClientGone, ctx, err)
		return
	} else if err != nil {
		if ctx.Client.Status > 0 {
			// Returning an error to the client is not really possible here because we have to assume that we already
			// wrote content to the frontend
			ctx.Error = err
		} else {
			this.markDone(lctx.ResultFailedWithUnexpectedError, ctx, err)
		}
		return
	}

	this.markDone(lctx.ResultSuccess, ctx)
}

func (this *Proxy) switchStageAndCallInterceptors(stage lctx.Stage, ctx *lctx.Context) (bool, error) {
	ctx.Stage = stage
	return this.callInterceptors(ctx)
}

func (this *Proxy) callInterceptors(ctx *lctx.Context) (bool, error) {
	if i := this.Interceptors; i == nil {
		return true, nil
	} else {
		return i.Handle(ctx)
	}
}

func (this *Proxy) markDone(result lctx.Result, ctx *lctx.Context, err ...error) {
	ctx.Done(result, err...)
}

func (this *Proxy) selectRule(in rules.Rules) (out rules.Rule, err error) {
	return in.Any(), nil
}

func (this *Proxy) createBackendRequestFor(ctx *lctx.Context, r rules.Rule) (proceed bool, err error) {
	ctx.Stage = lctx.StagePrepareUpstreamRequest
	fReq := ctx.Client.Request
	u, err := url.Parse(fReq.URL.String())
	if err != nil {
		return false, err
	}

	if v := this.settings.Upstream.OverrideHost; v != "" {
		u.Host = v
	} else {
		u.Host = r.Backend().String()
	}
	if v := this.settings.Upstream.OverrideScheme; v != "" {
		u.Scheme = v
	} else {
		u.Scheme = "http"
	}

	bCtx := fReq.Context()
	var cancel context.CancelFunc
	bCtx, ctx.Upstream.Cancel = context.WithCancel(bCtx)
	ctx.Upstream.Cancel = cancel
	go func(cancel context.CancelFunc) {
		if cancel != nil {
			select {
			case <-ctx.Client.Request.Context().Done():
				cancel()
			case <-bCtx.Done():
			}
		}
	}(cancel)

	bReq := (&http.Request{
		Host:             u.Host,
		Method:           fReq.Method,
		URL:              u,
		Proto:            "HTTP/1.1",
		ProtoMajor:       1,
		ProtoMinor:       1,
		Header:           cloneHeader(fReq.Header),
		Trailer:          cloneHeader(fReq.Trailer),
		Close:            false,
		Body:             fReq.Body,
		ContentLength:    fReq.ContentLength,
		TransferEncoding: fReq.TransferEncoding,
	}).WithContext(bCtx)

	if fReq.ContentLength == 0 {
		bReq.Body = nil // Issue 16036: nil Body for http.Transport retries
	}
	if fReq.ContentLength < 0 && len(bReq.TransferEncoding) <= 0 {
		bReq.TransferEncoding = []string{"chunked"}
	}
	reqUpType := retrieveUpgradeType(bReq.Header)
	removeConnectionHeaders(bReq.Header)
	removeHopReqHeaders(bReq.Header)
	setConnectionUpgrades(bReq.Header, reqUpType)

	ctx.Upstream.Request = bReq

	return this.callInterceptors(ctx)
}

func (this *Proxy) execute(ctx *lctx.Context) error {
	if mc := this.MetricsCollector; mc != nil {
		finalize := mc.CollectUpstreamStarted()
		defer finalize()
	}

	ctx.Client.Status = 0
	ctx.Upstream.Status = 0
	ctx.Upstream.Started = time.Now()
	if proceed, err := this.switchStageAndCallInterceptors(lctx.StageSendRequestToUpstream, ctx); err != nil {
		return err
	} else if !proceed {
		return nil
	}
	bResp, err := this.Transport.RoundTrip(ctx.Upstream.Request)
	ctx.Upstream.Duration = time.Now().Sub(ctx.Upstream.Started)
	if err != nil {
		return err
	}
	ctx.Upstream.Response = bResp
	ctx.Upstream.Status = ctx.Upstream.Response.StatusCode
	ctx.Stage = lctx.StagePrepareClientResponse

	// Deal with 101 Switching Protocols responses: (WebSocket, h2c, etc)
	if bResp.StatusCode == http.StatusSwitchingProtocols {
		if proceed, err := this.switchStageAndCallInterceptors(lctx.StageSendResponseToClient, ctx); err != nil {
			_ = bResp.Body.Close()
			return err
		} else if !proceed {
			_ = bResp.Body.Close()
			return nil
		}
		return this.handleUpgradeResponse(ctx.Client.Response, ctx.Upstream.Request, ctx.Upstream.Response)
	}

	copyHeader(ctx.Client.Response.Header(), ctx.Upstream.Response.Header)
	removeConnectionHeaders(ctx.Upstream.Response.Header)
	removeHopRespHeaders(ctx.Upstream.Response.Header)

	// The "Trailer" header isn't included in the Transport's response,
	// at least for *http.Transport. Build it up from Trailer.
	announcedTrailers := len(ctx.Upstream.Response.Trailer)
	if announcedTrailers > 0 {
		trailerKeys := make([]string, 0, len(ctx.Upstream.Response.Trailer))
		for k := range ctx.Upstream.Response.Trailer {
			trailerKeys = append(trailerKeys, k)
		}
		ctx.Client.Response.Header().Add("Trailer", strings.Join(trailerKeys, ", "))
	}

	if proceed, err := this.callInterceptors(ctx); err != nil {
		_ = bResp.Body.Close()
		return err
	} else if !proceed {
		_ = bResp.Body.Close()
		return nil
	}

	ctx.Client.Response.WriteHeader(ctx.Upstream.Status)
	ctx.Client.Status = ctx.Upstream.Status
	if proceed, err := this.switchStageAndCallInterceptors(lctx.StageSendResponseToClient, ctx); err != nil {
		_ = bResp.Body.Close()
		return err
	} else if !proceed {
		_ = bResp.Body.Close()
		return nil
	}

	_, err = this.copyBuffered(ctx.Client.Response, ctx.Upstream.Response.Body)
	if err != nil {
		//noinspection GoUnhandledErrorResult
		defer ctx.Upstream.Response.Body.Close()
		return fmt.Errorf("%w: %v", http.ErrAbortHandler, err)
	}
	//noinspection GoUnhandledErrorResult
	ctx.Upstream.Response.Body.Close() // close now, instead of defer, to populate res.Trailer

	if len(ctx.Upstream.Response.Trailer) > 0 {
		// Force chunking if we saw a response trailer.
		// This prevents net/http from calculating the length for short
		// bodies and adding a Content-Length.
		ctx.Client.Response.(http.Flusher).Flush()
	}

	if len(ctx.Upstream.Response.Trailer) == announcedTrailers {
		copyHeader(ctx.Client.Response.Header(), ctx.Upstream.Response.Trailer)
		return nil
	}

	for k, vv := range ctx.Upstream.Response.Trailer {
		k = http.TrailerPrefix + k
		for _, v := range vv {
			ctx.Client.Response.Header().Add(k, v)
		}
	}

	return nil
}

func (this *Proxy) handleUpgradeResponse(rw http.ResponseWriter, req *http.Request, res *http.Response) error {
	reqUpType := retrieveUpgradeType(req.Header)
	resUpType := retrieveUpgradeType(res.Header)
	if reqUpType != resUpType {
		return fmt.Errorf("backend tried to switch protocol %q when %q was requested", resUpType, reqUpType)
	}

	copyHeader(res.Header, rw.Header())

	backConn, ok := res.Body.(io.ReadWriteCloser)
	if !ok {
		return fmt.Errorf("internal error: 101 switching protocols response with non-writable body")
	}
	//noinspection GoUnhandledErrorResult
	defer backConn.Close()
	conn, brw, err := rw.(http.Hijacker).Hijack()
	if err != nil {
		return fmt.Errorf("hijack failed on protocol switch: %v", err)
	}

	//noinspection GoUnhandledErrorResult
	defer conn.Close()
	res.Body = nil // so res.Write only writes the headers; we have res.Body in backConn above
	if err := res.Write(brw); err != nil {
		return fmt.Errorf("response write: %v", err)
	}
	if err := brw.Flush(); err != nil {
		return fmt.Errorf("response flush: %v", err)
	}
	errc := make(chan error, 1)
	spc := switchProtocolCopier{user: conn, backend: backConn}
	go spc.copyToBackend(errc)
	go spc.copyFromBackend(errc)
	<-errc
	return nil
}

// copyBuffered returns any write errors or non-EOF read errors, and the amount
// of bytes written.
func (this *Proxy) copyBuffered(dst io.Writer, src io.Reader) (int64, error) {
	buf := this.acquireBuffer()
	defer this.releaseBuffer(buf)
	var written int64
	for {
		nr, rErr := src.Read(*buf)
		if rErr != nil && rErr != io.EOF && rErr != context.Canceled {
			this.Logger.
				WithError(rErr).
				Warn("Read error during body copy.")
		}
		if nr > 0 {
			nw, wErr := dst.Write((*buf)[:nr])
			if nw > 0 {
				written += int64(nw)
			}
			if wErr != nil {
				return written, wErr
			}
			if nr != nw {
				return written, io.ErrShortWrite
			}
		}
		if rErr != nil {
			if rErr == io.EOF {
				rErr = nil
			}
			return written, rErr
		}
	}
}

func (this *Proxy) createBuffer() interface{} {
	result := make([]byte, 32*1024)
	return &result
}

func (this *Proxy) acquireBuffer() *[]byte {
	return this.bufferPool.Get().(*[]byte)
}

func (this *Proxy) releaseBuffer(buf *[]byte) {
	this.bufferPool.Put(buf)
}
