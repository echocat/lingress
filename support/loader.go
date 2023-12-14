package support

import (
	"errors"
	"fmt"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
	"path"
)

var (
	EmptyBundle = errors.New("empty bundle")
)

func LoadBundle(provider FileProvider) (bundle *i18n.Bundle, err error) {
	bundle = i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)
	bundle.RegisterUnmarshalFunc("yml", yaml.Unmarshal)

	var atLeastOneLanguageAvailable bool
	entries, err := provider.ReadDir(".")
	if err != nil {
		return nil, fmt.Errorf("cannot load contents of localization file provider: %w", err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if ext := path.Ext(entry.Name()); ext != ".yaml" && ext != ".yml" {
			continue
		}
		if content, err := provider.ReadFile(entry.Name()); err != nil {
			return nil, fmt.Errorf("file '%s' could not be loaded but should exists: %v", entry, err)
		} else if _, err := bundle.ParseMessageFileBytes(content, entry.Name()); err != nil {
			return nil, fmt.Errorf("cannot load localization file '%s': %v", entry, err)
		} else {
			atLeastOneLanguageAvailable = true
		}
	}
	if !atLeastOneLanguageAvailable {
		return nil, EmptyBundle
	}
	return
}
