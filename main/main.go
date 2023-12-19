package main

import (
	"fmt"
	"github.com/alecthomas/kingpin/v2"
	"github.com/echocat/lingress"
	"github.com/echocat/lingress/file/providers/def"
	"github.com/echocat/lingress/settings"
	"github.com/echocat/lingress/support"
	_ "github.com/echocat/lingress/support/slf4g_native"
	"github.com/echocat/slf4g"
	"github.com/echocat/slf4g/native"
	"github.com/echocat/slf4g/native/facade/value"
	"os"
	"os/signal"
)

const (
	appPrefix = "LINGRESS_"
)

func main() {
	vp := value.NewProvider(native.DefaultProvider)
	intSig := make(chan os.Signal, 1)

	rt := support.Runtime()
	app := kingpin.New("lingress", "Edge ingress implementation for Kubernetes")
	s := settings.MustNew()
	s.RegisterFlags(app, appPrefix)

	l, err := lingress.New(&s, def.Get())
	support.Must(err)

	app.Flag("version", "Shows the current version information").
		PreAction(func(*kingpin.ParseContext) error {
			fmt.Println(rt.LongString())
			os.Exit(0)
			return nil
		}).
		Bool()
	app.Flag("log.level", "").
		Envar(support.FlagEnvName(appPrefix, "LOG_LEVEL")).
		SetValue(&vp.Level)
	app.Flag("log.format", "").
		Default("text").
		Envar(support.FlagEnvName(appPrefix, "LOG_FORMAT")).
		SetValue(&vp.Consumer.Formatter)
	app.Flag("log.color", "").
		Default("auto").
		Envar(support.FlagEnvName(appPrefix, "LOG_COLOR")).
		SetValue(vp.Consumer.Formatter.ColorMode)

	stop := support.NewChannel()

	app.Action(func(_ *kingpin.ParseContext) error {
		support.ChannelDoOnEvent(stop, func() {
			close(intSig)
		})
		log.With("revision", rt.Revision).
			With("branch", rt.Branch).
			With("build", rt.Build).
			Info("lingress starting...")
		if err := l.Init(stop); err != nil {
			return err
		}
		<-intSig
		log.Info("Shutting down...")
		stop.Broadcast()
		log.Info("Bye!")
		return nil
	})

	signal.Notify(intSig, os.Interrupt, os.Kill)

	kingpin.MustParse(app.Parse(os.Args[1:]))
}
