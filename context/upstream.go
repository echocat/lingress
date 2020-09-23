package context

import (
	"context"
	"github.com/echocat/lingress/rules"
	"net"
	"net/http"
	"strings"
	"time"
)

const (
	FieldUpstreamAddress  = "address"
	FieldUpstreamStatus   = "status"
	FieldUpstreamStarted  = "started"
	FieldUpstreamDuration = "duration"
	FieldUpstreamUrl      = "url"
	FieldUpstreamMethod   = "method"
	FieldUpstreamProto    = "proto"
	FieldUpstreamSource   = "source"
	FieldUpstreamMatches  = "matches"
)

type Upstream struct {
	Response *http.Response
	Request  *http.Request
	Address  net.Addr
	Cancel   context.CancelFunc

	Status   int
	Started  time.Time
	Duration time.Duration
}

func (instance *Upstream) configure() {
	instance.Response = nil
	instance.Request = nil
	instance.Address = nil
	instance.Cancel = nil
	instance.Status = -1
	instance.Started = emptyTime
	instance.Duration = 0
}

func (instance *Upstream) clean() {
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
	instance.Response = nil
	instance.Request = nil
	instance.Address = nil
	instance.Cancel = nil
	instance.Status = -1
	instance.Started = emptyTime
	instance.Duration = 0
}

func (instance *Upstream) AsMap(r rules.Rule) map[string]interface{} {
	buf := make(map[string]interface{})

	instance.ApplyToMap(r, "", &buf)

	if len(buf) == 0 {
		return nil
	}

	return buf
}

func (instance *Upstream) ApplyToMap(r rules.Rule, prefix string, to *map[string]interface{}) {
	if addr := instance.Address; addr != nil {
		(*to)[prefix+FieldUpstreamAddress] = addr.String()
	}
	if s := instance.Status; s > 0 {
		(*to)[prefix+FieldUpstreamStatus] = s
	}
	if t := instance.Started; t != emptyTime {
		(*to)[prefix+FieldUpstreamStarted] = t
	}
	if d := instance.Duration; d > 0 {
		(*to)[prefix+FieldUpstreamDuration] = d / time.Microsecond
	}
	if req := instance.Request; req != nil {
		if u := req.URL; u != nil {
			(*to)[prefix+FieldUpstreamUrl] = u.String()
		}
		(*to)[prefix+FieldUpstreamMethod] = req.Method
		(*to)[prefix+FieldUpstreamProto] = req.Proto
	}
	if r != nil {
		(*to)[prefix+FieldUpstreamSource] = r.Source().String()
		(*to)[prefix+FieldUpstreamMatches] = r.Host() + "/" + strings.Join(r.Path(), "/")
	}
}
