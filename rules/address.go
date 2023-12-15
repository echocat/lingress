package rules

import (
	"encoding/json"
	"net"
)

type Address interface {
	Matches(net.IP) bool
}

type ipAddress net.IP

func (this ipAddress) Matches(candidate net.IP) bool {
	if candidate == nil {
		return false
	}
	return net.IP(this).Equal(candidate)
}

func (this ipAddress) MarshalJSON() ([]byte, error) {
	return json.Marshal(net.IP(this).String())
}

type networkAddress struct {
	*net.IPNet
}

func (this *networkAddress) Matches(candidate net.IP) bool {
	if candidate == nil {
		return false
	}
	return this.Contains(candidate)
}

func (this *networkAddress) MarshalJSON() ([]byte, error) {
	return json.Marshal(this.IPNet.String())
}
