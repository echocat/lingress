package support

import (
	"github.com/alecthomas/kingpin/v2"
	"strings"
	"unicode"
)

type FlagEnabled interface {
	Flag(name, help string) *kingpin.FlagClause
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
