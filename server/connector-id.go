package server

import (
	"errors"
	"fmt"
	"regexp"
)

type ConnectorId string

var (
	connectorIdRegexp = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])$`)

	ErrIllegalConnectorId = errors.New("illegal connector-id")
)

func (this *ConnectorId) Set(plain string) error {
	return this.UnmarshalText([]byte(plain))
}

func (this ConnectorId) String() string {
	v, _ := this.MarshalText()
	return string(v)
}

func (this ConnectorId) MarshalText() (text []byte, err error) {
	if len(this) > 0 && (!connectorIdRegexp.MatchString(string(this)) || len(this) > 64) {
		return []byte(fmt.Sprintf("illegal-connector-id-%s", string(this))),
			fmt.Errorf("%w: %s", ErrIllegalConnectorId, string(this))
	}
	return []byte(this), nil
}

func (this *ConnectorId) UnmarshalText(text []byte) error {
	v := ConnectorId(text)
	if _, err := this.MarshalText(); err != nil {
		return err
	}
	*this = v
	return nil
}
