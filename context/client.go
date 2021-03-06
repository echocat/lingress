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

func (instance *Client) configure(connector server.ConnectorId, fromOtherReverseProxy bool, resp http.ResponseWriter, req *http.Request) {
	instance.Connector = connector
	instance.FromOtherReverseProxy = fromOtherReverseProxy
	instance.Response = resp
	instance.Request = req
	instance.Status = -1
	instance.Started = emptyTime
	instance.Duration = -1

	instance.requestedUrl = nil
	instance.origin = nil
	instance.address = nil
}

func (instance *Client) clean() {
	if req := instance.Request; req != nil {
		if b := req.Body; b != nil {
			_ = b.Close()
		}
	}
	if resp := instance.Request; resp != nil {
		if b := resp.Body; b != nil {
			_ = b.Close()
		}
	}
	instance.Connector = ""
	instance.FromOtherReverseProxy = false
	instance.Response = nil
	instance.Request = nil
	instance.Status = -1
	instance.Started = emptyTime
	instance.Duration = -1

	instance.requestedUrl = nil
	instance.origin = nil
	instance.address = nil
}

func (instance *Client) AsMap() map[string]interface{} {
	buf := make(map[string]interface{})
	instance.ApplyToMap("", &buf)
	return buf
}

func (instance *Client) ApplyToMap(prefix string, to *map[string]interface{}) {
	req := instance.Request

	(*to)[prefix+FieldClientMethod] = req.Method
	(*to)[prefix+FieldClientProto] = req.Proto
	(*to)[prefix+FieldClientUserAgent] = support.UserAgentOfRequest(req)
	if u, _ := instance.RequestedUrl(); u != nil {
		(*to)[prefix+FieldClientUrl] = u.String()
	}

	if r, err := instance.Address(); err == nil {
		(*to)[prefix+FieldClientAddress] = r
	}
	if s := instance.Status; s > 0 {
		(*to)[prefix+FieldClientStatus] = s
	}
	if t := instance.Started; t != emptyTime {
		(*to)[prefix+FieldClientStarted] = t
	}
	if d := instance.Duration; d > -1 {
		(*to)[prefix+FieldClientDuration] = d / time.Microsecond
	}
}

func (instance Client) schemeOf(req *http.Request) string {
	if req.TLS != nil {
		return "https"
	}

	if instance.FromOtherReverseProxy {
		if x := req.Header.Get("X-Forwarded-Proto"); x != "" {
			return x
		}
		if x := req.Header.Get("X-Scheme"); x != "" {
			return x
		}
	}

	return "http"
}

func (instance Client) Host() string {
	req := instance.Request
	if req == nil {
		return ""
	}

	return instance.hostOf(req)
}

func (instance Client) hostOf(req *http.Request) string {
	host := req.Host

	if x := req.Header.Get("Host"); x != "" {
		host = x
	}

	if instance.FromOtherReverseProxy {
		if x := req.Header.Get("X-Forwarded-Host"); x != "" {
			host = x
		}
	}

	return host
}

func (instance Client) uriOf(req *http.Request) string {
	result := req.RequestURI

	if instance.FromOtherReverseProxy {
		if x := req.Header.Get("X-Forwarded-Prefix"); x != "" {
			result = x + result
		}
		if x := req.Header.Get("X-Original-URI"); x != "" {
			result = x
		}
	}

	return result
}

func (instance *Client) RequestedUrl() (*url.URL, error) {
	if ru := instance.requestedUrl; ru != nil {
		return ru, nil
	}

	req := instance.Request
	if req == nil {
		return nil, ErrNoRequestSet
	}
	inUrl := req.URL
	if inUrl == nil {
		return nil, ErrNoRequestSet
	}

	scheme := instance.schemeOf(req)
	host := instance.hostOf(req)
	uri := instance.uriOf(req)

	raw := fmt.Sprintf("%s://%s%s", scheme, host, uri)

	ru, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}

	ru.User = inUrl.User
	ru.Fragment = inUrl.Fragment

	instance.requestedUrl = ru

	return ru, nil
}

func (instance *Client) Origin() (*url.URL, error) {
	if ou := instance.origin; ou != nil {
		return ou, nil
	}

	req := instance.Request
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

	instance.origin = ou

	return ou, nil
}

func (instance *Client) Address() (string, error) {
	if r := instance.address; r != nil {
		return *r, nil
	}

	req := instance.Request
	if req == nil {
		return "", ErrNoRequestSet
	}

	r, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		return "", fmt.Errorf("illegal remote address (%s): %v", req.RemoteAddr, err)
	}

	if instance.FromOtherReverseProxy {
		if x := req.Header.Get("X-Forwarded-For"); x != "" {
			r = x
		}
	}

	instance.address = &r

	return r, nil
}
