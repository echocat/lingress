package support

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

var headerNewlineToSpace = strings.NewReplacer("\n", " ", "\r", " ")

func NormalizeHeaderContent(val string) string {
	return headerNewlineToSpace.Replace(val)
}

func RequestIdOfRequest(req *http.Request) string {
	if x := req.Header.Get("X-Request-ID"); len(x) > 0 && len(x) <= 256 {
		if len(x) > 256 {
			return x[:255]
		}
		return x
	}
	return ""
}

func HostOfRequest(req *http.Request) string {
	if x := req.Header.Get("X-Forwarded-Host"); len(x) > 0 {
		return x
	}
	return req.Host
}

func RemoteIpOfRequest(req *http.Request) string {
	if x := req.Header.Get("X-Forwarded-For"); len(x) > 0 {
		return x
	}
	if x := req.Header.Get("X-Real-IP"); len(x) > 0 {
		return x
	}
	remote := req.RemoteAddr
	ld := strings.LastIndexByte(remote, ':')
	if ld > 0 {
		remote = remote[:ld]
	}
	return remote
}

func UriOfRequest(req *http.Request) string {
	if x := req.Header.Get("X-Original-URI"); len(x) > 0 {
		return x
	}
	result := req.RequestURI
	if x := req.Header.Get("X-Forwarded-Prefix"); len(x) > 0 {
		return x + result
	}
	return result
}

func UserAgentOfRequest(req *http.Request) string {
	if x := req.Header.Get("User-Agent"); len(x) > 0 {
		if len(x) > 256 {
			return x[:255]
		}
		return x
	}
	return ""
}

func AcceptLanguageOfRequest(req *http.Request) string {
	if x := req.Header.Get("Accept-Language"); len(x) > 0 {
		return x
	}
	return ""
}

func PathOfRequest(req *http.Request) string {
	return req.URL.Path
}

func PathOfUri(uri string) (string, error) {
	path, _, err := PathAndQueryOfUri(uri)
	return path, err
}

func PathAndQueryOfUri(uri string) (path, query string, err error) {
	if uri != "" {
		if parsedUri, pErr := url.ParseRequestURI(uri); pErr != nil {
			return "", "", pErr
		} else {
			path = parsedUri.Path
			query = parsedUri.RawQuery
		}
	}
	return
}

func RequestBasedLazyStringerFor(req *http.Request, delegate func(*http.Request) string) fmt.Stringer {
	return ToLazyStringer(func() string {
		return delegate(req)
	})
}
