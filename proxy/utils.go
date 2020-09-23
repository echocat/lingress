package proxy

import (
	"errors"
	"golang.org/x/net/http/httpguts"
	"io"
	"net"
	"net/http"
	"strings"
)

func isDialError(err error) bool {
	if ne, ok := err.(*net.OpError); ok && ne.Op == "dial" {
		return true
	}
	return false
}

func isClientGoneError(err error) bool {
	return err == http.ErrAbortHandler || errors.Unwrap(err) == http.ErrAbortHandler
}

func retrieveUpgradeType(h http.Header) string {
	if !httpguts.HeaderValuesContainsToken(h["Connection"], "Upgrade") {
		return ""
	}
	return strings.ToLower(h.Get("Upgrade"))
}

// removeConnectionHeaders removes hop-by-hop headers listed in the "Connection" header of h.
// See RFC 7230, section 6.1
func removeConnectionHeaders(h http.Header) {
	if c := h.Get("Connection"); c != "" {
		for _, f := range strings.Split(c, ",") {
			if f = strings.TrimSpace(f); f != "" {
				h.Del(f)
			}
		}
	}
}

// removeHopReqHeaders removes hop-by-hop headers to the backend. Especially
// important is "Connection" because we want a persistent
// connection, regardless of what the client sent to us.
func removeHopReqHeaders(h http.Header) {
	for _, candidate := range hopHeaders {
		hv := h.Get(candidate)
		if hv == "" {
			continue
		}
		if candidate == "Te" && hv == "trailers" {
			// Issue 21096: tell backend applications that
			// care about trailer support that we support
			// trailers. (We do, but we don't go out of
			// our way to advertise that unless the
			// incoming client request thought it was
			// worth mentioning)
			continue
		}
		h.Del(candidate)
	}
}

func removeHopRespHeaders(h http.Header) {
	for _, hh := range hopHeaders {
		h.Del(hh)
	}
}

// setConnectionUpgrades handles after stripping all the hop-by-hop connection headers above, add back any
// necessary for protocol upgrades, such as for websockets.
func setConnectionUpgrades(h http.Header, reqUpType string) {
	if reqUpType != "" {
		h.Set("Connection", "Upgrade")
		h.Set("Upgrade", reqUpType)
	}
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func cloneHeader(h http.Header) http.Header {
	h2 := make(http.Header, len(h))
	for k, vv := range h {
		vv2 := make([]string, len(vv))
		copy(vv2, vv)
		h2[k] = vv2
	}
	return h2
}

// Hop-by-hop headers. These are removed when sent to the backend.
// As of RFC 7230, hop-by-hop headers are required to appear in the
// Connection header field. These are the headers defined by the
// obsoleted RFC 2616 (section 13.5.1) and are used for backward
// compatibility.
var hopHeaders = []string{
	"Connection",
	"Proxy-Connection", // non-standard but still sent by libcurl and rejected by e.g. google
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"Te",      // canonicalized version of "TE"
	"Trailer", // not Trailers per URL above; https://www.rfc-editor.org/errata_search.php?eid=4522
	"Transfer-Encoding",
	"Upgrade",
}

func stripPrefix(path, prefix []string) []string {
	if len(path) < len(prefix) {
		return path
	}
	for i, elem := range prefix {
		if path[i] != elem {
			return path
		}
	}
	return path[len(prefix):]
}

// switchProtocolCopier exists so goroutines proxying data back and
// forth have nice names in stacks.
type switchProtocolCopier struct {
	user, backend io.ReadWriter
}

func (c switchProtocolCopier) copyFromBackend(errc chan<- error) {
	_, err := io.Copy(c.user, c.backend)
	errc <- err
}

func (c switchProtocolCopier) copyToBackend(errc chan<- error) {
	_, err := io.Copy(c.backend, c.user)
	errc <- err
}
