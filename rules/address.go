package rules

import (
	"encoding/json"
	"net"
)

type Address interface {
	Matches(net.IP) bool
}

type ipAddress net.IP

func (instance ipAddress) Matches(candidate net.IP) bool {
	if candidate == nil {
		return false
	}
	return net.IP(instance).Equal(candidate)
}

func (instance ipAddress) MarshalJSON() ([]byte, error) {
	return json.Marshal(net.IP(instance).String())
}

type networkAddress struct {
	*net.IPNet
}

func (instance *networkAddress) Matches(candidate net.IP) bool {
	if candidate == nil {
		return false
	}
	return instance.Contains(candidate)
}

func (instance *networkAddress) MarshalJSON() ([]byte, error) {
	return json.Marshal(instance.IPNet.String())
}
