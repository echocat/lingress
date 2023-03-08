package main

import (
	"github.com/alecthomas/kingpin/v2"
	"github.com/echocat/lingress"
	"github.com/echocat/lingress/support"
	_ "github.com/echocat/lingress/support"
	"github.com/echocat/slf4g"
	"github.com/gobuffalo/packr"
	"os"
	"os/signal"
)

const (
	appPrefix = "LINGRESS_"
)

var (
	localizations = packr.NewBox("../resources/localization")
	static        = packr.NewBox("../resources/static")
	templates     = packr.NewBox("../resources/templates")

	fps = support.DefaultFileProviders{
		Localization: localizations,
		Static:       static,
		Templates:    templates,
	}
)

func main() {
	intSig := make(chan os.Signal, 1)

	rt := support.Runtime()
	app := kingpin.New(rt.Name(), "Edge ingress implementation for Kubernetes")

	l, err := lingress.New(fps)
	support.Must(err)

	support.Must(l.RegisterFlag(app, appPrefix))
	support.MustRegisterGlobalFalgs(app, appPrefix)

	stop := support.NewChannel()

	app.Action(func(_ *kingpin.ParseContext) error {
		support.ChannelDoOnEvent(stop, func() {
			close(intSig)
		})
		log.With("revision", rt.Revision).
			With("branch", rt.Branch).
			With("build", rt.Build).
			Infof("starting %s...", rt.Name())
		if err := l.Init(stop); err != nil {
			return err
		}
		<-intSig
		log.Info("shutting down...")
		stop.Broadcast()
		log.Info("bye!")
		return nil
	})

	signal.Notify(intSig, os.Interrupt, os.Kill)

	kingpin.MustParse(app.Parse(os.Args[1:]))
}
