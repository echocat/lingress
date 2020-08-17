package server

import (
	"net"
	"sync"
	"time"
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

type limitedListener struct {
	sync.Mutex
	net.Listener
	sem chan bool
}

func newLimitedListener(count uint16, l net.Listener) *limitedListener {
	sem := make(chan bool, count)
	for i := uint16(0); i < count; i++ {
		sem <- true
	}
	return &limitedListener{
		Listener: l,
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
		result := &limitedConn{c, instance}
		success = true
		return result, nil
	}
}

type limitedConn struct {
	net.Conn
	parent *limitedListener
}

func (instance *limitedConn) Close() error {
	defer func() {
		instance.parent.sem <- true // release
	}()
	return instance.Conn.Close()
}
