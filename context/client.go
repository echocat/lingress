package context

import (
	"errors"
	"fmt"
	"github.com/echocat/lingress/server"
	"github.com/echocat/lingress/support"
	"net"
	"net/http"
	"net/url"
	"time"
)

const (
	FieldClientMethod    = "method"
	FieldClientProto     = "proto"
	FieldClientUserAgent = "userAgent"
	FieldClientUrl       = "url"
	FieldClientAddress   = "address"
	FieldClientStatus    = "status"
	FieldClientStarted   = "started"
	FieldClientDuration  = "duration"
)

var (
	ErrNoRequestSet = errors.New("no request set")
)

type Client struct {
	Connector             server.ConnectorId
	FromOtherReverseProxy bool
	Response              http.ResponseWriter
	Request               *http.Request

	Status   int
	Started  time.Time
	Duration time.Duration

	requestedUrl *url.URL
	origin       *url.URL
	address      *string
}

func (this *Client) configure(connector server.ConnectorId, fromOtherReverseProxy bool, resp http.ResponseWriter, req *http.Request) {
	this.Connector = connector
	this.FromOtherReverseProxy = fromOtherReverseProxy
	this.Response = resp
	this.Request = req
	this.Status = -1
	this.Started = emptyTime
	this.Duration = -1

	this.requestedUrl = nil
	this.origin = nil
	this.address = nil
}

func (this *Client) clean() {
	if req := this.Request; req != nil {
		if b := req.Body; b != nil {
			_ = b.Close()
		}
	}
	if resp := this.Request; resp != nil {
		if b := resp.Body; b != nil {
			_ = b.Close()
		}
	}
	this.Connector = ""
	this.FromOtherReverseProxy = false
	this.Response = nil
	this.Request = nil
	this.Status = -1
	this.Started = emptyTime
	this.Duration = -1

	this.requestedUrl = nil
	this.origin = nil
	this.address = nil
}

func (this *Client) AsMap() map[string]interface{} {
	buf := make(map[string]interface{})
	this.ApplyToMap("", &buf)
	return buf
}

func (this *Client) ApplyToMap(prefix string, to *map[string]interface{}) {
	req := this.Request

	(*to)[prefix+FieldClientMethod] = req.Method
	(*to)[prefix+FieldClientProto] = req.Proto
	(*to)[prefix+FieldClientUserAgent] = support.UserAgentOfRequest(req)
	if u, _ := this.RequestedUrl(); u != nil {
		(*to)[prefix+FieldClientUrl] = u.String()
	}

	if r, err := this.Address(); err == nil {
		(*to)[prefix+FieldClientAddress] = r
	}
	if s := this.Status; s > 0 {
		(*to)[prefix+FieldClientStatus] = s
	}
	if t := this.Started; t != emptyTime {
		(*to)[prefix+FieldClientStarted] = t
	}
	if d := this.Duration; d > -1 {
		(*to)[prefix+FieldClientDuration] = d / time.Microsecond
	}
}

func (this Client) schemeOf(req *http.Request) string {
	if req.TLS != nil {
		return "https"
	}

	if this.FromOtherReverseProxy {
		if x := req.Header.Get("X-Forwarded-Proto"); x != "" {
			return x
		}
		if x := req.Header.Get("X-Scheme"); x != "" {
			return x
		}
	}

	return "http"
}

func (this Client) Host() string {
	req := this.Request
	if req == nil {
		return ""
	}

	return this.hostOf(req)
}

func (this Client) hostOf(req *http.Request) string {
	host := req.Host

	if x := req.Header.Get("Host"); x != "" {
		host = x
	}

	if this.FromOtherReverseProxy {
		if x := req.Header.Get("X-Forwarded-Host"); x != "" {
			host = x
		}
	}

	return host
}

func (this Client) uriOf(req *http.Request) string {
	result := req.RequestURI

	if this.FromOtherReverseProxy {
		if x := req.Header.Get("X-Forwarded-Prefix"); x != "" {
			result = x + result
		}
		if x := req.Header.Get("X-Original-URI"); x != "" {
			result = x
		}
	}

	return result
}

func (this *Client) RequestedUrl() (*url.URL, error) {
	if ru := this.requestedUrl; ru != nil {
		return ru, nil
	}

	req := this.Request
	if req == nil {
		return nil, ErrNoRequestSet
	}
	inUrl := req.URL
	if inUrl == nil {
		return nil, ErrNoRequestSet
	}

	scheme := this.schemeOf(req)
	host := this.hostOf(req)
	uri := this.uriOf(req)

	raw := fmt.Sprintf("%s://%s%s", scheme, host, uri)

	ru, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}

	ru.User = inUrl.User
	ru.Fragment = inUrl.Fragment

	this.requestedUrl = ru

	return ru, nil
}

func (this *Client) Origin() (*url.URL, error) {
	if ou := this.origin; ou != nil {
		return ou, nil
	}

	req := this.Request
	if req == nil {
		return nil, ErrNoRequestSet
	}

	raw := req.Header.Get("Origin")
	if raw == "" {
		return nil, nil
	}

	ou, err := url.Parse(raw)
	if err != nil {
		return nil, nil // We ignore these errors and just tread it as no Origin.
	}

	this.origin = ou

	return ou, nil
}

func (this *Client) Address() (string, error) {
	if r := this.address; r != nil {
		return *r, nil
	}

	req := this.Request
	if req == nil {
		return "", ErrNoRequestSet
	}

	r, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		return "", fmt.Errorf("illegal remote address (%s): %v", req.RemoteAddr, err)
	}

	if this.FromOtherReverseProxy {
		if x := req.Header.Get("X-Forwarded-For"); x != "" {
			r = x
		}
	}

	this.address = &r

	return r, nil
}
