package proxy

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	lctx "github.com/echocat/lingress/context"
	"github.com/echocat/lingress/rules"
	"github.com/echocat/lingress/support"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Proxy struct {
	Dialer          *net.Dialer
	Transport       *http.Transport
	RulesRepository rules.Repository

	OverrideHost   string
	OverrideScheme string

	BehindOtherReverseProxy bool
	ResultHandler           lctx.ResultHandler
	AccessLogger            AccessLogger
	Interceptors            Interceptors
}

type AccessLogger func(*lctx.Context)

func New(rules rules.Repository) (*Proxy, error) {
	dialer := &net.Dialer{
		Timeout:   10 * time.Second,
		KeepAlive: 30 * time.Second,
	}
	transport := &http.Transport{
		MaxIdleConnsPerHost:    20,
		MaxConnsPerHost:        250,
		IdleConnTimeout:        1 * time.Minute,
		MaxResponseHeaderBytes: 10 << 20,
		DialContext:            dialer.DialContext,
		TLSClientConfig: &tls.Config{
			RootCAs: support.Pool,
		},
	}
	return &Proxy{
		Dialer:          dialer,
		Transport:       transport,
		RulesRepository: rules,
		Interceptors:    DefaultInterceptors.Clone(),
	}, nil
}

func (instance *Proxy) RegisterFlag(fe support.FlagEnabled, appPrefix string) error {
	fe.Flag("behindOtherReverseProxy", "If true also X-Forwarded headers are evaluated before send to upstream.").
		PlaceHolder(fmt.Sprint(instance.BehindOtherReverseProxy)).
		Envar(support.FlagEnvName(appPrefix, "BEHIND_OTHER_REVERSE_PROXY")).
		BoolVar(&instance.BehindOtherReverseProxy)
	fe.Flag("upstream.maxIdleConnectionsPerHost", "Controls the maximum idle (keep-alive) connections to keep per-host.").
		PlaceHolder(fmt.Sprint(instance.Transport.MaxIdleConnsPerHost)).
		Envar(support.FlagEnvName(appPrefix, "UPSTREAM_MAX_IDLE_CONNECTIONS_PER_HOST")).
		IntVar(&instance.Transport.MaxIdleConnsPerHost)
	fe.Flag("upstream.maxConnectionsPerHost", "Limits the total number of connections per host, including connections in the dialing, active, and idle states. On limit violation, dials will block.").
		PlaceHolder(fmt.Sprint(instance.Transport.MaxConnsPerHost)).
		Envar(support.FlagEnvName(appPrefix, "UPSTREAM_MAX_CONNECTIONS_PER_HOST")).
		IntVar(&instance.Transport.MaxConnsPerHost)
	fe.Flag("upstream.idleConnectionTimeout", "Maximum amount of time an idle (keep-alive) connection will remain idle before closing itself. Zero means no limit.").
		PlaceHolder(fmt.Sprint(instance.Transport.IdleConnTimeout)).
		Envar(support.FlagEnvName(appPrefix, "UPSTREAM_IDLE_CONNECTION_TIMEOUT")).
		DurationVar(&instance.Transport.IdleConnTimeout)
	fe.Flag("upstream.maxResponseHeaderSize", "Limit on how many response bytes are allowed in the server's response header.").
		PlaceHolder(fmt.Sprint(instance.Transport.MaxResponseHeaderBytes)).
		Envar(support.FlagEnvName(appPrefix, "UPSTREAM_MAX_RESPONSE_HEADER_SIZE")).
		Int64Var(&instance.Transport.MaxResponseHeaderBytes)
	fe.Flag("upstream.dialTimeout", "Maximum amount of time a dial will wait for a connect to complete. If Deadline is also set, it may fail earlier.").
		PlaceHolder(fmt.Sprint(instance.Dialer.Timeout)).
		Envar(support.FlagEnvName(appPrefix, "UPSTREAM_DIAL_TIMEOUT")).
		DurationVar(&instance.Dialer.Timeout)
	fe.Flag("upstream.keepAlive", "Keep-alive period for an active network connection. If zero, keep-alives are enabled if supported by the protocol and operating system. Network protocols or operating systems that do not support keep-alives ignore this field. If negative, keep-alives are disabled.").
		PlaceHolder(fmt.Sprint(instance.Dialer.KeepAlive)).
		Envar(support.FlagEnvName(appPrefix, "UPSTREAM_KEEP_ALIVE")).
		DurationVar(&instance.Dialer.KeepAlive)
	fe.Flag("upstream.override.host", "Overrides the target host always with this value.").
		PlaceHolder(instance.OverrideHost).
		Envar(support.FlagEnvName(appPrefix, "UPSTREAM_OVERRIDE_HOST")).
		StringVar(&instance.OverrideHost)
	fe.Flag("upstream.override.scheme", "Overrides the target scheme always with this value.").
		PlaceHolder(instance.OverrideScheme).
		Envar(support.FlagEnvName(appPrefix, "UPSTREAM_OVERRIDE_SCHEME")).
		StringVar(&instance.OverrideScheme)

	if i := instance.Interceptors; i != nil {
		if err := i.RegisterFlag(fe, appPrefix); err != nil {
			return err
		}
	}

	return nil
}

