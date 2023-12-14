package main

import (
	"embed"
	"fmt"
	"github.com/alecthomas/kingpin/v2"
	"github.com/echocat/lingress"
	"github.com/echocat/lingress/support"
	_ "github.com/echocat/lingress/support/slf4g_native"
	"github.com/echocat/slf4g"
	"os"
	"os/signal"
)

const (
	appPrefix = "LINGRESS_"
)

var (
	//go:embed localization
	localizations embed.FS
	//go:embed templates
	templates embed.FS

	fps = support.DefaultFileProviders{
		Localization: support.FileProviderStrippingPrefix(localizations, "localization"),
		Static:       support.NoopFileProvider(),
		Templates:    support.FileProviderStrippingPrefix(templates, "templates"),
	}
)

func main() {
	intSig := make(chan os.Signal, 1)

	rt := support.Runtime()
	app := kingpin.New(rt.Name(), "Edge ingress implementation for Kubernetes")

	l, err := lingress.New(fps)
	support.Must(err)

	app.Flag("version", "Shows the current version information").
		PreAction(func(*kingpin.ParseContext) error {
			fmt.Println(rt.LongString())
			os.Exit(0)
			return nil
		}).
		Bool()

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
