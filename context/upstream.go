package context

import (
	"context"
	"net"
	"net/http"
	"time"
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
	instance.Response = nil
	instance.Request = nil
	instance.Address = nil
	instance.Cancel = nil
	instance.Status = -1
	instance.Started = emptyTime
	instance.Duration = 0
}

func (instance Upstream) AsMap() map[string]interface{} {
	buf := map[string]interface{}{}
	if addr := instance.Address; addr != nil {
		buf["address"] = addr.String()
	}
	if s := instance.Status; s > 0 {
		buf["status"] = s
	}
	if t := instance.Started; t != emptyTime {
		buf["started"] = t
	}
	if d := instance.Duration; d > 0 {
		buf["duration"] = d / time.Microsecond
	}
	return buf
}
