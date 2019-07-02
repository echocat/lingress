package fallback

import (
	"fmt"
	"github.com/echocat/lingress/support"
	"net/http"
)

func localizeStatus(statusCode int, lc *support.LocalizationContext) string {
	return lc.MessageOrDefault("status-message.default", fmt.Sprintf("status-message.%d", statusCode))
}

func isStatusTemporaryIssue(code int) bool {
	switch code {
	case http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusLoopDetected,
		http.StatusNetworkAuthenticationRequired:
		return true
	default:
		return false
	}
}

func isStatusCodeAnIssue(code int) bool {
	return code >= 400
}

func isStatusClientSideIssue(code int) bool {
	return code >= 400 && code <= 499
}

func isStatusServerSideIssue(code int) bool {
	return code >= 500 && code <= 599
}
