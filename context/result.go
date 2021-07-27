package context

import (
	"fmt"
	"net/http"
)

type Result interface {
	Name() string
	String() string
	Status() int
	WasResponseSendToClient() bool
}

type SimpleResult uint8

type ResultHandler func(ctx *Context)

var (
	ResultUnknown                       SimpleResult = 0
	ResultSuccess                       SimpleResult = 1
	ResultOk                            SimpleResult = 2
	ResultFailedWithUnexpectedError     SimpleResult = 3
	ResultFailedWithRuleNotFound        SimpleResult = 4
	ResultFailedWithUpstreamUnavailable SimpleResult = 5
	ResultFailedWithAccessDenied        SimpleResult = 6
	ResultFallback                      SimpleResult = 7
	ResultFailedWithUnauthorized        SimpleResult = 8
	ResultFailedWithClientGone          SimpleResult = 9
	ResultFailedWithIllegalHost         SimpleResult = 10
)

var (
	resultToName = map[Result]string{
		ResultUnknown:                       "unknown",
		ResultSuccess:                       "success",
		ResultOk:                            "ok",
		ResultFailedWithUnexpectedError:     "unexpectedError",
		ResultFailedWithRuleNotFound:        "ruleNotFound",
		ResultFailedWithUpstreamUnavailable: "upstreamUnavailable",
		ResultFailedWithAccessDenied:        "accessDenied",
		ResultFailedWithUnauthorized:        "unauthorized",
		ResultFallback:                      "fallback",
		ResultFailedWithClientGone:          "clientGone",
		ResultFailedWithIllegalHost:         "illegalHost",
	}

	resultToStatus = map[Result]int{
		ResultUnknown:                       http.StatusInternalServerError,
		ResultSuccess:                       http.StatusOK,
		ResultOk:                            http.StatusOK,
		ResultFailedWithUnexpectedError:     http.StatusInternalServerError,
		ResultFailedWithRuleNotFound:        http.StatusNotFound,
		ResultFailedWithUpstreamUnavailable: http.StatusServiceUnavailable,
		ResultFailedWithAccessDenied:        http.StatusForbidden,
		ResultFailedWithUnauthorized:        http.StatusUnauthorized,
		ResultFallback:                      http.StatusOK,
		ResultFailedWithClientGone:          http.StatusGone,
		ResultFailedWithIllegalHost:         http.StatusUnprocessableEntity,
	}
)

func (instance SimpleResult) WasResponseSendToClient() bool {
	return instance == ResultSuccess
}

func (instance SimpleResult) Status() int {
	if status, ok := resultToStatus[instance]; ok {
		return status
	} else {
		return http.StatusInternalServerError
	}
}

func (instance SimpleResult) Name() string {
	if name, ok := resultToName[instance]; ok {
		return name
	} else {
		return fmt.Sprintf("unknown-result-%d", instance)
	}
}

func (instance SimpleResult) String() string {
	return instance.Name()
}

type RedirectResult struct {
	StatusCode int
	Target     string
}

func (instance RedirectResult) WasResponseSendToClient() bool {
	return false
}

func (instance RedirectResult) Status() int {
	return instance.StatusCode
}

func (instance RedirectResult) Name() string {
	return fmt.Sprintf("redirect-%d", instance.StatusCode)
}

func (instance RedirectResult) String() string {
	return instance.Name() + ":" + instance.Target
}
