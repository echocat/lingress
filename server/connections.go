package server

import (
	"fmt"
	"net"
	"net/http"
	"sync"
)

type limitedListener struct {
	linger int16
	sync.Mutex
	net.Listener
	sem chan bool
}

func newLimitedListener(count uint16, linger int16, l net.Listener) *limitedListener {
	sem := make(chan bool, count)
	for i := uint16(0); i < count; i++ {
		sem <- true
	}
	return &limitedListener{
		Listener: l,
		linger:   linger,
		sem:      sem,
	}
}

func (instance *limitedListener) Accept() (net.Conn, error) {
	success := false
	<-instance.sem // acquire
	defer func() {
		if !success {
			instance.sem <- true
		}
	}()
	if c, err := instance.Listener.Accept(); err != nil {
		return nil, err
	} else {
		if err := c.(*net.TCPConn).SetLinger(int(instance.linger)); err != nil {
			return nil, fmt.Errorf("cannot set the SO_LINGER to %d", instance.linger)
		}
		result := &limitedConn{
			Conn:                c,
			annotatedRemoteAddr: &annotatedAddr{Addr: c.RemoteAddr()},
			parent:              instance,
		}
		success = true
		return result, nil
	}
}

type limitedConn struct {
	net.Conn
	annotatedRemoteAddr AnnotatedAddr
	parent              *limitedListener
	annotatedAddr
}

func (instance *limitedConn) RemoteAddr() net.Addr {
	return instance.annotatedRemoteAddr
}

func (instance *limitedConn) Close() error {
	defer func() {
		instance.parent.sem <- true // release
	}()
	return instance.Conn.Close()
}

type AnnotatedAddr interface {
	net.Addr
	GetState() *http.ConnState
	SetState(*http.ConnState)
}

type annotatedAddr struct {
	net.Addr
	state *http.ConnState
}

func (instance *annotatedAddr) GetState() *http.ConnState {
	return instance.state
}

func (instance *annotatedAddr) SetState(v *http.ConnState) {
	instance.state = v
}
