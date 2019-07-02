package support

import (
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"time"
)

var (
	logLevelFacadeVal = &logLevelFacade{}

	_ = RegisterFlagRegistrar(logLevelFacadeVal)
)

type logLevelFacade struct{}

func (instance logLevelFacade) String() string {
	return log.GetLevel().String()
}

func (instance *logLevelFacade) Set(plain string) error {
	var n log.Level
	if err := n.UnmarshalText([]byte(plain)); err != nil {
		return err
	}
	log.SetLevel(n)
	return nil
}

func (instance *logLevelFacade) RegisterFlag(fe FlagEnabled, appPrefix string) error {
	fe.Flag("logLevel", "On which level the output should be logged").
		PlaceHolder("<log level; default: " + instance.String() + ">").
		Envar(FlagEnvName(appPrefix, "LOG_LEVEL")).
		SetValue(instance)
	return nil
}

func init() {
	log.SetFormatter(&log.JSONFormatter{
		TimestampFormat: time.RFC3339Nano,
		FieldMap: log.FieldMap{
			log.FieldKeyTime:  "@timestamp",
			log.FieldKeyLevel: "@level",
			log.FieldKeyMsg:   "@message",
			log.FieldKeyFunc:  "@caller",
		},
	})
	log.SetLevel(log.InfoLevel)
	log.SetOutput(os.Stderr)
}

func LogForRequest(req *http.Request) log.FieldLogger {
	return log.WithFields(log.Fields{
		"runtime":    Runtime(),
		"requestId":  RequestBasedLazyStringerFor(req, RequestIdOfRequest),
		"remoteIp":   RequestBasedLazyStringerFor(req, RemoteIpOfRequest),
		"host":       RequestBasedLazyStringerFor(req, HostOfRequest),
		"method":     req.Method,
		"requestUri": RequestBasedLazyStringerFor(req, UriOfRequest),
		"userAgent":  RequestBasedLazyStringerFor(req, UserAgentOfRequest),
	})
}
