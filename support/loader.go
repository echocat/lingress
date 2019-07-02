package support

import (
	"errors"
	"fmt"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v2"
	"path"
)

var (
	EmptyBundle = errors.New("empty bundle")
)

type FileProvider interface {
	List() []string
	Find(filename string) ([]byte, error)
}

func LoadBundle(provider FileProvider) (bundle *i18n.Bundle, err error) {
	bundle = &i18n.Bundle{DefaultLanguage: language.English}
	bundle.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)
	bundle.RegisterUnmarshalFunc("yml", yaml.Unmarshal)

	var atLeastOneLanguageAvailable bool
	for _, fn := range provider.List() {
		ext := path.Ext(fn)
		dir := path.Dir(fn)
		if (ext == ".yaml" || ext == ".yml") && dir == "." {
			if content, err := provider.Find(fn); err != nil {
				return nil, fmt.Errorf("file '%s' could not be loaded but should exists: %v", fn, err)
			} else if _, err := bundle.ParseMessageFileBytes(content, fn); err != nil {
				return nil, fmt.Errorf("cannot load localization file '%s': %v", fn, err)
			} else {
				atLeastOneLanguageAvailable = true
			}
		}
	}
	if !atLeastOneLanguageAvailable {
		return nil, EmptyBundle
	}
	return
}