func (instance *Proxy) Init(stop support.Channel) error {
	if init, ok := instance.RulesRepository.(support.Initializable); ok {
		if err := init.Init(stop); err != nil {
			return err
		}
	}
	if i := instance.Interceptors; i != nil {
		if err := i.Init(stop); err != nil {
			return err
		}
	}
}

func (instance *Proxy) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	ctx := lctx.AcquireContext(instance.BehindOtherReverseProxy, resp, req)
	defer ctx.Release()
	ctx.Client.Started = time.Now()
	defer func(ctx *lctx.Context) {
		if rh := instance.ResultHandler; rh != nil {
			rh(ctx)
		}
		ctx.Client.Duration = time.Now().Sub(ctx.Client.Started)
		if r := ctx.Rule; r != nil {
			r.Statistics().MarkUsed(ctx.Client.Duration)
		}
		_, _ = instance.switchStageAndCallInterceptors(lctx.StageDone, ctx)
		if al := instance.AccessLogger; al != nil {
			al(ctx)
		}
	}(ctx)

	query := rules.Query{
		Host: ctx.Client.Host(),
		Path: ctx.Client.Request.RequestURI,
	}

	ctx.Stage = lctx.StageEvaluateClientRequest
	rs, err := instance.RulesRepository.FindBy(query)
	if err != nil {
		instance.markDone(lctx.ResultFailedWithUnexpectedError, ctx, err)
		return
	} else if rs == nil || rs.Len() == 0 {
		instance.markDone(lctx.ResultFailedWithRuleNotFound, ctx, err)
		return
	}

	r, err := instance.selectRule(rs)
	if err != nil {
		instance.markDone(lctx.ResultFailedWithUnexpectedError, ctx, err)
		return
	}

	ctx.Rule = r

	if proceed, err := instance.callInterceptors(ctx); err != nil {
		instance.markDone(lctx.ResultFailedWithUnexpectedError, ctx, err)
		return
	} else if !proceed {
		return
	}

	ctx.Upstream.Address = r.Backend()

	if proceed, err := instance.createBackendRequestFor(ctx, r); err != nil {
		instance.markDone(lctx.ResultFailedWithUnexpectedError, ctx, err)
		return
	} else if !proceed {
		return
	}
	if cancel := ctx.Upstream.Cancel; cancel != nil {
		defer cancel()
	}

	if err := instance.execute(ctx); isDialError(err) {
		instance.markDone(lctx.ResultFailedWithUpstreamUnavailable, ctx, err)
		return
	} else if err != nil {
		if ctx.Client.Status > 0 {
			// Returning an error to the client is not really possible here because we have to assume that we already
			// wrote content to the frontend
			ctx.Error = err
		} else {
			instance.markDone(lctx.ResultFailedWithUnexpectedError, ctx, err)
		}
		return
	}

	instance.markDone(lctx.ResultSuccess, ctx)
}

func (instance *Proxy) switchStageAndCallInterceptors(stage lctx.Stage, ctx *lctx.Context) (bool, error) {
	ctx.Stage = stage
	return instance.callInterceptors(ctx)
}

func (instance *Proxy) callInterceptors(ctx *lctx.Context) (bool, error) {
	if i := instance.Interceptors; i == nil {
		return true, nil
	} else {
		return i.Handle(ctx)
	}
}

func (instance *Proxy) markDone(result lctx.Result, ctx *lctx.Context, err ...error) {
	ctx.Done(result, err...)
}

func (instance *Proxy) selectRule(in rules.Rules) (out rules.Rule, err error) {
	return in.Any(), nil
}

func (instance *Proxy) createBackendRequestFor(ctx *lctx.Context, r rules.Rule) (proceed bool, err error) {
	ctx.Stage = lctx.StagePrepareUpstreamRequest
	fReq := ctx.Client.Request
	u, err := url.Parse(fReq.URL.String())
	if err != nil {
		return false, err
	}

	if instance.OverrideHost != "" {
		u.Host = instance.OverrideHost
	} else {
		u.Host = r.Backend().String()
	}
	if instance.OverrideScheme != "" {
		u.Scheme = instance.OverrideScheme
	} else {
		u.Scheme = "http"
	}

	bCtx := fReq.Context()
	//noinspection GoDeprecation
	if cn, ok := ctx.Client.Response.(http.CloseNotifier); ok {
		var cancel context.CancelFunc
		bCtx, ctx.Upstream.Cancel = context.WithCancel(bCtx)
		ctx.Upstream.Cancel = cancel
		notifyChan := cn.CloseNotify()
		go func(cancel context.CancelFunc) {
			if cancel != nil {
				select {
				case <-notifyChan:
					cancel()
				case <-bCtx.Done():
				}
			}
		}(cancel)
	}

	bReq := (&http.Request{
		Host:       u.Host,
		Method:     fReq.Method,
		URL:        u,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     cloneHeader(fReq.Header),
		Close:      false,
	}).WithContext(bCtx)

	if fReq.ContentLength == 0 {
		bReq.Body = nil // Issue 16036: nil Body for http.Transport retries
	}
	reqUpType := retrieveUpgradeType(bReq.Header)
	removeConnectionHeaders(bReq.Header)
	removeHopReqHeaders(bReq.Header)
	setConnectionUpgrades(bReq.Header, reqUpType)

	ctx.Upstream.Request = bReq

	return instance.callInterceptors(ctx)
}

