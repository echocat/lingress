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

func (instance *ConnectorId) Set(plain string) error {
	return instance.UnmarshalText([]byte(plain))
}

func (instance ConnectorId) String() string {
	v, _ := instance.MarshalText()
	return string(v)
}

func (instance ConnectorId) MarshalText() (text []byte, err error) {
	if len(instance) > 0 && (!connectorIdRegexp.MatchString(string(instance)) || len(instance) > 64) {
		return []byte(fmt.Sprintf("illegal-connector-id-%s", string(instance))),
			fmt.Errorf("%w: %s", ErrIllegalConnectorId, string(instance))
	}
	return []byte(instance), nil
}

func (instance *ConnectorId) UnmarshalText(text []byte) error {
	v := ConnectorId(text)
	if _, err := instance.MarshalText(); err != nil {
		return err
	}
	*instance = v
	return nil
}
