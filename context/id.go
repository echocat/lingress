package context

import (
	"encoding/base32"
	"encoding/base64"
	"encoding/json"
	"github.com/google/uuid"
	"net/http"
)

type Id uuid.UUID

var (
	NilRequestId = Id(uuid.Nil)
	idEncoding   = base64.StdEncoding.WithPadding(base32.NoPadding)
)

func NewId(fromOtherReverseProxy bool, req *http.Request) Id {
	return newId(fromOtherReverseProxy, "X-Request-ID", req)
}

func NewCorrelationId(fromOtherReverseProxy bool, req *http.Request) Id {
	return newId(fromOtherReverseProxy, "X-Correlation-ID", req)
}

func newId(fromOtherReverseProxy bool, fromHeader string, req *http.Request) Id {
	if fromOtherReverseProxy {
		if x := req.Header.Get(fromHeader); len(x) > 0 && len(x) <= 256 {
			if id, err := ParseId(x); err == nil {
				return id
			}
		}
	}
	val := uuid.New()
	return Id(val)
}

func (instance Id) String() string {
	return idEncoding.EncodeToString(instance[:])
}

func (instance Id) MarshalJSON() ([]byte, error) {
	return json.Marshal(instance.String())
}

func ParseId(plain string) (Id, error) {
	val, err := uuid.Parse(plain)
	if err != nil {
		return Id{}, err
	}
	return Id(val), nil
}
