package lingress

import (
	"github.com/echocat/lingress/support"
	"net"
	"net/http"
	"sync"
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

type ConnectionInformation struct {
	all map[net.Conn]http.ConnState

	mutex *sync.Mutex
}

func NewConnectionInformation() *ConnectionInformation {
	return &ConnectionInformation{
		all:   make(map[net.Conn]http.ConnState),
		mutex: new(sync.Mutex),
	}
}

func (instance *ConnectionInformation) SetState(
	conn net.Conn,
	newState http.ConnState,
	keepFunc func(http.ConnState) bool,
) (previousState http.ConnState) {
	instance.mutex.Lock()
	defer instance.mutex.Unlock()

	keep := newState >= 0 || keepFunc(newState)

	result, exists := instance.all[conn]
	if exists {
		if !keep {
			delete(instance.all, conn)
			return result
		}
		previousState = result
		instance.all[conn] = newState
		return previousState
	}

	if keep {
		instance.all[conn] = newState
	}
	return -1
}
