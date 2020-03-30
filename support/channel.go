package support

import (
	"sync"
)

type Channel interface {
	Wait()
	Broadcast()
}

type channel struct {
	mutex sync.Mutex
	cond  *sync.Cond
}

func NewChannel() Channel {
	result := channel{}
	result.cond = sync.NewCond(&result.mutex)
	return &result
}

func (instance *channel) Wait() {
	instance.mutex.Lock()
	defer instance.mutex.Unlock()
	instance.cond.Wait()
}

func (instance *channel) Broadcast() {
	instance.mutex.Lock()
	defer instance.mutex.Unlock()
	instance.cond.Broadcast()
}

func ToChan(channel Channel) chan struct{} {
	result := make(chan struct{})
	go func() {
		channel.Wait()
		result <- struct{}{}
	}()
	return result
}

func ChannelDoOnEvent(of Channel, what func()) {
	go func() {
		of.Wait()
		what()
	}()
}
