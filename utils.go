package lingress

import (
	"github.com/echocat/lingress/support"
	"net"
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
