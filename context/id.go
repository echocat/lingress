package context

import (
	"encoding/base32"
	"encoding/base64"
	"encoding/json"
	"github.com/satori/go.uuid"
	"net/http"
)

type Id uuid.UUID

var (
	NilRequestId = Id(uuid.Nil)
	idEncoding   = base64.StdEncoding.WithPadding(base32.NoPadding)
)

func NewId(_ *http.Request) Id {
	val := uuid.NewV4()
	return Id(val)
}

func (instance Id) String() string {
	return idEncoding.EncodeToString(instance[:])
}

func (instance Id) MarshalJSON() ([]byte, error) {
	return json.Marshal(instance.String())
}
