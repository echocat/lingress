package context

import (
	"encoding/base64"
	"fmt"
	"github.com/google/uuid"
	"net/http"
)

type Id uuid.UUID

var (
	NilRequestId = Id(uuid.Nil)
	idEncoding   = base64.StdEncoding.WithPadding(base64.NoPadding)
)

func NewId(fromOtherReverseProxy bool, req *http.Request) (Id, error) {
	return newId(fromOtherReverseProxy, "X-Request-ID", req)
}

func NewCorrelationId(req *http.Request) (Id, error) {
	return newId(true, "X-Correlation-ID", req)
}

func newId(acceptUpstreamHeaderIfAny bool, fromHeader string, req *http.Request) (Id, error) {
	if acceptUpstreamHeaderIfAny {
		if x := req.Header.Get(fromHeader); len(x) > 0 && len(x) <= 256 {
			var id Id
			if err := id.UnmarshalText([]byte(x)); err == nil {
				return id, nil
			}
		}
	}
	val, err := uuid.NewRandom()
	if err != nil {
		return Id{}, err
	}
	return Id(val), nil
}

func (this Id) String() string {
	return idEncoding.EncodeToString(this[:])
}

func (this Id) MarshalText() ([]byte, error) {
	return []byte(this.String()), nil
}

func (this *Id) UnmarshalText(in []byte) error {
	decoded, err := idEncoding.DecodeString(string(in))
	if err != nil {
		return fmt.Errorf("illegal id: %q", string(in))
	}
	var buf uuid.UUID
	if err := buf.UnmarshalBinary(decoded); err != nil {
		return fmt.Errorf("illegal id: %q", string(in))
	}
	*this = Id(buf)
	return nil
}
