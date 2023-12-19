package rules

import (
	"errors"
	"fmt"
	networkingv1 "k8s.io/api/networking/v1"
	"strings"
)

const (
	ingressClass = "lingress"
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
		return nil, fmt.Errorf("%w: %s", ErrIllegalPath, in)
	}
	if len(out) > 0 && out[0] == "" {
		out = out[1:]
	} else {
		return nil, fmt.Errorf("%w: %s", ErrIllegalPath, in)
	}

	if len(out) == 1 && out[0] == "" {
		return []string{}, nil
	}

	if len(out) > 0 && out[len(out)-1] == "" {
		out = out[:len(out)-1]
	}

	for _, element := range out {
		if element == "" && !faultTolerant {
			return nil, fmt.Errorf("%w: %s", ErrIllegalPath, in)
		}
	}
	return out, nil
}

func ParsePathType(in *networkingv1.PathType) (PathType, error) {
	if in == nil {
		return PathTypePrefix, nil
	}
	switch *in {
	case "Prefix", "ImplementationSpecific":
		return PathTypePrefix, nil
	case "Exact":
		return PathTypeExact, nil
	default:
		return 0, fmt.Errorf("cannot handle path type: %s", *in)
	}
}
