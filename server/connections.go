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

func (this *limitedListener) Accept() (net.Conn, error) {
	success := false
	<-this.sem // acquire
	defer func() {
		if !success {
			this.sem <- true
		}
	}()
	if c, err := this.Listener.Accept(); err != nil {
		return nil, err
	} else {
		if err := c.(*net.TCPConn).SetLinger(int(this.linger)); err != nil {
			return nil, fmt.Errorf("cannot set the SO_LINGER to %d", this.linger)
		}
		result := &limitedConn{
			Conn:                c,
			annotatedRemoteAddr: &annotatedAddr{Addr: c.RemoteAddr()},
			parent:              this,
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

func (this *limitedConn) RemoteAddr() net.Addr {
	return this.annotatedRemoteAddr
}

func (this *limitedConn) Close() error {
	defer func() {
		this.parent.sem <- true // release
	}()
	return this.Conn.Close()
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

func (this *annotatedAddr) GetState() *http.ConnState {
	return this.state
}

func (this *annotatedAddr) SetState(v *http.ConnState) {
	this.state = v
}
