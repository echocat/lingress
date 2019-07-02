package main

import (
	"github.com/alecthomas/kingpin"
	log "github.com/sirupsen/logrus"
	"github.com/echocat/lingress"
	"github.com/echocat/lingress/support"
	_ "github.com/echocat/lingress/support"
	"os"
	"os/signal"
)

const (
	appPrefix = "LINGRESS_"
)

func main() {
	intSig := make(chan os.Signal, 1)

	app := kingpin.New(support.Runtime().Name(), "Edge ingress implementation for Kubernetes")

	l, err := lingress.New()
	support.Must(err)

	support.Must(l.RegisterFlag(app, appPrefix))
	support.MustRegisterGlobalFalgs(app, appPrefix)

	stopCh := make(chan struct{})

	app.Action(func(_ *kingpin.ParseContext) error {
		defer close(stopCh)
		if err := l.Init(stopCh); err != nil {
			return err
		}
		<-intSig
		log.Info("shutting down...")
		stopCh <- struct{}{}
		log.Info("bye!")
		return nil
	})

	signal.Notify(intSig, os.Interrupt, os.Kill)

	kingpin.MustParse(app.Parse(os.Args[1:]))
}