func (instance *Proxy) execute(ctx *lctx.Context) error {
	ctx.Client.Status = 0
	ctx.Upstream.Status = 0
	ctx.Upstream.Started = time.Now()
	if proceed, err := instance.switchStageAndCallInterceptors(lctx.StageSendRequestToUpstream, ctx); err != nil {
		return err
	} else if !proceed {
		return nil
	}
	bResp, err := instance.Transport.RoundTrip(ctx.Upstream.Request)
	ctx.Upstream.Duration = time.Now().Sub(ctx.Upstream.Started)
	if err != nil {
		return err
	}
	ctx.Upstream.Response = bResp
	ctx.Upstream.Status = ctx.Upstream.Response.StatusCode
	ctx.Stage = lctx.StagePrepareClientResponse

	// Deal with 101 Switching Protocols responses: (WebSocket, h2c, etc)
	if bResp.StatusCode == http.StatusSwitchingProtocols {
		if proceed, err := instance.switchStageAndCallInterceptors(lctx.StageSendResponseToClient, ctx); err != nil {
			_ = bResp.Body.Close()
			return err
		} else if !proceed {
			_ = bResp.Body.Close()
			return nil
		}
		return instance.handleUpgradeResponse(ctx.Client.Response, ctx.Upstream.Request, ctx.Upstream.Response)
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

	if proceed, err := instance.callInterceptors(ctx); err != nil {
		_ = bResp.Body.Close()
		return err
	} else if !proceed {
		_ = bResp.Body.Close()
		return nil
	}

	ctx.Client.Response.WriteHeader(ctx.Upstream.Status)
	ctx.Client.Status = ctx.Upstream.Status
	if proceed, err := instance.switchStageAndCallInterceptors(lctx.StageSendResponseToClient, ctx); err != nil {
		_ = bResp.Body.Close()
		return err
	} else if !proceed {
		_ = bResp.Body.Close()
		return nil
	}

	err = instance.copyResponse(ctx.Client.Response, ctx.Upstream.Response.Body)
	if err != nil {
		//noinspection GoUnhandledErrorResult
		defer ctx.Upstream.Response.Body.Close()
		// Since we're streaming the response, if we run into an error all we can do
		// is abort the request. Proxy should use ErrAbortHandler on read error while copying body.
		if !shouldPanicOnCopyError(ctx.Upstream.Request) {
			return errors.New("error; not a panic")
		}
		panic(http.ErrAbortHandler)
	}
	//noinspection GoUnhandledErrorResult
	ctx.Upstream.Response.Body.Close() // close now, instead of defer, to populate res.Trailer

	if len(ctx.Upstream.Response.Trailer) > 0 {
		// Force chunking if we saw a response trailer.
		// This prevents net/http from calculating the length for short
		// bodies and adding a Content-Length.
		if fl, ok := ctx.Client.Response.(http.Flusher); ok {
			fl.Flush()
		}
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

func (instance *Proxy) handleUpgradeResponse(rw http.ResponseWriter, req *http.Request, res *http.Response) error {
	reqUpType := retrieveUpgradeType(req.Header)
	resUpType := retrieveUpgradeType(res.Header)
	if reqUpType != resUpType {
		return fmt.Errorf("backend tried to switch protocol %q when %q was requested", resUpType, reqUpType)
	}

	copyHeader(res.Header, rw.Header())

	hj, ok := rw.(http.Hijacker)
	if !ok {
		return fmt.Errorf("can't switch protocols using non-Hijacker ResponseWriter type %T", rw)
	}
	backConn, ok := res.Body.(io.ReadWriteCloser)
	if !ok {
		return fmt.Errorf("internal error: 101 switching protocols response with non-writable body")
	}
	//noinspection GoUnhandledErrorResult
	defer backConn.Close()
	conn, brw, err := hj.Hijack()
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

func (instance *Proxy) copyResponse(dst io.Writer, src io.Reader) error {
	var buf []byte
	_, err := instance.copyBuffer(dst, src, buf)
	return err
}

// copyBuffer returns any write errors or non-EOF read errors, and the amount
// of bytes written.
func (instance *Proxy) copyBuffer(dst io.Writer, src io.Reader, buf []byte) (int64, error) {
	if len(buf) == 0 {
		buf = make([]byte, 32*1024)
	}
	var written int64
	for {
		nr, rErr := src.Read(buf)
		if rErr != nil && rErr != io.EOF && rErr != context.Canceled {
			log.WithError(rErr).
				Warn("read error during body copy")
		}
		if nr > 0 {
			nw, wErr := dst.Write(buf[:nr])
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
