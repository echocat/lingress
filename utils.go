package lingress

import (
	"fmt"
	"github.com/echocat/lingress/support"
	"net"
	"net/http"
	"reflect"
	"sync/atomic"
	"time"
)

var (
	privateNetworks = func(ins ...string) (result []*net.IPNet) {
		result = make([]*net.IPNet, len(ins))
		for i, in := range ins {
			_, n, err := net.ParseCIDR(in)
			support.Must(err)
			result[i] = n
		}
		return
	}(
		"10.0.0.0/8", "100.64.0.0/10", "172.16.0.0/12", "192.0.0.0/24", "192.168.0.0/16", "198.18.0.0/15", "127.0.0.0/8",
		"fe80::/10", "fc00::/7", "::1/128",
	)
)

type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (ln tcpKeepAliveListener) Accept() (net.Conn, error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return nil, err
	}
	_ = tc.SetKeepAlive(true)
	_ = tc.SetKeepAlivePeriod(2 * time.Minute)
	return tc, nil
}

type stateTrackingListener struct {
	net.Listener
}

func (ln stateTrackingListener) Accept() (net.Conn, error) {
	conn, err := ln.Listener.Accept()
	if err != nil {
		return nil, err
	}
	return ConnectionWithConnStateTracking(conn), nil
}

type stateTrackingConnection struct {
	net.Conn
	current int64
}

func (instance *stateTrackingConnection) SetConnState(newState http.ConnState) (previousState http.ConnState) {
	for {
		current := atomic.LoadInt64(&instance.current)
		if atomic.CompareAndSwapInt64(&instance.current, current, int64(newState)) {
			return http.ConnState(current)
		}
	}
}

func SetConnState(conn net.Conn, newState http.ConnState) (previousState http.ConnState) {
	stc, ok := conn.(*stateTrackingConnection)
	if !ok {
		panic(fmt.Sprintf("%v is not of type %v", conn, reflect.TypeOf(&stateTrackingConnection{})))
	}
	return stc.SetConnState(newState)
}

func ConnectionWithConnStateTracking(in net.Conn) net.Conn {
	stc, ok := in.(*stateTrackingConnection)
	if ok {
		return stc
	}
	return &stateTrackingConnection{
		Conn:    in,
		current: -1,
	}
}
