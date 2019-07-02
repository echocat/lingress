package rules

import (
	"errors"
	"strings"
)

const (
	annotationIngressClass = "kubernetes.io/ingress.class"
	ingressClass           = "lingress"
)

var (
	ErrIllegalPath = errors.New("illegal path")
)

func normalizeHostname(in string) string {
	return strings.ToLower(
		strings.TrimSpace(in),
	)
}

func ParsePath(in string, faultTolerant bool) ([]string, error) {
	out := strings.Split(in, "/")
	if len(out) == 1 && !faultTolerant {
		return nil, ErrIllegalPath
	}
	if len(out) > 0 && out[0] == "" {
		out = out[1:]
	} else {
		return nil, ErrIllegalPath
	}

	if len(out) == 1 && out[0] == "" {
		return []string{}, nil
	}

	if len(out) > 0 && out[len(out)-1] == "" {
		out = out[:len(out)-1]
	}

	for _, element := range out {
		if element == "" && !faultTolerant {
			return nil, ErrIllegalPath
		}
	}
	return out, nil
}
