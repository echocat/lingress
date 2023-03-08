package support

import (
	"github.com/alecthomas/kingpin/v2"
	"strings"
	"unicode"
)

var (
	globalFlagRegistrars []FlagRegistrar
)

type FlagEnabled interface {
	Flag(name, help string) *kingpin.FlagClause
}

type FlagRegistrar interface {
	RegisterFlag(fe FlagEnabled, appPrefix string) error
}

func RegisterFlagRegistrar(fr FlagRegistrar) FlagRegistrar {
	globalFlagRegistrars = append(globalFlagRegistrars, fr)
	return fr
}

func RegisterGlobalFlags(fe FlagEnabled, appPrefix string) error {
	for _, fr := range globalFlagRegistrars {
		if err := fr.RegisterFlag(fe, appPrefix); err != nil {
			return err
		}
	}
	return nil
}

func MustRegisterGlobalFalgs(fe FlagEnabled, appPrefix string) {
	Must(RegisterGlobalFlags(fe, appPrefix))
}

func FlagEnvNameNormalize(what string) string {
	return strings.Map(func(what rune) rune {
		if (what >= 'A' && what <= 'Z') || (what >= '0' && what <= '9') {
			return what
		} else if what >= 'a' && what <= 'z' {
			return unicode.ToUpper(what)
		} else {
			return '_'
		}
	}, what)
}

func FlagEnvName(appPrefix, name string) string {
	return FlagEnvNameNormalize(appPrefix + name)
}
