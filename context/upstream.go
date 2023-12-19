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

func (this *Upstream) configure() {
	this.Response = nil
	this.Request = nil
	this.Address = nil
	this.Cancel = nil
	this.Status = -1
	this.Started = emptyTime
	this.Duration = 0
}

func (this *Upstream) clean() {
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
	this.Response = nil
	this.Request = nil
	this.Address = nil
	this.Cancel = nil
	this.Status = -1
	this.Started = emptyTime
	this.Duration = 0
}

func (this *Upstream) AsMap(r rules.Rule) map[string]interface{} {
	buf := make(map[string]interface{})

	this.ApplyToMap(r, "", &buf)

	if len(buf) == 0 {
		return nil
	}

	return buf
}

func (this *Upstream) ApplyToMap(r rules.Rule, prefix string, to *map[string]interface{}) {
	if addr := this.Address; addr != nil {
		(*to)[prefix+FieldUpstreamAddress] = addr.String()
	}
	if s := this.Status; s > 0 {
		(*to)[prefix+FieldUpstreamStatus] = s
	}
	if t := this.Started; t != emptyTime {
		(*to)[prefix+FieldUpstreamStarted] = t
	}
	if d := this.Duration; d > 0 {
		(*to)[prefix+FieldUpstreamDuration] = d / time.Microsecond
	}
	if req := this.Request; req != nil {
		if u := req.URL; u != nil {
			(*to)[prefix+FieldUpstreamUrl] = u.String()
		}
		(*to)[prefix+FieldUpstreamMethod] = req.Method
		(*to)[prefix+FieldUpstreamProto] = req.Proto
	}
	if r != nil {
		(*to)[prefix+FieldUpstreamSource] = r.Source().String()
		(*to)[prefix+FieldUpstreamMatches] = r.Host().String() + "/" + strings.Join(r.Path(), "/")
	}
}
